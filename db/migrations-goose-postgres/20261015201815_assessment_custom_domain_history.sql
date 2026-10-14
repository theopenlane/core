-- +goose Up
-- create "assessment_domain_history" table
CREATE TABLE "assessment_domain_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "owner_id" character varying NULL, "custom_domain_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "assessmentdomainhistory_history_time" to table: "assessment_domain_history"
CREATE INDEX "assessmentdomainhistory_history_time" ON "assessment_domain_history" ("history_time");

-- +goose Down
-- reverse: create index "assessmentdomainhistory_history_time" to table: "assessment_domain_history"
DROP INDEX "assessmentdomainhistory_history_time";
-- reverse: create "assessment_domain_history" table
DROP TABLE "assessment_domain_history";
