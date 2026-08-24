-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "auditor_reference_id" character varying NULL;
-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "organization_job_runner_creators", DROP COLUMN "organization_job_runner_registration_token_creators", DROP COLUMN "organization_job_runner_token_creators", DROP COLUMN "organization_job_template_creators", DROP COLUMN "organization_scheduled_job_creators", DROP COLUMN "organization_scheduled_job_run_creators";
