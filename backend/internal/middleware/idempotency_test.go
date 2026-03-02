package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestIdempotency_NoPOST_Passthrough(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Idempotency-Key", "key-123")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("GET request status = %d, want 200", resp.StatusCode)
	}
	if resp.Header.Get("X-Idempotency-Replay") != "" {
		t.Error("GET request should not be replayed")
	}
}

func TestIdempotency_POSTWithoutKey_Passthrough(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Post("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(201).SendString("created")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	if callCount != 3 {
		t.Errorf("without idempotency key: handler called %d times, want 3", callCount)
	}
}

func TestIdempotency_POSTWithKey_CachesResponse(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Post("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(201).JSON(fiber.Map{"id": "abc-123"})
	})

	key := "unique-key-456"

	// First request
	req1 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req1.Header.Set("Idempotency-Key", key)
	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatal(err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	if resp1.StatusCode != 201 {
		t.Errorf("first request status = %d, want 201", resp1.StatusCode)
	}
	if resp1.Header.Get("X-Idempotency-Replay") != "" {
		t.Error("first request should not be a replay")
	}

	// Second request with same key
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req2.Header.Set("Idempotency-Key", key)
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatal(err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	if resp2.StatusCode != 201 {
		t.Errorf("second request status = %d, want 201", resp2.StatusCode)
	}
	if resp2.Header.Get("X-Idempotency-Replay") != "true" {
		t.Error("second request should be marked as replay")
	}

	if string(body1) != string(body2) {
		t.Errorf("response bodies differ:\n  first:  %s\n  second: %s", body1, body2)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1 (cached on retry)", callCount)
	}
}

func TestIdempotency_DifferentKeys_ExecuteSeparately(t *testing.T) {
	cache := NewIdempotencyCache(5 * time.Minute)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Post("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(201).SendString("ok")
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
		req.Header.Set("Idempotency-Key", "key-"+string(rune('a'+i)))
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	if callCount != 3 {
		t.Errorf("different keys: handler called %d times, want 3", callCount)
	}
}

func TestIdempotency_ExpiredEntry_ReExecutes(t *testing.T) {
	cache := NewIdempotencyCache(50 * time.Millisecond)
	callCount := 0
	app := fiber.New()
	app.Use(Idempotency(cache))
	app.Post("/test", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(201).SendString("ok")
	})

	key := "expire-test"

	// First request
	req1 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req1.Header.Set("Idempotency-Key", key)
	resp1, _ := app.Test(req1, -1)
	resp1.Body.Close()

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Second request - should re-execute because entry expired
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req2.Header.Set("Idempotency-Key", key)
	resp2, _ := app.Test(req2, -1)
	resp2.Body.Close()

	if callCount != 2 {
		t.Errorf("after expiry: handler called %d times, want 2", callCount)
	}
}

func TestNewIdempotencyCache_Initializes(t *testing.T) {
	cache := NewIdempotencyCache(time.Minute)
	if cache == nil {
		t.Fatal("NewIdempotencyCache returned nil")
	}
	if cache.ttl != time.Minute {
		t.Errorf("TTL = %v, want %v", cache.ttl, time.Minute)
	}
	if cache.entries == nil {
		t.Error("entries map not initialized")
	}
}
