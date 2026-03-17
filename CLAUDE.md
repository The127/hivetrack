# Hivetrack — Claude Instructions

Hivetrack is a lean, self-hosted task planning tool for high-performing software development teams. It is a Jira alternative built on strong opinions: fewer features, zero management bloat, fast, open source, self-hostable.

---

## Tech Stack

### Backend (`hivetrack/`)
- **Language**: Go 1.24+
- **Router**: Gorilla Mux
- **Database**: PostgreSQL (hand-crafted SQL via `huandu/go-sqlbuilder`)
- **CQRS Mediator**: `github.com/The127/mediatr`
- **DI Container**: `github.com/The127/ioc`
- **Config**: Koanf (YAML + env var overrides)
- **Logging**: Uber Zap (structured)
- **Auth**: OIDC (JWT validation middleware); email tokens for customer portal
- **Email**: SMTP (configurable provider)

### Frontend (`hivetrack-ui/`)
- **Framework**: Vue 3 (Composition API only, no Options API)
- **Build**: Vite
- **State**: TanStack Query (server state), no Vuex/Pinia
- **Router**: Vue Router 4 with auth guards in route `meta`
- **Styling**: Tailwind CSS v4
- **Auth**: `oidc-client-ts`
- **Icons**: Lucide Vue Next
- **HTTP**: Native `fetch` API via `apiFetch` composable wrapper (no Axios)
- **Components**: Custom CVA-based components

---

## Architecture Patterns

### CQRS + Mediator
All operations go through the mediator. Write operations are **Commands**, read operations are **Queries**. Never mix them.

```
Command → []Behavior (auth, audit, validation) → Handler → (optional) Event
Query   → Handler → Response
```

- Commands return minimal data (IDs, success). Queries return full response shapes.
- Cross-cutting concerns (permission checks, audit logging) live in **Behaviors**, not handlers.
- The mediator is registered in the IoC container and injected into HTTP handlers.

### IoC Dependency Injection
Use `github.com/The127/ioc` with three lifetimes:
- `Singleton` — DB connection, config, email client, OIDC verifier
- `Scoped` — per-request (DB transaction context, current user)
- `Transient` — new each time (rarely needed)

Build the provider at startup and treat it as immutable after that. Never pass the provider itself around — inject concrete dependencies.

### Repository Pattern
Each aggregate root has a repository interface with a corresponding PostgreSQL implementation. Use **Filter objects** for composable queries — never build query strings ad-hoc in handlers or command handlers.

```go
filter := repositories.NewIssueFilter().
    ByProjectID(projectID).
    ByStatus(StatusInProgress).
    WithAssignee(userID)
issues, err := repo.Issues().List(ctx, filter)
```

### Behaviors (Middleware for Commands)
Behaviors wrap command handlers in a pipeline. Register them per-command or globally. Standard behaviors:
- `AuthBehavior` — verifies caller has required permission
- `AuditBehavior` — writes audit log entry
- `ValidationBehavior` — validates command fields before handler runs

### Outbox Pattern
Commands that produce side effects (emails, events) write to an `outbox_messages` table in the same transaction. A background job delivers them. This ensures consistency — if the transaction rolls back, the email is never sent.

---

## Design Decisions

**1. No multi-tenancy.**
One deployment = one organization. Simplifies auth, URLs, and the entire data model. This is a self-hosted tool. URL prefix is `/api/v1/...` not `/api/v1/tenants/{t}/...`.

**2. On-hold is orthogonal to status.**
`status` represents workflow position. `on_hold` is a separate boolean dimension with a reason and timestamp. An issue can be `in_progress` AND `waiting_on_customer` simultaneously. These never mix into a single enum. The board renders on-hold issues with a visual indicator regardless of status. When a blocking issue resolves, `blocked_by_issue` hold is automatically cleared.

**3. Project archetypes, not custom workflows.**
Projects have one of two archetypes: `software` or `support`. The archetype determines default statuses, available features, and visibility model. No custom workflows. Teams cannot misconfigure what they don't need.

- `software` archetype: team-only, statuses = `backlog | todo | in_progress | in_review | done | cancelled`, features = sprints + backlog + board + milestones
- `support` archetype: public-facing, statuses = `open | in_progress | resolved | closed`, features = email submission + token tracking + link to software issues

**4. Fixed status sets per archetype.**
No custom statuses. This is an intentional constraint. Teams waste enormous time debating workflow states. Two sets of statuses cover 95% of real work.

**5. Epic → Task hierarchy, subtasks as checklists.**
Two issue types: `epic` and `task`. An epic groups tasks (parent/child, one level deep). A task has a `[]ChecklistItem` field for subtasks — these are not entities, they live in the task. This keeps the query model simple and avoids recursive issue trees.

**6. Simple multi-assignee.**
Issues have `[]assignees` (many-to-many with users). No "assignee group" entity. One person = one entry. Team-owned = multiple entries. This is the minimum that solves the real use case.

**7. Restricted issue visibility.**
An issue can be `visibility: normal | restricted`. Restricted issues are only visible to project admins and explicitly listed users. This covers sensitive HR/security tickets. Project membership does not grant access to restricted issues.

**8. Customer portal via email tokens.**
External customers submit issues to `support` archetype projects via a public form (no account needed). On submission, an email is sent with a magic token URL. The token allows read-only access to that specific issue. No OIDC, no accounts for external users. Internal team links customer issues to software issues.

**9. Roles are per-project, with one global override.**
- `instance_admin` — global, manages users, instance config, sees everything
- `project_admin` — manages a specific project, sees all issues including restricted
- `project_member` — creates and works on issues, normal visibility rules apply
- `viewer` — read-only on a specific project, normal visibility rules apply

A user can have different roles on different projects. `instance_admin` supersedes project roles.

**10. Milestones for long-term, sprints for short-term.**
Sprints are time-boxed (start/end date + goal text). Issues are in exactly one active sprint or in the backlog. Milestones are target-date goals (like GitHub milestones). Issues belong to one milestone. Sprints and milestones are independent — a sprint's issues may contribute to one or multiple milestones.

**11. OIDC for everything — no custom API keys.**
Internal users authenticate via OIDC Authorization Code + PKCE. Machine callers (CI, bots, AI agents) authenticate via OAuth 2.0 Client Credentials grant (service accounts in the OIDC provider). Hivetrack validates JWTs — one auth path for all callers, zero custom credential management. OIDC provider config is fetched at runtime from `GET /api/v1/auth/oidc-config`.

**12. Triage inbox for low-friction capture.**
Issues with `triaged = false` are in the inbox — visible but not on the board. Quick-capture creates issues with title only, landing in triage. External integrations (CI, monitoring, git webhooks) also land here. Triaging means placing the issue into the workflow.

**13. No mandatory fields except title.**
Creating an issue requires only a `title`. Everything else — priority, assignee, sprint, labels, estimate — is optional and can be set later. Mandatory fields kill the habit of capturing work.

**14. T-shirt sizing for estimates, never story points.**
Estimate values: `none | xs | s | m | l | xl`. Low ceremony, high adoption. Story points are not supported.

**15. DbContext as the unit of work.**
Command handlers receive a `DbContext` (not individual repositories). The DbContext wraps a transaction and exposes all repositories. All writes in a command — including outbox enqueues — happen in the same transaction. Inspired by .NET Entity Framework's DbContext pattern. See `docs/engineering-principles.md`.

**16. Sprint carry count surfaces stuck work.**
`sprint_carry_count` on Issue increments each time an issue is moved to a new sprint without reaching a terminal state. The board shows a badge after 2+ carries. No report, just a signal visible in context.

**17. Git integration is webhook-based and platform-agnostic.**
Inbound webhooks from any git host (GitHub, GitLab, Gitea, etc.) are parsed for issue references. No platform-specific integrations, no OAuth scopes to manage. Auto-transitions (PR merged → issue done) are opt-in per project.

---

## Domain Model Summary

```
User
├── id, email, display_name, avatar_url
└── [role assignments per project]

Project
├── id, slug, name, description, archetype (software|support)
├── created_by, created_at
└── [members with roles]

Sprint (software projects only)
├── id, project_id, name, goal, start_date, end_date, status (planning|active|completed)

Milestone
├── id, project_id, title, description, target_date, closed_at

Label
├── id, project_id, name, color (hex)

Issue
├── id, project_id, number (human-readable per-project), title, description (markdown)
├── type: epic | task
├── status: [per archetype]
├── hold: null | { reason: waiting_on_customer|waiting_on_external|blocked_by_issue, since, note }
├── priority: none | low | medium | high | critical
├── assignees: []User
├── reporter: User
├── parent_id: ?Issue (task → epic only)
├── milestone_id: ?Milestone
├── sprint_id: ?Sprint  (null = backlog)
├── labels: []Label
├── visibility: normal | restricted
├── restricted_viewers: []User  (only when visibility=restricted)
├── checklist: []{ id, text, done }  (tasks only, not entities)
├── customer_email: ?string  (support archetype, external submissions)
├── customer_token: ?uuid  (support archetype, for token-based tracking)
└── links: []{ type: blocks|is_blocked_by|duplicates|relates_to, target_issue_id }

Comment
├── id, issue_id, author: User|{email, name} (external), body (markdown), created_at

AuditLog
├── id, entity_type, entity_id, action, actor, diff (json), created_at

OutboxMessage
├── id, type, payload (json), status, created_at, delivered_at
```

---

## File Structure

### Backend
```
hivetrack/
├── cmd/hivetrack/main.go
├── internal/
│   ├── setup/          # IoC container wiring (one file per concern)
│   ├── server/         # Route registration, HTTP server
│   ├── config/         # Config struct + loading
│   ├── commands/       # Write operations + handlers
│   ├── queries/        # Read operations + handlers
│   ├── handlers/       # HTTP handlers (thin, delegate to mediator)
│   ├── repositories/
│   │   ├── interfaces.go
│   │   └── postgres/
│   ├── database/       # Migrations, connection
│   ├── middlewares/    # HTTP middleware (auth, cors, recovery, logging)
│   ├── authentication/ # JWT validation, OIDC verifier, token auth
│   ├── behaviors/      # Mediator behaviors (auth, audit, validation)
│   ├── events/         # Domain events + outbox delivery
│   ├── email/          # Email templates + SMTP sender
│   └── models/         # Shared domain types
├── migrations/         # SQL migration files (numbered)
├── go.mod
└── config.yaml
```

### Frontend
```
hivetrack-ui/
├── src/
│   ├── main.js
│   ├── App.vue
│   ├── router/index.js       # Routes with meta (requiresAuth, layout, breadcrumb)
│   ├── composables/
│   │   ├── useAuth.js        # OIDC user manager
│   │   ├── useCurrentUser.js # Current user query
│   │   └── useApi.js         # Axios instance with auth headers
│   ├── api/                  # Per-resource API functions (called by TanStack Query)
│   │   ├── projects.js
│   │   ├── issues.js
│   │   ├── sprints.js
│   │   └── ...
│   ├── views/                # Page components (one per route)
│   ├── components/           # Reusable UI components
│   │   ├── ui/               # Base components (Button, Badge, Input, etc.)
│   │   └── [feature]/        # Feature-specific components
│   ├── layouts/              # Layout components (main, minimal, public)
│   └── style.css             # Global styles + Tailwind imports
├── index.html
├── vite.config.js
└── package.json
```

---

## API Conventions

- Base: `/api/v1/`
- Resources: plural nouns (`/projects`, `/issues`, `/sprints`)
- Nested: `/projects/{slug}/issues`, `/projects/{slug}/sprints`
- Customer portal: `/public/support/{project-slug}/...` (no auth)
- Auth config: `GET /api/v1/auth/oidc-config`
- Error shape: `{ "errors": [{ "code": "...", "message": "...", "field": "..." }] }`
- Pagination: `{ "items": [], "total": N, "limit": N, "offset": N }`

## Documentation

Full design docs live in `docs/`:
- `docs/architecture.md` — system architecture, auth flows, deployment model
- `docs/domain-model.md` — full entity definitions, SQL indexes, status values
- `docs/api-and-ai.md` — API-first design, API keys, webhooks, AI integration
- `docs/engineering-principles.md` — TDD, testability, architecture patterns (CQRS, IoC, repository), maintainability, usability, ease of installation

**Read `docs/engineering-principles.md` before writing any code.**

---

## Development

All tasks are documented in `justfile` at the repo root. Use `just` as the entry point for everything.

```bash
just              # list all recipes
just dev          # start full local stack (postgres + backend + frontend)
just run          # backend only
just ui-dev       # frontend dev server only
just test         # unit tests + architecture tests (no DB)
just test-arch    # architecture constraint tests only
just test-unit    # unit tests only (fast, in-memory repos)
just test-integration  # integration tests (requires postgres)
just check        # lint + test-arch + test-unit (run before committing)
just release      # production build (frontend embedded in binary)
just db-shell     # psql shell to local database
```

Config is loaded from `config.yaml` with env var overrides using `HIVETRACK_` prefix (e.g. `HIVETRACK_DATABASE_URL`).

### Architecture Tests
`hivetrack/internal/architecture/architecture_test.go` enforces layer boundaries automatically. These run with `just test-arch`. When adding a new layer or package, update the architecture tests to encode its allowed dependencies.

### Testing
New functionality is written TDD (test first). Command and query handlers use in-memory repository fakes — no database needed for unit tests. See `docs/engineering-principles.md` for the full approach.
