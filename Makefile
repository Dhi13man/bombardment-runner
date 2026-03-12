.PHONY: build test test-cover lint run run-cli docker docker-run swagger clean help

GO_DIR := ./app
BINARY := bombardment

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	cd $(GO_DIR) && CGO_ENABLED=0 go build -o $(BINARY) .

test: ## Run tests with race detector
	cd $(GO_DIR) && go test -v -race -count=1 ./...

test-cover: ## Run tests with coverage report
	cd $(GO_DIR) && go test -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	cd $(GO_DIR) && golangci-lint run ./...

run: ## Run the server locally
	cd $(GO_DIR) && go run main.go server

run-cli: ## Show CLI help
	cd $(GO_DIR) && go run main.go cli --help

docker: ## Build Docker image
	docker build -t $(BINARY) .

docker-run: ## Run with Docker Compose
	docker-compose up --build

swagger: ## Regenerate Swagger docs
	cd $(GO_DIR) && swag init -g main.go --parseDependency --parseInternal

clean: ## Remove build artifacts
	rm -f $(GO_DIR)/$(BINARY) $(GO_DIR)/coverage.out $(GO_DIR)/coverage.html
	rm -f $(GO_DIR)/__debug_bin*
