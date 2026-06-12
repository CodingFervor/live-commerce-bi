.PHONY: build run test clean docker-up docker-down migrate

APP_NAME=live-commerce-bi
BUILD_DIR=bin

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/server ./cmd/api

run:
	go run ./cmd/api

dev:
	go run ./cmd/api -mode=debug

test:
	go test -v -race ./...

clean:
	rm -rf $(BUILD_DIR)

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app

migrate:
	docker-compose exec postgres psql -U biuser -d live_commerce_bi -f /docker-entrypoint-initdb.d/init.sql

lint:
	golangci-lint run ./...

deps:
	go mod tidy
	go mod download

.DEFAULT_GOAL := build
