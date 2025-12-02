-- +goose Up
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "script", DROP COLUMN "cadence";
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "script", DROP COLUMN "cadence";

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "cadence" jsonb NULL, ADD COLUMN "script" character varying NULL;
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "cadence" jsonb NULL, ADD COLUMN "script" character varying NULL;
