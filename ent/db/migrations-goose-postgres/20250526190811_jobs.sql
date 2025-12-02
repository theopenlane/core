-- +goose Up
-- create "control_scheduled_job_history" table
CREATE TABLE "control_scheduled_job_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "job_id" character varying NOT NULL, "configuration" jsonb NOT NULL, "cadence" jsonb NULL, "cron" character varying NULL, "job_runner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "controlscheduledjobhistory_history_time" to table: "control_scheduled_job_history"
CREATE INDEX "controlscheduledjobhistory_history_time" ON "control_scheduled_job_history" ("history_time");
-- create "scheduled_job_history" table
CREATE TABLE "scheduled_job_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "title" character varying NOT NULL, "description" character varying NULL, "job_type" character varying NOT NULL DEFAULT 'SSL', "script" character varying NULL, "configuration" jsonb NOT NULL, "cadence" jsonb NULL, "cron" character varying NULL, PRIMARY KEY ("id"));
-- create index "scheduledjobhistory_history_time" to table: "scheduled_job_history"
CREATE INDEX "scheduledjobhistory_history_time" ON "scheduled_job_history" ("history_time");
-- create "scheduled_jobs" table
CREATE TABLE "scheduled_jobs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "title" character varying NOT NULL, "description" character varying NULL, "job_type" character varying NOT NULL DEFAULT 'SSL', "script" character varying NULL, "configuration" jsonb NOT NULL, "cadence" jsonb NULL, "cron" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "scheduled_jobs_organizations_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "scheduledjob_display_id_owner_id" to table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_display_id_owner_id" ON "scheduled_jobs" ("display_id", "owner_id");
-- create index "scheduledjob_id" to table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_id" ON "scheduled_jobs" ("id");
-- create "control_scheduled_jobs" table
CREATE TABLE "control_scheduled_jobs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "configuration" jsonb NOT NULL, "cadence" jsonb NULL, "cron" character varying NULL, "job_id" character varying NOT NULL, "job_runner_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "control_scheduled_jobs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "control_scheduled_jobs_organizations_scheduled_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "control_scheduled_jobs_scheduled_jobs_job" FOREIGN KEY ("job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "controlscheduledjob_id" to table: "control_scheduled_jobs"
CREATE UNIQUE INDEX "controlscheduledjob_id" ON "control_scheduled_jobs" ("id");
-- create "control_scheduled_job_controls" table
CREATE TABLE "control_scheduled_job_controls" ("control_scheduled_job_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("control_scheduled_job_id", "control_id"), CONSTRAINT "control_scheduled_job_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_scheduled_job_controls_control_scheduled_job_id" FOREIGN KEY ("control_scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_scheduled_job_subcontrols" table
CREATE TABLE "control_scheduled_job_subcontrols" ("control_scheduled_job_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("control_scheduled_job_id", "subcontrol_id"), CONSTRAINT "control_scheduled_job_subcontrols_control_scheduled_job_id" FOREIGN KEY ("control_scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_scheduled_job_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "job_results" table
CREATE TABLE "job_results" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "status" character varying NOT NULL, "exit_code" bigint NOT NULL, "finished_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "scheduled_job_id" character varying NOT NULL, "file_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "job_results_control_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "job_results_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "job_results_organizations_job_results" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "jobresult_id" to table: "job_results"
CREATE UNIQUE INDEX "jobresult_id" ON "job_results" ("id");
-- create "scheduled_job_runs" table
CREATE TABLE "scheduled_job_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "job_runner_id" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "owner_id" character varying NULL, "scheduled_job_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "scheduled_job_runs_control_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "scheduled_job_runs_organizations_scheduled_job_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "scheduledjobrun_id" to table: "scheduled_job_runs"
CREATE UNIQUE INDEX "scheduledjobrun_id" ON "scheduled_job_runs" ("id");

-- +goose Down
-- reverse: create index "scheduledjobrun_id" to table: "scheduled_job_runs"
DROP INDEX "scheduledjobrun_id";
-- reverse: create "scheduled_job_runs" table
DROP TABLE "scheduled_job_runs";
-- reverse: create index "jobresult_id" to table: "job_results"
DROP INDEX "jobresult_id";
-- reverse: create "job_results" table
DROP TABLE "job_results";
-- reverse: create "control_scheduled_job_subcontrols" table
DROP TABLE "control_scheduled_job_subcontrols";
-- reverse: create "control_scheduled_job_controls" table
DROP TABLE "control_scheduled_job_controls";
-- reverse: create index "controlscheduledjob_id" to table: "control_scheduled_jobs"
DROP INDEX "controlscheduledjob_id";
-- reverse: create "control_scheduled_jobs" table
DROP TABLE "control_scheduled_jobs";
-- reverse: create index "scheduledjob_id" to table: "scheduled_jobs"
DROP INDEX "scheduledjob_id";
-- reverse: create index "scheduledjob_display_id_owner_id" to table: "scheduled_jobs"
DROP INDEX "scheduledjob_display_id_owner_id";
-- reverse: create "scheduled_jobs" table
DROP TABLE "scheduled_jobs";
-- reverse: create index "scheduledjobhistory_history_time" to table: "scheduled_job_history"
DROP INDEX "scheduledjobhistory_history_time";
-- reverse: create "scheduled_job_history" table
DROP TABLE "scheduled_job_history";
-- reverse: create index "controlscheduledjobhistory_history_time" to table: "control_scheduled_job_history"
DROP INDEX "controlscheduledjobhistory_history_time";
-- reverse: create "control_scheduled_job_history" table
DROP TABLE "control_scheduled_job_history";
