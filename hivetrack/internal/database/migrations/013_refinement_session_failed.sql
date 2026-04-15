-- +migrate Up

-- Allow refinement sessions to transition to a failed state when the
-- upstream agent reports a terminal error (e.g. Claude API 401).
ALTER TABLE refinement_sessions
    DROP CONSTRAINT refinement_sessions_status_check;
ALTER TABLE refinement_sessions
    ADD CONSTRAINT refinement_sessions_status_check
    CHECK (status IN ('active', 'completed', 'abandoned', 'failed'));
