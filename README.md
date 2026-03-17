# Hivetrack

Lean, self-hosted task planning for high-performing software development teams.

A Jira alternative that helps the people doing the work — not the people watching it. Fast, opinionated, open source.

## Quick start

```bash
# 1. Start postgres
just dev-deps

# 2. Start backend + frontend
just dev
```

Backend: `http://localhost:8080`
Frontend: `http://localhost:5173`

You'll need an OIDC provider for auth. [Keyline](https://github.com/the127/keyline) is the recommended companion — see [docs/architecture.md](docs/architecture.md) for setup.

## Repository layout

```
hivetrack/       Go backend (HTTP API)
hivetrack-ui/    Vue 3 frontend
docs/            Architecture and design documentation
scripts/         Git hooks and tooling scripts
justfile         Task runner — run `just` to see all recipes
```

## Development

All tasks are in the `justfile`. Run `just` to list them.

```bash
just dev              # start everything locally
just test             # run all tests (no DB needed)
just check            # lint + tests — run before pushing
just install-hooks    # install git pre-push hook
```

## Documentation

| Doc | Contents |
|---|---|
| [docs/architecture.md](docs/architecture.md) | System architecture, auth, data model, API |
| [docs/domain-model.md](docs/domain-model.md) | Full entity definitions and relationships |
| [docs/engineering-principles.md](docs/engineering-principles.md) | TDD, patterns, coding conventions |
| [docs/api-and-ai.md](docs/api-and-ai.md) | API-first design, webhooks, AI integration |
| [hivetrack/CLAUDE.md](hivetrack/CLAUDE.md) | Backend recipes for adding features |
| [hivetrack-ui/CLAUDE.md](hivetrack-ui/CLAUDE.md) | Frontend recipes for adding views |

## Stack

**Backend:** Go 1.24, PostgreSQL, CQRS + mediator, IoC DI
**Frontend:** Vue 3, Vite, TanStack Query, Tailwind CSS v4
**Auth:** OIDC (Keyline / Keycloak / Authentik)
