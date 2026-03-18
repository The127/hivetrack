-- +migrate Up

ALTER TABLE projects
    ADD COLUMN wip_limit_in_progress integer,
    ADD COLUMN wip_limit_in_review   integer;
