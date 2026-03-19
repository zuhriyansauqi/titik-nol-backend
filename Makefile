.PHONY: build run test test-v test-cover clean tidy lint \
       docker-up docker-down docker-build docker-logs \
       migrate-up migrate-down migrate-create \
       vuln-check sec-scan docker-scan security help

# ─── Build & Run ──────────────────────────────────────────
build: ## Build the API binary
	go build -o bin/api cmd/api/main.go

run: ## Run the API locally
	go run cmd/api/main.go

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out

tidy: ## Tidy and verify Go modules
	go mod tidy
	go mod verify

# ─── Testing ──────────────────────────────────────────────
test: ## Run all tests
	go test ./...

test-v: ## Run all tests with verbose output
	go test -v -count=1 ./...

test-cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "\n📊 Open HTML report with: go tool cover -html=coverage.out"

# ─── Linting ──────────────────────────────────────────────
lint: ## Run golangci-lint
	golangci-lint run

# ─── Security ────────────────────────────────────────────
vuln-check: ## Scan Go dependencies for known vulnerabilities (govulncheck)
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

sec-scan: ## Static-analysis security scan on Go source (gosec)
	go run github.com/securego/gosec/v2/cmd/gosec@latest ./...

docker-scan: docker-build ## Scan Docker image for OS/lib CVEs (Trivy)
	trivy image --severity HIGH,CRITICAL titik-nol-backend-api

security: vuln-check sec-scan ## Run all security checks (vuln-check + sec-scan)
	@echo "\n🔒 All security checks passed"

# ─── Docker ───────────────────────────────────────────────
docker-up: ## Start all Docker services in background
	docker compose up -d

docker-down: ## Stop and remove Docker services
	docker compose down

docker-build: ## Build Docker image from scratch
	docker compose build --no-cache

docker-logs: ## Tail Docker service logs
	docker compose logs -f

# ─── Migrations ───────────────────────────────────────────
migrate-up: ## Run all pending migrations
	go run cmd/api/main.go migrate up

migrate-down: ## Rollback the last migration
	go run cmd/api/main.go migrate down

migrate-create: ## Create a new migration (usage: make migrate-create name=<name>)
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(name)

# ─── Help ─────────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

