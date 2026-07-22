-- +goose Up
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "observation_period_start_date" timestamptz NULL, ADD COLUMN "observation_period_end_date" timestamptz NULL, ADD COLUMN "fieldwork_start_date" timestamptz NULL, ADD COLUMN "fieldwork_end_date" timestamptz NULL;
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "is_suggested" boolean NOT NULL DEFAULT false, ADD COLUMN "priority" bigint NOT NULL DEFAULT 0, ADD COLUMN "source" character varying NULL, ADD COLUMN "source_key" character varying NULL;

-- +goose Down
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "source_key", DROP COLUMN "source", DROP COLUMN "priority", DROP COLUMN "is_suggested", DROP COLUMN "metadata";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "fieldwork_end_date", DROP COLUMN "fieldwork_start_date", DROP COLUMN "observation_period_end_date", DROP COLUMN "observation_period_start_date";
