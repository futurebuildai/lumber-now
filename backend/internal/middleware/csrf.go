package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// CSRFProtection validates that state-changing requests (POST/PUT/DELETE)
// include a custom header (X-Requested-With). Since browsers block custom
// headers in cross-origin requests unless CORS is explicitly configured,
// this acts as a lightweight CSRF protection without tokens.
//
// This is the "double submit" / "custom header" pattern recommended for
// APIs consumed by SPAs — the frontend always sends the header, and
// cross-origin requests from malicious sites cannot set custom headers
// due to the same-origin policy.
func CSRFProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// Require X-Requested-With header for state-changing operations
		if c.Get("X-Requested-With") == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "missing X-Requested-With header",
			})
		}

		return c.Next()
	}
}
