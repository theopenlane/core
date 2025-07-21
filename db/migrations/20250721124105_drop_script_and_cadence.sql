-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "script", DROP COLUMN "cadence";
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "script", DROP COLUMN "cadence";
