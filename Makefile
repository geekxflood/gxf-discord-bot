.PHONY: help build run test clean docker-build docker-run generate validate lint security ci

# Variables
BINARY_NAME=gxf-discord-bot
DOCKER_IMAGE=gxf-discord-bot
VERSION?=latest
GOLANGCI_LINT_VERSION?=v1.61

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete!"

run: ## Run the bot (requires config.yaml)
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) --config config.yaml

run-debug: ## Run the bot with debug logging
	@echo "Running $(BINARY_NAME) in debug mode..."
	./$(BINARY_NAME) --config config.yaml --debug

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-watch: ## Run tests in watch mode (requires entr)
	@echo "Running tests in watch mode..."
	@which entr > /dev/null || (echo "Install entr: brew install entr" && exit 1)
	find . -name "*.go" | entr -c go test -v ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded!"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	go mod tidy
	@echo "Tidy complete!"

generate: build ## Generate sample config file
	@echo "Generating sample config..."
	./$(BINARY_NAME) generate --output config.yaml --force
	@echo "Config generated: config.yaml"

validate: ## Validate configuration file
	@echo "Validating config..."
	./$(BINARY_NAME) validate --config config.yaml

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(VERSION)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -d --name $(BINARY_NAME) \
		-e DISCORD_BOT_TOKEN="${DISCORD_BOT_TOKEN}" \
		-v $(PWD)/config.yaml:/app/config/config.yaml \
		$(DOCKER_IMAGE):$(VERSION)
	@echo "Container started!"

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME)
	docker rm $(BINARY_NAME)
	@echo "Container stopped!"

docker-logs: ## View Docker container logs
	docker logs -f $(BINARY_NAME)

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete!"

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	golangci-lint run
	@echo "Lint complete!"

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix
	@echo "Lint fix complete!"

install-lint: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "golangci-lint installed!"

security: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -fmt=json -out=gosec-report.json ./...
	@echo "Security scan complete! Report: gosec-report.json"

ci: deps lint test-coverage security ## Run all CI checks locally
	@echo "All CI checks passed!"

.DEFAULT_GOAL := help
