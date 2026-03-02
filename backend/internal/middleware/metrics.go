package middleware

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PoolStats holds database connection pool statistics.
type PoolStats struct {
	TotalConns    int32
	IdleConns     int32
	AcquiredConns int32
	MaxConns      int32
}

// PoolStatsFunc returns current pool statistics.
type PoolStatsFunc func() PoolStats

// Metrics provides lightweight request metrics without external dependencies.
// Counters are exposed via the /metrics endpoint in Prometheus text format.
type Metrics struct {
	requestsTotal   atomic.Int64
	errorsTotal     atomic.Int64
	inFlightGauge   atomic.Int64
	durationSumMs   atomic.Int64
	durationCount   atomic.Int64
	startTime       time.Time

	// Per-status-code counters
	statusMu     sync.Mutex
	statusCounts map[int]int64

	// Per-method counters
	methodMu     sync.Mutex
	methodCounts map[string]int64

	// Per-path counters for top endpoints
	pathMu     sync.Mutex
	pathCounts map[string]int64

	// Duration histogram buckets (ms): 10, 50, 100, 250, 500, 1000, 5000
	histogramMu      sync.Mutex
	histogramBuckets [7]int64

	// Per-input-type counters (text, image, pdf, voice)
	inputTypeMu     sync.Mutex
	requestsByInputType map[string]*atomic.Int64

	// Per-request-status counters (pending, processing, parsed, confirmed, sent, failed)
	requestStatusMu     sync.Mutex
	requestsByStatus    map[string]*atomic.Int64

	// Optional DB pool stats provider
	poolStats PoolStatsFunc
}

var histogramBounds = [7]int64{10, 50, 100, 250, 500, 1000, 5000}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	inputTypes := []string{"text", "image", "pdf", "voice"}
	inputTypeMap := make(map[string]*atomic.Int64, len(inputTypes))
	for _, t := range inputTypes {
		inputTypeMap[t] = &atomic.Int64{}
	}

	statuses := []string{"pending", "processing", "parsed", "confirmed", "sent", "failed"}
	statusMap := make(map[string]*atomic.Int64, len(statuses))
	for _, s := range statuses {
		statusMap[s] = &atomic.Int64{}
	}

	return &Metrics{
		startTime:           time.Now(),
		statusCounts:        make(map[int]int64),
		methodCounts:        make(map[string]int64),
		pathCounts:          make(map[string]int64),
		requestsByInputType: inputTypeMap,
		requestsByStatus:    statusMap,
	}
}

// RecordInputType increments the counter for a given input type (text, image, pdf, voice).
func (m *Metrics) RecordInputType(inputType string) {
	m.inputTypeMu.Lock()
	defer m.inputTypeMu.Unlock()
	if counter, ok := m.requestsByInputType[inputType]; ok {
		counter.Add(1)
	} else {
		c := &atomic.Int64{}
		c.Add(1)
		m.requestsByInputType[inputType] = c
	}
}

// RecordRequestStatus increments the counter for a given request status (pending, processing, parsed, confirmed, sent, failed).
func (m *Metrics) RecordRequestStatus(status string) {
	m.requestStatusMu.Lock()
	defer m.requestStatusMu.Unlock()
	if counter, ok := m.requestsByStatus[status]; ok {
		counter.Add(1)
	} else {
		c := &atomic.Int64{}
		c.Add(1)
		m.requestsByStatus[status] = c
	}
}

// SetPoolStatsFunc sets a function to retrieve DB pool statistics.
func (m *Metrics) SetPoolStatsFunc(fn PoolStatsFunc) {
	m.poolStats = fn
}

// Handler returns a Fiber middleware that tracks request metrics.
func (m *Metrics) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		m.inFlightGauge.Add(1)
		start := time.Now()

		err := c.Next()

		durationMs := time.Since(start).Milliseconds()
		m.inFlightGauge.Add(-1)
		m.requestsTotal.Add(1)
		m.durationSumMs.Add(durationMs)
		m.durationCount.Add(1)

		status := c.Response().StatusCode()
		if status >= 500 {
			m.errorsTotal.Add(1)
		}

		// Track per-status and per-method
		m.statusMu.Lock()
		m.statusCounts[status]++
		m.statusMu.Unlock()

		m.methodMu.Lock()
		m.methodCounts[c.Method()]++
		m.methodMu.Unlock()

		// Track per-path (normalize parameterized paths)
		path := normalizePath(c.Route().Path)
		m.pathMu.Lock()
		m.pathCounts[path]++
		m.pathMu.Unlock()

		// Track duration histogram
		m.histogramMu.Lock()
		for i, bound := range histogramBounds {
			if durationMs <= bound {
				m.histogramBuckets[i]++
				break
			}
			if i == len(histogramBounds)-1 {
				m.histogramBuckets[i]++
			}
		}
		m.histogramMu.Unlock()

		return err
	}
}

// normalizePath reduces parameterized Fiber route paths into static patterns.
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

// Endpoint returns a Fiber handler that exposes metrics in Prometheus text format.
func (m *Metrics) Endpoint() fiber.Handler {
	return func(c *fiber.Ctx) error {
		total := m.requestsTotal.Load()
		errors := m.errorsTotal.Load()
		inFlight := m.inFlightGauge.Load()
		durSum := m.durationSumMs.Load()
		durCount := m.durationCount.Load()
		uptimeSeconds := time.Since(m.startTime).Seconds()

		avgDurationMs := float64(0)
		if durCount > 0 {
			avgDurationMs = float64(durSum) / float64(durCount)
		}

		var b strings.Builder
		b.WriteString("# HELP http_requests_total Total number of HTTP requests.\n")
		b.WriteString("# TYPE http_requests_total counter\n")
		b.WriteString("http_requests_total " + strconv.FormatInt(total, 10) + "\n")
		b.WriteString("# HELP http_errors_total Total number of HTTP 5xx errors.\n")
		b.WriteString("# TYPE http_errors_total counter\n")
		b.WriteString("http_errors_total " + strconv.FormatInt(errors, 10) + "\n")
		b.WriteString("# HELP http_requests_in_flight Current number of in-flight requests.\n")
		b.WriteString("# TYPE http_requests_in_flight gauge\n")
		b.WriteString("http_requests_in_flight " + strconv.FormatInt(inFlight, 10) + "\n")
		b.WriteString("# HELP http_request_duration_ms_avg Average request duration in milliseconds.\n")
		b.WriteString("# TYPE http_request_duration_ms_avg gauge\n")
		b.WriteString("http_request_duration_ms_avg " + fmt.Sprintf("%.2f", avgDurationMs) + "\n")
		b.WriteString("# HELP http_request_duration_ms_sum Total request duration in milliseconds.\n")
		b.WriteString("# TYPE http_request_duration_ms_sum counter\n")
		b.WriteString("http_request_duration_ms_sum " + strconv.FormatInt(durSum, 10) + "\n")
		b.WriteString("# HELP http_request_duration_count Total number of timed requests.\n")
		b.WriteString("# TYPE http_request_duration_count counter\n")
		b.WriteString("http_request_duration_count " + strconv.FormatInt(durCount, 10) + "\n")
		b.WriteString("# HELP process_uptime_seconds Seconds since process start.\n")
		b.WriteString("# TYPE process_uptime_seconds gauge\n")
		b.WriteString("process_uptime_seconds " + fmt.Sprintf("%.0f", uptimeSeconds) + "\n")

		// Duration histogram
		b.WriteString("# HELP http_request_duration_ms_bucket Request duration histogram.\n")
		b.WriteString("# TYPE http_request_duration_ms_bucket counter\n")
		m.histogramMu.Lock()
		cumulative := int64(0)
		for i, bound := range histogramBounds {
			cumulative += m.histogramBuckets[i]
			b.WriteString(fmt.Sprintf("http_request_duration_ms_bucket{le=\"%d\"} %d\n", bound, cumulative))
		}
		b.WriteString(fmt.Sprintf("http_request_duration_ms_bucket{le=\"+Inf\"} %d\n", durCount))
		m.histogramMu.Unlock()

		// Per-status-code breakdown
		b.WriteString("# HELP http_requests_by_status Total requests by HTTP status code.\n")
		b.WriteString("# TYPE http_requests_by_status counter\n")
		m.statusMu.Lock()
		for code, count := range m.statusCounts {
			b.WriteString(fmt.Sprintf("http_requests_by_status{code=\"%d\"} %d\n", code, count))
		}
		m.statusMu.Unlock()

		// Per-method breakdown
		b.WriteString("# HELP http_requests_by_method Total requests by HTTP method.\n")
		b.WriteString("# TYPE http_requests_by_method counter\n")
		m.methodMu.Lock()
		for method, count := range m.methodCounts {
			b.WriteString(fmt.Sprintf("http_requests_by_method{method=\"%s\"} %d\n", method, count))
		}
		m.methodMu.Unlock()

		// Top paths
		b.WriteString("# HELP http_requests_by_path Total requests by path.\n")
		b.WriteString("# TYPE http_requests_by_path counter\n")
		m.pathMu.Lock()
		type pathCount struct {
			path  string
			count int64
		}
		paths := make([]pathCount, 0, len(m.pathCounts))
		for p, cnt := range m.pathCounts {
			paths = append(paths, pathCount{p, cnt})
		}
		m.pathMu.Unlock()
		sort.Slice(paths, func(i, j int) bool { return paths[i].count > paths[j].count })
		for i, pc := range paths {
			if i >= 20 {
				break
			}
			b.WriteString(fmt.Sprintf("http_requests_by_path{path=\"%s\"} %d\n", pc.path, pc.count))
		}

		// Per-input-type breakdown
		b.WriteString("# HELP business_requests_by_input_type Total requests by input type.\n")
		b.WriteString("# TYPE business_requests_by_input_type counter\n")
		m.inputTypeMu.Lock()
		for inputType, counter := range m.requestsByInputType {
			b.WriteString(fmt.Sprintf("business_requests_by_input_type{type=\"%s\"} %d\n", inputType, counter.Load()))
		}
		m.inputTypeMu.Unlock()

		// Per-request-status breakdown
		b.WriteString("# HELP business_requests_by_status Total requests by status.\n")
		b.WriteString("# TYPE business_requests_by_status counter\n")
		m.requestStatusMu.Lock()
		for status, counter := range m.requestsByStatus {
			b.WriteString(fmt.Sprintf("business_requests_by_status{status=\"%s\"} %d\n", status, counter.Load()))
		}
		m.requestStatusMu.Unlock()

		// DB connection pool metrics (if available)
		if m.poolStats != nil {
			stats := m.poolStats()
			b.WriteString("# HELP db_pool_total_conns Total number of connections in the pool.\n")
			b.WriteString("# TYPE db_pool_total_conns gauge\n")
			b.WriteString(fmt.Sprintf("db_pool_total_conns %d\n", stats.TotalConns))
			b.WriteString("# HELP db_pool_idle_conns Number of idle connections.\n")
			b.WriteString("# TYPE db_pool_idle_conns gauge\n")
			b.WriteString(fmt.Sprintf("db_pool_idle_conns %d\n", stats.IdleConns))
			b.WriteString("# HELP db_pool_acquired_conns Number of acquired connections.\n")
			b.WriteString("# TYPE db_pool_acquired_conns gauge\n")
			b.WriteString(fmt.Sprintf("db_pool_acquired_conns %d\n", stats.AcquiredConns))
			b.WriteString("# HELP db_pool_max_conns Maximum number of connections.\n")
			b.WriteString("# TYPE db_pool_max_conns gauge\n")
			b.WriteString(fmt.Sprintf("db_pool_max_conns %d\n", stats.MaxConns))
		}

		c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return c.SendString(b.String())
	}
}
