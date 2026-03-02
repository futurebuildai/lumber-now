package handler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type PlatformHandler struct {
	store    *store.Store
	authSvc  *service.AuthService
	mediaSvc *service.MediaService
}

func NewPlatformHandler(s *store.Store, authSvc *service.AuthService, mediaSvc *service.MediaService) *PlatformHandler {
	return &PlatformHandler{store: s, authSvc: authSvc, mediaSvc: mediaSvc}
}

func (h *PlatformHandler) ListDealers(c *fiber.Ctx) error {
	limit := clampLimit(int32(c.QueryInt("limit", 50)))
	offset := clampOffset(int32(c.QueryInt("offset", 0)))

	dealers, err := h.store.Queries.ListDealers(c.Context(), db.ListDealersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list dealers"})
	}

	return c.JSON(fiber.Map{"dealers": dealers})
}

func (h *PlatformHandler) CreateDealer(c *fiber.Ctx) error {
	var body struct {
		Name           string `json:"name"`
		Slug           string `json:"slug"`
		Subdomain      string `json:"subdomain"`
		LogoURL        string `json:"logo_url"`
		PrimaryColor   string `json:"primary_color"`
		SecondaryColor string `json:"secondary_color"`
		ContactEmail   string `json:"contact_email"`
		ContactPhone   string `json:"contact_phone"`
		Address        string `json:"address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.Name == "" || body.Slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and slug are required"})
	}

	dealer, err := h.store.Queries.CreateDealer(c.Context(), db.CreateDealerParams{
		Name:           body.Name,
		Slug:           body.Slug,
		Subdomain:      body.Subdomain,
		LogoUrl:        body.LogoURL,
		PrimaryColor:   body.PrimaryColor,
		SecondaryColor: body.SecondaryColor,
		ContactEmail:   body.ContactEmail,
		ContactPhone:   body.ContactPhone,
		Address:        body.Address,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to create dealer"})
	}

	return c.Status(fiber.StatusCreated).JSON(dealer)
}

func (h *PlatformHandler) UpdateDealer(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dealer ID"})
	}

	var body struct {
		Name           string `json:"name"`
		LogoURL        string `json:"logo_url"`
		PrimaryColor   string `json:"primary_color"`
		SecondaryColor string `json:"secondary_color"`
		ContactEmail   string `json:"contact_email"`
		ContactPhone   string `json:"contact_phone"`
		Address        string `json:"address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	dealer, err := h.store.Queries.UpdateDealer(c.Context(), db.UpdateDealerParams{
		ID:             id,
		Name:           body.Name,
		LogoUrl:        body.LogoURL,
		PrimaryColor:   body.PrimaryColor,
		SecondaryColor: body.SecondaryColor,
		ContactEmail:   body.ContactEmail,
		ContactPhone:   body.ContactPhone,
		Address:        body.Address,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update dealer"})
	}

	return c.JSON(dealer)
}

func (h *PlatformHandler) ActivateDealer(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dealer ID"})
	}

	err = h.store.Queries.SetDealerActive(c.Context(), db.SetDealerActiveParams{
		ID:     id,
		Active: true,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to activate dealer"})
	}

	return c.JSON(fiber.Map{"status": "activated"})
}

func (h *PlatformHandler) DeactivateDealer(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dealer ID"})
	}

	err = h.store.Queries.SetDealerActive(c.Context(), db.SetDealerActiveParams{
		ID:     id,
		Active: false,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to deactivate dealer"})
	}

	return c.JSON(fiber.Map{"status": "deactivated"})
}

func (h *PlatformHandler) ListBuilds(c *fiber.Ctx) error {
	// Placeholder - build tracking would be stored separately
	return c.JSON(fiber.Map{"builds": []interface{}{}})
}

func (h *PlatformHandler) TriggerBuild(c *fiber.Ctx) error {
	var body struct {
		DealerSlug string `json:"dealer_slug"`
		Platform   string `json:"platform"` // "android" or "ios"
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Placeholder - would trigger generate-build.sh via worker
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":      "queued",
		"dealer_slug": body.DealerSlug,
		"platform":    body.Platform,
	})
}

// UploadLogo handles logo upload without requiring tenant context.
// Used during dealer onboarding before a tenant is fully set up.
func (h *PlatformHandler) UploadLogo(c *fiber.Ctx) error {
	if h.mediaSvc == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "media uploads not configured"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required"})
	}

	// Enforce file size limit (5MB for logos)
	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "file too large (max 5MB)"})
	}

	// Validate content type: only image types allowed for logos
	allowedLogoTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/webp":    true,
		"image/svg+xml": true,
	}
	ct := file.Header.Get("Content-Type")
	if ct == "" || !allowedLogoTypes[ct] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported file type (allowed: jpeg, png, webp, svg)"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer f.Close()

	// Sanitize filename: keep only base name, strip path traversal
	safeName := filepath.Base(file.Filename)
	safeName = strings.ReplaceAll(safeName, "..", "")
	filename := fmt.Sprintf("%s-%s", uuid.New().String()[:8], safeName)
	contentType := ct

	// Store under a platform-level path since there's no dealer context yet
	platformID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	key, err := h.mediaSvc.Upload(c.Context(), platformID, filename, f, contentType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "upload failed"})
	}

	// Generate a presigned URL for immediate use
	url, _ := h.mediaSvc.GetPresignedURL(c.Context(), key)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"key": key,
		"url": url,
	})
}

// CreateDealerUser creates the first admin user for a newly created dealer.
func (h *PlatformHandler) CreateDealerUser(c *fiber.Ctx) error {
	dealerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dealer ID"})
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.Email == "" || body.Password == "" || body.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email, password, and full_name are required"})
	}

	user, err := h.authSvc.Register(c.Context(), service.RegisterInput{
		DealerID: dealerID,
		Email:    body.Email,
		Password: body.Password,
		FullName: body.FullName,
		Phone:    body.Phone,
		Role:     domain.RoleDealerAdmin,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        user.ID,
		"dealer_id": user.DealerID,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
		"active":    user.Active,
	})
}
