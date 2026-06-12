.PHONY: build run test clean docker-up docker-down migrate lint deps swagger

APP_NAME=live-commerce-bi
BUILD_DIR=bin
VERSION=2.0.0
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/server ./cmd/api

run:
	go run ./cmd/api

dev:
	GIN_MODE=debug go run ./cmd/api

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

test-short:
	go test -v -short ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app

docker-ps:
	docker-compose ps

migrate:
	docker-compose exec postgres psql -U biuser -d live_commerce_bi -f /docker-entrypoint-initdb.d/02-advanced.sql

seed:
	@echo "Seeding demo data..."
	docker-compose exec postgres psql -U biuser -d live_commerce_bi -c "\
		INSERT INTO streamers (name, platform, category, follower_count) VALUES \
		('李佳琦Austin', 'taobao_live', '美妆', 5000000), \
		('疯狂小杨哥', 'douyin', '全品类', 8000000), \
		('辛有志辛巴', 'kuaishou', '全品类', 6000000), \
		('罗永浩', 'douyin', '数码', 3000000), \
		('薇娅viya', 'taobao_live', '全品类', 4500000) \
		ON CONFLICT DO NOTHING;"

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

deps:
	go mod tidy
	go mod download

swagger:
	swag init -g cmd/api/main.go -o docs/swagger

benchmark:
	go test -bench=. -benchmem ./...

.DEFAULT_GOAL := build
