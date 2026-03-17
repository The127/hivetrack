# Hivetrack — task runner
# Install just: https://github.com/casey/just
#
# Usage:
#   just          → list all available recipes
#   just <recipe> → run a recipe

# Show available recipes by default
default:
    @just --list

# ─── Backend ──────────────────────────────────────────────────────────────────

# Run the backend server (requires postgres, see: just dev-deps)
run:
    cd hivetrack && go run ./cmd/hivetrack

# Build the backend binary (embeds frontend assets if built first)
build:
    cd hivetrack && go build -o ../bin/hivetrack ./cmd/hivetrack

# Build a production Docker image
docker-build:
    docker build -t hivetrack:latest .

# Run all backend tests (unit + architecture, no DB required)
test:
    cd hivetrack && go test ./...

# Run only unit tests (fast, in-memory repositories, no DB)
test-unit:
    cd hivetrack && go test ./... -tags unit -count=1

# Run architecture constraint tests only
test-arch:
    cd hivetrack && go test ./internal/architecture/... -v -count=1

# Run integration tests (requires running postgres — run: just dev-deps first)
test-integration:
    cd hivetrack && go test ./... -tags integration -count=1

# Run all tests including integration
test-all: dev-deps
    cd hivetrack && go test ./... -tags 'unit integration' -count=1

# Run the linter
lint:
    cd hivetrack && golangci-lint run ./...

# Format all Go source files
fmt:
    cd hivetrack && gofmt -w .
    cd hivetrack && goimports -w .

# Check for outdated dependencies
deps-check:
    cd hivetrack && go list -u -m all

# Update all dependencies
deps-update:
    cd hivetrack && go get -u ./... && go mod tidy

# ─── Database ─────────────────────────────────────────────────────────────────

# Start only the database (for running the backend locally without full stack)
db-start:
    docker compose up -d postgres

# Stop the database
db-stop:
    docker compose stop postgres

# Open a psql shell to the local development database
db-shell:
    docker compose exec postgres psql -U hivetrack -d hivetrack

# Reset the database (drops and recreates — destroys all data)
db-reset:
    docker compose down -v postgres
    docker compose up -d postgres

# ─── Frontend ─────────────────────────────────────────────────────────────────

# Install frontend dependencies
ui-install:
    cd hivetrack-ui && npm install

# Run the frontend dev server (proxies /api to the backend)
ui-dev:
    cd hivetrack-ui && npm run dev

# Build the frontend for production (output goes to hivetrack-ui/dist)
ui-build:
    cd hivetrack-ui && npm run build

# Run frontend type checking
ui-check:
    cd hivetrack-ui && npm run type-check

# Run frontend linter
ui-lint:
    cd hivetrack-ui && npm run lint

# ─── Local Development ────────────────────────────────────────────────────────

# Start all development dependencies (postgres + keyline) in the background
dev-deps:
    docker compose up -d
    @echo "Waiting for postgres to be ready..."
    @until pg_isready -h localhost -p 5432 -U hivetrack > /dev/null 2>&1; do sleep 1; done
    @echo "Postgres is ready."

# Start the full local development stack (deps + backend + frontend)
# Opens backend on :8080, frontend dev server on :5173
dev: dev-deps
    #!/usr/bin/env bash
    trap 'kill %1 %2 2>/dev/null' EXIT
    just run &
    just ui-dev &
    wait

# Stop all docker compose services
dev-stop:
    docker compose down

# ─── Release ──────────────────────────────────────────────────────────────────

# Build frontend and embed into backend binary (production build)
release: ui-build build
    @echo "Release binary built at bin/hivetrack"

# Run the release binary locally (full stack in one process)
release-run: release
    ./bin/hivetrack

# ─── Git Hooks ────────────────────────────────────────────────────────────────

# Install git hooks (run once after cloning)
install-hooks:
    chmod +x scripts/hooks/pre-push
    cp scripts/hooks/pre-push .git/hooks/pre-push
    @echo "✓ pre-push hook installed"

# Remove installed git hooks
uninstall-hooks:
    rm -f .git/hooks/pre-push
    @echo "✓ hooks removed"

# ─── Utilities ────────────────────────────────────────────────────────────────

# Print the current version (from git tag or commit)
version:
    @git describe --tags --always --dirty 2>/dev/null || echo "dev"

# Verify the full project is in a healthy state (lint + test-arch + test-unit)
check: lint test-arch test-unit
    @echo "All checks passed."
