-- Modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" DROP COLUMN "cadence";
-- Modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" DROP COLUMN "cadence";
-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "platform" character varying NOT NULL;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "platform" character varying NOT NULL;
