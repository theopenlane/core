-- Modify "job_template_history" table
ALTER TABLE "job_template_history" ALTER COLUMN "windmill_path" DROP NOT NULL;
-- Modify "job_templates" table
ALTER TABLE "job_templates" ALTER COLUMN "windmill_path" DROP NOT NULL;
-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "active" boolean NOT NULL DEFAULT true;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP CONSTRAINT "scheduled_jobs_job_templates_job_template", ADD COLUMN "active" boolean NOT NULL DEFAULT true, ADD CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs" FOREIGN KEY ("job_id") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
