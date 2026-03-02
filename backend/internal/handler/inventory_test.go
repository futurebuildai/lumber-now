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

func TestNewInventoryHandler_NilDeps(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	if h == nil {
		t.Fatal("NewInventoryHandler(nil, nil) returned nil")
	}
	if h.invSvc != nil {
		t.Error("expected invSvc field to be nil when constructed with nil")
	}
	if h.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
}

// ---------------------------------------------------------------------------
// Update — invalid UUID param returns 400
// ---------------------------------------------------------------------------

func TestInventoryUpdate_InvalidUUID(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Put("/inventory/:id", h.Update)

	req := httptest.NewRequest("PUT", "/inventory/not-a-uuid", strings.NewReader(`{"name":"x"}`))
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
	if result["error"] != "invalid item ID" {
		t.Errorf("expected error 'invalid item ID', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// Delete — invalid UUID param returns 400
// ---------------------------------------------------------------------------

func TestInventoryDelete_InvalidUUID(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Delete("/inventory/:id", h.Delete)

	req := httptest.NewRequest("DELETE", "/inventory/bad-id", nil)

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
	if result["error"] != "invalid item ID" {
		t.Errorf("expected error 'invalid item ID', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// Create — missing/invalid body returns 400
// ---------------------------------------------------------------------------

func TestInventoryCreate_MissingBody(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Post("/inventory", h.Create)

	req := httptest.NewRequest("POST", "/inventory", nil)
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

func TestInventoryCreate_InvalidJSON(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Post("/inventory", h.Create)

	req := httptest.NewRequest("POST", "/inventory", strings.NewReader(`{invalid json`))
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
// Update — missing body returns 400 (after valid UUID)
// ---------------------------------------------------------------------------

func TestInventoryUpdate_ValidUUID_MissingBody(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Put("/inventory/:id", h.Update)

	req := httptest.NewRequest("PUT", "/inventory/550e8400-e29b-41d4-a716-446655440000", nil)
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
// List — no dealer ID in locals returns 400
// ---------------------------------------------------------------------------

func TestInventoryList_NoDealerID(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Get("/inventory", h.List)

	req := httptest.NewRequest("GET", "/inventory", nil)

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
	if result["error"] != "tenant required" {
		t.Errorf("expected error 'tenant required', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// ImportCSV — no dealer ID in locals returns 400
// ---------------------------------------------------------------------------

func TestInventoryImportCSV_NoDealerID(t *testing.T) {
	h := NewInventoryHandler(nil, nil)
	app := fiber.New()
	app.Post("/inventory/import", h.ImportCSV)

	req := httptest.NewRequest("POST", "/inventory/import", nil)

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
	if result["error"] != "tenant required" {
		t.Errorf("expected error 'tenant required', got %v", result["error"])
	}
}
