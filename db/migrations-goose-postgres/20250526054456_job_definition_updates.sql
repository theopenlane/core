-- +goose Up
-- modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" DROP CONSTRAINT "control_scheduled_jobs_job_runners_job_runner", ALTER COLUMN "job_runner_id" DROP NOT NULL, ADD CONSTRAINT "control_scheduled_jobs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" DROP CONSTRAINT "control_scheduled_jobs_job_runners_job_runner", ALTER COLUMN "job_runner_id" SET NOT NULL, ADD CONSTRAINT "control_scheduled_jobs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
