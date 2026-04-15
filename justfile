# Hivetrack — task runner
# Install just: https://github.com/casey/just
#
# Usage:
#   just          → list all available recipes
#   just <recipe> → run a recipe

# Go binary — override with: HIVETRACK_GO=/usr/bin/go just <recipe>
go := env("HIVETRACK_GO", `command -v go 2>/dev/null || echo go`)

# Show available recipes by default
default:
    @just --list

# ─── Backend ──────────────────────────────────────────────────────────────────

# Run the backend server (requires postgres, see: just dev-deps)
run:
    cd hivetrack && {{go}} run ./cmd/hivetrack

# Build the backend binary (embeds frontend assets if built first)
build:
    cd hivetrack && {{go}} build -o ../bin/hivetrack ./cmd/hivetrack

# Build a production Docker image
docker-build:
    docker build -t hivetrack:latest .

# Run all backend tests (unit + architecture, no DB required)
test:
    cd hivetrack && {{go}} test ./...

# Run only unit tests (fast, in-memory repositories, no DB)
test-unit:
    cd hivetrack && {{go}} test ./... -tags unit -count=1

# Run architecture constraint tests only
test-arch:
    cd hivetrack && {{go}} test ./internal/architecture/... -v -count=1

# Run integration tests (requires running postgres — run: just dev-deps first)
test-integration:
    cd hivetrack && {{go}} test ./... -tags integration -count=1

# Run all tests including integration
test-all: dev-deps
    cd hivetrack && {{go}} test ./... -tags 'unit integration' -count=1

# Run the linter
lint:
    cd hivetrack && golangci-lint run ./...

# Format all Go source files
fmt:
    cd hivetrack && gofmt -w .
    cd hivetrack && goimports -w .

# Check for outdated dependencies
deps-check:
    cd hivetrack && {{go}} list -u -m all

# Update all dependencies
deps-update:
    cd hivetrack && {{go}} get -u ./... && {{go}} mod tidy

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

# Start services needed for E2E tests (idempotent — skips already-running services)
_e2e-services:
    #!/usr/bin/env bash
    set -euo pipefail
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [ -s "$NVM_DIR/nvm.sh" ] && source "$NVM_DIR/nvm.sh"
    HIVEMIND_DIR="${HIVEMIND_DIR:-$HOME/code/github.com/The127/hivemind}"

    # Start postgres + keyline
    just dev-deps

    # Start hivetrack backend
    if ! lsof -i :8086 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "Starting hivetrack backend..."
        just run >/tmp/hivetrack-e2e.log 2>&1 &
        until lsof -i :8086 -sTCP:LISTEN -t >/dev/null 2>&1; do sleep 1; done
        echo "Backend ready."
    fi

    # Start frontend dev server
    if ! lsof -i :5173 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "Starting frontend dev server..."
        (cd hivetrack-ui && nvm use 2>/dev/null || true && [ -d node_modules ] || npm install && npm run dev) >/tmp/hivetrack-ui-e2e.log 2>&1 &
        until lsof -i :5173 -sTCP:LISTEN -t >/dev/null 2>&1; do sleep 1; done
        echo "Frontend ready."
    fi

    # Start hivemind + a Claude drone registered for e2e-test (for integration tests)
    if [ -f "$HIVEMIND_DIR/justfile" ]; then
        if ! lsof -i :50051 -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo "Starting hivemind..."
            docker compose -f "$HIVEMIND_DIR/compose.yml" up hivemind -d
            until lsof -i :50051 -sTCP:LISTEN -t >/dev/null 2>&1; do sleep 1; done
            echo "Hivemind ready."
        fi
        echo "Waiting for hivemind management API..."
        until curl -sf http://localhost:8080/api/v1/drones >/dev/null 2>&1; do sleep 1; done
        echo "Registering e2e drone for e2e-test project..."
        E2E_TOKEN=$(curl -sf -X POST http://localhost:8080/api/v1/drones/tokens \
          -H "Content-Type: application/json" \
          -d '{"project_slug":"e2e-test","capabilities":["refinement"],"max_concurrency":1}' \
          | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
        echo "Starting Claude drone for e2e-test..."
        just -f "$HIVEMIND_DIR/justfile" claude-drone "$E2E_TOKEN" "e2e-claude-drone" "localhost:50051" >/tmp/hivemind-e2e-drone.log 2>&1 &
        echo "Drone started (PID $!)."
    else
        echo "Warning: HIVEMIND_DIR=$HIVEMIND_DIR not found — integration tests will be skipped."
    fi

# Run Playwright E2E tests — starts all required services if not already running
ui-e2e:
    #!/usr/bin/env bash
    set -euo pipefail
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [ -s "$NVM_DIR/nvm.sh" ] && source "$NVM_DIR/nvm.sh"
    just _e2e-services
    cd hivetrack-ui
    nvm use 2>/dev/null || true
    [ -d node_modules ] || npm install
    npx playwright install --with-deps chromium 2>/dev/null || true
    npm run e2e

# Run Playwright E2E tests with interactive UI — starts all required services if not already running
ui-e2e-ui:
    #!/usr/bin/env bash
    set -euo pipefail
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [ -s "$NVM_DIR/nvm.sh" ] && source "$NVM_DIR/nvm.sh"
    just _e2e-services
    cd hivetrack-ui
    nvm use 2>/dev/null || true
    [ -d node_modules ] || npm install
    npx playwright install --with-deps chromium 2>/dev/null || true
    npm run e2e:ui

# Open the Playwright HTML report from the last E2E run
ui-e2e-report:
    #!/usr/bin/env bash
    set -euo pipefail
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    [ -s "$NVM_DIR/nvm.sh" ] && source "$NVM_DIR/nvm.sh"
    cd hivetrack-ui
    nvm use 2>/dev/null || true
    npx playwright show-report

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
    @until docker compose exec -T postgres pg_isready -U hivetrack > /dev/null 2>&1; do sleep 1; done
    @echo "Postgres is ready."

# Start the full local development stack (deps + backend + frontend)
# Opens backend on :8080, frontend dev server on :5173
dev: dev-deps ui-install
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

# ─── MCP Server ──────────────────────────────────────────────────────────────

# Run the MCP server (stdio transport, for Claude Code integration)
mcp:
    cd mcp && {{go}} run ./cmd/hivetrack-mcp

# Build the MCP server binary
mcp-build:
    cd mcp && {{go}} build -o ../bin/hivetrack-mcp ./cmd/hivetrack-mcp

# Run MCP server tests
mcp-test:
    cd mcp && {{go}} test ./... -count=1

# Launch the MCP Inspector UI against the MCP server binary
# Override server URL with: just mcp-inspect url=https://your-instance.example.com
mcp-inspect url="https://hivetrack.karo.gay": mcp-build
    npx @modelcontextprotocol/inspector -e HIVETRACK_URL={{url}} {{justfile_directory()}}/bin/hivetrack-mcp

# ─── Utilities ────────────────────────────────────────────────────────────────

# Print the current version (from git tag or commit)
version:
    @git describe --tags --always --dirty 2>/dev/null || echo "dev"

# Verify the full project is in a healthy state (lint + test-arch + test-unit)
check: lint test-arch test-unit
    @echo "All checks passed."
