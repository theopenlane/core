-- +goose Up
-- modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" ALTER COLUMN "configuration" DROP NOT NULL, DROP COLUMN "cadence";
-- modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" ALTER COLUMN "configuration" DROP NOT NULL, DROP COLUMN "cadence";
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ALTER COLUMN "configuration" DROP NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "windmill_path" character varying NOT NULL, ADD COLUMN "download_url" character varying NOT NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ALTER COLUMN "configuration" DROP NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "windmill_path" character varying NOT NULL, ADD COLUMN "download_url" character varying NOT NULL;

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "download_url", DROP COLUMN "windmill_path", DROP COLUMN "platform", ALTER COLUMN "configuration" SET NOT NULL;
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "download_url", DROP COLUMN "windmill_path", DROP COLUMN "platform", ALTER COLUMN "configuration" SET NOT NULL;
-- reverse: modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" ADD COLUMN "cadence" jsonb NULL, ALTER COLUMN "configuration" SET NOT NULL;
-- reverse: modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" ADD COLUMN "cadence" jsonb NULL, ALTER COLUMN "configuration" SET NOT NULL;
