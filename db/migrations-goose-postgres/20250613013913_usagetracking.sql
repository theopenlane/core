-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "organization_id" character varying NULL;
-- create "usages" table
CREATE TABLE "usages" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "organization_id" character varying NOT NULL, "resource_type" character varying NOT NULL, "used" bigint NOT NULL DEFAULT 0, "limit" bigint NOT NULL DEFAULT 0, PRIMARY KEY ("id"));
-- create index "usage_id" to table: "usages"
CREATE UNIQUE INDEX "usage_id" ON "usages" ("id");
-- create "usage_history" table
CREATE TABLE "usage_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "organization_id" character varying NOT NULL, "resource_type" character varying NOT NULL, "used" bigint NOT NULL DEFAULT 0, "limit" bigint NOT NULL DEFAULT 0, PRIMARY KEY ("id"));
-- create index "usagehistory_history_time" to table: "usage_history"
CREATE INDEX "usagehistory_history_time" ON "usage_history" ("history_time");
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "organization_id" character varying NULL, ADD CONSTRAINT "files_organizations_files" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_organizations_files", DROP COLUMN "organization_id";
-- reverse: create index "usagehistory_history_time" to table: "usage_history"
DROP INDEX "usagehistory_history_time";
-- reverse: create "usage_history" table
DROP TABLE "usage_history";
-- reverse: create index "usage_id" to table: "usages"
DROP INDEX "usage_id";
-- reverse: create "usages" table
DROP TABLE "usages";
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "organization_id";
