# Hivetrack Domain Model

## Entity Reference

### User
```
id            uuid        PK
sub           string      OIDC subject claim (unique)
email         string      unique
display_name  string
avatar_url    string?
created_at    timestamp
last_login_at timestamp
```

### ProjectMember (join table)
```
project_id  uuid   FK → Project
user_id     uuid   FK → User
role        enum   project_admin | project_member | viewer
```

### Project
```
id          uuid    PK
slug        string  unique, URL-safe (e.g. "backend-platform")
name        string
description string?
archetype   enum    software | support
archived    bool    default false
created_by  uuid    FK → User
created_at  timestamp
-- Settings
auto_archive_done_after_days  int?  null = disabled. When set, issues in terminal status are archived after N days.
```

### Sprint (software projects only)
```
id          uuid    PK
project_id  uuid    FK → Project
name        string
goal        string?
start_date  date
end_date    date
status      enum    planning | active | completed
created_at  timestamp
```
Constraint: at most one `active` sprint per project.

### Milestone
```
id          uuid    PK
project_id  uuid    FK → Project
title       string
description string?
target_date date?
closed_at   timestamp?
created_at  timestamp
```

### Label
```
id         uuid   PK
project_id uuid   FK → Project
name       string
color      string (hex, e.g. "#e2b340")
```

### Issue
```
id                  uuid      PK
project_id          uuid      FK → Project
number              int       sequential per-project (e.g. 42 for "HT-42")
type                enum      epic | task
title               string
description         text?     markdown
status              string    value constrained by project archetype

-- Hold (orthogonal to status)
on_hold             bool      default false
hold_reason         enum?     waiting_on_customer | waiting_on_external | blocked_by_issue
hold_since          timestamp?
hold_note           string?

priority            enum      none | low | medium | high | critical   default none
estimate            enum      none | xs | s | m | l | xl              default none
reporter_id         uuid?     FK → User (null for external customer submissions)
parent_id           uuid?     FK → Issue (task → epic only)
milestone_id        uuid?     FK → Milestone
sprint_id           uuid?     FK → Sprint (null = backlog)
sprint_carry_count  int       default 0  incremented each time issue moves to a new sprint without being done

-- Triage (issues created externally or via quick-capture land here first)
triaged             bool      default true  (false = in triage inbox, not yet placed)

-- Visibility
visibility          enum      normal | restricted   default normal

-- Customer portal fields (support archetype only)
customer_email      string?
customer_name       string?
customer_token      uuid?     unique, random, for token-based access

-- Subtasks (tasks only, stored as JSONB, not queryable as entities)
checklist           jsonb     []{ id: uuid, text: string, done: bool }

created_at          timestamp
updated_at          timestamp
```

Issues with `triaged = false` form the **Triage Inbox**. They are visible to project members but do not appear on the board or backlog until triaged. Triage inbox issues are created when:
- An issue is submitted via CI/CD webhook or monitoring alert integration
- An issue is created via quick-capture (title only, no other fields set)
- A customer support submission arrives (support archetype)

Quick-capture flow: user types a title, hits Enter — issue created with `triaged = false`. They can triage it immediately or later.

### IssueAssignee (join table)
```
issue_id  uuid  FK → Issue
user_id   uuid  FK → User
```

### IssueRestrictedViewer (join table)
```
issue_id  uuid  FK → Issue
user_id   uuid  FK → User
```
Only relevant when `issue.visibility = restricted`.

### IssueLabel (join table)
```
issue_id  uuid  FK → Issue
label_id  uuid  FK → Label
```

### IssueWatcher (join table)
```
issue_id  uuid  FK → Issue
user_id   uuid  FK → User
```
Users can follow/watch any issue. Watchers receive notifications for events on the issue (comments, status changes, assignment changes). The reporter and all assignees are automatically added as watchers on issue creation. Users can manually follow or unfollow any issue they can see.

### UserFavorite
```
id          uuid   PK
user_id     uuid   FK → User
entity_type string  "issue" | "project" | "view"
entity_id   uuid
created_at  timestamp
```
Users can favorite issues, projects, and saved views. Favorites appear in the personal dashboard.

### UserRecentIssue
```
user_id    uuid       FK → User
issue_id   uuid       FK → Issue
viewed_at  timestamp
```
Tracks the last 20 issues a user has viewed. Used for the "recently viewed" section of the personal dashboard. Updated on every issue detail view. Composite primary key `(user_id, issue_id)`, `viewed_at` updated on conflict.

### IssueTemplate
```
id          uuid    PK
project_id  uuid    FK → Project
name        string  e.g. "Bug Report", "Feature Request", "Incident"
description string? explains when to use this template
icon        string? emoji or icon name
-- Pre-filled fields
default_type      enum?   epic | task
default_priority  enum?
default_labels    []uuid  FK → Label
default_assignees []uuid  FK → User
-- Description template (markdown with placeholders)
description_template  text?  markdown, e.g. "## Steps to reproduce\n\n## Expected\n\n## Actual"
checklist_template    jsonb  []{ text: string, done: false }
created_at  timestamp
```
Templates are per-project. When a user creates an issue using a template, all default fields are pre-filled and the description is pre-populated with the template markdown. User can override anything.

### SavedView
```
id          uuid    PK
project_id  uuid?   FK → Project (null = instance-wide / cross-project)
owner_id    uuid    FK → User
name        string
shared      bool    default false  (shared views visible to all project members)
-- Filter state stored as JSON
filters     jsonb   { status: [], priority: [], assignee: [], label: [], sprint: [], milestone: [], on_hold: bool?, triaged: bool?, text: string? }
-- Display preferences
sort_by     string  default "updated_at"
sort_dir    string  default "desc"
created_at  timestamp
```
Saved views are named, persistent filter+sort combinations. Shared views are visible to all members of the project (or instance, if cross-project). Personal views are visible only to the owner. Examples: "My open bugs", "Unassigned critical issues", "Everything blocked this sprint".

### Webhook
```
id         uuid      PK
project_id uuid?     FK → Project (null = instance-wide)
name       string
url        string
secret     string    HMAC-SHA256 signing secret
events     []string  e.g. ["issue.created", "issue.status_changed", "sprint.started"]
active     bool      default true
created_at timestamp
```
See `docs/api-and-ai.md` for event payload shape and delivery guarantees.

### IssueLink
```
id              uuid  PK
source_issue_id uuid  FK → Issue
target_issue_id uuid  FK → Issue
link_type       enum  blocks | duplicates | relates_to
```
`is_blocked_by` and `is_duplicated_by` are the inverse views of `blocks` and `duplicates` — stored as one record, rendered bidirectionally. `relates_to` is symmetric.

When a `blocks` link is created: target issue's `on_hold` is set to `true`, `hold_reason` to `blocked_by_issue` (if not already on hold for another reason).
When source issue reaches `done` or `cancelled`: scan for `blocks` links originating from it. For each target, **only clear the hold if all other blocking issues are also in a terminal state** (`done` or `cancelled`). An issue blocked by both A and B must not be unblocked when only A closes.

### Comment
```
id            uuid      PK
issue_id      uuid      FK → Issue
-- Author is either an internal user or an external customer (not both)
author_id     uuid?     FK → User
author_email  string?   for external customer comments
author_name   string?   for external customer comments
body          text      markdown
created_at    timestamp
updated_at    timestamp
```

### AuditLog
```
id           uuid      PK
entity_type  string    (e.g. "issue", "project", "sprint")
entity_id    uuid
action       string    (e.g. "created", "status_changed", "assigned")
actor_id     uuid?     FK → User (null for system/background actions)
actor_email  string?   for external customer actions
diff         jsonb     { field: { from, to } }
created_at   timestamp
```

### OutboxMessage
```
id           uuid      PK
type         string    (e.g. "send_email", "clear_blocked_hold")
payload      jsonb
status       enum      pending | delivered | failed
created_at   timestamp
delivered_at timestamp?
error        string?
```

---

## Status Values by Archetype

### software
```
todo         → default status for new issues (scheduled, not started)
in_progress  → actively being worked on
in_review    → in code review / QA
done         → complete
cancelled    → will not be completed
```
Terminal states: `done`, `cancelled`.

**Note:** There is no `backlog` status. "Backlog" is not a workflow state — it is the set of issues where `sprint_id IS NULL`. An issue is in the backlog by virtue of not being assigned to a sprint, not by having a special status. The board has a Backlog section filtered by `sprint_id IS NULL AND triaged = true`. This avoids the ambiguity of `status = backlog` conflicting with `sprint_id = NULL`.

### support
```
open         → submitted, not yet triaged
in_progress  → team is actively working on it
resolved     → team considers it done, awaiting customer confirmation
closed       → complete (either confirmed resolved or timed out)
```
Terminal states: `closed`.

---

## Issue Number Generation

Issue numbers are sequential integers scoped to a project. The human-readable identifier is `{PROJECT_SLUG_UPPERCASE}-{number}` (e.g. `HT-42`, `BACKEND-107`).

**Implementation:** A `project_issue_counters` table with one row per project:
```sql
CREATE TABLE project_issue_counters (
    project_id uuid PRIMARY KEY REFERENCES projects(id),
    next_number int NOT NULL DEFAULT 1
);
```
On issue creation, within the same transaction:
```sql
UPDATE project_issue_counters SET next_number = next_number + 1
WHERE project_id = $1
RETURNING next_number - 1 AS issue_number;
```
The `FOR UPDATE` is implicit via the transaction. No DDL per project, no sequences to manage.

Numbers are immutable. If an issue is deleted, its number is not reused.

---

## Indexes (performance-critical)
```sql
-- Issues by project (most common query)
CREATE INDEX idx_issues_project_id ON issues(project_id);

-- Issues by sprint (board view)
CREATE INDEX idx_issues_sprint_id ON issues(sprint_id);

-- Issues by milestone
CREATE INDEX idx_issues_milestone_id ON issues(milestone_id);

-- Issues by parent (epic view)
CREATE INDEX idx_issues_parent_id ON issues(parent_id);

-- On-hold issues (dashboard)
CREATE INDEX idx_issues_on_hold ON issues(project_id, on_hold) WHERE on_hold = true;

-- Full text search
CREATE INDEX idx_issues_search ON issues USING GIN(to_tsvector('english', title || ' ' || coalesce(description, '')));

-- Customer token lookup (unique, sparse)
CREATE UNIQUE INDEX idx_issues_customer_token ON issues(customer_token) WHERE customer_token IS NOT NULL;

-- Outbox polling
CREATE INDEX idx_outbox_pending ON outbox_messages(created_at) WHERE status = 'pending';
```
