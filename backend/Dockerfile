FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/builderwire/lumber-now/backend/internal/handler.Version=${VERSION}" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget curl && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

# Install golang-migrate for running migrations on deploy
RUN curl -fsSL https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xz -C /usr/local/bin

COPY --from=builder /api /usr/local/bin/api
COPY --from=builder /worker /usr/local/bin/worker
COPY migrations /migrations
COPY start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget --spider -q http://localhost:8080/v1/liveness || exit 1

USER appuser
EXPOSE 8080
CMD ["start.sh"]
