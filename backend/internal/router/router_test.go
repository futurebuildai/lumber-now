package router

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/handler"
	"github.com/builderwire/lumber-now/backend/internal/middleware"
	"github.com/builderwire/lumber-now/backend/internal/service"
)

// TestHandlersStruct_ZeroValue verifies that the Handlers struct fields are
// properly nil when zero-initialized.
func TestHandlersStruct_ZeroValue(t *testing.T) {
	var h Handlers
	if h.Health != nil {
		t.Error("zero-value Health should be nil")
	}
	if h.Auth != nil {
		t.Error("zero-value Auth should be nil")
	}
	if h.Tenant != nil {
		t.Error("zero-value Tenant should be nil")
	}
	if h.Request != nil {
		t.Error("zero-value Request should be nil")
	}
	if h.Inventory != nil {
		t.Error("zero-value Inventory should be nil")
	}
	if h.Media != nil {
		t.Error("zero-value Media should be nil")
	}
	if h.Admin != nil {
		t.Error("zero-value Admin should be nil")
	}
	if h.Platform != nil {
		t.Error("zero-value Platform should be nil")
	}
}

// TestHandlersStruct_PopulatedFields verifies that the Handlers struct can
// hold non-nil handler pointers.
func TestHandlersStruct_PopulatedFields(t *testing.T) {
	h := Handlers{
		Health:    handler.NewHealthHandler(nil),
		Admin:     handler.NewAdminHandler(nil),
		Media:     handler.NewMediaHandler(nil),
		Auth:      handler.NewAuthHandler(nil, nil),
		Tenant:    handler.NewTenantHandler(nil),
		Inventory: handler.NewInventoryHandler(nil, nil),
	}

	if h.Health == nil {
		t.Error("Health should not be nil after construction")
	}
	if h.Admin == nil {
		t.Error("Admin should not be nil after construction")
	}
	if h.Media == nil {
		t.Error("Media should not be nil after construction")
	}
	if h.Auth == nil {
		t.Error("Auth should not be nil after construction")
	}
	if h.Tenant == nil {
		t.Error("Tenant should not be nil after construction")
	}
	if h.Inventory == nil {
		t.Error("Inventory should not be nil after construction")
	}
}

// TestSetup_DoesNotPanic verifies that Setup registers routes without
// panicking, even with nil store and authSvc (the middleware closures
// only execute at request time, not during route registration).
func TestSetup_DoesNotPanic(t *testing.T) {
	app := fiber.New()
	metrics := middleware.NewMetrics()
	authSvc := service.NewAuthService(nil, "test-secret")

	h := Handlers{
		Health:    handler.NewHealthHandler(nil),
		Auth:      handler.NewAuthHandler(nil, nil),
		Tenant:    handler.NewTenantHandler(nil),
		Request:   handler.NewRequestHandler(nil, nil),
		Inventory: handler.NewInventoryHandler(nil, nil),
		Media:     handler.NewMediaHandler(nil),
		Admin:     handler.NewAdminHandler(nil),
		Platform:  handler.NewPlatformHandler(nil, nil, nil),
	}

	// Setup should not panic when registering routes
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Setup panicked: %v", r)
		}
	}()

	Setup(app, nil, authSvc, h, "*", metrics)
}

// TestSetup_LivenessEndpoint verifies that the /v1/liveness route returns
// a 200 response with {"status":"ok"}. This endpoint is a simple inline
// handler that does not require any dependencies.
func TestSetup_LivenessEndpoint(t *testing.T) {
	app := fiber.New()
	metrics := middleware.NewMetrics()
	authSvc := service.NewAuthService(nil, "test-secret")

	h := Handlers{
		Health:    handler.NewHealthHandler(nil),
		Auth:      handler.NewAuthHandler(nil, nil),
		Tenant:    handler.NewTenantHandler(nil),
		Request:   handler.NewRequestHandler(nil, nil),
		Inventory: handler.NewInventoryHandler(nil, nil),
		Media:     handler.NewMediaHandler(nil),
		Admin:     handler.NewAdminHandler(nil),
		Platform:  handler.NewPlatformHandler(nil, nil, nil),
	}

	Setup(app, nil, authSvc, h, "*", metrics)

	req := httptest.NewRequest("GET", "/v1/liveness", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("liveness status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("liveness status = %v, want ok", result["status"])
	}
}

// TestSetup_SecurityHeaders verifies that security headers are set on
// responses from the liveness endpoint.
func TestSetup_SecurityHeaders(t *testing.T) {
	app := fiber.New()
	metrics := middleware.NewMetrics()
	authSvc := service.NewAuthService(nil, "test-secret")

	h := Handlers{
		Health:    handler.NewHealthHandler(nil),
		Auth:      handler.NewAuthHandler(nil, nil),
		Tenant:    handler.NewTenantHandler(nil),
		Request:   handler.NewRequestHandler(nil, nil),
		Inventory: handler.NewInventoryHandler(nil, nil),
		Media:     handler.NewMediaHandler(nil),
		Admin:     handler.NewAdminHandler(nil),
		Platform:  handler.NewPlatformHandler(nil, nil, nil),
	}

	Setup(app, nil, authSvc, h, "*", metrics)

	req := httptest.NewRequest("GET", "/v1/liveness", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"X-XSS-Protection":      "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expected := range expectedHeaders {
		got := resp.Header.Get(header)
		if got != expected {
			t.Errorf("header %s = %q, want %q", header, got, expected)
		}
	}
}

// TestSetup_UnknownRoute verifies that unregistered routes do not return 200.
// The tenant middleware intercepts requests under /v1 and returns 400 when
// no tenant header is present, so unknown tenant-scoped routes get 400.
func TestSetup_UnknownRoute(t *testing.T) {
	app := fiber.New()
	metrics := middleware.NewMetrics()
	authSvc := service.NewAuthService(nil, "test-secret")

	h := Handlers{
		Health:    handler.NewHealthHandler(nil),
		Auth:      handler.NewAuthHandler(nil, nil),
		Tenant:    handler.NewTenantHandler(nil),
		Request:   handler.NewRequestHandler(nil, nil),
		Inventory: handler.NewInventoryHandler(nil, nil),
		Media:     handler.NewMediaHandler(nil),
		Admin:     handler.NewAdminHandler(nil),
		Platform:  handler.NewPlatformHandler(nil, nil, nil),
	}

	Setup(app, nil, authSvc, h, "*", metrics)

	// A completely unknown top-level route (outside /v1) returns 404.
	req := httptest.NewRequest("GET", "/unknown-top-level", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("unknown top-level route status = %d, want 404", resp.StatusCode)
	}
}

// TestSetup_AuthEndpointsRequireTenant verifies that auth endpoints respond
// with an error when no tenant header is provided.
func TestSetup_AuthEndpointsRequireTenant(t *testing.T) {
	app := fiber.New()
	metrics := middleware.NewMetrics()
	authSvc := service.NewAuthService(nil, "test-secret")

	h := Handlers{
		Health:    handler.NewHealthHandler(nil),
		Auth:      handler.NewAuthHandler(nil, nil),
		Tenant:    handler.NewTenantHandler(nil),
		Request:   handler.NewRequestHandler(nil, nil),
		Inventory: handler.NewInventoryHandler(nil, nil),
		Media:     handler.NewMediaHandler(nil),
		Admin:     handler.NewAdminHandler(nil),
		Platform:  handler.NewPlatformHandler(nil, nil, nil),
	}

	Setup(app, nil, authSvc, h, "*", metrics)

	// POST /v1/auth/login without tenant header should fail
	req := httptest.NewRequest("POST", "/v1/auth/login", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	// Should get a 400 (bad request) or similar because tenant is missing
	if resp.StatusCode == 200 {
		t.Error("auth/login without tenant should not return 200")
	}
}
