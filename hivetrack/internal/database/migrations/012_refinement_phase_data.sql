-- +migrate Up

-- Structured phase data for refinement messages (actor/goal, scenario steps, etc.)
ALTER TABLE refinement_messages ADD COLUMN phase_data JSONB;
