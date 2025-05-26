-- +goose Up
-- create "scheduled_job_runs" table
CREATE TABLE "scheduled_job_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "job_runner_id" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "owner_id" character varying NULL, "scheduled_job_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "scheduled_job_runs_control_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "control_scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "scheduled_job_runs_organizations_scheduled_job_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "scheduledjobrun_id" to table: "scheduled_job_runs"
CREATE UNIQUE INDEX "scheduledjobrun_id" ON "scheduled_job_runs" ("id");

-- +goose Down
-- reverse: create index "scheduledjobrun_id" to table: "scheduled_job_runs"
DROP INDEX "scheduledjobrun_id";
-- reverse: create "scheduled_job_runs" table
DROP TABLE "scheduled_job_runs";
