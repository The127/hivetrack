# Code Smell Report

Generated: 2026-04-04

## Backend (Go)

| # | Smell | Location | Summary |
|---|-------|----------|---------|
| 1 | ~~**Long Method**~~ | ~~`handlers/issue_handler.go:27-107`~~ | ~~`ListIssues` — 80 lines of repetitive query-param-to-struct mapping~~ **Fixed in #88** |
| 2 | ~~**Long Method**~~ | ~~`handlers/issue_handler.go:176-277`~~ | ~~`UpdateIssue` — 100 lines, double-decodes JSON to detect explicit nulls~~ **Fixed in #89** |
| 3 | ~~**Long Method**~~ | ~~`commands/update_issue.go:44-190`~~ | ~~`HandleUpdateIssue` — 145 lines, field-by-field patching + auto-assign + auto-triage + hold-clearing in one function~~ **Fixed in #90** |
| 4 | ~~**Long Method**~~ | ~~`repositories/postgres/issue_repository.go:78-233`~~ | ~~`ExecuteUpdate` — 155 lines of per-field `HasChange` checks~~ **Fixed in #91** |
| 5 | ~~**Long Method**~~ | ~~`repositories/postgres/issue_repository.go:307-414`~~ | ~~`List` — 107 lines of filter-to-SQL mapping~~ **Fixed in #92** |
| 6 | ~~**Long Parameter List**~~ | ~~`models/issue.go:195-206`~~ | ~~`NewIssueFromDB` — 33 positional params~~ **Fixed in #93** |
| 7 | ~~**Long Parameter List**~~ | ~~`repositories/postgres/issue_repository.go:543-554`~~ | ~~`buildIssue` — 32 positional params mirroring above~~ **Fixed in #93** |
| 8 | ~~**Duplicated Code**~~ | ~~`repositories/postgres/issue_repository.go:450-541`~~ | ~~`scanIssue` vs `scanIssueRow` — identical variable declarations and scan args, differ only by receiver type~~ **Fixed in #94** |
| 9 | ~~**Duplicated Code**~~ | ~~`commands/batch_update_issues.go:53-82` vs `commands/update_issue.go:58-129`~~ | ~~Field-by-field patching logic duplicated between single and batch update~~ **Fixed in #95** |
| 10 | ~~**Duplicated Code**~~ | ~~`handlers/issue_handler.go` (8 locations)~~ | ~~Issue resolution from slug+number repeated 8 times~~ **Already resolved — `resolveIssueID` shared helper exists** |
| 11 | ~~**Duplicated Code**~~ | ~~`handlers/sprint_handler.go:80-95, 162-177`~~ | ~~Sprint-belongs-to-project verification duplicated in Update and Delete~~ **Fixed in #96** |
| 12 | ~~**Duplicated Code**~~ | ~~`queries/get_issues.go:126-156` vs `queries/get_my_issues.go:46-77`~~ | ~~Building `IssueSummary` from `*models.Issue` duplicated~~ **Fixed in #96** |
| 13 | ~~**Data Clump**~~ | ~~`commands/update_issue.go`, `batch_update_issues.go`, `handlers/issue_handler.go`~~ | ~~`{OnHold, HoldReason, HoldNote}` triple travels together across commands and DTOs~~ **Fixed in #96** |
| 14 | ~~**Switch Statement**~~ | ~~`repositories/postgres/db_context.go:145-208`~~ | ~~`applyChange` — 6x3 = 18 branches dispatching on entity type x change type~~ **Fixed in #97** |
| 15 | ~~**Duplicated Code / Oddball Solution**~~ | ~~`commands/update_issue.go:200`, `update_sprint.go:129`, `models/issue.go:402`~~ | ~~Three separate "is terminal status" definitions with inconsistent behavior (`IsTerminal` omits `Resolved`)~~ **Fixed in #98** |
| 16 | **Shotgun Surgery** | Issue entity across 6+ files | Adding one field requires ~10 edit sites across models, repos, queries, commands |

## Frontend (Vue)

| # | Smell | Location | Summary |
|---|-------|----------|---------|
| 17 | ~~**Duplicated Code**~~ | ~~6-7 views~~ | ~~`PRIORITY_BORDER` / `ESTIMATE_LABEL` maps copy-pasted across BoardView, BacklogView, SprintDetailView, EpicsView, TriageView, HomeView, IssueDetailView~~ **Fixed in #99** |
| 18 | ~~**Duplicated Code**~~ | ~~3 views~~ | ~~`TERMINAL_STATUSES` sets defined independently in Board, Backlog, Overview~~ **Fixed in #100** |
| 19 | **Duplicated Code** | 5 views | Status column/color/label config defined ad-hoc in Board, Overview, Triage, SprintDetail, Home |
| 20 | **Duplicated Code** | 3 views | `computeRank()` fractional-indexing function identical in Board, Backlog, Home |
| 21 | **Duplicated Code** | 3 views | Drag-and-drop boilerplate (~100 lines each) in Board, Backlog, Home |
| 22 | **Duplicated Code** | 5 views | `formatDate` / `formatDateRange` reimplemented in Board, Overview, SprintDetail, Sprints, Milestones |
| 23 | **Duplicated Code** | 6+ views | Optimistic mutation cancel-snapshot-rollback-invalidate skeleton repeated everywhere |
| 24 | **Duplicated Code** | 3 views | Project header markup (slug badge + name + archetype badge) repeated |
| 25 | **Large Class** | `IssueDetailView.vue` (822 lines) | 11 mutations, inline editing, hold modal, link form, split modal — too many responsibilities |
| 26 | **Large Class** | `ProjectBacklogView.vue` (600+ lines) | Epic filtering, sprint CRUD, drag-and-drop, inline creation, status updates |
| 27 | **Large Class** | `ProjectOverviewView.vue` (615 lines) | Members, labels, WIP settings, drones, burndown — five entity types managed in one component |

## MCP Server

| # | Smell | Location | Summary |
|---|-------|----------|---------|
| 28 | **Long Method** | `tools_issues.go:562-675` | `makeBatchUpdateIssues` — 113 lines with pointless intermediate map |
| 29 | **Duplicated Code** | `tools_issues.go`, `tools_sprints.go`, `tools_milestones.go` | `setOrNull` / `setStr` / `setFromMap` helper lambdas — same logic redeclared 4 times |
| 30 | **Duplicated Code** | `tools_issues.go` (2 locations) | UUID list parsing + label resolution duplicated between single and batch update |
| 31 | **Oddball Solution** | `tools_issues.go:573-585` | Hand-rolled character-by-character integer parsing when `strconv.Atoi` and `intArg` exist elsewhere |
| 32 | **Temporary Field** | `tools_issues.go:592-668` | Intermediate `body` map populated then immediately read back — vestige of old raw HTTP client |

## Themes

1. **Issue entity is the epicenter.** Most backend smells stem from Issue's high field count cascading through every CQRS layer. The 33-param constructors and manual scan functions make adding fields painful and error-prone.

2. **Frontend duplication is pervasive.** Domain constants (priorities, statuses, estimates), utility functions (dates, ranks), and structural patterns (drag-and-drop, optimistic mutations) are copy-pasted across views. Extracting shared modules/composables would cut the most duplication.

3. **Three god components** (`IssueDetailView`, `ProjectBacklogView`, `ProjectOverviewView`) each manage 4-6 independent concerns and would benefit from decomposition into composables and sub-components.
