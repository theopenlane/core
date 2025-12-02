-- +goose Up
-- create "trust_center_compliances" table
CREATE TABLE "trust_center_compliances" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- create "trust_center_compliance_history" table
CREATE TABLE "trust_center_compliance_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- create index "trustcentercompliancehistory_history_time" to table: "trust_center_compliance_history"
CREATE INDEX "trustcentercompliancehistory_history_time" ON "trust_center_compliance_history" ("history_time");

-- +goose Down
-- reverse: create index "trustcentercompliancehistory_history_time" to table: "trust_center_compliance_history"
DROP INDEX "trustcentercompliancehistory_history_time";
-- reverse: create "trust_center_compliance_history" table
DROP TABLE "trust_center_compliance_history";
-- reverse: create "trust_center_compliances" table
DROP TABLE "trust_center_compliances";
