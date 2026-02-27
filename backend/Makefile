.PHONY: dev build test lint sqlc migrate-up migrate-down docker-up docker-down

DATABASE_URL ?= postgres://lumber:lumber@localhost:5432/lumber_now?sslmode=disable
MIGRATE_CMD = migrate -path migrations -database "$(DATABASE_URL)"

dev:
	go run ./cmd/api

worker:
	go run ./cmd/worker

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

test:
	go test ./...

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
