-- Modify "exports" table
ALTER TABLE "exports" ADD COLUMN "filters" character varying NULL, ADD COLUMN "error_message" character varying NULL;
-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "job_type";
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "job_type";
