package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestMetrics_CountsRequests(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/ok", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}

	if m.requestsTotal.Load() != 5 {
		t.Errorf("expected 5 total requests, got %d", m.requestsTotal.Load())
	}
	if m.errorsTotal.Load() != 0 {
		t.Errorf("expected 0 errors, got %d", m.errorsTotal.Load())
	}
}

func TestMetrics_Counts5xxErrors(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/fail", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("error")
	})

	req := httptest.NewRequest("GET", "/fail", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if m.errorsTotal.Load() != 1 {
		t.Errorf("expected 1 error, got %d", m.errorsTotal.Load())
	}
	if m.requestsTotal.Load() != 1 {
		t.Errorf("expected 1 total request, got %d", m.requestsTotal.Load())
	}
}

func TestMetrics_DoesNotCount4xxAsError(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/notfound", func(c *fiber.Ctx) error {
		return c.Status(404).SendString("not found")
	})

	req := httptest.NewRequest("GET", "/notfound", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if m.errorsTotal.Load() != 0 {
		t.Errorf("expected 0 errors for 404, got %d", m.errorsTotal.Load())
	}
}

func TestMetrics_InFlightReturnToZero(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if m.inFlightGauge.Load() != 0 {
		t.Errorf("expected 0 in-flight after request, got %d", m.inFlightGauge.Load())
	}
}

func TestMetrics_TracksDuration(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if m.durationCount.Load() != 1 {
		t.Errorf("expected 1 duration count, got %d", m.durationCount.Load())
	}
	if m.durationSumMs.Load() < 0 {
		t.Errorf("expected non-negative duration sum, got %d", m.durationSumMs.Load())
	}
}

func TestMetrics_TracksStatusCodes(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/200", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/404", func(c *fiber.Ctx) error {
		return c.Status(404).SendString("not found")
	})
	app.Get("/500", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("error")
	})

	for _, path := range []string{"/200", "/200", "/404", "/500"} {
		req := httptest.NewRequest("GET", path, nil)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}

	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	if m.statusCounts[200] != 2 {
		t.Errorf("expected 2 for status 200, got %d", m.statusCounts[200])
	}
	if m.statusCounts[404] != 1 {
		t.Errorf("expected 1 for status 404, got %d", m.statusCounts[404])
	}
	if m.statusCounts[500] != 1 {
		t.Errorf("expected 1 for status 500, got %d", m.statusCounts[500])
	}
}

func TestMetrics_TracksMethodCounts(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/method-count-test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Send 3 GET requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/method-count-test", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}

	// Verify method counts through the endpoint output
	app.Get("/method-metrics", m.Endpoint())
	req := httptest.NewRequest("GET", "/method-metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	// Should contain GET method counter
	if !strings.Contains(body, `http_requests_by_method{method="GET"}`) {
		t.Errorf("expected method metrics in output, got:\n%s", body)
	}

	// Verify total request count is correct (3 + 1 for metrics endpoint = 4)
	if m.requestsTotal.Load() != 4 {
		t.Errorf("expected 4 total requests, got %d", m.requestsTotal.Load())
	}
}

func TestMetrics_HistogramBuckets(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/fast", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/fast", nil)
	resp, _ := app.Test(req, -1)
	resp.Body.Close()

	// Fast requests should fall into the first bucket (<=10ms)
	m.histogramMu.Lock()
	total := int64(0)
	for _, count := range m.histogramBuckets {
		total += count
	}
	m.histogramMu.Unlock()

	if total != 1 {
		t.Errorf("expected 1 entry in histogram, got %d", total)
	}
}

func TestMetrics_NormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/v1/health", "/v1/health"},
		{"/v1/requests/:id", "/v1/requests/:id"},
	}

	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.expected {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMetrics_Endpoint(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/metrics", m.Endpoint())
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Make a request to generate metrics
	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// Check metrics endpoint
	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %s", ct)
	}

	cc := resp.Header.Get("Cache-Control")
	if !strings.Contains(cc, "no-cache") {
		t.Errorf("expected Cache-Control no-cache, got %s", cc)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(bodyBytes)

	expectedMetrics := []string{
		"http_requests_total",
		"http_errors_total",
		"http_requests_in_flight",
		"http_request_duration_ms_avg",
		"http_request_duration_ms_bucket",
		"http_requests_by_status",
		"http_requests_by_method",
		"http_requests_by_path",
		"process_uptime_seconds",
	}
	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("expected metrics body to contain %q", metric)
		}
	}
}

func TestMetrics_EndpointShowsHistogramBuckets(t *testing.T) {
	m := NewMetrics()
	app := fiber.New()
	app.Use(m.Handler())
	app.Get("/metrics", m.Endpoint())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Generate some metrics
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req, -1)
	resp.Body.Close()

	// Fetch metrics
	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req, -1)
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	// Should contain bucket boundaries
	bucketLabels := []string{
		`le="10"`,
		`le="50"`,
		`le="100"`,
		`le="250"`,
		`le="500"`,
		`le="1000"`,
		`le="5000"`,
		`le="+Inf"`,
	}
	for _, label := range bucketLabels {
		if !strings.Contains(body, label) {
			t.Errorf("expected histogram bucket %s in output", label)
		}
	}
}

func TestMetrics_PoolStatsIncludedWhenSet(t *testing.T) {
	m := NewMetrics()
	m.SetPoolStatsFunc(func() PoolStats {
		return PoolStats{
			TotalConns:    10,
			IdleConns:     5,
			AcquiredConns: 3,
			MaxConns:      25,
		}
	})

	app := fiber.New()
	app.Get("/metrics", m.Endpoint())

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	poolMetrics := []string{
		"db_pool_total_conns 10",
		"db_pool_idle_conns 5",
		"db_pool_acquired_conns 3",
		"db_pool_max_conns 25",
	}
	for _, metric := range poolMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("expected pool metric %q in output", metric)
		}
	}
}

func TestMetrics_PoolStatsOmittedWhenNotSet(t *testing.T) {
	m := NewMetrics()
	// Do NOT set pool stats func

	app := fiber.New()
	app.Get("/metrics", m.Endpoint())

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	if strings.Contains(body, "db_pool_total_conns") {
		t.Error("did not expect pool metrics when no pool stats func is set")
	}
}
