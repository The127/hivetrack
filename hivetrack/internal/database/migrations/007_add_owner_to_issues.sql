-- +migrate Up

ALTER TABLE issues ADD COLUMN owner_id UUID REFERENCES users(id);

-- Backfill: set owner to reporter for existing issues
UPDATE issues SET owner_id = reporter_id WHERE owner_id IS NULL;
