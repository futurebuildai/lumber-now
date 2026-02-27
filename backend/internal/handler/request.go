package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type RequestHandler struct {
	reqSvc *service.RequestService
	store  *store.Store
}

func NewRequestHandler(reqSvc *service.RequestService, s *store.Store) *RequestHandler {
	return &RequestHandler{reqSvc: reqSvc, store: s}
}

func (h *RequestHandler) List(c *fiber.Ctx) error {
	claims, err := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	dealerID, _ := domain.DealerIDFromLocals(c.Locals(domain.LocalsDealerID))
	limit := int32(c.QueryInt("limit", 50))
	offset := int32(c.QueryInt("offset", 0))

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

	return c.JSON(fiber.Map{"requests": requests})
}

func (h *RequestHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	req, err := h.store.Queries.GetRequest(c.Context(), id)
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

	req, err := h.reqSvc.Create(c.Context(), service.CreateRequestInput{
		DealerID:     dealerID,
		ContractorID: claims.UserID,
		InputType:    domain.InputType(body.InputType),
		RawText:      body.RawText,
		MediaURL:     body.MediaURL,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *RequestHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	var body struct {
		Notes string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	err = h.store.Queries.UpdateRequestNotes(c.Context(), db.UpdateRequestNotesParams{
		ID:    id,
		Notes: body.Notes,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update request"})
	}

	req, _ := h.store.Queries.GetRequest(c.Context(), id)
	return c.JSON(req)
}

func (h *RequestHandler) Process(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	req, err := h.reqSvc.Process(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(req)
}

func (h *RequestHandler) Confirm(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	var body struct {
		Items json.RawMessage `json:"items"`
	}
	c.BodyParser(&body)

	req, err := h.reqSvc.Confirm(c.Context(), id, body.Items)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(req)
}

func (h *RequestHandler) Send(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	req, err := h.reqSvc.Send(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(req)
}
