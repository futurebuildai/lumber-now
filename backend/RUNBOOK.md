# LumberNow Operations Runbook

## Service Architecture

- **API** (port 8080): Go/Fiber HTTP server handling all REST endpoints
- **Worker** (port 9090): Background processor for AI parsing, email delivery
- **PostgreSQL** (port 5432): Primary datastore with pgx connection pool
- **MinIO/S3**: Object storage for media uploads (images, PDFs, audio)

## Health Checks

| Service | Endpoint | Expected |
|---------|----------|----------|
| API | `GET /v1/health` | `{"status":"ok","db":"ok"}` |
| API | `GET /v1/readiness` | `{"status":"ready"}` with circuit breaker states |
| API | `GET /v1/liveness` | `{"status":"ok"}` |
| Worker | `GET /healthz` | `{"status":"ok"}` |
| Worker | `GET /metrics` | Prometheus text format |

## Startup Sequence

1. PostgreSQL must be healthy (`pg_isready`)
2. Migrations run (`migrate` service)
3. API starts (depends on DB + migrations)
4. Worker starts (depends on DB + migrations)

## Common Alerts and Responses

### HighErrorRate (API 5xx > 5%)
1. Check API logs: `docker compose logs api --tail 100`
2. Check DB connectivity: `docker compose exec db pg_isready`
3. Check circuit breaker states: `curl localhost:8080/v1/readiness`
4. If AI circuit open: Anthropic API may be down, requests queue for retry

### HighLatencyP99 (> 5s)
1. Check DB pool: `curl localhost:8080/v1/metrics | grep db_pool`
2. If `acquired_conns` near `max_conns`: increase pool size or check slow queries
3. Check worker backlog: pending requests accumulating

### WorkerHighFailureRate (> 10%)
1. Check worker logs: `docker compose logs worker --tail 100`
2. Common cause: Anthropic API rate limiting or outage
3. Failed requests auto-retry up to 3 times with dead-letter alerting

### DatabasePoolExhausted (> 90%)
1. Check active queries: `SELECT * FROM pg_stat_activity WHERE state = 'active'`
2. Kill long-running queries if needed: `SELECT pg_terminate_backend(pid)`
3. Increase `MaxConns` in pool config if sustained

### CircuitBreakerOpen
1. Check which circuit: AI, S3, or Email
2. Circuit auto-recovers after 60s timeout (half-open → closed after 2 successes)
3. If persistent: check external service status

## Scaling

### Horizontal (API)
- API is stateless; add replicas behind load balancer
- Rate limiting is per-instance (fiber limiter uses in-memory store)
- Idempotency cache is in-memory; upgrade to Redis for multi-instance

### Vertical
- Increase `deploy.resources.limits.memory` in docker-compose
- Increase DB pool `MaxConns` (default: determined by pgxpool)

## Database Operations

### Run migrations
```bash
docker compose run --rm migrate
```

### Rollback last migration
```bash
docker compose exec db psql -U lumber lumber_now -f /migrations/000009_add_inventory_version.down.sql
```

### Check table sizes
```sql
SELECT relname, pg_size_pretty(pg_total_relation_size(relid))
FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC;
```

## Environment Variables

### Required
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Min 32 characters, used for HS256 token signing
- `CORS_ORIGINS`: Comma-separated allowed origins

### Optional
- `ANTHROPIC_API_KEY`: Enables AI parsing
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_REGION`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`: Media storage
- `GCLOUD_CREDENTIALS_FILE`: Voice transcription (Google Cloud STT)
- `RESEND_API_KEY`, `EMAIL_FROM`: Email delivery via Resend

### Worker-specific
- `WORKER_CONCURRENCY`: Parallel processing slots (default: 3)
- `WORKER_POLL_INTERVAL`: Polling frequency (default: 10s)
- `WORKER_BATCH_SIZE`: Requests per poll (default: 10)
- `WORKER_STUCK_TIMEOUT`: Recovery threshold (default: 15m)

## Graceful Shutdown

Both API and Worker handle SIGTERM/SIGINT:
1. Stop accepting new requests
2. Wait for in-flight email goroutines
3. Close external clients (GCloud STT)
4. Drain DB connections
5. 30s timeout before forced exit

## Security Notes

- JWT tokens: 15m access, 7d refresh, HS256 signed
- Account lockout: 5 failed attempts → 15min lock (DB-persisted)
- Token blacklist: in-memory (cleared on restart)
- SSRF prevention: DNS resolution check on media URLs
- CSRF: X-Requested-With header required on authenticated endpoints
- Rate limiting: 60 req/min global, 5 req/min auth endpoints
