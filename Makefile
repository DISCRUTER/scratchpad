build:
	@go build -o ./build/web ./cmd/web

fmt:
	@go fmt ./...

run: build
	@./build/web

test:
	@go test -v ./...