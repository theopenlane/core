-- Modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "finding_tasks", DROP COLUMN "vulnerability_tasks";
-- Create "finding_tasks" table
CREATE TABLE "finding_tasks" ("finding_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "task_id"), CONSTRAINT "finding_tasks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "vulnerability_tasks" table
CREATE TABLE "vulnerability_tasks" ("vulnerability_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "task_id"), CONSTRAINT "vulnerability_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_tasks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
