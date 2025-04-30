-- Create "scheduled_job_history" table
CREATE TABLE "scheduled_job_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "title" character varying NOT NULL, "description" character varying NULL, "job_type" character varying NOT NULL DEFAULT 'SSL', "environment" character varying NOT NULL DEFAULT 'OPENLANE', "script" character varying NULL, "is_active" boolean NOT NULL DEFAULT true, PRIMARY KEY ("id"));
-- Create index "scheduledjobhistory_history_time" to table: "scheduled_job_history"
CREATE INDEX "scheduledjobhistory_history_time" ON "scheduled_job_history" ("history_time");
-- Create "scheduled_jobs" table
CREATE TABLE "scheduled_jobs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "job_type" character varying NOT NULL DEFAULT 'SSL', "environment" character varying NOT NULL DEFAULT 'OPENLANE', "script" character varying NULL, "is_active" boolean NOT NULL DEFAULT true, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "scheduled_jobs_organizations_scheduled_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "scheduledjob_display_id_owner_id" to table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_display_id_owner_id" ON "scheduled_jobs" ("display_id", "owner_id");
-- Create index "scheduledjob_id" to table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_id" ON "scheduled_jobs" ("id");
-- Create "scheduled_job_settings" table
CREATE TABLE "scheduled_job_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "configuration" jsonb NOT NULL, "cadence" jsonb NULL, "cron" character varying NULL, "scheduled_job_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "scheduled_job_settings_scheduled_jobs_scheduled_job_setting" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "scheduled_job_settings_scheduled_job_id_key" to table: "scheduled_job_settings"
CREATE UNIQUE INDEX "scheduled_job_settings_scheduled_job_id_key" ON "scheduled_job_settings" ("scheduled_job_id");
-- Create index "scheduledjobsetting_id" to table: "scheduled_job_settings"
CREATE UNIQUE INDEX "scheduledjobsetting_id" ON "scheduled_job_settings" ("id");
