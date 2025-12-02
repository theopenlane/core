-- +goose Up
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "program_type" character varying NOT NULL DEFAULT 'FRAMEWORK', ADD COLUMN "framework_name" character varying NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "program_type" character varying NOT NULL DEFAULT 'FRAMEWORK', ADD COLUMN "framework_name" character varying NULL;
-- modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "description";
-- modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "description";

-- +goose Down
-- reverse: modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "description" character varying NULL;
-- reverse: modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "description" character varying NULL;
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "framework_name", DROP COLUMN "program_type";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "framework_name", DROP COLUMN "program_type";
