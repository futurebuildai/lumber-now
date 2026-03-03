package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/handler"
	"github.com/builderwire/lumber-now/backend/internal/middleware"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
)

type Handlers struct {
	Health    *handler.HealthHandler
	Auth      *handler.AuthHandler
	Tenant    *handler.TenantHandler
	Request   *handler.RequestHandler
	Inventory *handler.InventoryHandler
	Media     *handler.MediaHandler
	Admin     *handler.AdminHandler
	Platform  *handler.PlatformHandler
}

func Setup(app *fiber.App, s *store.Store, authSvc *service.AuthService, h Handlers, corsOrigins string, metrics *middleware.Metrics) {
	// Global middleware: Recovery -> RequestID -> Metrics -> CORS -> Logging
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(metrics.Handler())
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Tenant-ID, X-Request-ID, X-Requested-With, Idempotency-Key",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(middleware.Logging())

	// Rate limiting: 60 requests per minute per IP
	app.Use(limiter.New(limiter.Config{
		Max:               60,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	// Security headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		return c.Next()
	})

	v1 := app.Group("/v1")

	// Public routes (no tenant, no auth)
	v1.Get("/health", h.Health.Check)
	v1.Get("/readiness", h.Health.Readiness)
	v1.Get("/liveness", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	// Metrics: protected - require Authorization header with platform admin token
	v1.Get("/metrics", middleware.Auth(authSvc), middleware.RequireRole(domain.RolePlatformAdmin), metrics.Endpoint())
	v1.Get("/tenant/config", h.Tenant.GetConfig)

	// Stricter rate limit for auth endpoints (5 req/min per IP)
	authLimiter := limiter.New(limiter.Config{
		Max:               5,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
	})

	// Body size limiter for auth endpoints (1KB max - credentials only)
	authBodyLimit := func(c *fiber.Ctx) error {
		if len(c.Body()) > 1024 {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "request body too large",
			})
		}
		return c.Next()
	}

	// Tenant-scoped public routes
	tenanted := v1.Group("", middleware.Tenant(s))
	tenanted.Post("/auth/login", authLimiter, authBodyLimit, h.Auth.Login)
	tenanted.Post("/auth/register", authLimiter, authBodyLimit, h.Auth.Register)

	// Idempotency cache for POST requests (24h TTL)
	idempotencyCache := middleware.NewIdempotencyCache(24 * time.Hour)

	// Authenticated routes (tenant + auth + CSRF + request timeout + idempotency)
	authed := tenanted.Group("", middleware.Auth(authSvc), middleware.CSRFProtection(), middleware.RequestTimeout(30*time.Second), middleware.Idempotency(idempotencyCache))
	authed.Get("/auth/me", h.Auth.Me)
	authed.Post("/auth/refresh", authLimiter, h.Auth.Refresh)
	authed.Post("/auth/logout", h.Auth.Logout)

	// Requests (any authenticated user)
	authed.Get("/requests", h.Request.List)
	authed.Post("/requests", h.Request.Create)
	authed.Get("/requests/:id", h.Request.Get)
	authed.Put("/requests/:id", h.Request.Update)
	authed.Post("/requests/:id/process", h.Request.Process)
	authed.Post("/requests/:id/confirm", h.Request.Confirm)
	authed.Post("/requests/:id/send", h.Request.Send)

	// Inventory
	authed.Get("/inventory", h.Inventory.List)
	authed.Post("/inventory", middleware.RequireRole(domain.RoleDealerAdmin, domain.RoleSalesRep), h.Inventory.Create)
	authed.Put("/inventory/:id", middleware.RequireRole(domain.RoleDealerAdmin, domain.RoleSalesRep), h.Inventory.Update)
	authed.Delete("/inventory/:id", middleware.RequireRole(domain.RoleDealerAdmin), h.Inventory.Delete)
	authed.Post("/inventory/import", middleware.RequireRole(domain.RoleDealerAdmin), h.Inventory.ImportCSV)

	// Media
	authed.Post("/media/upload", h.Media.Upload)
	authed.Get("/media/:key", h.Media.Download)

	// Dealer admin routes
	admin := authed.Group("/admin", middleware.RequireRole(domain.RoleDealerAdmin), middleware.AuditLog())
	admin.Get("/requests", h.Admin.ListRequests)
	admin.Get("/users", h.Admin.ListUsers)
	admin.Post("/users/:id/assign", h.Admin.AssignContractorToRep)
	admin.Put("/routing", h.Admin.UpdateRouting)
	admin.Get("/settings", h.Admin.GetSettings)
	admin.Put("/settings", h.Admin.UpdateSettings)

	// Platform admin routes
	platform := v1.Group("/platform", middleware.Auth(authSvc), middleware.RequireRole(domain.RolePlatformAdmin), middleware.RequestTimeout(30*time.Second), middleware.AuditLog())
	platform.Get("/dealers", h.Platform.ListDealers)
	platform.Get("/dealers/:id", h.Platform.GetDealer)
	platform.Post("/dealers", h.Platform.CreateDealer)
	platform.Put("/dealers/:id", h.Platform.UpdateDealer)
	platform.Post("/dealers/:id/activate", h.Platform.ActivateDealer)
	platform.Post("/dealers/:id/deactivate", h.Platform.DeactivateDealer)
	platform.Get("/builds", h.Platform.ListBuilds)
	platform.Post("/builds", h.Platform.TriggerBuild)
	platform.Post("/media/upload", h.Platform.UploadLogo)
	platform.Post("/dealers/:id/users", h.Platform.CreateDealerUser)
}
