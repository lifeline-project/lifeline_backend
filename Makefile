APP_NAME=lifeline_backend
BUILD_DIR=./bin
MAIN=./cmd/server/main.go

.PHONY: run build lint migrate-up migrate-down tidy

## run: start the server locally
run:
	go run $(MAIN)

## build: compile the binary to ./bin/
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

## tidy: sync go.mod/go.sum
tidy:
	go mod tidy

## lint: run golangci-lint (install separately)
lint:
	golangci-lint run ./...

## migrate-up: apply all pending migrations
migrate-up:
	@echo "migrate-up: not implemented yet"

## migrate-down: roll back the last migration
migrate-down:
	@echo "migrate-down: not implemented yet"
