-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "download_url" character varying NOT NULL;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "download_url" character varying NOT NULL;
