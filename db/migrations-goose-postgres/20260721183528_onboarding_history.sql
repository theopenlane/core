-- +goose Up
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "observation_period_start_date" timestamptz NULL, ADD COLUMN "observation_period_end_date" timestamptz NULL, ADD COLUMN "fieldwork_start_date" timestamptz NULL, ADD COLUMN "fieldwork_end_date" timestamptz NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;

-- +goose Down
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "priority";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "fieldwork_end_date", DROP COLUMN "fieldwork_start_date", DROP COLUMN "observation_period_end_date", DROP COLUMN "observation_period_start_date";
