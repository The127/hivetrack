# Hivetrack Backend — Claude Instructions

Go HTTP API. Read the root `CLAUDE.md` and `docs/` before making changes here.

**Module:** `github.com/the127/hivetrack`
**Entry point:** `cmd/hivetrack/main.go`
**Run:** `just run` from repo root, or `go run ./cmd/hivetrack` from this directory.

---

## Adding a new feature (the full recipe)

Every feature follows the same pattern. Example: "close a sprint".

### 1. Migration (if schema changes needed)
Add `migrations/NNN_description.sql`. Migrations run automatically on startup.

### 2. Repository method (if new data access needed)
Add method to the interface in `internal/repositories/interfaces.go`:
```go
type SprintRepository interface {
    // existing methods...
    Close(ctx context.Context, id uuid.UUID) error
}
```
Implement in both:
- `internal/repositories/postgres/sprint_repository.go`
- `internal/repositories/inmemory/sprint_repository.go`

### 3. Command (write operation)
Create `internal/commands/close_sprint.go`:
```go
package commands

type CloseSprintCommand struct {
    SprintID  uuid.UUID
    ProjectID uuid.UUID
    ActorID   uuid.UUID
}

type CloseSprintResult struct {
    // minimal — just what callers need
}
```

Create `internal/commands/close_sprint_handler.go`:
```go
package commands

type CloseSprintHandler struct{}

func (h *CloseSprintHandler) Handle(ctx context.Context, cmd CloseSprintCommand) (CloseSprintResult, error) {
    db := ioc.Get[DbContext](ctx)

    sprint, err := db.Sprints().GetByID(ctx, cmd.SprintID)
    if err != nil {
        return CloseSprintResult{}, err
    }
    // ... business logic ...

    if err := db.Sprints().Close(ctx, sprint.ID); err != nil {
        return CloseSprintResult{}, err
    }
    return CloseSprintResult{}, db.Commit(ctx)
}
```

Write the test FIRST (`close_sprint_handler_test.go`) — see TDD section below.

### 4. Register handler with mediator
In `internal/setup/commands.go`:
```go
mediatr.RegisterHandler[commands.CloseSprintCommand, commands.CloseSprintResult](
    dp.Get[*commands.CloseSprintHandler](),
)
```

### 5. Register behavior (auth check)
In `internal/setup/behaviors.go`:
```go
mediatr.RegisterBehavior[commands.CloseSprintCommand](
    behaviors.NewAuthBehavior(permissions.CloseSprint),
)
```

### 6. HTTP handler (thin — just parse, delegate, respond)
In `internal/handlers/sprint_handler.go`:
```go
func (h *SprintHandler) CloseSprint(w http.ResponseWriter, r *http.Request) {
    sprintID := // parse from vars
    cmd := commands.CloseSprintCommand{
        SprintID:  sprintID,
        ProjectID: // from vars,
        ActorID:   middleware.CurrentUserID(r.Context()),
    }
    result, err := h.mediator.Send(r.Context(), cmd)
    if err != nil {
        respondError(w, err)
        return
    }
    respondJSON(w, http.StatusOK, result)
}
```

### 7. Register route
In `internal/server/server.go`:
```go
r.Handle("POST /api/v1/projects/{slug}/sprints/{id}/close",
    authMiddleware(http.HandlerFunc(sprintHandler.CloseSprint)))
```

---

## DbContext — unit of work

All repository access in command handlers goes through `DbContext`. Never inject repositories directly into handlers.

```go
db := ioc.Get[DbContext](r.Context())  // in HTTP handler (scoped to request)
// or
db := ioc.Get[DbContext](ctx)           // in command handler
```

The DbContext wraps a Postgres transaction. Call `db.Commit(ctx)` at the end of a successful command. On error, the scoped DI will roll back automatically.

Outbox writes go through `db.Outbox().Enqueue(...)` — same transaction, same commit.

---

## TDD

Write the test before the implementation. Command handler tests use in-memory repositories:

```go
func TestCloseSprintHandler_Success(t *testing.T) {
    // Arrange
    sprint := fixtures.ActiveSprint(projectID)
    db := inmemory.NewDbContext()
    db.Sprints().Insert(ctx, sprint)

    handler := commands.NewCloseSprintHandler()
    ctx := ioc.WithValue(ctx, db)

    // Act
    _, err := handler.Handle(ctx, commands.CloseSprintCommand{
        SprintID: sprint.ID,
    })

    // Assert
    require.NoError(t, err)
    updated, _ := db.Sprints().GetByID(ctx, sprint.ID)
    assert.Equal(t, models.SprintStatusCompleted, updated.Status)
}
```

---

## File layout

```
internal/
├── architecture/       # Architecture constraint tests — keep updated
├── authentication/     # JWT validation, OIDC verifier, current user extraction
├── behaviors/          # Mediator pipeline behaviors (auth, audit, validation)
├── commands/           # One file per command, one file per handler
├── config/             # Config struct, loading via koanf
├── database/           # Connection, migration runner
├── email/              # Templates, SMTP sender
├── events/             # Domain event types, outbox delivery
├── handlers/           # HTTP handlers (thin controllers)
├── middlewares/        # HTTP middleware (auth, cors, recovery, request logging)
├── models/             # Domain types — no imports from other internal packages
├── queries/            # One file per query, one file per handler
├── repositories/
│   ├── interfaces.go   # All repository interfaces + DbContext interface
│   ├── inmemory/       # In-memory implementations for tests
│   └── postgres/       # PostgreSQL implementations
├── server/             # Route registration, HTTP server setup
└── setup/              # IoC wiring (one file per concern: commands, queries, repos, services)
migrations/             # NNN_description.sql — immutable once merged
cmd/hivetrack/main.go   # Startup: load config → wire DI → run migrations → start server
```

---

## Key rules

- `models` package imports nothing from `internal/`
- Commands never import queries; queries never import commands
- Handlers never import `repositories/postgres` or `repositories/inmemory`
- `setup` is the only package that imports both handlers and postgres implementations
- Run `just test-arch` to verify — it will catch violations

---

## Error handling

Return errors up the call stack. The HTTP handler converts them:
```go
// Define typed errors in models/errors.go
var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

// In handlers/respond.go
func respondError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, models.ErrNotFound):
        respondJSON(w, http.StatusNotFound, ...)
    case errors.Is(err, models.ErrForbidden):
        respondJSON(w, http.StatusForbidden, ...)
    default:
        respondJSON(w, http.StatusInternalServerError, ...)
    }
}
```

Never log and return. Either log (top-level) or return. Not both.
