package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

func Tenant(s *store.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try X-Tenant-ID header first (UUID)
		tenantHeader := c.Get("X-Tenant-ID")
		if tenantHeader != "" {
			dealerID, err := uuid.Parse(tenantHeader)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid X-Tenant-ID header",
				})
			}
			c.Locals(domain.LocalsDealerID, dealerID)
			return c.Next()
		}

		// Try subdomain resolution
		host := c.Hostname()
		parts := strings.Split(host, ".")
		if len(parts) >= 3 {
			subdomain := parts[0]
			dealer, err := s.Queries.GetDealerBySubdomain(c.Context(), subdomain)
			if err == nil {
				c.Locals(domain.LocalsDealerID, dealer.ID)
				return c.Next()
			}
		}

		// Try slug query parameter
		slug := c.Query("tenant")
		if slug != "" {
			dealer, err := s.Queries.GetDealerBySlug(c.Context(), slug)
			if err == nil {
				c.Locals(domain.LocalsDealerID, dealer.ID)
				return c.Next()
			}
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tenant identification required: set X-Tenant-ID header, use subdomain, or pass ?tenant=slug",
		})
	}
}
