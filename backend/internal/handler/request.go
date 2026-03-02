package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

// MetricsRecorder records business-level metrics for observability.
type MetricsRecorder interface {
	RecordInputType(inputType string)
	RecordRequestStatus(status string)
}

type RequestHandler struct {
	reqSvc  *service.RequestService
	store   *store.Store
	metrics MetricsRecorder
}

func NewRequestHandler(reqSvc *service.RequestService, s *store.Store) *RequestHandler {
	return &RequestHandler{reqSvc: reqSvc, store: s}
}

// SetMetrics sets the optional metrics recorder for business metrics.
func (h *RequestHandler) SetMetrics(m MetricsRecorder) {
	h.metrics = m
}

func (h *RequestHandler) List(c *fiber.Ctx) error {
	claims, err := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	limit := int32(c.QueryInt("limit", 50))
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}
	offset := int32(c.QueryInt("offset", 0))
	if offset < 0 {
		offset = 0
	}

	var requests []db.Request

	switch claims.Role {
	case domain.RoleContractor:
		requests, err = h.store.Queries.ListRequestsByContractor(c.Context(), db.ListRequestsByContractorParams{
			ContractorID: claims.UserID,
			Limit:        limit,
			Offset:       offset,
		})
	case domain.RoleSalesRep:
		requests, err = h.store.Queries.ListRequestsByRep(c.Context(), db.ListRequestsByRepParams{
			AssignedRepID: pgtype.UUID{Bytes: claims.UserID, Valid: true},
			Limit:         limit,
			Offset:        offset,
		})
	default:
		requests, err = h.store.Queries.ListRequestsByDealer(c.Context(), db.ListRequestsByDealerParams{
			DealerID: dealerID,
			Limit:    limit,
			Offset:   offset,
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list requests"})
	}

	// Include total count for pagination
	total, _ := h.store.Queries.CountRequestsByDealer(c.Context(), dealerID)

	return c.JSON(fiber.Map{"requests": requests, "total": total})
}

func (h *RequestHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	req, err := h.store.Queries.GetRequestByDealer(c.Context(), db.GetRequestByDealerParams{
		ID:       id,
		DealerID: dealerID,
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
	}

	return c.JSON(req)
}

type createRequestBody struct {
	InputType string `json:"input_type"`
	RawText   string `json:"raw_text"`
	MediaURL  string `json:"media_url"`
}

func (h *RequestHandler) Create(c *fiber.Ctx) error {
	var body createRequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	claims, err := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	// Validate input type
	if !domain.InputType(body.InputType).Valid() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input_type: must be text, voice, image, or pdf"})
	}

	// Validate input size limits
	if len(body.RawText) > 50000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "raw_text exceeds maximum length of 50000 characters"})
	}

	// Validate media URL if provided (SSRF prevention)
	if body.MediaURL != "" {
		if err := validateMediaURL(body.MediaURL); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid media URL"})
		}
	}

	req, err := h.reqSvc.Create(c.Context(), service.CreateRequestInput{
		DealerID:     dealerID,
		ContractorID: claims.UserID,
		InputType:    domain.InputType(body.InputType),
		RawText:      body.RawText,
		MediaURL:     body.MediaURL,
	})
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to create request"})
	}

	if h.metrics != nil {
		h.metrics.RecordInputType(body.InputType)
		h.metrics.RecordRequestStatus("pending")
	}

	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *RequestHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	var body struct {
		Notes string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if len(body.Notes) > 10000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "notes exceeds maximum length of 10000 characters"})
	}

	err = h.store.Queries.UpdateRequestNotesByDealer(c.Context(), db.UpdateRequestNotesByDealerParams{
		ID:       id,
		Notes:    body.Notes,
		DealerID: dealerID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update request"})
	}

	req, err := h.store.Queries.GetRequestByDealer(c.Context(), db.GetRequestByDealerParams{
		ID:       id,
		DealerID: dealerID,
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
	}
	return c.JSON(req)
}

func (h *RequestHandler) Process(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	// Verify request belongs to this tenant
	if _, err := h.store.Queries.GetRequestByDealer(c.Context(), db.GetRequestByDealerParams{
		ID: id, DealerID: dealerID,
	}); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
	}

	req, err := h.reqSvc.Process(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVersionConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "request was modified by another process, please retry"})
		}
		slog.Error("failed to process request", "id", id, "error", err)
		if h.metrics != nil {
			h.metrics.RecordRequestStatus("failed")
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to process request"})
	}

	if h.metrics != nil {
		h.metrics.RecordRequestStatus("parsed")
	}

	return c.JSON(req)
}

func (h *RequestHandler) Confirm(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	claims, _ := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	// Verify request belongs to this tenant
	if _, err := h.store.Queries.GetRequestByDealer(c.Context(), db.GetRequestByDealerParams{
		ID: id, DealerID: dealerID,
	}); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
	}

	var body struct {
		Items json.RawMessage `json:"items"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate items JSON structure if provided
	if body.Items != nil && len(body.Items) > 0 {
		var items []domain.StructuredItem
		if err := json.Unmarshal(body.Items, &items); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid items format: must be an array of structured items"})
		}
		for i, item := range items {
			if item.Name == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("item %d: name is required", i)})
			}
			if item.Quantity <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("item %d: quantity must be positive", i)})
			}
		}
	}

	req, err := h.reqSvc.Confirm(c.Context(), id, body.Items)
	if err != nil {
		if errors.Is(err, domain.ErrVersionConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "request was modified concurrently, please refresh and retry"})
		}
		slog.Error("failed to confirm request", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to confirm request"})
	}

	if h.metrics != nil {
		h.metrics.RecordRequestStatus("confirmed")
	}

	slog.Info("request confirmed",
		"request_id", id,
		"user_id", claims.UserID,
		"dealer_id", dealerID,
	)

	return c.JSON(req)
}

func (h *RequestHandler) Send(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	claims, _ := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))

	// Verify request belongs to this tenant
	if _, err := h.store.Queries.GetRequestByDealer(c.Context(), db.GetRequestByDealerParams{
		ID: id, DealerID: dealerID,
	}); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
	}

	req, err := h.reqSvc.Send(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVersionConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "request was modified concurrently, please refresh and retry"})
		}
		slog.Error("failed to send request", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to send request"})
	}

	if h.metrics != nil {
		h.metrics.RecordRequestStatus("sent")
	}

	slog.Info("request_sent",
		"request_id", id,
		"user_id", claims.UserID,
		"dealer_id", dealerID,
	)

	return c.JSON(req)
}

// validateMediaURL ensures media URLs are safe (SSRF prevention).
// It validates the scheme, blocks known-bad hostnames, checks parsed IPs,
// and resolves DNS to prevent DNS rebinding attacks where a hostname
// resolves to a private/internal IP address.
func validateMediaURL(rawURL string) error {
	if rawURL == "" {
		return nil // caller is responsible for checking empty
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	// Only allow http/https schemes
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" && scheme != "" {
		return domain.ErrInvalidInput
	}

	// Block internal/private hostnames
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return domain.ErrInvalidInput
	}
	blocked := []string{"localhost", "metadata.google.internal", "metadata.google", "169.254.169.254"}
	for _, b := range blocked {
		if host == b {
			return domain.ErrInvalidInput
		}
	}

	// Block private IP ranges by prefix (conservative: blocks all of 172.x.x.x,
	// not just 172.16-31.x.x, because there's no legitimate reason for media
	// URLs to point at any 172.x address).
	if strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "172.") {
		return domain.ErrInvalidInput
	}

	// Check if host is a literal IP
	if ip := net.ParseIP(host); ip != nil {
		return checkIPSafe(ip)
	}

	// DNS resolution check: resolve hostname and verify all IPs are public.
	// This prevents DNS rebinding attacks where a hostname initially resolves
	// to a public IP but later rebinds to an internal address.
	ips, err := net.LookupHost(host)
	if err != nil {
		// DNS resolution failure — block to be safe
		return domain.ErrInvalidInput
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return domain.ErrInvalidInput
		}
		if err := checkIPSafe(ip); err != nil {
			return err
		}
	}

	return nil
}

// checkIPSafe returns an error if the IP is private, loopback, link-local,
// or otherwise unsafe for server-side requests.
func checkIPSafe(ip net.IP) error {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return domain.ErrInvalidInput
	}
	// Block cloud metadata IPs (169.254.x.x link-local range)
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 169 && ip4[1] == 254 {
		return domain.ErrInvalidInput
	}
	return nil
}
