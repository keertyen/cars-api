.PHONY: build run emulator stop test test-cover lint tidy docker-up docker-down

BINARY      := cars-api
PROJECT_ID  := cars-api-local
EMULATOR    := localhost:8081

## build: compile to bin/cars-api
build:
	go build -o bin/$(BINARY) ./cmd/server

## run: start the Firestore emulator then the API server
run: emulator
	FIRESTORE_EMULATOR_HOST=$(EMULATOR) \
	GCP_PROJECT_ID=$(PROJECT_ID) \
	PORT=8080 \
	go run ./cmd/server

## emulator: start the Firestore emulator via Firebase CLI (runs in background)
emulator:
	firebase emulators:start --only firestore --project=$(PROJECT_ID) &
	@echo "Waiting for Firestore emulator on $(EMULATOR)..."
	@until curl -sf http://$(EMULATOR) > /dev/null 2>&1; do sleep 1; done
	@echo "Firestore emulator ready"

## stop: stop all docker-compose services
stop:
	docker compose down

## test: run all unit tests with race detector
test:
	go test -v -race ./...

## test-cover: run tests and open HTML coverage report
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

## lint: run golangci-lint (install: brew install golangci-lint)
lint:
	golangci-lint run ./...

## tidy: sync go.mod / go.sum
tidy:
	go mod tidy

## docker-up: build image and start everything via docker compose
docker-up:
	docker compose up --build

## docker-down: tear down all containers
docker-down:
	docker compose down --volumes
