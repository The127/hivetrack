# Hivetrack — API-First Design & AI Integration

## API-First

Hivetrack is built API-first. This means:

- The frontend is a first-class API consumer, not a privileged client. It uses the exact same endpoints that external tools, scripts, and integrations use.
- Every capability of Hivetrack is available via the API. If you can do it in the UI, you can do it via the API.
- The API is versioned (`/api/v1/`) from day one. Breaking changes require a new version prefix.
- The API is the contract. The UI is one implementation of that contract.

### No API Keys — OIDC Service Accounts Instead
Hivetrack has no custom API key system. Machine callers (CI pipelines, bots, AI agents, monitoring integrations) authenticate using **OAuth 2.0 Client Credentials** — a standard feature of any OIDC provider (Keyline, Keycloak, Authentik).

The operator creates a service account in the OIDC provider. The service exchanges its client credentials for a short-lived JWT and calls Hivetrack with `Authorization: Bearer <token>` — the same auth path as a human user. Hivetrack validates the JWT and resolves the caller's identity and permissions from its `sub` claim.

**Benefits:**
- Zero custom credential management in Hivetrack
- Credentials are managed in one place (the OIDC provider), not two
- Service accounts expire, rotate, and are revoked via the OIDC provider — not via Hivetrack admin
- The same permission model applies to humans and machines — no special cases

### Webhook Outbound
Projects can register webhook endpoints. When issue events occur, Hivetrack POSTs a JSON payload to the registered URL.

```
Webhook
├── id          uuid
├── project_id  uuid      FK → Project (or null for instance-wide)
├── url         string
├── secret      string    used to sign payload (HMAC-SHA256 in X-Hivetrack-Signature header)
├── events      []string  e.g. ["issue.created", "issue.status_changed", "sprint.started"]
├── active      bool
├── created_at  timestamp
```

Webhook delivery also uses the outbox pattern — guaranteed at-least-once delivery with retry.

**Event payload shape:**
```json
{
  "event": "issue.status_changed",
  "timestamp": "2026-03-17T10:00:00Z",
  "project": { "slug": "ht", "name": "Hivetrack" },
  "issue": { "number": 42, "title": "...", "status": "in_progress" },
  "actor": { "email": "user@example.com" },
  "diff": { "status": { "from": "todo", "to": "in_progress" } }
}
```

---

## AI Integration

AI is a first-class concept in Hivetrack, not a bolt-on feature. The integration philosophy:

- **AI augments the human, it does not replace decisions.** AI suggests, humans approve.
- **AI operates through the same API.** No special AI-only endpoints. An AI agent has an API key and uses the same REST API a human would use.
- **Transparency.** AI-authored content (issue descriptions, summaries) is visibly marked.
- **Opt-in per instance.** AI features require explicit configuration. Instances without an LLM API key simply do not show AI UI elements.

### AI-Assisted Issue Creation
When creating an issue, a user can paste raw text (a Slack message, a bug report, an email) and request AI structuring. The backend sends the raw text to the configured LLM and returns a structured draft:
- `title` suggestion
- `description` as clean markdown
- suggested `type`, `priority`, `labels` (from existing labels in the project)

The user reviews and confirms before the issue is created. The AI never creates issues autonomously via this flow.

### AI Issue Summarization
For long issues with many comments, an endpoint returns an AI-generated summary:
```
GET /api/v1/projects/{slug}/issues/{number}/summary
```
Summary is cached (invalidated when new comments are added). Rendered in the UI as a collapsible "AI summary" section above the comment thread.

### AI Sprint Planning Assistance
Given a backlog, an AI endpoint can suggest which issues to pull into the next sprint based on:
- Priority
- Milestone target dates
- Estimated complexity (if issues have been sized)
- Current team capacity (number of active members)

```
POST /api/v1/projects/{slug}/sprints/suggest
Body: { "capacity": 5, "milestone_id": "..." }
```
Returns a suggested list of issue IDs with reasoning. The user drags issues from the suggestion into the sprint — no auto-assignment.

### AI in the Customer Portal
When a customer submits a support issue, AI can:
- Detect if the submission is a duplicate of an existing open issue (semantic similarity)
- Draft an initial response for the team to review and send

### AI Configuration
```yaml
ai:
  enabled: true
  provider: anthropic          # anthropic | openai | ollama
  model: claude-sonnet-4-6
  api_key: sk-ant-...
  # For ollama (self-hosted):
  # base_url: http://localhost:11434
```

When `ai.enabled: false` (default), no LLM calls are made and AI UI elements are hidden.

### AI Agent Access (API-first implication)
Because Hivetrack is API-first and uses OIDC for all callers, external AI agents (Claude, Copilot, custom bots) interact with Hivetrack via a service account JWT — exactly like any other machine caller. Example use cases:

- A CI/CD pipeline automatically creates a `bug` issue when a test suite fails, assigning it to the last committer on the failed test
- A bot comments on issues with deployment status ("This issue was deployed to staging at 14:32")
- An AI agent triages incoming support issues and sets initial priority and labels
- An LLM-powered tool reads the backlog and proposes epic breakdowns

These are all achievable without any special AI integration code in Hivetrack itself — they are API consumers. The AI features built into Hivetrack core are the ones that benefit from having access to internal context (existing issues, labels, members) and from being integrated into the creation/editing flow.

---

## Integration Patterns

### Inbound (webhooks → Hivetrack)
Via the public API with an API key. Useful for:
- Creating issues from external monitoring alerts
- Syncing issues from other tools

### Outbound (Hivetrack → webhooks)
Via the webhook system. Useful for:
- Posting to Slack/Discord on issue updates
- Triggering CI/CD on sprint start
- Syncing status to external dashboards

### Embedding & Headless Use
The API supports embedding Hivetrack data into other tools. A project's public roadmap (milestones + issues in `done`) can be fetched without auth:

```
GET /public/projects/{slug}/roadmap
```

This allows teams to embed a live roadmap on their website or docs without exposing internal project details.
