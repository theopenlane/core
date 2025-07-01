-- +goose Up
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "windmill_path" character varying NOT NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "windmill_path" character varying NOT NULL;

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "windmill_path";
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "windmill_path";
