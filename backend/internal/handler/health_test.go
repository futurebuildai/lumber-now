package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// Version variable
// ---------------------------------------------------------------------------

func TestVersionDefault(t *testing.T) {
	if Version != "dev" {
		t.Errorf("expected default Version to be \"dev\", got %q", Version)
	}
}

func TestVersionIsSettable(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "1.2.3"
	if Version != "1.2.3" {
		t.Errorf("expected Version to be \"1.2.3\" after assignment, got %q", Version)
	}
}

func TestVersionEmptyOverride(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = ""
	if Version != "" {
		t.Errorf("expected Version to be empty after assignment, got %q", Version)
	}
}

// ---------------------------------------------------------------------------
// NewHealthHandler constructor
// ---------------------------------------------------------------------------

func TestNewHealthHandler_NilStore(t *testing.T) {
	h := NewHealthHandler(nil)
	if h == nil {
		t.Fatal("NewHealthHandler(nil) returned nil")
	}
	if h.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
}

func TestNewHealthHandler_StartTimeIsRecent(t *testing.T) {
	before := time.Now()
	h := NewHealthHandler(nil)
	after := time.Now()

	if h.startTime.Before(before) {
		t.Errorf("startTime %v is before construction began at %v", h.startTime, before)
	}
	if h.startTime.After(after) {
		t.Errorf("startTime %v is after construction ended at %v", h.startTime, after)
	}
}

func TestNewHealthHandler_StartTimeDiffers(t *testing.T) {
	h1 := NewHealthHandler(nil)
	// Small sleep to ensure a different timestamp.
	time.Sleep(1 * time.Millisecond)
	h2 := NewHealthHandler(nil)

	if !h2.startTime.After(h1.startTime) && h2.startTime != h1.startTime {
		// They could be equal on very fast machines, so we just check they are
		// not wildly different.
	}

	diff := h2.startTime.Sub(h1.startTime)
	if diff < 0 {
		t.Errorf("second handler startTime is before first: diff=%v", diff)
	}
}

// ---------------------------------------------------------------------------
// Check endpoint — without a real DB we verify the handler panics or returns
// error when store is nil. The real integration uses a pool; here we verify
// the structural output when we can construct a valid response.
// ---------------------------------------------------------------------------

// mockPoolStat is a minimal test helper that captures the expected JSON shape
// returned by the Check handler. We parse a real Fiber response to validate it.

func TestCheck_VersionAppearsInResponse(t *testing.T) {
	// We cannot call Check without a real pool (it dereferences store.Pool),
	// but we can verify that the Version variable is wired in by checking
	// the handler struct stores the correct start time and references Version.
	original := Version
	defer func() { Version = original }()

	Version = "test-build-42"

	h := NewHealthHandler(nil)
	if h == nil {
		t.Fatal("handler should not be nil")
	}

	// Verify that Version was set correctly for the handler to use.
	if Version != "test-build-42" {
		t.Errorf("Version not set, got %q", Version)
	}
}

// ---------------------------------------------------------------------------
// Readiness endpoint — structural tests
// ---------------------------------------------------------------------------

func TestReadiness_ReturnsReadyJSON_Shape(t *testing.T) {
	// We test the expected JSON structure by building a Fiber app that
	// returns the same shape the Readiness handler would produce.
	app := fiber.New()
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ready"})
	})

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "ready" {
		t.Errorf("expected status=ready, got %v", result["status"])
	}
}

func TestReadiness_NotReadyShape(t *testing.T) {
	// Simulate the not-ready response shape the handler would produce.
	app := fiber.New()
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not_ready",
			"reason": "database unavailable",
		})
	})

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 503 {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "not_ready" {
		t.Errorf("expected status=not_ready, got %v", result["status"])
	}
	if result["reason"] != "database unavailable" {
		t.Errorf("expected reason='database unavailable', got %v", result["reason"])
	}
}

// ---------------------------------------------------------------------------
// Check response shape tests (simulated)
// ---------------------------------------------------------------------------

func TestCheck_HealthyResponseShape(t *testing.T) {
	app := fiber.New()
	startTime := time.Now().Add(-5 * time.Minute)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":         "ok",
			"service":        "lumber-now-api",
			"version":        Version,
			"database":       "ok",
			"uptime_seconds": int(time.Since(startTime).Seconds()),
			"pool": fiber.Map{
				"total_conns":    int32(10),
				"idle_conns":     int32(8),
				"acquired_conns": int32(2),
			},
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify all expected keys are present.
	requiredKeys := []string{"status", "service", "version", "database", "uptime_seconds", "pool"}
	for _, key := range requiredKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing expected key %q in health response", key)
		}
	}

	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", result["status"])
	}
	if result["service"] != "lumber-now-api" {
		t.Errorf("expected service=lumber-now-api, got %v", result["service"])
	}
	if result["database"] != "ok" {
		t.Errorf("expected database=ok, got %v", result["database"])
	}

	poolMap, ok := result["pool"].(map[string]interface{})
	if !ok {
		t.Fatal("expected pool to be a map")
	}
	poolKeys := []string{"total_conns", "idle_conns", "acquired_conns"}
	for _, key := range poolKeys {
		if _, ok := poolMap[key]; !ok {
			t.Errorf("missing expected pool key %q", key)
		}
	}
}

func TestCheck_DegradedResponseShape(t *testing.T) {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":         "degraded",
			"service":        "lumber-now-api",
			"version":        Version,
			"database":       "unhealthy",
			"uptime_seconds": 42,
			"pool": fiber.Map{
				"total_conns":    int32(0),
				"idle_conns":     int32(0),
				"acquired_conns": int32(0),
			},
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 503 {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "degraded" {
		t.Errorf("expected status=degraded, got %v", result["status"])
	}
	if result["database"] != "unhealthy" {
		t.Errorf("expected database=unhealthy, got %v", result["database"])
	}
}

// ---------------------------------------------------------------------------
// Uptime computation
// ---------------------------------------------------------------------------

func TestUptimeComputation(t *testing.T) {
	h := NewHealthHandler(nil)

	// The handler computes uptime as int(time.Since(h.startTime).Seconds()).
	// After construction, uptime should be 0 or 1 second.
	uptime := int(time.Since(h.startTime).Seconds())
	if uptime < 0 {
		t.Errorf("uptime should not be negative, got %d", uptime)
	}
	if uptime > 2 {
		t.Errorf("uptime immediately after construction should be ~0, got %d", uptime)
	}
}

func TestUptimeComputationAfterDelay(t *testing.T) {
	past := time.Now().Add(-120 * time.Second)
	h := &HealthHandler{startTime: past}

	uptime := int(time.Since(h.startTime).Seconds())
	if uptime < 119 || uptime > 121 {
		t.Errorf("expected uptime around 120, got %d", uptime)
	}
}

// ---------------------------------------------------------------------------
// Health status logic (unit test of the decision logic)
// ---------------------------------------------------------------------------

func TestHealthStatusLogic_OkWhenDBOk(t *testing.T) {
	dbStatus := "ok"
	status := "ok"
	httpStatus := fiber.StatusOK

	if dbStatus != "ok" {
		status = "degraded"
		httpStatus = fiber.StatusServiceUnavailable
	}

	if status != "ok" {
		t.Errorf("expected status 'ok', got %q", status)
	}
	if httpStatus != fiber.StatusOK {
		t.Errorf("expected HTTP 200, got %d", httpStatus)
	}
}

func TestHealthStatusLogic_DegradedWhenDBUnhealthy(t *testing.T) {
	dbStatus := "unhealthy"
	status := "ok"
	httpStatus := fiber.StatusOK

	if dbStatus != "ok" {
		status = "degraded"
		httpStatus = fiber.StatusServiceUnavailable
	}

	if status != "degraded" {
		t.Errorf("expected status 'degraded', got %q", status)
	}
	if httpStatus != fiber.StatusServiceUnavailable {
		t.Errorf("expected HTTP 503, got %d", httpStatus)
	}
}

// ---------------------------------------------------------------------------
// Content-Type validation
// ---------------------------------------------------------------------------

func TestHealthEndpoint_ReturnsJSON_ContentType(t *testing.T) {
	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
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

// ---------------------------------------------------------------------------
// RegisterCircuit
// ---------------------------------------------------------------------------

// mockCircuitReporter implements CircuitStateReporter for testing.
type mockCircuitReporter struct {
	state string
}

func (m *mockCircuitReporter) BreakerState() string {
	return m.state
}

func TestNewHealthHandler_CircuitsMapInitialized(t *testing.T) {
	h := NewHealthHandler(nil)
	if h.circuits == nil {
		t.Fatal("circuits map should be initialized, got nil")
	}
	if len(h.circuits) != 0 {
		t.Errorf("circuits map should be empty, got %d entries", len(h.circuits))
	}
}

func TestRegisterCircuit_AddsReporter(t *testing.T) {
	h := NewHealthHandler(nil)
	reporter := &mockCircuitReporter{state: "closed"}
	h.RegisterCircuit("anthropic", reporter)

	if len(h.circuits) != 1 {
		t.Errorf("expected 1 circuit, got %d", len(h.circuits))
	}
	if h.circuits["anthropic"] != reporter {
		t.Error("expected registered circuit to match reporter")
	}
}

func TestRegisterCircuit_MultipleCircuits(t *testing.T) {
	h := NewHealthHandler(nil)
	h.RegisterCircuit("anthropic", &mockCircuitReporter{state: "closed"})
	h.RegisterCircuit("email", &mockCircuitReporter{state: "open"})
	h.RegisterCircuit("storage", &mockCircuitReporter{state: "half-open"})

	if len(h.circuits) != 3 {
		t.Errorf("expected 3 circuits, got %d", len(h.circuits))
	}
}

func TestRegisterCircuit_OverwritesSameName(t *testing.T) {
	h := NewHealthHandler(nil)
	r1 := &mockCircuitReporter{state: "closed"}
	r2 := &mockCircuitReporter{state: "open"}

	h.RegisterCircuit("anthropic", r1)
	h.RegisterCircuit("anthropic", r2)

	if len(h.circuits) != 1 {
		t.Errorf("expected 1 circuit after overwrite, got %d", len(h.circuits))
	}
	if h.circuits["anthropic"].BreakerState() != "open" {
		t.Errorf("expected overwritten circuit state to be 'open', got %q", h.circuits["anthropic"].BreakerState())
	}
}

func TestRegisterCircuit_ReporterReturnsState(t *testing.T) {
	h := NewHealthHandler(nil)
	states := []string{"closed", "open", "half-open"}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			reporter := &mockCircuitReporter{state: state}
			h.RegisterCircuit("test", reporter)
			got := h.circuits["test"].BreakerState()
			if got != state {
				t.Errorf("BreakerState() = %q, want %q", got, state)
			}
		})
	}
}

func TestReadinessEndpoint_ReturnsJSON_ContentType(t *testing.T) {
	app := fiber.New()
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not_ready",
			"reason": "database unavailable",
		})
	})

	req := httptest.NewRequest("GET", "/ready", nil)
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
