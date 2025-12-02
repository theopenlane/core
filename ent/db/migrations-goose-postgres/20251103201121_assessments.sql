-- +goose Up
-- create "assessment_history" table
CREATE TABLE "assessment_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "template_id" character varying NOT NULL, "assessment_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmenthistory_history_time" to table: "assessment_history"
CREATE INDEX "assessmenthistory_history_time" ON "assessment_history" ("history_time");
-- create "assessment_response_history" table
CREATE TABLE "assessment_response_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "assessment_id" character varying NOT NULL, "email" character varying NOT NULL, "send_attempts" bigint NOT NULL DEFAULT 1, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "document_data_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
CREATE INDEX "assessmentresponsehistory_history_time" ON "assessment_response_history" ("history_time");
-- create "assessments" table
CREATE TABLE "assessments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "assessment_owner_id" character varying NULL, "template_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessments_organizations_assessments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "assessments_templates_template" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "assessment_name_owner_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "assessment_owner_id" to table: "assessments"
CREATE INDEX "assessment_owner_id" ON "assessments" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "assessments_assessment_owner_id_key" to table: "assessments"
CREATE UNIQUE INDEX "assessments_assessment_owner_id_key" ON "assessments" ("assessment_owner_id");
-- create "assessment_responses" table
CREATE TABLE "assessment_responses" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "email" character varying NOT NULL, "send_attempts" bigint NOT NULL DEFAULT 1, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "assessment_id" character varying NOT NULL, "document_data_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessment_responses_assessments_assessment_responses" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "assessment_responses_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "assessment_responses_organizations_assessment_responses" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "assessmentresponse_assessment_id_email" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_email" ON "assessment_responses" ("assessment_id", "email") WHERE (deleted_at IS NULL);
-- create index "assessmentresponse_assigned_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_assigned_at" ON "assessment_responses" ("assigned_at");
-- create index "assessmentresponse_completed_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_completed_at" ON "assessment_responses" ("completed_at");
-- create index "assessmentresponse_due_date" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_due_date" ON "assessment_responses" ("due_date");
-- create index "assessmentresponse_owner_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_owner_id" ON "assessment_responses" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "assessmentresponse_status" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_status" ON "assessment_responses" ("status");
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "assessment_blocked_groups" character varying NULL, ADD COLUMN "assessment_editors" character varying NULL, ADD COLUMN "assessment_viewers" character varying NULL, ADD CONSTRAINT "groups_assessments_blocked_groups" FOREIGN KEY ("assessment_blocked_groups") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_editors" FOREIGN KEY ("assessment_editors") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_viewers" FOREIGN KEY ("assessment_viewers") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_assessments_viewers", DROP CONSTRAINT "groups_assessments_editors", DROP CONSTRAINT "groups_assessments_blocked_groups", DROP COLUMN "assessment_viewers", DROP COLUMN "assessment_editors", DROP COLUMN "assessment_blocked_groups";
-- reverse: create index "assessmentresponse_status" to table: "assessment_responses"
DROP INDEX "assessmentresponse_status";
-- reverse: create index "assessmentresponse_owner_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_owner_id";
-- reverse: create index "assessmentresponse_due_date" to table: "assessment_responses"
DROP INDEX "assessmentresponse_due_date";
-- reverse: create index "assessmentresponse_completed_at" to table: "assessment_responses"
DROP INDEX "assessmentresponse_completed_at";
-- reverse: create index "assessmentresponse_assigned_at" to table: "assessment_responses"
DROP INDEX "assessmentresponse_assigned_at";
-- reverse: create index "assessmentresponse_assessment_id_email" to table: "assessment_responses"
DROP INDEX "assessmentresponse_assessment_id_email";
-- reverse: create "assessment_responses" table
DROP TABLE "assessment_responses";
-- reverse: create index "assessments_assessment_owner_id_key" to table: "assessments"
DROP INDEX "assessments_assessment_owner_id_key";
-- reverse: create index "assessment_owner_id" to table: "assessments"
DROP INDEX "assessment_owner_id";
-- reverse: create index "assessment_name_owner_id" to table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- reverse: create "assessments" table
DROP TABLE "assessments";
-- reverse: create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
DROP INDEX "assessmentresponsehistory_history_time";
-- reverse: create "assessment_response_history" table
DROP TABLE "assessment_response_history";
-- reverse: create index "assessmenthistory_history_time" to table: "assessment_history"
DROP INDEX "assessmenthistory_history_time";
-- reverse: create "assessment_history" table
DROP TABLE "assessment_history";
