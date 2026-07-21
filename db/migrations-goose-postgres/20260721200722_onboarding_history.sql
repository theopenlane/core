-- +goose Up
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "observation_period_start_date" timestamptz NULL, ADD COLUMN "observation_period_end_date" timestamptz NULL, ADD COLUMN "fieldwork_start_date" timestamptz NULL, ADD COLUMN "fieldwork_end_date" timestamptz NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;
-- modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "available_at";

-- +goose Down
-- reverse: modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "available_at" timestamptz NULL;
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "priority";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "fieldwork_end_date", DROP COLUMN "fieldwork_start_date", DROP COLUMN "observation_period_end_date", DROP COLUMN "observation_period_start_date";
