package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCSRFProtection_AllowsGET(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("GET without header: status = %d, want 200", resp.StatusCode)
	}
}

func TestCSRFProtection_AllowsHEAD(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Head("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("HEAD without header: status = %d, want 200", resp.StatusCode)
	}
}

func TestCSRFProtection_AllowsOPTIONS(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Options("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("OPTIONS without header: status = %d, want 204", resp.StatusCode)
	}
}

func TestCSRFProtection_BlocksPOSTWithoutHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("POST without X-Requested-With: status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "X-Requested-With") {
		t.Errorf("expected error message about X-Requested-With header, got: %s", body)
	}
}

func TestCSRFProtection_AllowsPOSTWithHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("POST with X-Requested-With: status = %d, want 200", resp.StatusCode)
	}
}

func TestCSRFProtection_BlocksPUTWithoutHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Put("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("PUT", "/test", strings.NewReader(`{}`))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("PUT without X-Requested-With: status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestCSRFProtection_BlocksDELETEWithoutHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("DELETE without X-Requested-With: status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestCSRFProtection_AllowsDELETEWithHeader(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFProtection())
	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	req.Header.Set("X-Requested-With", "fetch")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("DELETE with X-Requested-With: status = %d, want 204", resp.StatusCode)
	}
}
