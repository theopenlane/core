-- +goose Up
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "system_owned" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "system_owned";
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "system_owned";
