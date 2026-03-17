-- +migrate Up

CREATE TABLE users (
    id            UUID        PRIMARY KEY,
    sub           TEXT        NOT NULL UNIQUE,
    email         TEXT        NOT NULL UNIQUE,
    display_name  TEXT        NOT NULL DEFAULT '',
    avatar_url    TEXT,
    is_admin      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE projects (
    id            UUID        PRIMARY KEY,
    slug          TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    description   TEXT,
    archetype     TEXT        NOT NULL CHECK (archetype IN ('software', 'support')),
    archived      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_by    UUID        NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    auto_archive_done_after_days INT
);

CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       TEXT NOT NULL CHECK (role IN ('project_admin', 'project_member', 'viewer')),
    PRIMARY KEY (project_id, user_id)
);

CREATE TABLE project_issue_counters (
    project_id  UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    next_number INT  NOT NULL DEFAULT 1
);

CREATE TABLE sprints (
    id         UUID        PRIMARY KEY,
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    goal       TEXT,
    start_date DATE        NOT NULL,
    end_date   DATE        NOT NULL,
    status     TEXT        NOT NULL CHECK (status IN ('planning', 'active', 'completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE milestones (
    id          UUID        PRIMARY KEY,
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    description TEXT,
    target_date DATE,
    closed_at   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE labels (
    id         UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    color      TEXT NOT NULL,
    UNIQUE (project_id, name)
);

CREATE TABLE issues (
    id                   UUID        PRIMARY KEY,
    project_id           UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number               INT         NOT NULL,
    type                 TEXT        NOT NULL CHECK (type IN ('epic', 'task')),
    title                TEXT        NOT NULL,
    description          TEXT,
    status               TEXT        NOT NULL,

    on_hold              BOOLEAN     NOT NULL DEFAULT FALSE,
    hold_reason          TEXT        CHECK (hold_reason IN ('waiting_on_customer', 'waiting_on_external', 'blocked_by_issue')),
    hold_since           TIMESTAMPTZ,
    hold_note            TEXT,

    priority             TEXT        NOT NULL DEFAULT 'none' CHECK (priority IN ('none', 'low', 'medium', 'high', 'critical')),
    estimate             TEXT        NOT NULL DEFAULT 'none' CHECK (estimate IN ('none', 'xs', 's', 'm', 'l', 'xl')),

    reporter_id          UUID        REFERENCES users(id),
    parent_id            UUID        REFERENCES issues(id),
    milestone_id         UUID        REFERENCES milestones(id),
    sprint_id            UUID        REFERENCES sprints(id),
    sprint_carry_count   INT         NOT NULL DEFAULT 0,

    triaged              BOOLEAN     NOT NULL DEFAULT TRUE,

    visibility           TEXT        NOT NULL DEFAULT 'normal' CHECK (visibility IN ('normal', 'restricted')),

    customer_email       TEXT,
    customer_name        TEXT,
    customer_token       UUID        UNIQUE,

    checklist            JSONB       NOT NULL DEFAULT '[]',

    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (project_id, number)
);

CREATE TABLE issue_assignees (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, user_id)
);

CREATE TABLE issue_labels (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, label_id)
);

CREATE TABLE issue_restricted_viewers (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, user_id)
);

CREATE TABLE issue_links (
    id              UUID PRIMARY KEY,
    source_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    link_type       TEXT NOT NULL CHECK (link_type IN ('blocks', 'duplicates', 'relates_to'))
);

CREATE TABLE issue_watchers (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, user_id)
);

CREATE TABLE comments (
    id           UUID        PRIMARY KEY,
    issue_id     UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id    UUID        REFERENCES users(id),
    author_email TEXT,
    author_name  TEXT,
    body         TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE outbox_messages (
    id           UUID        PRIMARY KEY,
    type         TEXT        NOT NULL,
    payload      JSONB       NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'delivered', 'failed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    error        TEXT
);

-- Indexes
CREATE INDEX idx_issues_project_id    ON issues(project_id);
CREATE INDEX idx_issues_sprint_id     ON issues(sprint_id);
CREATE INDEX idx_issues_milestone_id  ON issues(milestone_id);
CREATE INDEX idx_issues_parent_id     ON issues(parent_id);
CREATE INDEX idx_issues_on_hold       ON issues(project_id, on_hold) WHERE on_hold = TRUE;
CREATE INDEX idx_issues_search        ON issues USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));
CREATE UNIQUE INDEX idx_issues_customer_token ON issues(customer_token) WHERE customer_token IS NOT NULL;
CREATE INDEX idx_outbox_pending       ON outbox_messages(created_at) WHERE status = 'pending';

-- +migrate Down

DROP INDEX IF EXISTS idx_outbox_pending;
DROP INDEX IF EXISTS idx_issues_customer_token;
DROP INDEX IF EXISTS idx_issues_search;
DROP INDEX IF EXISTS idx_issues_on_hold;
DROP INDEX IF EXISTS idx_issues_parent_id;
DROP INDEX IF EXISTS idx_issues_milestone_id;
DROP INDEX IF EXISTS idx_issues_sprint_id;
DROP INDEX IF EXISTS idx_issues_project_id;

DROP TABLE IF EXISTS outbox_messages;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS issue_watchers;
DROP TABLE IF EXISTS issue_links;
DROP TABLE IF EXISTS issue_restricted_viewers;
DROP TABLE IF EXISTS issue_labels;
DROP TABLE IF EXISTS issue_assignees;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS sprints;
DROP TABLE IF EXISTS project_issue_counters;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
