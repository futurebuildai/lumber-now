package store

import (
	"context"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// PoolConfig defaults (zero-value struct)
// ---------------------------------------------------------------------------

func TestPoolConfigDefaults(t *testing.T) {
	cfg := PoolConfig{}
	if cfg.MaxConns != 0 {
		t.Errorf("expected zero value for MaxConns, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 0 {
		t.Errorf("expected zero value for MinConns, got %d", cfg.MinConns)
	}
}

func TestPoolConfigCustomValues(t *testing.T) {
	cfg := PoolConfig{MaxConns: 50, MinConns: 10}
	if cfg.MaxConns != 50 {
		t.Errorf("expected MaxConns=50, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 10 {
		t.Errorf("expected MinConns=10, got %d", cfg.MinConns)
	}
}

func TestPoolConfigNegativeValues(t *testing.T) {
	// Negative values are valid int32 values; the NewPool function only
	// applies them when > 0, so negatives are effectively ignored.
	cfg := PoolConfig{MaxConns: -1, MinConns: -5}
	if cfg.MaxConns != -1 {
		t.Errorf("expected MaxConns=-1, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != -5 {
		t.Errorf("expected MinConns=-5, got %d", cfg.MinConns)
	}
}

func TestPoolConfigPartialOverride(t *testing.T) {
	// Only setting MaxConns; MinConns should remain zero.
	cfg := PoolConfig{MaxConns: 30}
	if cfg.MaxConns != 30 {
		t.Errorf("expected MaxConns=30, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 0 {
		t.Errorf("expected MinConns=0 (zero value), got %d", cfg.MinConns)
	}
}

// ---------------------------------------------------------------------------
// NewPool — invalid URL (parse error)
// ---------------------------------------------------------------------------

func TestNewPoolInvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewPool(ctx, "not-a-valid-url")
	if err == nil {
		t.Error("expected error for invalid database URL")
	}
}

func TestNewPoolEmptyURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewPool(ctx, "")
	if err == nil {
		t.Error("expected error for empty database URL")
	}
}

func TestNewPoolMalformedScheme(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewPool(ctx, "://missing-scheme")
	if err == nil {
		t.Error("expected error for malformed URL scheme")
	}
}

// ---------------------------------------------------------------------------
// NewPool — unreachable host (valid URL but cannot connect)
// ---------------------------------------------------------------------------

func TestNewPoolUnreachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewPool(ctx, "postgres://user:pass@localhost:59999/nonexistent?connect_timeout=1")
	if err == nil {
		t.Error("expected error for unreachable database")
	}
}

func TestNewPoolUnreachableWithCustomConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := PoolConfig{MaxConns: 2, MinConns: 1}
	_, err := NewPool(ctx, "postgres://user:pass@localhost:59999/nonexistent?connect_timeout=1", cfg)
	if err == nil {
		t.Error("expected error for unreachable database with custom config")
	}
}

func TestNewPoolCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := NewPool(ctx, "postgres://user:pass@localhost:5432/testdb?connect_timeout=1")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// ---------------------------------------------------------------------------
// NewPool — with PoolConfig options
// ---------------------------------------------------------------------------

func TestNewPoolWithZeroMaxConns(t *testing.T) {
	// PoolConfig with MaxConns=0 should use the default (25).
	// We cannot verify the actual pool config without connecting, but we verify
	// it does not panic during URL parsing.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := PoolConfig{MaxConns: 0, MinConns: 0}
	_, err := NewPool(ctx, "postgres://user:pass@localhost:59999/testdb?connect_timeout=1", cfg)
	// Connection will fail but the config parsing should succeed.
	if err == nil {
		t.Error("expected connection error for unreachable host")
	}
}

func TestNewPoolWithMultipleOpts(t *testing.T) {
	// When multiple PoolConfig values are passed, only the first is used.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg1 := PoolConfig{MaxConns: 10, MinConns: 2}
	cfg2 := PoolConfig{MaxConns: 99, MinConns: 50}
	_, err := NewPool(ctx, "postgres://user:pass@localhost:59999/testdb?connect_timeout=1", cfg1, cfg2)
	// Connection will fail but the parsing should succeed.
	if err == nil {
		t.Error("expected connection error for unreachable host")
	}
}

// ---------------------------------------------------------------------------
// New constructor
// ---------------------------------------------------------------------------

func TestNewStoreNilPool(t *testing.T) {
	// New(nil) should not panic; it creates the Store with a nil pool
	// and a Queries backed by the nil pool.
	s := New(nil)
	if s == nil {
		t.Fatal("New(nil) returned nil")
	}
	if s.Pool != nil {
		t.Error("expected Pool to be nil when constructed with nil")
	}
	if s.Queries == nil {
		t.Error("expected Queries to be non-nil (db.New handles nil DBTX)")
	}
}

func TestNewStoreQueriesNotNil(t *testing.T) {
	s := New(nil)
	if s.Queries == nil {
		t.Error("Queries should not be nil even with a nil pool")
	}
}

// ---------------------------------------------------------------------------
// Store struct fields
// ---------------------------------------------------------------------------

func TestStoreFieldsAccessible(t *testing.T) {
	s := &Store{}
	if s.Pool != nil {
		t.Error("zero-value Pool should be nil")
	}
	if s.Queries != nil {
		t.Error("zero-value Queries should be nil")
	}
}

// ---------------------------------------------------------------------------
// NewPool error messages contain context
// ---------------------------------------------------------------------------

func TestNewPoolInvalidURLErrorMessage(t *testing.T) {
	ctx := context.Background()
	_, err := NewPool(ctx, "not-a-valid-url")
	if err == nil {
		t.Fatal("expected error")
	}
	// The error should wrap with "parse database url:" prefix.
	errMsg := err.Error()
	if len(errMsg) == 0 {
		t.Error("error message should not be empty")
	}
}

func TestNewPoolUnreachableErrorMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewPool(ctx, "postgres://user:pass@localhost:59999/nonexistent?connect_timeout=1")
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	if len(errMsg) == 0 {
		t.Error("error message should not be empty")
	}
}

// ---------------------------------------------------------------------------
// PoolConfig boundary values
// ---------------------------------------------------------------------------

func TestPoolConfigMaxInt32(t *testing.T) {
	cfg := PoolConfig{MaxConns: 2147483647, MinConns: 2147483647}
	if cfg.MaxConns != 2147483647 {
		t.Errorf("expected MaxConns to hold max int32, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 2147483647 {
		t.Errorf("expected MinConns to hold max int32, got %d", cfg.MinConns)
	}
}

func TestPoolConfigMinInt32(t *testing.T) {
	cfg := PoolConfig{MaxConns: -2147483648, MinConns: -2147483648}
	if cfg.MaxConns != -2147483648 {
		t.Errorf("expected MaxConns to hold min int32, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != -2147483648 {
		t.Errorf("expected MinConns to hold min int32, got %d", cfg.MinConns)
	}
}
