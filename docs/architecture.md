# Hivetrack Architecture

## Vision

Hivetrack is a task planning tool built for self-organizing software development teams who want to get things done — not manage stakeholders. It is deliberately opinionated, deliberately lean, and deliberately not trying to be Jira.

The guiding question for every feature decision: *does this help the people doing the work, or does it help the people watching the work?* If the latter, it does not belong in Hivetrack.

---

## Core Principles

- **No bloat.** If a feature exists primarily for reporting, compliance theatre, or enterprise audit trails — it does not ship in core.
- **Opinionated.** Fixed status sets, fixed issue hierarchy, two project archetypes. Teams should not spend time configuring their tools; they should spend time using them.
- **Fast by default.** Both in UI performance and in time-to-value for a new team. Setup should take minutes.
- **Self-hostable.** One binary + one Postgres database + one OIDC provider. No cloud dependencies.
- **Reporting is orthogonal.** Hivetrack exposes a clean API. Reporting tools can be built on top. Hivetrack itself does not become a reporting tool.
- **Customer portal is the same coin.** Customer-facing support and internal development are two sides of the same issue-tracking model. They share data, not a codebase fork.

---

## System Architecture

```
┌─────────────────────────────────────────────┐
│                  Clients                     │
│                                              │
│  hivetrack-ui (Vue 3)   Public Support Form  │
│  OIDC-authenticated     Email token access   │
└──────────────┬──────────────────┬────────────┘
               │                  │
               ▼                  ▼
┌─────────────────────────────────────────────┐
│            hivetrack (Go HTTP API)           │
│                                             │
│  /api/v1/...         (OIDC JWT required)    │
│  /public/support/... (no auth)              │
│  /api/v1/auth/...    (OIDC config endpoint) │
└──────────────┬──────────────────────────────┘
               │
       ┌───────┴───────┐
       ▼               ▼
  PostgreSQL        SMTP Server
  (primary store)   (email delivery)
```

The backend is a single Go binary. It serves both the authenticated API and the public customer portal endpoints. There is no microservice split. A reverse proxy (nginx/caddy) sits in front for TLS.

---

## Authentication

### Internal Users — OIDC (Authorization Code + PKCE)
Internal team members authenticate via OIDC. The flow:

1. Frontend loads OIDC config from `GET /api/v1/auth/oidc-config` (authority URL, client ID)
2. `oidc-client-ts` initiates the authorization code flow with PKCE
3. Access token (JWT) is attached to all API requests as `Authorization: Bearer <token>`
4. Backend middleware validates the JWT signature and claims on every request
5. User identity is resolved from the JWT `sub` claim; user record is created on first login

The OIDC provider is external (Keyline, Keycloak, Authentik, etc.). Hivetrack does not implement its own IdP.

**Why dynamic config fetch instead of baked-in env vars?**
Allows the frontend to be deployed as static assets independent of OIDC provider config. The backend is the single source of truth for environment-specific config.

### Service Accounts / Bots — OIDC Client Credentials
CI pipelines, bots, AI agents, and any machine caller authenticate via the **OAuth 2.0 Client Credentials grant** — a standard feature of any compliant OIDC provider (Keyline, Keycloak, Authentik).

The operator creates a service account in the OIDC provider (e.g. "ci-bot", "deploy-notifier") with a client ID and private key. The service exchanges its credentials for a JWT via the token endpoint, then calls Hivetrack with `Authorization: Bearer <token>` — exactly like a human user.

**Hivetrack has no API key concept.** There is no custom credential to manage, no `ApiKey` table, no special auth code path. The OIDC provider manages all credentials for both humans and machines. This is a deliberate simplification — one auth system, zero special cases.

The service account is added to projects in Hivetrack the same way a human is: the project admin adds its `sub` claim as a project member with the appropriate role.

### External Customers — Email Tokens
External customers do not have accounts. For `support` archetype projects:

1. Customer submits issue via public form (`POST /public/support/{project-slug}/issues`)
2. Backend creates the issue and generates a UUID `customer_token`
3. Outbox job sends a confirmation email with a link: `https://{host}/support/track/{token}`
4. Token is presented as a query/header parameter to token-protected endpoints
5. Token grants read-only access to that specific issue only

There is no session, no cookie, no account. The token is the credential. This is intentional — it is the minimum viable auth for external users, and it works with every email client.

---

## Data Model

### Users
Users are created on first OIDC login. Profile data (name, email, avatar) is synced from JWT claims. Users cannot self-register; they must authenticate via the OIDC provider first.

### Projects
A project is the primary organizational unit. Every issue belongs to exactly one project.

**Archetype: `software`**
- Visibility: internal (team members only)
- Statuses: `backlog`, `todo`, `in_progress`, `in_review`, `done`, `cancelled`
- Features: board view, backlog view, sprints, milestones, labels
- Issue types: epic, task

**Archetype: `support`**
- Visibility: public submission endpoint, internal management
- Statuses: `open`, `in_progress`, `resolved`, `closed`
- Features: public submission form, email confirmation, token tracking, link to software issues
- Issue types: task only (no epics in support projects)

A project can be archived. Archived projects are read-only.

### Issues
The central entity. Key design decisions:

**Status vs On-Hold**
Status represents *where in the workflow* an issue is. On-hold represents *why it is not moving*. These are independent dimensions and must never be collapsed into a single enum.

```
on_hold reasons:
- waiting_on_customer  — need information or action from a customer
- waiting_on_external  — blocked by a third party (vendor, API, legal, etc.)
- blocked_by_issue     — auto-set/cleared when a "blocks" link exists and the blocking issue is unresolved
```

When rendering a board, on-hold issues display with a visual hold indicator independent of their column. A filter "show me all on-hold issues across all projects" is a first-class use case.

**Issue Hierarchy**
Two levels only:
- `epic` — a grouping of related tasks. Has a title and description. No sprint assignment (epics span sprints). Can have a milestone. Cannot have a parent.
- `task` — the unit of work. Belongs to one epic or standalone. Has sprint, milestone, assignees, labels, checklist. Cannot have children (subtasks live as checklist items in the task).

A checklist item is `{ id: uuid, text: string, done: bool }` stored as JSONB in the task row. Checklist items are not queryable as independent entities. If a checklist item grows into a full task, the team creates a new task and removes the checklist item.

**Issue Numbers**
Each issue has a human-readable number scoped to the project (e.g. `HT-42`). The project slug forms the prefix. Numbers are sequential and immutable.

**Links**
Issues can be linked with a typed relationship:
- `blocks` / `is_blocked_by` (inverse pair, stored as one record with direction)
- `duplicates` / `is_duplicated_by` (inverse pair)
- `relates_to` (symmetric)

Cross-project links are allowed. This is the mechanism for connecting a customer-reported support issue to an internal software issue.

When a `blocks` link is created between issues A→B, issue B automatically receives `on_hold: { reason: blocked_by_issue }`. When issue A reaches `done` or `cancelled`, the hold on B is automatically cleared.

**Restricted Visibility**
By default, all project members can see all issues in their projects. An issue can be marked `restricted`, in which case only:
- The reporter
- Explicitly listed `restricted_viewers`
- Project admins
- Instance admins

...can see it. This covers security, HR, and sensitive operational tickets.

### Sprints
Sprints belong to `software` projects. An issue is either in a sprint or in the backlog (no sprint). A project can have one active sprint at a time, and many planned/completed sprints.

Sprint statuses: `planning`, `active`, `completed`. Only one sprint per project can be `active`.

When a sprint is completed, unfinished issues are moved to the backlog or to the next sprint (user choice).

### Milestones
Milestones are goal markers with a target date. They belong to a project. Issues reference one milestone (or none). Milestones can be closed when all their issues are done.

Milestones span sprints. A milestone represents "what we're shipping," a sprint represents "what we're doing this week."

---

## Permissions

### Role Hierarchy
```
instance_admin
  └── Supersedes all project roles. Manages users, instance config, OIDC settings.
      Sees all issues in all projects including restricted.

project_admin (per project)
  └── Manages project settings, members, and roles within the project.
      Sees all issues in the project including restricted.

project_member (per project)
  └── Creates, edits, and manages issues. Normal visibility rules apply.
      Can be assigned issues. Can comment.

viewer (per project)
  └── Read-only. Sees issues subject to normal visibility rules.
      Cannot create or modify anything.
```

A user can hold different roles in different projects simultaneously.

### Permission Checks
Permission checks live in **Behaviors** (CQRS middleware), not in handlers. A handler never directly checks permissions. Instead, commands declare their required permission, and the `AuthBehavior` enforces it before the handler runs.

---

## Triage Inbox

The triage inbox is a staging area for issues that need attention before entering the board or backlog. An issue with `triaged = false` is in the inbox — visible to the team but not cluttering the active workspace.

Issues land in triage when:
- Created via quick-capture (title only — the lowest-friction path to capturing work)
- Submitted by external integrations (CI failure webhook, monitoring alert, external API)
- Customer support submissions arrive on a `support` project (same mechanism, different view)

Triaging an issue means assigning it a status, sprint, and/or epic and flipping `triaged = true`. This can be done inline in the triage view without opening the full issue. Bulk triage (select many, assign sprint/status in one action) is a first-class operation.

**Why not just use `status = backlog`?** Because backlog implies "we know about this and have chosen not to do it yet." Triage means "we haven't even looked at this properly." They're distinct mental states. Conflating them makes backlogs noisy.

---

## Notifications

Notifications are opt-in by default. The goal is to surface what matters without creating noise that trains people to ignore everything.

### Default notification triggers (automatic, no setup needed)
- Someone `@mentions` you in a comment
- An issue assigned to you changes status
- An issue blocking yours (on-hold: `blocked_by_issue`) is resolved
- A sprint you're part of starts or is completed

### Opt-in triggers (user subscribes per issue or project)
- Any comment added to an issue you're watching
- Any field change on an issue you're watching
- New issue created in a project you're following

### Following/watching
- Reporter and all assignees are auto-added as watchers on creation
- Any user can follow/unfollow any issue they can see
- Following a project subscribes to new issues only (not all changes)

### Delivery
v1 delivers notifications in-app (notification bell) and optionally via email. Email notifications are batched (max one email per 15 minutes per user, summarising all pending notifications) — never one email per event.

---

## Personal Dashboard ("My Work")

The personal dashboard is the default landing page after login. It answers: **what should I be working on right now?**

Sections:
- **My open issues** — all issues assigned to me across all projects, grouped by project, sorted by priority
- **Watching** — issues I'm watching that have recent activity
- **Triage inbox** — issues awaiting triage across my projects
- **Recent** — last 20 issues I've viewed (persisted, survives page reload)
- **Favorites** — starred issues, projects, and saved views

**AI-powered "what's next" suggestion** (planned, not v1): given my current assignments, sprint priorities, and what's blocking others, suggest the single most valuable issue to pick up. This requires the AI integration to be enabled and enough project context to be useful.

---

## Background Jobs

The following operations run asynchronously via the outbox pattern:

- **Email delivery** — confirmation emails for customer submissions, mention notifications
- **Blocked-by auto-resolution** — when an issue is closed, scan for issues blocked by it and clear their hold

The outbox table stores pending messages. A background goroutine polls and delivers them. In the initial version this is a single-instance background goroutine (no distributed lock needed without clustering).

---

## Board Performance

The board is the most performance-sensitive view. It must feel instant with 200+ issues.

**Frontend:**
- Board columns are virtualized — only render cards in or near the viewport
- Drag-and-drop uses optimistic updates (card moves immediately, API call in background)
- Cards are minimal by default — title, assignee avatar, priority dot, label chips. Full detail on hover/expand.
- Column issue counts are always visible so users know when a column is not fully loaded

**Backend:**
- The board query is a single indexed query: `WHERE project_id = $1 AND sprint_id = $2 AND triaged = true`, returning all columns in one round trip
- No N+1 queries. Assignees, labels, and on-hold state are loaded with the issue in one query via JOIN or array aggregation
- Done/cancelled issues are excluded from the board query by default (they're in archive). The "Show done" toggle adds them back with a separate query.

---

## Git Integration

Git integration is platform-agnostic: it works via inbound webhooks, not deep API integrations. This means it works with GitHub, GitLab, Gitea, Forgejo, Bitbucket, or anything that can send webhooks.

### How it works
1. A developer creates a branch named `ht-42-fix-login-bug` or commits with message `fix: resolve login timeout HT-42`
2. The git host sends a push/PR webhook to `POST /api/v1/integrations/git`
3. Hivetrack parses the payload for issue references (`HT-\d+` pattern, configurable)
4. Matching issues get a linked git event: branch created, PR opened, PR merged, etc.

### Issue detail shows
- Linked branches
- Linked PRs with their status (open/merged/closed)
- Optionally: auto-transition trigger (PR merged → issue moves to `in_review` or `done`)

### Auto-transitions (configurable per project)
```
PR opened  → move issue to in_review  (optional)
PR merged  → move issue to done       (optional)
```
Auto-transitions are off by default. Teams opt in per project.

### Configuration
```yaml
integrations:
  git:
    enabled: true
    # Webhook secret for HMAC validation (shared with git host)
    webhook_secret: changeme
    # Regex pattern to find issue references in branch names and commit messages
    issue_pattern: "[A-Z]+-\\d+"
```

---

## API Design

### Authenticated Endpoints (`/api/v1/`)
All require `Authorization: Bearer <jwt>`.

```
GET    /api/v1/auth/oidc-config

GET    /api/v1/users/me
GET    /api/v1/users

GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/{slug}
PUT    /api/v1/projects/{slug}
DELETE /api/v1/projects/{slug}

GET    /api/v1/projects/{slug}/members
POST   /api/v1/projects/{slug}/members
DELETE /api/v1/projects/{slug}/members/{user-id}

GET    /api/v1/projects/{slug}/issues
POST   /api/v1/projects/{slug}/issues
GET    /api/v1/projects/{slug}/issues/{number}
PUT    /api/v1/projects/{slug}/issues/{number}
DELETE /api/v1/projects/{slug}/issues/{number}

GET    /api/v1/projects/{slug}/issues/{number}/comments
POST   /api/v1/projects/{slug}/issues/{number}/comments

GET    /api/v1/projects/{slug}/sprints
POST   /api/v1/projects/{slug}/sprints
GET    /api/v1/projects/{slug}/sprints/{id}
PUT    /api/v1/projects/{slug}/sprints/{id}

GET    /api/v1/projects/{slug}/milestones
POST   /api/v1/projects/{slug}/milestones
PUT    /api/v1/projects/{slug}/milestones/{id}
DELETE /api/v1/projects/{slug}/milestones/{id}

GET    /api/v1/projects/{slug}/labels
POST   /api/v1/projects/{slug}/labels
PUT    /api/v1/projects/{slug}/labels/{id}
DELETE /api/v1/projects/{slug}/labels/{id}

GET    /api/v1/search?q=...&project=...&status=...&assignee=...

GET    /api/v1/me/issues          # My Work — all issues assigned to me, cross-project
GET    /api/v1/me/watching        # Issues I'm watching with recent activity
GET    /api/v1/me/recent          # Last 20 issues I've viewed
GET    /api/v1/me/favorites       # My favorited issues, projects, views
POST   /api/v1/me/favorites
DELETE /api/v1/me/favorites/{id}

GET    /api/v1/projects/{slug}/triage         # Triage inbox for a project
POST   /api/v1/projects/{slug}/issues/{number}/triage  # Triage an issue (set triaged=true + fields)

GET    /api/v1/projects/{slug}/issues/{number}/activity  # Audit log for an issue
GET    /api/v1/projects/{slug}/issues/{number}/watchers
POST   /api/v1/projects/{slug}/issues/{number}/watch
DELETE /api/v1/projects/{slug}/issues/{number}/watch

GET    /api/v1/projects/{slug}/templates
POST   /api/v1/projects/{slug}/templates
PUT    /api/v1/projects/{slug}/templates/{id}
DELETE /api/v1/projects/{slug}/templates/{id}

GET    /api/v1/views              # Saved views (personal + shared)
POST   /api/v1/views
PUT    /api/v1/views/{id}
DELETE /api/v1/views/{id}

GET    /api/v1/projects/{slug}/webhooks
POST   /api/v1/projects/{slug}/webhooks
PUT    /api/v1/projects/{slug}/webhooks/{id}
DELETE /api/v1/projects/{slug}/webhooks/{id}

POST   /api/v1/integrations/git   # Inbound git webhook (push, PR open/merge)
```

### Public Customer Portal (`/public/support/`)
No authentication required.

```
GET  /public/support/{project-slug}          # project info (name, description only)
POST /public/support/{project-slug}/issues   # submit a new customer issue

GET  /public/support/track/{token}           # get issue status by token
POST /public/support/track/{token}/comments  # customer adds a comment
```

---

## Frontend Architecture

### Auth Flow
1. App loads → fetches OIDC config from backend
2. Router guard checks `meta.requiresAuth` on navigation
3. If not authenticated → `mgr.signinRedirect()`
4. On return → `mgr.signinRedirectCallback()` → token stored
5. All API requests attach token via Axios interceptor

### Layouts
Three layouts:
- `MainLayout` — full app chrome (nav, sidebar), requires auth
- `MinimalLayout` — centered content, requires auth (for onboarding, settings)
- `PublicLayout` — no auth, used for customer portal and support tracking pages

### State Management
- **Server state**: TanStack Query exclusively. No manual fetch/store cycles.
- **UI state**: `ref` / `reactive` in component or composable. No global store.
- **Auth state**: Managed by `oidc-client-ts` via `useAuth` composable.

### Search
The search input does instant filtering client-side for small result sets, and debounced server queries for full-text search. The UI presents a unified search experience — no "advanced query language" exposed to users. Filters (project, status, assignee, label) are presented as UI controls, not query syntax.

---

## Configuration

Config file (`config.yaml`) with `HIVETRACK_` env var overrides via Koanf.

```yaml
server:
  host: 0.0.0.0
  port: 8080
  external_url: https://hivetrack.example.com
  allowed_origins:
    - https://hivetrack.example.com

database:
  url: postgres://user:pass@localhost:5432/hivetrack

oidc:
  authority: https://auth.example.com
  client_id: hivetrack

email:
  smtp_host: smtp.example.com
  smtp_port: 587
  smtp_user: noreply@example.com
  smtp_password: secret
  from_address: "Hivetrack <noreply@example.com>"

initial_admin:
  email: admin@example.com   # This user gets instance_admin on first login
```

---

## Deployment

Minimal footprint:
- One Go binary
- One PostgreSQL database
- One OIDC provider (external, e.g. Keyline)
- One SMTP server (external)
- Static frontend assets (served by the Go binary or a CDN)

Docker Compose for local development. A production deployment is a single container + managed Postgres.

Migrations run automatically on startup (using sequential numbered SQL files).
