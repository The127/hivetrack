-- +migrate Up

CREATE TABLE refinement_sessions (
    id UUID PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_refinement_sessions_active ON refinement_sessions(issue_id) WHERE status = 'active';

CREATE TABLE refinement_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES refinement_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    message_type TEXT NOT NULL DEFAULT 'message' CHECK (message_type IN ('message', 'proposal')),
    proposal JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refinement_messages_session ON refinement_messages(session_id, created_at);
