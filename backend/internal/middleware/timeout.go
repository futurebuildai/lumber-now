package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

// RequestTimeout returns middleware that enforces a per-request timeout.
func RequestTimeout(d time.Duration) fiber.Handler {
	return timeout.NewWithContext(func(c *fiber.Ctx) error {
		return c.Next()
	}, d)
}
