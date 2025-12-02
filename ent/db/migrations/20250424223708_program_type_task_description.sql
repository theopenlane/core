-- Modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "program_type" character varying NOT NULL DEFAULT 'FRAMEWORK', ADD COLUMN "framework_name" character varying NULL;
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "program_type" character varying NOT NULL DEFAULT 'FRAMEWORK', ADD COLUMN "framework_name" character varying NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "description";
-- Modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "description";
