-- +goose Up
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "parent_task_id" character varying NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "parent_task_id" character varying NULL, ADD CONSTRAINT "tasks_tasks_tasks" FOREIGN KEY ("parent_task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_tasks_tasks", DROP COLUMN "parent_task_id";
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "parent_task_id";
