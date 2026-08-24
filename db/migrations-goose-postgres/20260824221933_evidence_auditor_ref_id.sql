-- +goose Up
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "auditor_reference_id" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "organization_job_runner_creators", DROP COLUMN "organization_job_runner_registration_token_creators", DROP COLUMN "organization_job_runner_token_creators", DROP COLUMN "organization_job_template_creators", DROP COLUMN "organization_scheduled_job_creators", DROP COLUMN "organization_scheduled_job_run_creators";

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_scheduled_job_run_creators" character varying NULL, ADD COLUMN "organization_scheduled_job_creators" character varying NULL, ADD COLUMN "organization_job_template_creators" character varying NULL, ADD COLUMN "organization_job_runner_token_creators" character varying NULL, ADD COLUMN "organization_job_runner_registration_token_creators" character varying NULL, ADD COLUMN "organization_job_runner_creators" character varying NULL;
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "auditor_reference_id";
