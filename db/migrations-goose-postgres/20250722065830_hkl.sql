-- +goose Up
-- modify "job_template_history" table
ALTER TABLE "job_template_history" ALTER COLUMN "windmill_path" DROP NOT NULL;
-- modify "job_templates" table
ALTER TABLE "job_templates" ALTER COLUMN "windmill_path" DROP NOT NULL;
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "active" boolean NOT NULL DEFAULT true;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP CONSTRAINT "scheduled_jobs_job_templates_job_template", ADD COLUMN "active" boolean NOT NULL DEFAULT true, ADD CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs" FOREIGN KEY ("job_id") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs", DROP COLUMN "active", ADD CONSTRAINT "scheduled_jobs_job_templates_job_template" FOREIGN KEY ("job_id") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "active";
-- reverse: modify "job_templates" table
ALTER TABLE "job_templates" ALTER COLUMN "windmill_path" SET NOT NULL;
-- reverse: modify "job_template_history" table
ALTER TABLE "job_template_history" ALTER COLUMN "windmill_path" SET NOT NULL;
