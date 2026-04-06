-- +migrate Up

-- Track current phase per refinement session
ALTER TABLE refinement_sessions
    ADD COLUMN current_phase TEXT NOT NULL DEFAULT 'actor_goal'
    CHECK (current_phase IN ('actor_goal', 'main_scenario', 'extensions', 'acceptance_criteria'));

-- Track which phase each message belongs to
ALTER TABLE refinement_messages
    ADD COLUMN phase TEXT NOT NULL DEFAULT 'actor_goal'
    CHECK (phase IN ('actor_goal', 'main_scenario', 'extensions', 'acceptance_criteria'));

-- Expand message_type to include 'phase_result'
ALTER TABLE refinement_messages
    DROP CONSTRAINT refinement_messages_message_type_check;
ALTER TABLE refinement_messages
    ADD CONSTRAINT refinement_messages_message_type_check
    CHECK (message_type IN ('message', 'proposal', 'phase_result'));
