-- +goose Up
-- create "subprocessors" table
CREATE TABLE "subprocessors" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- create "subprocessor_history" table
CREATE TABLE "subprocessor_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- create index "subprocessorhistory_history_time" to table: "subprocessor_history"
CREATE INDEX "subprocessorhistory_history_time" ON "subprocessor_history" ("history_time");
-- create "trust_center_subprocessors" table
CREATE TABLE "trust_center_subprocessors" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- create "trust_center_subprocessor_history" table
CREATE TABLE "trust_center_subprocessor_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- create index "trustcentersubprocessorhistory_history_time" to table: "trust_center_subprocessor_history"
CREATE INDEX "trustcentersubprocessorhistory_history_time" ON "trust_center_subprocessor_history" ("history_time");

-- +goose Down
-- reverse: create index "trustcentersubprocessorhistory_history_time" to table: "trust_center_subprocessor_history"
DROP INDEX "trustcentersubprocessorhistory_history_time";
-- reverse: create "trust_center_subprocessor_history" table
DROP TABLE "trust_center_subprocessor_history";
-- reverse: create "trust_center_subprocessors" table
DROP TABLE "trust_center_subprocessors";
-- reverse: create index "subprocessorhistory_history_time" to table: "subprocessor_history"
DROP INDEX "subprocessorhistory_history_time";
-- reverse: create "subprocessor_history" table
DROP TABLE "subprocessor_history";
-- reverse: create "subprocessors" table
DROP TABLE "subprocessors";
