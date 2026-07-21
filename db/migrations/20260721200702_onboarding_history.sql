-- Modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "observation_period_start_date" timestamptz NULL, ADD COLUMN "observation_period_end_date" timestamptz NULL, ADD COLUMN "fieldwork_start_date" timestamptz NULL, ADD COLUMN "fieldwork_end_date" timestamptz NULL;
-- Modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;
-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "is_suggested" boolean NOT NULL DEFAULT false, ADD COLUMN "priority" bigint NOT NULL DEFAULT 0, ADD COLUMN "source" character varying NULL, ADD COLUMN "source_key" character varying NULL;
