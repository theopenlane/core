-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "is_template" boolean NOT NULL DEFAULT false;
