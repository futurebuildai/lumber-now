package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/service"
)

func Auth(authSvc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		claims, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Cross-check dealer_id from JWT against tenant middleware
		if dealerVal := c.Locals(domain.LocalsDealerID); dealerVal != nil {
			tenantDealerID, ok := dealerVal.(uuid.UUID)
			if ok && claims.DealerID != tenantDealerID {
				// Platform admins can access any tenant
				if claims.Role != domain.RolePlatformAdmin {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "token dealer does not match tenant",
					})
				}
			}
		}

		c.Locals(domain.LocalsClaims, claims)
		return c.Next()
	}
}
