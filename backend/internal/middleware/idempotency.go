package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type idempotencyEntry struct {
	status    int
	body      []byte
	expiresAt time.Time
}

// IdempotencyCache stores responses keyed by Idempotency-Key header.
// Entries expire after the configured TTL. This is an in-memory cache
// suitable for single-instance deployments; for multi-instance, replace
// with Redis.
type IdempotencyCache struct {
	mu      sync.RWMutex
	entries map[string]*idempotencyEntry
	ttl     time.Duration
}

// NewIdempotencyCache creates a cache with the given TTL and starts a
// background cleanup goroutine.
func NewIdempotencyCache(ttl time.Duration) *IdempotencyCache {
	ic := &IdempotencyCache{
		entries: make(map[string]*idempotencyEntry),
		ttl:     ttl,
	}
	go ic.cleanup()
	return ic
}

func (ic *IdempotencyCache) cleanup() {
	ticker := time.NewTicker(ic.ttl)
	defer ticker.Stop()
	for range ticker.C {
		ic.mu.Lock()
		now := time.Now()
		for k, v := range ic.entries {
			if now.After(v.expiresAt) {
				delete(ic.entries, k)
			}
		}
		ic.mu.Unlock()
	}
}

// Idempotency returns middleware that caches responses for POST requests
// that include an Idempotency-Key header. If the same key is seen again
// within the TTL, the cached response is returned without re-executing
// the handler.
func Idempotency(cache *IdempotencyCache) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only apply to POST requests
		if c.Method() != "POST" {
			return c.Next()
		}

		key := c.Get("Idempotency-Key")
		if key == "" {
			return c.Next()
		}

		// Check cache for existing response
		cache.mu.RLock()
		entry, exists := cache.entries[key]
		cache.mu.RUnlock()

		if exists && time.Now().Before(entry.expiresAt) {
			c.Set("X-Idempotency-Replay", "true")
			return c.Status(entry.status).Send(entry.body)
		}

		// Execute the handler
		err := c.Next()
		if err != nil {
			return err
		}

		// Cache the response
		cache.mu.Lock()
		cache.entries[key] = &idempotencyEntry{
			status:    c.Response().StatusCode(),
			body:      append([]byte(nil), c.Response().Body()...),
			expiresAt: time.Now().Add(cache.ttl),
		}
		cache.mu.Unlock()

		return nil
	}
}
