-- +goose Up
-- modify "exports" table
ALTER TABLE "exports" ADD COLUMN "filters" character varying NULL, ADD COLUMN "error_message" character varying NULL;
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "job_type";
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "job_type";

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "job_type" character varying NOT NULL DEFAULT 'SSL';
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "job_type" character varying NOT NULL DEFAULT 'SSL';
-- reverse: modify "exports" table
ALTER TABLE "exports" DROP COLUMN "error_message", DROP COLUMN "filters";
