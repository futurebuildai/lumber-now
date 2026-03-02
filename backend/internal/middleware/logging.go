package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/domain"
)

func Logging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		rid, _ := c.Locals(domain.LocalsRequestID).(string)
		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", rid,
			"ip", c.IP(),
			"bytes_in", len(c.Body()),
			"bytes_out", len(c.Response().Body()),
		}

		// Include tenant/user context if available
		if claims, cErr := domain.ClaimsFromLocals(c.Locals(domain.LocalsClaims)); cErr == nil {
			attrs = append(attrs, "user_id", claims.UserID, "dealer_id", claims.DealerID, "role", claims.Role)
		}

		// Use appropriate log level based on status code
		switch {
		case status >= 500:
			slog.Error("request", attrs...)
		case status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}

		return err
	}
}
