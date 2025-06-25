-- Create "assessment_history" table
CREATE TABLE "assessment_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "questionnaire_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "assessmenthistory_history_time" to table: "assessment_history"
CREATE INDEX "assessmenthistory_history_time" ON "assessment_history" ("history_time");
-- Create "assessment_response_history" table
CREATE TABLE "assessment_response_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NULL, "started_at" timestamptz NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
CREATE INDEX "assessmentresponsehistory_history_time" ON "assessment_response_history" ("history_time");
-- Create "assessments" table
CREATE TABLE "assessments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "questionnaire_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessments_organizations_assessments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "assessment_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_id" ON "assessments" ("id");
-- Create index "assessment_name_owner_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "assessment_owner_id" to table: "assessments"
CREATE INDEX "assessment_owner_id" ON "assessments" ("owner_id") WHERE (deleted_at IS NULL);
-- Create "assessment_responses" table
CREATE TABLE "assessment_responses" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NULL, "started_at" timestamptz NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "assessment_responses_assessments_assessment_responses" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "assessment_responses_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "assessmentresponse_assessment_id_user_id" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_user_id" ON "assessment_responses" ("assessment_id", "user_id") WHERE (deleted_at IS NULL);
-- Create index "assessmentresponse_completed_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_completed_at" ON "assessment_responses" ("completed_at");
-- Create index "assessmentresponse_due_date" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_due_date" ON "assessment_responses" ("due_date");
-- Create index "assessmentresponse_id" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_id" ON "assessment_responses" ("id");
-- Create index "assessmentresponse_status" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_status" ON "assessment_responses" ("status");
-- Create "assessment_users" table
CREATE TABLE "assessment_users" ("assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("assessment_id", "user_id"), CONSTRAINT "assessment_users_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "assessment_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
