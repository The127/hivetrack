# Hivetrack Backend

Go HTTP API for Hivetrack. See the [root README](../README.md) for project overview.

## Requirements

- Go 1.24+
- PostgreSQL 15+
- An OIDC provider (Keyline, Keycloak, Authentik)

## Running locally

```bash
# From repo root — starts postgres + backend + frontend
just dev

# Backend only (requires postgres running)
just run

# Or directly
go run ./cmd/hivetrack
```

Config is loaded from `config.yaml`. Override any field with `HIVETRACK_` env vars:
```bash
HIVETRACK_SERVER_PORT=9090 go run ./cmd/hivetrack
HIVETRACK_DATABASE_URL=postgres://... go run ./cmd/hivetrack
```

## Testing

```bash
just test           # unit tests + architecture tests (no DB needed)
just test-unit      # unit tests only
just test-arch      # architecture constraint tests only
just test-integration   # requires running postgres
```

## Adding dependencies

```bash
go get github.com/gorilla/mux
go mod tidy
```

## Architecture

See [`../docs/architecture.md`](../docs/architecture.md) and [`CLAUDE.md`](./CLAUDE.md) for the full architecture guide and recipes for adding new features.

**Pattern:** CQRS (commands + queries) via mediator → DbContext (unit of work) → repository interfaces → postgres or in-memory implementations.

**Never** add business logic to HTTP handlers. Handlers parse requests, call the mediator, and write responses.
