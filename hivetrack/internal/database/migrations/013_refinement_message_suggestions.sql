-- +migrate Up

-- Add suggestions array to refinement_messages for structured quick-reply chips
ALTER TABLE refinement_messages ADD COLUMN suggestions JSONB;
