package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/domain"
)

func RequireRole(roles ...domain.Role) fiber.Handler {
	allowed := make(map[domain.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *fiber.Ctx) error {
		claims, err := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		if !allowed[claims.Role] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}
