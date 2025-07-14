-- +goose Up
-- modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" DROP COLUMN "cadence";
-- modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" DROP COLUMN "cadence";
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "platform" character varying NOT NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "platform" character varying NOT NULL;

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "platform";
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "platform";
-- reverse: modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" ADD COLUMN "cadence" jsonb NULL;
-- reverse: modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" ADD COLUMN "cadence" jsonb NULL;
