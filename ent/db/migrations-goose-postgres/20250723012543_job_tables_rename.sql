-- +goose Up
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "tags", DROP COLUMN "system_owned", DROP COLUMN "title", DROP COLUMN "description", DROP COLUMN "platform", DROP COLUMN "windmill_path", DROP COLUMN "download_url", ADD COLUMN "job_id" character varying NOT NULL, ADD COLUMN "active" boolean NOT NULL DEFAULT true, ADD COLUMN "job_runner_id" character varying NULL;
-- create "job_template_history" table
CREATE TABLE "job_template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "title" character varying NOT NULL, "description" character varying NULL, "platform" character varying NOT NULL, "windmill_path" character varying NULL, "download_url" character varying NOT NULL, "configuration" jsonb NULL, "cron" character varying NULL, PRIMARY KEY ("id"));
-- create index "jobtemplatehistory_history_time" to table: "job_template_history"
CREATE INDEX "jobtemplatehistory_history_time" ON "job_template_history" ("history_time");
-- create "job_templates" table
CREATE TABLE "job_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "title" character varying NOT NULL, "description" character varying NULL, "platform" character varying NOT NULL, "windmill_path" character varying NULL, "download_url" character varying NOT NULL, "configuration" jsonb NULL, "cron" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "job_templates_organizations_job_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "jobtemplate_display_id_owner_id" to table: "job_templates"
CREATE UNIQUE INDEX "jobtemplate_display_id_owner_id" ON "job_templates" ("display_id", "owner_id");
-- create index "jobtemplate_owner_id" to table: "job_templates"
CREATE INDEX "jobtemplate_owner_id" ON "job_templates" ("owner_id") WHERE (deleted_at IS NULL);
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP CONSTRAINT "scheduled_jobs_organizations_jobs", DROP COLUMN "tags", DROP COLUMN "system_owned", DROP COLUMN "title", DROP COLUMN "description", DROP COLUMN "platform", DROP COLUMN "windmill_path", DROP COLUMN "download_url", ADD COLUMN "active" boolean NOT NULL DEFAULT true, ADD COLUMN "job_id" character varying NOT NULL, ADD COLUMN "job_runner_id" character varying NULL, ADD CONSTRAINT "scheduled_jobs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs" FOREIGN KEY ("job_id") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "scheduled_jobs_organizations_scheduled_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "job_results" table
ALTER TABLE "job_results" DROP CONSTRAINT "job_results_control_scheduled_jobs_scheduled_job", ADD CONSTRAINT "job_results_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "scheduled_job_controls" table
CREATE TABLE "scheduled_job_controls" ("scheduled_job_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("scheduled_job_id", "control_id"), CONSTRAINT "scheduled_job_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "scheduled_job_controls_scheduled_job_id" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP CONSTRAINT "scheduled_job_runs_control_scheduled_jobs_scheduled_job", ADD CONSTRAINT "scheduled_job_runs_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "scheduled_job_subcontrols" table
CREATE TABLE "scheduled_job_subcontrols" ("scheduled_job_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("scheduled_job_id", "subcontrol_id"), CONSTRAINT "scheduled_job_subcontrols_scheduled_job_id" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "scheduled_job_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "scheduled_job_subcontrols" table
DROP TABLE "scheduled_job_subcontrols";
-- reverse: modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP CONSTRAINT "scheduled_job_runs_scheduled_jobs_scheduled_job", ADD CONSTRAINT "scheduled_job_runs_control_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: create "scheduled_job_controls" table
DROP TABLE "scheduled_job_controls";
-- reverse: modify "job_results" table
ALTER TABLE "job_results" DROP CONSTRAINT "job_results_scheduled_jobs_scheduled_job", ADD CONSTRAINT "job_results_control_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP CONSTRAINT "scheduled_jobs_organizations_scheduled_jobs", DROP CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs", DROP CONSTRAINT "scheduled_jobs_job_runners_job_runner", DROP COLUMN "job_runner_id", DROP COLUMN "job_id", DROP COLUMN "active", ADD COLUMN "download_url" character varying NOT NULL, ADD COLUMN "windmill_path" character varying NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "description" character varying NULL, ADD COLUMN "title" character varying NOT NULL, ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "tags" jsonb NULL, ADD CONSTRAINT "scheduled_jobs_organizations_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: create index "jobtemplate_owner_id" to table: "job_templates"
DROP INDEX "jobtemplate_owner_id";
-- reverse: create index "jobtemplate_display_id_owner_id" to table: "job_templates"
DROP INDEX "jobtemplate_display_id_owner_id";
-- reverse: create "job_templates" table
DROP TABLE "job_templates";
-- reverse: create index "jobtemplatehistory_history_time" to table: "job_template_history"
DROP INDEX "jobtemplatehistory_history_time";
-- reverse: create "job_template_history" table
DROP TABLE "job_template_history";
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "job_runner_id", DROP COLUMN "active", DROP COLUMN "job_id", ADD COLUMN "download_url" character varying NOT NULL, ADD COLUMN "windmill_path" character varying NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "description" character varying NULL, ADD COLUMN "title" character varying NOT NULL, ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "tags" jsonb NULL;
