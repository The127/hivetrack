-- +migrate Up

ALTER TABLE issues ADD COLUMN cancel_reason TEXT;
