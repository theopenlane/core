-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "is_template" boolean NOT NULL DEFAULT false;
