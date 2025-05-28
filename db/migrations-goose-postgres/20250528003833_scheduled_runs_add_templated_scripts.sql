-- +goose Up
-- modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ADD COLUMN "expected_execution_time" timestamptz NOT NULL, ADD COLUMN "script" character varying NOT NULL;

-- +goose Down
-- reverse: modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP COLUMN "script", DROP COLUMN "expected_execution_time";
