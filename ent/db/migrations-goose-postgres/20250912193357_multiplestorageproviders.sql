-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "storage_region" character varying NULL, ADD COLUMN "storage_provider" character varying NULL, ADD COLUMN "last_accessed_at" timestamptz NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "storage_region" character varying NULL, ADD COLUMN "storage_provider" character varying NULL, ADD COLUMN "last_accessed_at" timestamptz NULL, ADD COLUMN "integration_files" character varying NULL;
-- modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "credential_set" jsonb NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "last_used_at" timestamptz NULL, ADD COLUMN "expires_at" timestamptz NULL;
-- modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "credential_set" jsonb NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "last_used_at" timestamptz NULL, ADD COLUMN "expires_at" timestamptz NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "integration_type" character varying NULL, ADD COLUMN "metadata" jsonb NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "integration_type" character varying NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "file_integrations" character varying NULL;
-- create "file_secrets" table
CREATE TABLE "file_secrets" ("file_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("file_id", "hush_id"));
-- modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_integrations_files" FOREIGN KEY ("integration_files") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_files_integrations" FOREIGN KEY ("file_integrations") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "file_secrets" table
ALTER TABLE "file_secrets" ADD CONSTRAINT "file_secrets_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "file_secrets" table
ALTER TABLE "file_secrets" DROP CONSTRAINT "file_secrets_hush_id", DROP CONSTRAINT "file_secrets_file_id";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_files_integrations";
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_integrations_files";
-- reverse: create "file_secrets" table
DROP TABLE "file_secrets";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "file_integrations", DROP COLUMN "metadata", DROP COLUMN "integration_type";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "metadata", DROP COLUMN "integration_type";
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "expires_at", DROP COLUMN "last_used_at", DROP COLUMN "metadata", DROP COLUMN "credential_set";
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "expires_at", DROP COLUMN "last_used_at", DROP COLUMN "metadata", DROP COLUMN "credential_set";
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "integration_files", DROP COLUMN "last_accessed_at", DROP COLUMN "storage_provider", DROP COLUMN "storage_region", DROP COLUMN "metadata";
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "last_accessed_at", DROP COLUMN "storage_provider", DROP COLUMN "storage_region", DROP COLUMN "metadata";
