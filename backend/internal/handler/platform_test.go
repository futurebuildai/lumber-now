package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewPlatformHandler_NilDeps(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	if h == nil {
		t.Fatal("NewPlatformHandler(nil, nil, nil) returned nil")
	}
	if h.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
	if h.authSvc != nil {
		t.Error("expected authSvc field to be nil when constructed with nil")
	}
	if h.mediaSvc != nil {
		t.Error("expected mediaSvc field to be nil when constructed with nil")
	}
}

// ---------------------------------------------------------------------------
// UpdateDealer — invalid UUID returns 400
// ---------------------------------------------------------------------------

func TestPlatformUpdateDealer_InvalidUUID(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Put("/platform/dealers/:id", h.UpdateDealer)

	req := httptest.NewRequest("PUT", "/platform/dealers/not-a-uuid", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid dealer ID" {
		t.Errorf("expected error 'invalid dealer ID', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// ActivateDealer — invalid UUID returns 400
// ---------------------------------------------------------------------------

func TestPlatformActivateDealer_InvalidUUID(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/dealers/:id/activate", h.ActivateDealer)

	req := httptest.NewRequest("POST", "/platform/dealers/bad-id/activate", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid dealer ID" {
		t.Errorf("expected error 'invalid dealer ID', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// DeactivateDealer — invalid UUID returns 400
// ---------------------------------------------------------------------------

func TestPlatformDeactivateDealer_InvalidUUID(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/dealers/:id/deactivate", h.DeactivateDealer)

	req := httptest.NewRequest("POST", "/platform/dealers/xyz/deactivate", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid dealer ID" {
		t.Errorf("expected error 'invalid dealer ID', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// CreateDealerUser — invalid UUID returns 400
// ---------------------------------------------------------------------------

func TestPlatformCreateDealerUser_InvalidUUID(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/dealers/:id/users", h.CreateDealerUser)

	req := httptest.NewRequest("POST", "/platform/dealers/not-uuid/users", strings.NewReader(`{"email":"a@b.com"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid dealer ID" {
		t.Errorf("expected error 'invalid dealer ID', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// CreateDealer — missing body returns 400
// ---------------------------------------------------------------------------

func TestPlatformCreateDealer_MissingBody(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/dealers", h.CreateDealer)

	req := httptest.NewRequest("POST", "/platform/dealers", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// CreateDealer — invalid JSON returns 400
// ---------------------------------------------------------------------------

func TestPlatformCreateDealer_InvalidJSON(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/dealers", h.CreateDealer)

	req := httptest.NewRequest("POST", "/platform/dealers", strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// TriggerBuild — missing body returns 400
// ---------------------------------------------------------------------------

func TestPlatformTriggerBuild_MissingBody(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/builds", h.TriggerBuild)

	req := httptest.NewRequest("POST", "/platform/builds", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// UpdateDealer — valid UUID but missing body returns 400
// ---------------------------------------------------------------------------

func TestPlatformUpdateDealer_ValidUUID_MissingBody(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Put("/platform/dealers/:id", h.UpdateDealer)

	req := httptest.NewRequest("PUT", "/platform/dealers/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// CreateDealerUser — valid UUID but missing body returns 400
// ---------------------------------------------------------------------------

func TestPlatformCreateDealerUser_ValidUUID_MissingBody(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/dealers/:id/users", h.CreateDealerUser)

	req := httptest.NewRequest("POST", "/platform/dealers/550e8400-e29b-41d4-a716-446655440000/users", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// ListBuilds — returns empty builds array
// ---------------------------------------------------------------------------

func TestPlatformListBuilds_ReturnsEmptyArray(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Get("/platform/builds", h.ListBuilds)

	req := httptest.NewRequest("GET", "/platform/builds", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	builds, ok := result["builds"].([]interface{})
	if !ok {
		t.Fatal("expected builds to be an array")
	}
	if len(builds) != 0 {
		t.Errorf("expected empty builds array, got %d elements", len(builds))
	}
}

// ---------------------------------------------------------------------------
// UploadLogo — nil mediaSvc returns 503
// ---------------------------------------------------------------------------

func TestPlatformUploadLogo_NilMediaSvc(t *testing.T) {
	h := NewPlatformHandler(nil, nil, nil)
	app := fiber.New()
	app.Post("/platform/logo", h.UploadLogo)

	req := httptest.NewRequest("POST", "/platform/logo", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "media uploads not configured" {
		t.Errorf("expected error 'media uploads not configured', got %v", result["error"])
	}
}
