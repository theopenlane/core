-- +goose Up
-- drop "job_results" table
DROP TABLE IF EXISTS "job_results" CASCADE;
-- drop "job_runner_job_runner_tokens" table
DROP TABLE IF EXISTS "job_runner_job_runner_tokens" CASCADE;
-- drop "job_runner_registration_tokens" table
DROP TABLE IF EXISTS "job_runner_registration_tokens" CASCADE;
-- drop "job_runner_tokens" table
DROP TABLE IF EXISTS "job_runner_tokens" CASCADE;
-- drop "job_runners" table
DROP TABLE IF EXISTS "job_runners" CASCADE;
-- drop "job_template_history" table
DROP TABLE IF EXISTS "job_template_history" CASCADE;
-- drop "job_templates" table
DROP TABLE IF EXISTS "job_templates" CASCADE;
-- drop "scheduled_job_controls" table
DROP TABLE IF EXISTS "scheduled_job_controls" CASCADE;
-- drop "scheduled_job_history" table
DROP TABLE IF EXISTS "scheduled_job_history" CASCADE;
-- drop "scheduled_job_runs" table
DROP TABLE IF EXISTS "scheduled_job_runs" CASCADE;
-- drop "scheduled_job_subcontrols" table
DROP TABLE IF EXISTS "scheduled_job_subcontrols" CASCADE;
-- drop "scheduled_jobs" table
DROP TABLE IF EXISTS "scheduled_jobs" CASCADE;

-- +goose Down
-- irreversible: the dropped tables cannot be restored from this migration
