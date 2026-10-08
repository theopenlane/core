-- Create "assessment_policy_history" table
CREATE TABLE "assessment_policy_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "owner_id" character varying NULL, "assessment_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, "policy_revision" character varying NULL, PRIMARY KEY ("id"));
-- Create index "assessmentpolicyhistory_history_time" to table: "assessment_policy_history"
CREATE INDEX "assessmentpolicyhistory_history_time" ON "assessment_policy_history" ("history_time");
