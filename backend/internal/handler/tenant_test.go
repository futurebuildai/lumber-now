package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewTenantHandler_NilStore(t *testing.T) {
	h := NewTenantHandler(nil)
	if h == nil {
		t.Fatal("NewTenantHandler(nil) returned nil")
	}
	if h.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
}

// ---------------------------------------------------------------------------
// GetConfig — missing slug query parameter returns 400
// ---------------------------------------------------------------------------

func TestTenantGetConfig_MissingSlug(t *testing.T) {
	h := NewTenantHandler(nil)
	app := fiber.New()
	app.Get("/tenant/config", h.GetConfig)

	req := httptest.NewRequest("GET", "/tenant/config", nil)

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
	if result["error"] != "slug query parameter required" {
		t.Errorf("expected error 'slug query parameter required', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// GetConfig — empty slug query parameter returns 400
// ---------------------------------------------------------------------------

func TestTenantGetConfig_EmptySlug(t *testing.T) {
	h := NewTenantHandler(nil)
	app := fiber.New()
	app.Get("/tenant/config", h.GetConfig)

	req := httptest.NewRequest("GET", "/tenant/config?slug=", nil)

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
	if result["error"] != "slug query parameter required" {
		t.Errorf("expected error 'slug query parameter required', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// GetConfig — returns JSON content type
// ---------------------------------------------------------------------------

func TestTenantGetConfig_ReturnsJSONContentType(t *testing.T) {
	h := NewTenantHandler(nil)
	app := fiber.New()
	app.Get("/tenant/config", h.GetConfig)

	// Without a real store, providing a slug will hit the store and panic or
	// return 404/500. We only verify that the missing-slug path returns JSON.
	req := httptest.NewRequest("GET", "/tenant/config", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
