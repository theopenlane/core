-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "parent_task_id" character varying NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "parent_task_id" character varying NULL, ADD CONSTRAINT "tasks_tasks_tasks" FOREIGN KEY ("parent_task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
