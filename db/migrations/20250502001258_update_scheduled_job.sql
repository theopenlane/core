-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "system_owned" boolean NULL DEFAULT false;
