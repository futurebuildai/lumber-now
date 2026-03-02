package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/builderwire/lumber-now/backend/internal/app"
	"github.com/builderwire/lumber-now/backend/internal/platform/anthropic"
	"github.com/builderwire/lumber-now/backend/internal/platform/email"
	"github.com/builderwire/lumber-now/backend/internal/platform/gcloud"
	s3platform "github.com/builderwire/lumber-now/backend/internal/platform/s3"
	"github.com/builderwire/lumber-now/backend/internal/service"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

// Worker metrics tracked via atomics for lightweight observability.
type workerMetrics struct {
	processed    atomic.Int64
	failed       atomic.Int64
	retried      atomic.Int64
	ticksSkipped atomic.Int64
	panics       atomic.Int64
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return fallback
}

func main() {
	godotenv.Load()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(cfg.LogLevel)})))

	// Configurable worker parameters
	concurrency := getEnvInt("WORKER_CONCURRENCY", 3)
	pollInterval := getEnvDuration("WORKER_POLL_INTERVAL", 10*time.Second)
	batchSize := getEnvInt("WORKER_BATCH_SIZE", 10)
	stuckTimeout := getEnvDuration("WORKER_STUCK_TIMEOUT", 15*time.Minute)

	pool, err := store.NewPool(context.Background(), cfg.DatabaseURL, store.PoolConfig{
		MaxConns: cfg.DBMaxConns,
		MinConns: cfg.DBMinConns,
	})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	s := store.New(pool)
	aiClient := anthropic.NewClient(cfg.AnthropicAPIKey)

	// Track closers for graceful shutdown
	var closers []io.Closer

	// Optional: S3/Media for voice downloads
	var mediaSvc *service.MediaService
	s3Client, err := s3platform.NewClient(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey)
	if err != nil {
		slog.Warn("S3 client init failed, voice transcription disabled", "error", err)
	} else {
		mediaSvc = service.NewMediaService(s3Client, cfg.S3Bucket)
	}

	// Optional: Google Cloud STT
	var transcriber service.Transcriber
	if cfg.GCloudCredentialsFile != "" {
		speechClient, err := gcloud.NewSpeechClient(context.Background(), cfg.GCloudCredentialsFile)
		if err != nil {
			slog.Warn("Google Cloud STT init failed", "error", err)
		} else {
			transcriber = speechClient
			closers = append(closers, speechClient)
		}
	}

	// Optional: Resend email
	var emailClient service.EmailSender
	if cfg.ResendAPIKey != "" {
		emailClient = email.NewClient(cfg.ResendAPIKey, cfg.EmailFrom)
	}

	reqSvc := service.NewRequestService(s, aiClient, transcriber, emailClient, mediaSvc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("worker started",
		"concurrency", concurrency,
		"poll_interval", pollInterval,
		"batch_size", batchSize,
		"stuck_timeout", stuckTimeout,
	)

	// Write health file so Docker can check worker liveness
	writeHealthFile()

	var wg sync.WaitGroup
	var metrics workerMetrics
	sem := make(chan struct{}, concurrency)

	// Start worker metrics HTTP server for Prometheus scraping
	metricsPort := os.Getenv("WORKER_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}
	metricsSrv := startMetricsServer(metricsPort, &metrics)
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		metricsSrv.Shutdown(shutCtx)
	}()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Periodic metrics reporting
	metricsTicker := time.NewTicker(60 * time.Second)
	defer metricsTicker.Stop()

	// Periodic stuck request recovery
	stuckTicker := time.NewTicker(5 * time.Minute)
	defer stuckTicker.Stop()

	// Adaptive backoff: when no work is found, increase poll interval up to maxBackoff
	currentInterval := pollInterval
	maxBackoff := 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down, waiting for in-flight requests...")
			// Remove health file on shutdown
			os.Remove("/tmp/worker-healthy")
			// Wait for in-flight email goroutines
			reqSvc.Close()
			wg.Wait()
			for _, c := range closers {
				c.Close()
			}
			slog.Info("worker shutdown complete",
				"total_processed", metrics.processed.Load(),
				"total_failed", metrics.failed.Load(),
				"total_retried", metrics.retried.Load(),
			)
			return

		case <-metricsTicker.C:
			slog.Info("worker_metrics",
				"processed", metrics.processed.Load(),
				"failed", metrics.failed.Load(),
				"retried", metrics.retried.Load(),
				"ticks_skipped", metrics.ticksSkipped.Load(),
				"panics", metrics.panics.Load(),
				"current_interval", currentInterval,
			)
			// Refresh health file
			writeHealthFile()

		case <-stuckTicker.C:
			recoverStuckRequests(ctx, s, stuckTimeout)

		case <-ticker.C:
			select {
			case sem <- struct{}{}:
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer func() { <-sem }()
					prevProcessed := metrics.processed.Load()
					safeProcessPendingRequests(ctx, s, reqSvc, int32(batchSize), &metrics)

					// Adaptive backoff: if work was found, reset; otherwise increase
					if metrics.processed.Load() > prevProcessed {
						if currentInterval != pollInterval {
							currentInterval = pollInterval
							ticker.Reset(currentInterval)
							slog.Debug("worker backoff reset", "interval", currentInterval)
						}
					} else {
						newInterval := currentInterval * 2
						if newInterval > maxBackoff {
							newInterval = maxBackoff
						}
						if newInterval != currentInterval {
							currentInterval = newInterval
							ticker.Reset(currentInterval)
							slog.Debug("worker backoff increased", "interval", currentInterval)
						}
					}
				}()
			default:
				metrics.ticksSkipped.Add(1)
				slog.Debug("worker tick skipped, previous batch still processing")
			}
		}
	}
}

func writeHealthFile() {
	if err := os.WriteFile("/tmp/worker-healthy", []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		slog.Warn("failed to write health file", "error", err)
	}
}

func recoverStuckRequests(ctx context.Context, s *store.Store, stuckTimeout time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in stuck request recovery", "panic", r)
		}
	}()

	recoverCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	count, err := s.Queries.RecoverStuckRequests(recoverCtx, db.RecoverStuckRequestsParams{
		StuckInterval: stuckTimeout.String(),
		Limit:         20,
	})
	if err != nil {
		slog.Error("failed to recover stuck requests", "error", err)
		return
	}
	if count > 0 {
		slog.Warn("recovered stuck processing requests", "count", count)
	}
}

func safeProcessPendingRequests(ctx context.Context, s *store.Store, reqSvc *service.RequestService, batchSize int32, metrics *workerMetrics) {
	defer func() {
		if r := recover(); r != nil {
			metrics.panics.Add(1)
			slog.Error("panic in worker processing", "panic", r)
		}
	}()

	// Use a scoped context with timeout for the batch to prevent hung queries
	batchCtx, batchCancel := context.WithTimeout(ctx, 60*time.Second)
	defer batchCancel()

	// Atomically claim up to batchSize pending requests using FOR UPDATE SKIP LOCKED
	requests, err := s.Queries.ClaimPendingRequests(batchCtx, batchSize)
	if err != nil {
		slog.Error("failed to claim pending requests", "error", err)
		return
	}

	for _, req := range requests {
		slog.Info("processing request", "id", req.ID, "dealer_id", req.DealerID, "type", req.InputType)
		// Per-request timeout (2 minutes for AI processing)
		reqCtx, reqCancel := context.WithTimeout(ctx, 2*time.Minute)
		if _, err := reqSvc.Process(reqCtx, req.ID); err != nil {
			metrics.failed.Add(1)
			slog.Error("failed to process request", "id", req.ID, "error", err)
		} else {
			metrics.processed.Add(1)
		}
		reqCancel()
	}

	// Retry failed requests (max 3 retries) - dead-letter queue pattern
	retried, err := s.Queries.RetryFailedRequests(batchCtx, db.RetryFailedRequestsParams{
		MaxRetries: 3,
		Limit:      5,
	})
	if err != nil {
		slog.Error("failed to retry failed requests", "error", err)
		return
	}
	if len(retried) > 0 {
		metrics.retried.Add(int64(len(retried)))
		slog.Info("retried failed requests", "count", len(retried))

		// Alert on requests hitting max retries (dead-letter threshold)
		for _, r := range retried {
			if r.RetryCount >= 2 {
				slog.Error("request approaching max retries (dead-letter)",
					"request_id", r.ID,
					"retry_count", r.RetryCount,
					"last_error", r.LastError,
					"dealer_id", r.DealerID,
				)
			}
		}
	}
}

func startMetricsServer(port string, m *workerMetrics) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		var b strings.Builder
		b.WriteString("# TYPE worker_requests_processed_total counter\n")
		b.WriteString("worker_requests_processed_total " + strconv.FormatInt(m.processed.Load(), 10) + "\n")
		b.WriteString("# TYPE worker_requests_failed_total counter\n")
		b.WriteString("worker_requests_failed_total " + strconv.FormatInt(m.failed.Load(), 10) + "\n")
		b.WriteString("# TYPE worker_requests_retried_total counter\n")
		b.WriteString("worker_requests_retried_total " + strconv.FormatInt(m.retried.Load(), 10) + "\n")
		b.WriteString("# TYPE worker_ticks_skipped_total counter\n")
		b.WriteString("worker_ticks_skipped_total " + strconv.FormatInt(m.ticksSkipped.Load(), 10) + "\n")
		b.WriteString("# TYPE worker_panics_total counter\n")
		b.WriteString("worker_panics_total " + strconv.FormatInt(m.panics.Load(), 10) + "\n")

		w.Write([]byte(b.String()))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		slog.Info("worker metrics server started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("worker metrics server error", "error", err)
		}
	}()
	return srv
}
