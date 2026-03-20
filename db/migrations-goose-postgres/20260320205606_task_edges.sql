-- +goose Up
-- modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "finding_tasks", DROP COLUMN "vulnerability_tasks";
-- create "finding_tasks" table
CREATE TABLE "finding_tasks" ("finding_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "task_id"), CONSTRAINT "finding_tasks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "vulnerability_tasks" table
CREATE TABLE "vulnerability_tasks" ("vulnerability_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "task_id"), CONSTRAINT "vulnerability_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_tasks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "vulnerability_tasks" table
DROP TABLE "vulnerability_tasks";
-- reverse: create "finding_tasks" table
DROP TABLE "finding_tasks";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "vulnerability_tasks" character varying NULL, ADD COLUMN "finding_tasks" character varying NULL;
