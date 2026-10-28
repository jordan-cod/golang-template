APP        := app
BIN_DIR    := bin
IMAGE      ?= golang-template
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -s -w -X main.version=$(VERSION)

AIR_VERSION           := v1.63.4
GOLANGCI_LINT_VERSION := v2.15.0

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev
dev: ## Run with hot reload (air)
	@command -v air >/dev/null || go install github.com/air-verse/air@$(AIR_VERSION)
	air -c .air.toml

.PHONY: run
run: ## Run without hot reload
	go run ./cmd/$(APP)

.PHONY: build
build: ## Build the binary into bin/
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP) ./cmd/$(APP)

.PHONY: test
test: ## Run tests with race detector and coverage
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -n 1

.PHONY: cover
cover: test ## Open the HTML coverage report
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	golangci-lint fmt ./...

.PHONY: tidy
tidy: ## Tidy and verify go modules
	go mod tidy
	go mod verify

.PHONY: docker-build
docker-build: ## Build the production image
	docker build --target prod --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) .

.PHONY: docker-run
docker-run: docker-build ## Run the production image
	docker run --rm -p 8080:8080 $(IMAGE):$(VERSION)

.PHONY: compose-up
compose-up: ## Start the dev stack with hot reload
	docker compose up --build

.PHONY: compose-down
compose-down: ## Stop the dev stack
	docker compose down

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) tmp coverage.out coverage.html build-errors.log
