.PHONY: dev build test test-race test-cover lint sqlc migrate-up migrate-down docker-up docker-down tidy

DATABASE_URL ?= postgres://lumber:lumber@localhost:5432/lumber_now?sslmode=disable
MIGRATE_CMD = migrate -path migrations -database "$(DATABASE_URL)"
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

dev:
	go run ./cmd/api

worker:
	go run ./cmd/worker

build:
	go build -ldflags="-s -w -X github.com/builderwire/lumber-now/backend/internal/handler.Version=$(VERSION)" -o bin/api ./cmd/api
	go build -ldflags="-s -w" -o bin/worker ./cmd/worker

test:
	go test ./...

test-race:
	go test -race ./...

test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

migrate-up:
	$(MIGRATE_CMD) up

migrate-down:
	$(MIGRATE_CMD) down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

tidy:
	go mod tidy
