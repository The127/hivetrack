-- +migrate Up
ALTER TABLE issues ADD COLUMN refined BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down
ALTER TABLE issues DROP COLUMN IF EXISTS refined;
