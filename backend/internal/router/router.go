package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

func Setup(app *fiber.App, s *store.Store, authSvc *service.AuthService, h Handlers, corsOrigins string) {
	// Global middleware: Recovery -> RequestID -> CORS -> Logging
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Tenant-ID, X-Request-ID",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(middleware.Logging())

	v1 := app.Group("/v1")

	// Public routes (no tenant, no auth)
	v1.Get("/health", h.Health.Check)
	v1.Get("/tenant/config", h.Tenant.GetConfig)

	// Tenant-scoped public routes
	tenanted := v1.Group("", middleware.Tenant(s))
	tenanted.Post("/auth/login", h.Auth.Login)
	tenanted.Post("/auth/register", h.Auth.Register)

	// Authenticated routes (tenant + auth)
	authed := tenanted.Group("", middleware.Auth(authSvc))
	authed.Get("/auth/me", h.Auth.Me)
	authed.Post("/auth/refresh", h.Auth.Refresh)

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
	admin := authed.Group("/admin", middleware.RequireRole(domain.RoleDealerAdmin))
	admin.Get("/requests", h.Admin.ListRequests)
	admin.Get("/users", h.Admin.ListUsers)
	admin.Post("/users/:id/assign", h.Admin.AssignContractorToRep)
	admin.Put("/routing", h.Admin.UpdateRouting)
	admin.Get("/settings", h.Admin.GetSettings)
	admin.Put("/settings", h.Admin.UpdateSettings)

	// Platform admin routes
	platform := v1.Group("/platform", middleware.Auth(authSvc), middleware.RequireRole(domain.RolePlatformAdmin))
	platform.Get("/dealers", h.Platform.ListDealers)
	platform.Post("/dealers", h.Platform.CreateDealer)
	platform.Put("/dealers/:id", h.Platform.UpdateDealer)
	platform.Post("/dealers/:id/activate", h.Platform.ActivateDealer)
	platform.Post("/dealers/:id/deactivate", h.Platform.DeactivateDealer)
	platform.Get("/builds", h.Platform.ListBuilds)
	platform.Post("/builds", h.Platform.TriggerBuild)
	platform.Post("/media/upload", h.Platform.UploadLogo)
	platform.Post("/dealers/:id/users", h.Platform.CreateDealerUser)
}
