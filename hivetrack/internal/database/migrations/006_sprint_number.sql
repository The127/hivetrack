-- +migrate Up
ALTER TABLE sprints ADD COLUMN number INT NOT NULL DEFAULT 0;

UPDATE sprints
SET number = sub.rn
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY created_at, id) AS rn
    FROM sprints
) sub
WHERE sprints.id = sub.id;

ALTER TABLE sprints ALTER COLUMN number DROP DEFAULT;
ALTER TABLE sprints ADD CONSTRAINT sprints_project_id_number_unique UNIQUE (project_id, number);

CREATE TABLE project_sprint_counters (
    project_id  UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    next_number INT  NOT NULL DEFAULT 1
);

INSERT INTO project_sprint_counters (project_id, next_number)
SELECT p.id, COALESCE((SELECT MAX(s.number) + 1 FROM sprints s WHERE s.project_id = p.id), 1)
FROM projects p;

-- +migrate Down
DROP TABLE IF EXISTS project_sprint_counters;
ALTER TABLE sprints DROP COLUMN IF EXISTS number;
