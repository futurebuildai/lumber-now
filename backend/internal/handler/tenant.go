package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

type TenantHandler struct {
	store *store.Store
}

func NewTenantHandler(s *store.Store) *TenantHandler {
	return &TenantHandler{store: s}
}

func (h *TenantHandler) GetConfig(c *fiber.Ctx) error {
	slug := c.Query("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug query parameter required"})
	}

	dealer, err := h.store.Queries.GetDealerBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dealer not found"})
	}

	return c.JSON(domain.TenantConfig{
		DealerID:       dealer.ID,
		Name:           dealer.Name,
		Slug:           dealer.Slug,
		LogoURL:        dealer.LogoUrl,
		PrimaryColor:   dealer.PrimaryColor,
		SecondaryColor: dealer.SecondaryColor,
		ContactEmail:   dealer.ContactEmail,
		ContactPhone:   dealer.ContactPhone,
	})
}
