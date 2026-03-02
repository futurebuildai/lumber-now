FROM golang:1.23-alpine AS builder

ARG VERSION=dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/builderwire/lumber-now/backend/internal/handler.Version=${VERSION}" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /api /usr/local/bin/api
COPY --from=builder /worker /usr/local/bin/worker
COPY migrations /migrations

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget --spider -q http://localhost:8080/v1/liveness || exit 1

USER appuser
EXPOSE 8080
CMD ["api"]
