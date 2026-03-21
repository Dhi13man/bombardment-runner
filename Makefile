.PHONY: build test test-cover lint run run-cli docker docker-run swagger clean help ui-install ui-build ui-watch ui-typecheck check-bundle-size landing-minify landing-webp

GO_DIR := ./app
BINARY := bombardment

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

ui-install: ## Install frontend dependencies
	cd $(GO_DIR) && npm install

ui-build: ## Build frontend assets (CSS + JS)
	cd $(GO_DIR) && npm run build

ui-watch: ## Watch mode for frontend development
	cd $(GO_DIR) && npm run watch

ui-typecheck: ## Run TypeScript type checking
	cd $(GO_DIR) && npm run typecheck

ui-test: ## Run frontend unit tests
	cd $(GO_DIR) && npm test

build: ui-build ## Build frontend assets then Go binary
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

check-bundle-size: ui-build ## Check frontend bundle sizes against budgets
	@echo "Checking bundle sizes..."
	@CSS_SIZE=$$(wc -c < $(GO_DIR)/src/app/ui/static/css/app.min.css); \
	JS_SIZE=$$(wc -c < $(GO_DIR)/src/app/ui/static/js/app.min.js); \
	ICON_SIZE=$$(wc -c < $(GO_DIR)/src/app/ui/static/icons/sprite.svg); \
	echo "  CSS:   $$CSS_SIZE bytes (budget: 40960 / 40KB)"; \
	echo "  JS:    $$JS_SIZE bytes (budget: 90112 / 88KB)"; \
	echo "  Icons: $$ICON_SIZE bytes (budget: 12288 / 12KB)"; \
	if [ $$CSS_SIZE -gt 40960 ]; then echo "FAIL: CSS exceeds 40KB budget" && exit 1; fi; \
	if [ $$JS_SIZE -gt 90112 ]; then echo "FAIL: JS exceeds 88KB budget" && exit 1; fi; \
	if [ $$ICON_SIZE -gt 12288 ]; then echo "FAIL: Icons exceed 12KB budget" && exit 1; fi; \
	echo "All bundle sizes within budget."

LANDING_DIR := ./landing-site

landing-minify: ## Minify landing site CSS and JS
	@echo "Minifying landing site assets..."
	@npx --yes cssnano-cli $(LANDING_DIR)/src/styles/global-styles.css $(LANDING_DIR)/src/styles/global-styles.min.css 2>/dev/null
	@npx --yes terser $(LANDING_DIR)/src/scripts/animations.js -o $(LANDING_DIR)/src/scripts/animations.min.js --compress --mangle 2>/dev/null
	@ORIG_CSS=$$(wc -c < $(LANDING_DIR)/src/styles/global-styles.css); \
	MIN_CSS=$$(wc -c < $(LANDING_DIR)/src/styles/global-styles.min.css); \
	ORIG_JS=$$(wc -c < $(LANDING_DIR)/src/scripts/animations.js); \
	MIN_JS=$$(wc -c < $(LANDING_DIR)/src/scripts/animations.min.js); \
	echo "  CSS: $$ORIG_CSS -> $$MIN_CSS bytes"; \
	echo "  JS:  $$ORIG_JS -> $$MIN_JS bytes"

landing-webp: ## Convert landing site PNGs to WebP
	@echo "Converting PNGs to WebP..."
	@for f in $(LANDING_DIR)/assets/ui-*.png; do \
		cwebp -q 82 "$$f" -o "$${f%.png}.webp" 2>/dev/null; \
	done
	@echo "Done."

clean: ## Remove build artifacts
	rm -f $(GO_DIR)/$(BINARY) $(GO_DIR)/coverage.out $(GO_DIR)/coverage.html
	rm -f $(GO_DIR)/__debug_bin*
	rm -f $(LANDING_DIR)/src/styles/global-styles.min.css $(LANDING_DIR)/src/scripts/animations.min.js
