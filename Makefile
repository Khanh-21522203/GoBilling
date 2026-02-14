.PHONY: help build run test clean migrate-up migrate-down docker-build docker-up docker-down

help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start Docker Compose services"
	@echo "  docker-down   - Stop Docker Compose services"

build:
	go build -o bin/gobilling ./cmd/server

run:
	go run ./cmd/server/main.go

test:
	go test -v -race -cover ./...

clean:
	rm -rf bin/
	go clean

migrate-up:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" up

migrate-down:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" down 1

docker-build:
	docker build -t gobilling:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
