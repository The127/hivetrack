-- +migrate Up
CREATE TABLE issue_status_log (
    id         UUID        PRIMARY KEY,
    issue_id   UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    status     TEXT        NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_issue_status_log_issue_id ON issue_status_log(issue_id);

-- +migrate Down
DROP TABLE IF EXISTS issue_status_log;
