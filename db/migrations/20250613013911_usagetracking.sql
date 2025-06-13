-- Modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "organization_id" character varying NULL;
-- Create "usages" table
CREATE TABLE "usages" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "organization_id" character varying NOT NULL, "resource_type" character varying NOT NULL, "used" bigint NOT NULL DEFAULT 0, "limit" bigint NOT NULL DEFAULT 0, PRIMARY KEY ("id"));
-- Create index "usage_id" to table: "usages"
CREATE UNIQUE INDEX "usage_id" ON "usages" ("id");
-- Create "usage_history" table
CREATE TABLE "usage_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "organization_id" character varying NOT NULL, "resource_type" character varying NOT NULL, "used" bigint NOT NULL DEFAULT 0, "limit" bigint NOT NULL DEFAULT 0, PRIMARY KEY ("id"));
-- Create index "usagehistory_history_time" to table: "usage_history"
CREATE INDEX "usagehistory_history_time" ON "usage_history" ("history_time");
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "organization_id" character varying NULL, ADD CONSTRAINT "files_organizations_files" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
