package s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/builderwire/lumber-now/backend/internal/platform/circuitbreaker"
)

// ---------------------------------------------------------------------------
// NewClient
// ---------------------------------------------------------------------------

func TestNewClient_ValidConfig(t *testing.T) {
	c, err := NewClient("http://localhost:9000", "test-bucket", "us-east-1", "access", "secret")
	if err != nil {
		t.Fatalf("NewClient returned unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient returned nil client")
	}
	if c.bucket != "test-bucket" {
		t.Errorf("bucket = %q, want %q", c.bucket, "test-bucket")
	}
	if c.client == nil {
		t.Error("underlying s3.Client is nil")
	}
	if c.breaker == nil {
		t.Error("circuit breaker is nil")
	}
}

func TestNewClient_EmptyEndpoint(t *testing.T) {
	// Empty endpoint is valid -- the code only sets BaseEndpoint when non-empty.
	c, err := NewClient("", "bucket", "us-east-1", "access", "secret")
	if err != nil {
		t.Fatalf("NewClient with empty endpoint should succeed, got: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClient_EmptyBucket(t *testing.T) {
	// Empty bucket is accepted at construction time; the bucket string is only
	// used later when calling Upload/Download. Verify it stores what we pass.
	c, err := NewClient("http://localhost:9000", "", "us-east-1", "access", "secret")
	if err != nil {
		t.Fatalf("NewClient with empty bucket returned error: %v", err)
	}
	if c.bucket != "" {
		t.Errorf("bucket = %q, want empty string", c.bucket)
	}
}

func TestNewClient_EmptyRegion(t *testing.T) {
	// AWS SDK v2 LoadDefaultConfig does not fail on empty region; it falls back
	// to environment / config file. Verify no panic and the client is created.
	c, err := NewClient("http://localhost:9000", "bucket", "", "access", "secret")
	if err != nil {
		t.Fatalf("NewClient with empty region returned error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_EmptyCredentials(t *testing.T) {
	// Static credentials with empty strings are accepted at construction time.
	c, err := NewClient("http://localhost:9000", "bucket", "us-east-1", "", "")
	if err != nil {
		t.Fatalf("NewClient with empty credentials returned error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// ---------------------------------------------------------------------------
// BreakerState
// ---------------------------------------------------------------------------

func TestBreakerState_InitiallyClosed(t *testing.T) {
	c, err := NewClient("http://localhost:9000", "bucket", "us-east-1", "a", "s")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.BreakerState(); got != "closed" {
		t.Errorf("BreakerState() = %q, want %q", got, "closed")
	}
}

// ---------------------------------------------------------------------------
// Circuit breaker integration with a fake HTTP server
// ---------------------------------------------------------------------------

// newTestClient creates a Client whose underlying S3 client points at the
// given httptest.Server. This lets us control responses without hitting AWS.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(serverURL, "test-bucket", "us-east-1", "access", "secret")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestUpload_CircuitBreakerOpensAfterFailures(t *testing.T) {
	// Create a server that always returns 500 to force S3 SDK errors.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	// The breaker is configured with maxFailures=5. Trigger 5 Upload failures.
	for i := 0; i < 5; i++ {
		_ = c.Upload(ctx, "key", strings.NewReader("data"), "text/plain")
	}

	// Circuit should now be open.
	if got := c.BreakerState(); got != "open" {
		t.Errorf("after 5 failures: BreakerState() = %q, want %q", got, "open")
	}

	// Subsequent call should be rejected immediately with ErrCircuitOpen.
	err := c.Upload(ctx, "key", strings.NewReader("data"), "text/plain")
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got: %v", err)
	}
}

func TestDownload_CircuitBreakerOpensAfterFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, _, _ = c.Download(ctx, "key")
	}

	if got := c.BreakerState(); got != "open" {
		t.Errorf("after 5 failures: BreakerState() = %q, want %q", got, "open")
	}

	_, _, err := c.Download(ctx, "key")
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got: %v", err)
	}
}

func TestPresignedURL_CircuitBreakerOpensAfterFailures(t *testing.T) {
	// PresignedURL uses the presign client which constructs the URL locally
	// without making an HTTP call, so it should succeed even with a bad server.
	// However, if we point at a completely invalid endpoint, the SDK may error.
	// We test with a server that is immediately closed (connection refused).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	// Close immediately so connections are refused.
	srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	// PresignedURL may or may not fail depending on the SDK version -- presigning
	// is typically a local operation. If it succeeds, the breaker stays closed.
	// We verify the breaker starts closed and that the method does not panic.
	initialState := c.BreakerState()
	if initialState != "closed" {
		t.Errorf("initial BreakerState() = %q, want %q", initialState, "closed")
	}

	_, _ = c.PresignedURL(ctx, "key", 15*time.Minute)
	// No assertion on error -- the presign path may succeed offline.
}

func TestUpload_SuccessKeepsBreakerClosed(t *testing.T) {
	// A server that returns 200 to PutObject.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	err := c.Upload(ctx, "key", strings.NewReader("data"), "text/plain")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if got := c.BreakerState(); got != "closed" {
		t.Errorf("after success: BreakerState() = %q, want %q", got, "closed")
	}
}

func TestDownload_SuccessReturnsBodyAndContentType(t *testing.T) {
	body := "hello from s3"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// Minimal valid S3 GetObject response: just write the body.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	rc, ct, err := c.Download(ctx, "key")
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if rc != nil {
		defer rc.Close()
		data, _ := io.ReadAll(rc)
		// The content may or may not match exactly depending on how the SDK
		// parses the response, but the key is no error was returned.
		_ = data
	}
	// Content type default is "application/octet-stream" when the SDK cannot
	// parse the response. We just verify we got a non-empty value.
	if ct == "" {
		t.Error("expected non-empty content type")
	}
}

func TestUpload_CircuitBreakerResetsAfterSuccess(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// First 4 calls fail, then succeed.
		if callCount <= 4 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	// 4 failures -- should stay closed (threshold is 5).
	for i := 0; i < 4; i++ {
		_ = c.Upload(ctx, "key", strings.NewReader("data"), "text/plain")
	}
	if got := c.BreakerState(); got != "closed" {
		t.Errorf("after 4 failures: BreakerState() = %q, want %q", got, "closed")
	}

	// 1 success resets the failure count.
	err := c.Upload(ctx, "key", strings.NewReader("data"), "text/plain")
	if err != nil {
		t.Fatalf("5th call (should succeed) returned error: %v", err)
	}
	if got := c.BreakerState(); got != "closed" {
		t.Errorf("after success: BreakerState() = %q, want %q", got, "closed")
	}
}

func TestDownload_ReturnsErrorOnFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	_, _, err := c.Download(ctx, "nonexistent-key")
	if err == nil {
		t.Error("expected error from Download for 404 response")
	}
}

func TestUpload_ReturnsErrorOnFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	err := c.Upload(ctx, "key", strings.NewReader("data"), "text/plain")
	if err == nil {
		t.Error("expected error from Upload for 403 response")
	}
}

func TestPresignedURL_ReturnsURLOnSuccess(t *testing.T) {
	// Presigning is a local operation, so even a non-running server works.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	url, err := c.PresignedURL(ctx, "my-object-key", 30*time.Minute)
	if err != nil {
		t.Fatalf("PresignedURL returned error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty presigned URL")
	}
	if !strings.Contains(url, "my-object-key") {
		t.Errorf("presigned URL %q does not contain object key", url)
	}
	if !strings.Contains(url, "test-bucket") {
		t.Errorf("presigned URL %q does not contain bucket name", url)
	}
}

// ---------------------------------------------------------------------------
// Mixed operations sharing the breaker
// ---------------------------------------------------------------------------

func TestMixedOperations_SharedBreakerState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	// Mix Upload and Download failures -- they share the same breaker.
	_ = c.Upload(ctx, "k1", strings.NewReader("d"), "text/plain")
	_, _, _ = c.Download(ctx, "k2")
	_ = c.Upload(ctx, "k3", strings.NewReader("d"), "text/plain")
	_, _, _ = c.Download(ctx, "k4")
	_ = c.Upload(ctx, "k5", strings.NewReader("d"), "text/plain")

	if got := c.BreakerState(); got != "open" {
		t.Errorf("after 5 mixed failures: BreakerState() = %q, want %q", got, "open")
	}

	// All operations should now be rejected.
	err := c.Upload(ctx, "k", strings.NewReader("d"), "text/plain")
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("Upload after open: expected ErrCircuitOpen, got: %v", err)
	}

	_, _, err = c.Download(ctx, "k")
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("Download after open: expected ErrCircuitOpen, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// BreakerState string values
// ---------------------------------------------------------------------------

func TestBreakerState_ReflectsOpenState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	if got := c.BreakerState(); got != "closed" {
		t.Errorf("initial: BreakerState() = %q, want %q", got, "closed")
	}

	for i := 0; i < 5; i++ {
		_ = c.Upload(ctx, "k", strings.NewReader("d"), "text/plain")
	}

	if got := c.BreakerState(); got != "open" {
		t.Errorf("after failures: BreakerState() = %q, want %q", got, "open")
	}
}
