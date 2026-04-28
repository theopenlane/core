-- Modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "directory_name" character varying NULL;
-- Modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "directory_name" character varying NULL;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "fix_available" boolean NULL;
-- Create "check_result_history" table
CREATE TABLE "check_result_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'UNKNOWN', "source" character varying NOT NULL, "last_observed_at" timestamptz NULL, "external_uri" character varying NULL, "details" text NULL, "parent_external_id" character varying NULL, "integration_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "checkresulthistory_history_time" to table: "check_result_history"
CREATE INDEX "checkresulthistory_history_time" ON "check_result_history" ("history_time");
