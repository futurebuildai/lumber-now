package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/builderwire/lumber-now/backend/internal/store"
)

// CircuitStateReporter returns the current circuit breaker state string.
type CircuitStateReporter interface {
	BreakerState() string
}

type HealthHandler struct {
	store     *store.Store
	startTime time.Time
	circuits  map[string]CircuitStateReporter
}

func NewHealthHandler(s *store.Store) *HealthHandler {
	return &HealthHandler{store: s, startTime: time.Now(), circuits: make(map[string]CircuitStateReporter)}
}

// RegisterCircuit registers a named circuit breaker for health reporting.
func (h *HealthHandler) RegisterCircuit(name string, c CircuitStateReporter) {
	h.circuits[name] = c
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	dbStatus := "ok"
	if err := h.store.Pool.Ping(c.Context()); err != nil {
		dbStatus = "unhealthy"
	}

	status := "ok"
	httpStatus := fiber.StatusOK
	if dbStatus != "ok" {
		status = "degraded"
		httpStatus = fiber.StatusServiceUnavailable
	}

	poolStats := h.store.Pool.Stat()

	result := fiber.Map{
		"status":   status,
		"service":  "lumber-now-api",
		"version":  Version,
		"database": dbStatus,
		"uptime_seconds": int(time.Since(h.startTime).Seconds()),
		"pool": fiber.Map{
			"total_conns":    poolStats.TotalConns(),
			"idle_conns":     poolStats.IdleConns(),
			"acquired_conns": poolStats.AcquiredConns(),
		},
	}

	if len(h.circuits) > 0 {
		circuitStates := fiber.Map{}
		for name, reporter := range h.circuits {
			state := reporter.BreakerState()
			circuitStates[name] = state
			if state == "open" {
				status = "degraded"
				if httpStatus == fiber.StatusOK {
					httpStatus = fiber.StatusOK // still 200, but status=degraded
				}
			}
		}
		result["circuits"] = circuitStates
		result["status"] = status
	}

	return c.Status(httpStatus).JSON(result)
}

// Readiness returns 200 only when the service is ready to accept traffic.
// Unlike Check (which is informational), this returns a hard 503 if DB is down.
func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	if err := h.store.Pool.Ping(c.Context()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not_ready",
			"reason": "database unavailable",
		})
	}
	return c.JSON(fiber.Map{"status": "ready"})
}

// Version is set at build time via ldflags.
var Version = "dev"
