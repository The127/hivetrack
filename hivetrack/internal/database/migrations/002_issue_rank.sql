-- +migrate Up
ALTER TABLE issues ADD COLUMN "rank" VARCHAR(255);

-- +migrate Down
ALTER TABLE issues DROP COLUMN IF EXISTS "rank";
