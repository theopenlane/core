-- +goose Up
-- modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ALTER COLUMN "job_runner_id" SET NOT NULL, ADD CONSTRAINT "scheduled_job_runs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- reverse: modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP CONSTRAINT "scheduled_job_runs_job_runners_job_runner", ALTER COLUMN "job_runner_id" DROP NOT NULL;
