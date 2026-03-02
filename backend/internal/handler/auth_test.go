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

func TestNewAuthHandler_NilDeps(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	if h == nil {
		t.Fatal("NewAuthHandler(nil, nil) returned nil")
	}
	if h.authSvc != nil {
		t.Error("expected authSvc field to be nil when constructed with nil")
	}
	if h.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
}

// ---------------------------------------------------------------------------
// Login — missing body returns 400
// ---------------------------------------------------------------------------

func TestAuthLogin_MissingBody(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/login", h.Login)

	req := httptest.NewRequest("POST", "/auth/login", nil)
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
// Login — invalid JSON returns 400
// ---------------------------------------------------------------------------

func TestAuthLogin_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/login", h.Login)

	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{bad json`))
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
// Register — missing body returns 400
// ---------------------------------------------------------------------------

func TestAuthRegister_MissingBody(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/register", h.Register)

	req := httptest.NewRequest("POST", "/auth/register", nil)
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
// Register — invalid JSON returns 400
// ---------------------------------------------------------------------------

func TestAuthRegister_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/register", h.Register)

	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(`not json`))
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
// Refresh — missing body returns 400
// ---------------------------------------------------------------------------

func TestAuthRefresh_MissingBody(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/refresh", h.Refresh)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
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
// Refresh — invalid JSON returns 400
// ---------------------------------------------------------------------------

func TestAuthRefresh_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/refresh", h.Refresh)

	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(`{broken`))
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
// Me — no claims in locals returns 401
// ---------------------------------------------------------------------------

func TestAuthMe_NoClaims(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Get("/auth/me", h.Me)

	req := httptest.NewRequest("GET", "/auth/me", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %v", result["error"])
	}
}

// ---------------------------------------------------------------------------
// Logout — returns success even without Authorization header
// ---------------------------------------------------------------------------

func TestAuthLogout_NoHeader(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/logout", h.Logout)

	req := httptest.NewRequest("POST", "/auth/logout", nil)

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
	if result["message"] != "logged out" {
		t.Errorf("expected message 'logged out', got %v", result["message"])
	}
}

// ---------------------------------------------------------------------------
// Login — returns JSON content type on error
// ---------------------------------------------------------------------------

func TestAuthLogin_ReturnsJSONContentType(t *testing.T) {
	h := NewAuthHandler(nil, nil)
	app := fiber.New()
	app.Post("/auth/login", h.Login)

	req := httptest.NewRequest("POST", "/auth/login", nil)
	req.Header.Set("Content-Type", "application/json")

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
