-- +goose Up
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "download_url" character varying NOT NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "download_url" character varying NOT NULL;

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "download_url";
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "download_url";
