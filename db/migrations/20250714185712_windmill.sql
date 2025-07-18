-- Modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" ALTER COLUMN "configuration" DROP NOT NULL, DROP COLUMN "cadence";
-- Modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" ALTER COLUMN "configuration" DROP NOT NULL, DROP COLUMN "cadence";
-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ALTER COLUMN "configuration" DROP NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "windmill_path" character varying NOT NULL, ADD COLUMN "download_url" character varying NOT NULL;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ALTER COLUMN "configuration" DROP NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "windmill_path" character varying NOT NULL, ADD COLUMN "download_url" character varying NOT NULL;
