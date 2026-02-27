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

		rid, _ := c.Locals(domain.LocalsRequestID).(string)
		slog.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
			"request_id", rid,
		)

		return err
	}
}
