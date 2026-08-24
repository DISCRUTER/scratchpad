.PHONY: build fmt run test test-race migrate-up migrate-down compose-up compose-down

DSN ?= postgres://scratchpad:password@localhost/scratchpad?sslmode=disable
MIGRATE := $(shell command -v migrate 2>/dev/null || echo "$$HOME/go/bin/migrate")

build:
	@go build -o ./build/web ./cmd/web

fmt:
	@go fmt ./...

run: build
	@./build/web

test:
	@go test -v ./...

test-race:
	@go test -race -count=1 ./...

migrate-up:
	@$(MIGRATE) -path=./migrations -database="$(DSN)" up

migrate-down:
	@$(MIGRATE) -path=./migrations -database="$(DSN)" down 1

compose-up:
	@podman-compose up -d --build

compose-down:
	@podman-compose down
