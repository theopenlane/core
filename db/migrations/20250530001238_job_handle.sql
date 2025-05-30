-- Modify "control_scheduled_job_history" table
ALTER TABLE "control_scheduled_job_history" ADD COLUMN "job_handle" character varying NOT NULL;
-- Modify "control_scheduled_jobs" table
ALTER TABLE "control_scheduled_jobs" ADD COLUMN "job_handle" character varying NOT NULL;
