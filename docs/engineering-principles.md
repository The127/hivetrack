# Hivetrack Engineering Principles

These principles are non-negotiable. They inform every implementation decision. When in doubt, come back here.

---

## Testability

Testability is not a feature added after the fact — it is a structural property of the architecture. The patterns below exist in large part because they make testing natural.

### The IoC container enables full unit testability
Every dependency is injected. Nothing is a global singleton that cannot be replaced. A command handler that needs a repository gets the repository injected — in tests, inject a fake.

```go
// In production
dc.RegisterSingleton(func(dp ioc.DependencyProvider) repositories.IssueRepository {
    return postgres.NewIssueRepository(dp.Get[*sql.DB]())
})

// In tests
dc.RegisterSingleton(func(dp ioc.DependencyProvider) repositories.IssueRepository {
    return inmemory.NewIssueRepository()
})
```

### Repository interfaces enable in-memory fakes
Every repository has a Go interface. There are two implementations:
- `postgres/` — production
- `inmemory/` — tests

In-memory implementations are simple maps. They do not mock — they implement the full interface with real behavior. This means tests exercise real logic without a database.

### Commands and queries are pure input/output
A command is a struct. A query is a struct. Their handlers are functions from input to output (plus side effects via injected dependencies). This makes them trivially testable: construct the input, call the handler, assert the output.

### Behaviors are tested independently
Each behavior (AuthBehavior, AuditBehavior) is tested in isolation with a mock next handler. This keeps behavior tests fast and focused.

### Test layers
```
Unit tests      → Command/query handlers with in-memory repos. Fast. No I/O.
Integration     → HTTP handlers against a real test database (Docker). Slower.
E2E             → Frontend + backend against a full stack. Slowest, fewest.
```

The pyramid should be wide at the bottom. Most tests are unit tests.

---

## TDD

New functionality is written test-first. The cycle:

1. **Red** — write a failing test that describes the desired behavior
2. **Green** — write the minimum code to make it pass
3. **Refactor** — clean up while keeping tests green

TDD applies to:
- Command handlers (the core logic of the application)
- Query handlers
- Behaviors
- Domain utility functions

TDD does not mean writing tests for HTTP handler wiring or DI setup — those are structural and tested by integration tests.

**A test is a design tool.** If a handler is hard to test, it is hard to understand. Restructure until it is easy to test. Difficulty testing is a signal, not an excuse to skip tests.

---

## Maintainability

### Explicit over implicit
Prefer explicit dependency injection over package-level globals. Prefer explicit error returns over panics. Prefer explicit config fields over environment variable guessing.

### Small, focused files
No 1000-line files. Each command handler lives in its own file. Each query handler lives in its own file. A file that is growing is a signal to decompose.

### Names that tell the truth
`CreateIssueCommand`, `GetIssuesByProjectQuery`, `AuthBehavior`. Names that describe what a thing does, not what it is. If you need to read the implementation to understand the name, the name is wrong.

### No magic
No reflection-based auto-wiring of routes or handlers. Routes are registered explicitly. Handlers are registered explicitly. The startup sequence in `main.go` is readable by a new developer.

### Migrations are immutable history
Database migrations are numbered SQL files (`001_initial.sql`, `002_add_sprints.sql`). Once merged, they are never modified. To change a schema, add a new migration. This means the migration history is a reliable audit of the database schema's evolution.

### The outbox pattern keeps transactions honest
A command that sends an email does NOT call the email client directly. It writes to the outbox in the same transaction as the data change. If the transaction rolls back, the email is not sent. This eliminates an entire class of inconsistency bugs.

---

## Usability

Usability is a first-class engineering concern. It is not the designer's problem alone. Specific commitments:

### The board is the default view
When you open a project, you see the board. Not a dashboard of charts. Not a list of settings. The board. This is where work happens.

### Command palette (Cmd+K)
Every action in the application is reachable via the command palette. No need to know a shortcut in advance — type "assign", "move to sprint", "create issue", "go to project" and it works. The command palette is:
- Always available via `Cmd+K` / `Ctrl+K`
- Context-aware (shows relevant actions for the current view)
- Searchable (fuzzy match on action names and recent items)
- The primary discovery mechanism for keyboard power users

The command palette is not a search box — it's an action dispatcher. Search lives at `/`.

### Keyboard-first navigation
Every common action has a keyboard shortcut. `C` creates an issue. `B` goes to the board. `K` / `J` navigate between issues in a list. `/` focuses search. `Cmd+K` opens command palette. These are inspired by GitHub's keyboard shortcuts which set the standard for developer tools.

### Optimistic updates
The UI updates immediately on user action. Server confirmation happens in the background. If the server rejects the action, the UI reverts with an error message. Users should never wait for a spinner to drag a card.

### No dead ends
Every empty state has a call to action. An empty backlog says "Add the first issue." An empty sprint says "Pull issues from the backlog." The UI guides the user toward the next action.

### Mobile is not the primary target, but it must not be broken
The primary use case is a developer at a laptop. The UI is not designed mobile-first. However, it must be usable on mobile for quick status checks. No horizontal scrolling. No tiny tap targets. Use responsive layouts.

---

## Ease of Installation

Installation complexity is a tax on every deployment. Keep it low.

### Minimum viable deployment
```
docker run -e DATABASE_URL=... -e OIDC_AUTHORITY=... -p 8080:8080 hivetrack/hivetrack
```

One container. One database. One OIDC provider. Done.

### The binary serves the frontend
The Go binary embeds the compiled frontend assets. There is no separate static file server needed. `GET /` serves the Vue app. `GET /api/v1/...` serves the API. This eliminates an entire category of deployment complexity.

### Database migrations run automatically
On startup, the binary checks for pending migrations and applies them. No manual `migrate` step. No migration CLI to install. The database is always at the right schema version.

### Sensible defaults
A minimal `config.yaml` is a few lines. Every config field has a documented default. A developer can get a local instance running in under 5 minutes.

### Docker Compose for local development
The repository includes a `docker-compose.yaml` that starts Postgres and (optionally) a local Keyline instance. `docker compose up && go run ./cmd/hivetrack` is the full local setup.

---

## Architecture Patterns (from Keyline/Dockyard)

These patterns are battle-tested in the existing codebase. Replicate them faithfully.

### CQRS + Mediator
```
Write path:  HTTP Handler → Command{} → Mediator.Send() → []Behavior → CommandHandler → DB
Read path:   HTTP Handler → Query{}  → Mediator.Query() → QueryHandler → DB → Response
```

Commands and queries are structs. Handlers are registered with the mediator at startup. The mediator dispatches to the correct handler at runtime. Cross-cutting concerns (auth, audit, validation) are Behaviors registered in a pipeline.

**Why:** Eliminates handler bloat. Auth and audit are enforced by the pipeline, not sprinkled through business logic. Commands and queries are independently testable without the HTTP layer.

### IoC Dependency Injection
```go
// Register (startup, once)
dc.RegisterSingleton(...)
dc.RegisterScoped(...)

// Resolve (at request time, in handler)
repo := ioc.Get[repositories.IssueRepository](dp)
```

Three lifetimes:
- `Singleton` — created once, shared for the lifetime of the process. Use for: DB connection, config, OIDC verifier, email client, mediator.
- `Scoped` — created once per HTTP request, cleaned up after. Use for: DB transaction context, current user, request logger.
- `Transient` — new instance each call. Rarely needed.

**Why:** Testable. Replaces any dependency without changing business logic. Makes startup sequence explicit and readable.

### DbContext — Unit of Work
The `DbContext` wraps a database transaction and exposes typed repository accessors. Command handlers receive a `DbContext`, not individual repositories. This ensures all reads and writes within a single command are part of the same transaction.

```go
type DbContext interface {
    Issues()      IssueRepository
    Projects()    ProjectRepository
    Sprints()     SprintRepository
    Milestones()  MilestoneRepository
    Labels()      LabelRepository
    Users()       UserRepository
    Watchers()    IssueWatcherRepository
    Outbox()      OutboxRepository
    // ... all aggregate roots

    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

The `DbContext` is created per-request (scoped lifetime). The IoC container creates it at the start of the request and disposes it (rollback if not committed) at the end.

```go
// In a command handler
func (h *CreateIssueHandler) Handle(ctx context.Context, cmd CreateIssueCommand) (CreateIssueResult, error) {
    db := ioc.Get[DbContext](ctx)

    issue := models.NewIssue(cmd)
    if err := db.Issues().Insert(ctx, issue); err != nil {
        return CreateIssueResult{}, err
    }
    if err := db.Outbox().Enqueue(ctx, "notify_watchers", ...); err != nil {
        return CreateIssueResult{}, err
    }
    if err := db.Commit(ctx); err != nil {
        return CreateIssueResult{}, err
    }
    return CreateIssueResult{IssueID: issue.ID}, nil
}
```

**Why:** Inspired by .NET Entity Framework's `DbContext`. Makes the unit of work explicit and natural. All side effects (outbox writes) happen in the same transaction as the data change. No risk of partial writes. The handler does not know or care whether it's talking to Postgres or an in-memory store.

### Repository Pattern with Filters
```go
// Define a filter
type IssueFilter struct {
    projectID  *uuid.UUID
    status     *string
    onHold     *bool
    assigneeID *uuid.UUID
    sprintID   *uuid.UUID
    search     *string
}

func (f *IssueFilter) ByProject(id uuid.UUID) *IssueFilter { f.projectID = &id; return f }
func (f *IssueFilter) OnHoldOnly() *IssueFilter            { v := true; f.onHold = &v; return f }
// etc.

// Use in a query handler
issues, err := repo.Issues().List(ctx, NewIssueFilter().ByProject(id).OnHoldOnly())
```

Repository interfaces are in `internal/repositories/interfaces.go`. PostgreSQL implementations are in `internal/repositories/postgres/`. In-memory implementations are in `internal/repositories/inmemory/`.

**Why:** Decouples query logic from persistence. SQL stays in one place. Filter objects compose cleanly. In-memory implementations make unit tests fast.

### Behavior Pipeline
```go
type Behavior[TCommand any] interface {
    Handle(ctx context.Context, cmd TCommand, next func(context.Context) error) error
}

// AuthBehavior: checks permission before calling next
// AuditBehavior: writes audit log after calling next
// ValidationBehavior: validates struct before calling next
```

Behaviors are registered per command type or globally. They run in registration order. `next()` calls the command handler (or the next behavior in the chain).

**Why:** Permission checks cannot be forgotten — they are enforced at the pipeline level. Adding audit logging to a new command requires one line of registration, not modifying the handler.

### Outbox Pattern
```go
// In a command handler (same transaction as data change)
err = outbox.Enqueue(ctx, tx, "send_email", SendConfirmationEmailPayload{
    To:      issue.CustomerEmail,
    Token:   issue.CustomerToken,
    IssueID: issue.ID,
})

// Background goroutine polls outbox and delivers
```

The outbox table is polled by a background goroutine. Failed deliveries are retried with backoff. Delivered messages are marked, not deleted (audit trail).

**Why:** Guarantees that side effects (emails, webhook calls) are consistent with data changes. A crash after committing but before sending does not cause lost emails — the outbox will deliver on the next poll.

### Structured Configuration
```go
type Config struct {
    Server   ServerConfig   `koanf:"server"`
    Database DatabaseConfig `koanf:"database"`
    OIDC     OIDCConfig     `koanf:"oidc"`
    Email    EmailConfig    `koanf:"email"`
    AI       AIConfig       `koanf:"ai"`
}
```

Loaded from YAML, overridden by environment variables (`HIVETRACK_SERVER_PORT=9090`). Validated at startup — missing required fields fail fast with a clear error message. Config is injected as a singleton.

**Why:** One canonical source of config. Env var overrides make container deployments easy. Fail-fast validation prevents mysterious runtime errors.

### Dynamic OIDC Config Endpoint
The frontend does NOT have the OIDC authority URL baked in at build time. Instead:

```
Frontend starts → fetches GET /api/v1/auth/oidc-config → { authority, client_id }
                → initializes oidc-client-ts with fetched config
```

**Why:** The frontend can be deployed as static assets on a CDN. Different deployments (dev, staging, prod) can point to different OIDC providers without rebuilding the frontend.

---

## Architecture Tests

Architecture tests live in `hivetrack/internal/architecture/architecture_test.go`. They run with `go test` like any other test — no special tooling required. They enforce layer boundaries automatically as the codebase grows.

**What they test:**

| Test | What it enforces |
|---|---|
| `TestArchitecturalConstraints` | Layer rules (see table below) |
| `TestNoCycles` | No import cycles within internal packages |
| `TestSetupIsCompositionRoot` | Only `setup` may import both handlers and postgres implementations |

**Layer rules enforced:**

```
handlers       → may NOT import repositories/postgres, repositories/inmemory, setup
commands       → may NOT import queries, handlers, middlewares, setup, repo implementations
queries        → may NOT import commands, handlers, middlewares, setup, repo implementations
behaviors      → may NOT import handlers, setup
repositories/* → may NOT import handlers, commands, queries, behaviors, middlewares, setup
models         → may NOT import any other internal package
email          → may NOT import handlers, commands, queries, behaviors, repositories, setup
authentication → may NOT import handlers, commands, queries, behaviors, repositories, setup
config         → may NOT import any other internal package
```

**Run architecture tests:**
```bash
just test-arch
```

**When to add a new rule:** When you find yourself enforcing a layer boundary in code review more than once, encode it as an architecture test instead.

---

## Task Runner (just)

All common development tasks are documented in `justfile` at the repo root. Use `just` to discover and run them.

```bash
just              # list all available recipes
just run          # start the backend
just ui-dev       # start the frontend dev server
just dev          # start the full local stack
just test         # run all backend tests (unit + arch, no DB)
just test-arch    # run only architecture constraint tests
just test-unit    # run only unit tests (fast, no DB)
just test-integration  # run integration tests (requires postgres)
just check        # lint + test-arch + test-unit (pre-commit verification)
just release      # build frontend and embed into backend binary
just db-shell     # open psql shell to local dev database
just db-reset     # destroy and recreate local database
```

The Justfile is the canonical reference for how to run, build, and test the project. If you need to run something that is not in the Justfile, add it there before documenting it elsewhere.

---

## What We Explicitly Do Not Do

- **No ORM.** Hand-crafted SQL via `go-sqlbuilder`. ORMs hide complexity and generate bad queries. We own our SQL.
- **No global state.** No package-level `var db *sql.DB`. Everything through DI.
- **No custom authentication.** OIDC only for internal users. Email tokens for external. No username/password.
- **No custom workflow engine.** Fixed status sets. No drag-and-drop workflow designers.
- **No built-in reporting.** Charts, burndowns, velocity tracking — these are orthogonal. Build them on the API.
- **No microservices.** One binary. Complexity grows linearly, not exponentially.
- **No runtime reflection magic.** Routes are registered explicitly. The codebase is grep-able.
