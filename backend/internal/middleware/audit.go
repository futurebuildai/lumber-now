package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/domain"
)

// AuditLog logs admin operations for audit trail.
func AuditLog() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		// Only log state-changing operations
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return err
		}

		claims, _ := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims))

		attrs := []any{
			"method", method,
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"ip", c.IP(),
		}

		if claims != nil {
			attrs = append(attrs,
				"user_id", claims.UserID,
				"dealer_id", claims.DealerID,
				"role", claims.Role,
			)
		}

		requestID, _ := c.Locals(domain.LocalsRequestID).(string)
		if requestID == "" {
			requestID = c.Get("X-Request-ID")
		}
		if requestID != "" {
			attrs = append(attrs, "request_id", requestID)
		}

		slog.Info("audit", attrs...)
		return err
	}
}
