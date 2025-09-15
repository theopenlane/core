-- Modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "storage_region" character varying NULL, ADD COLUMN "storage_provider" character varying NULL, ADD COLUMN "last_accessed_at" timestamptz NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "storage_region" character varying NULL, ADD COLUMN "storage_provider" character varying NULL, ADD COLUMN "last_accessed_at" timestamptz NULL, ADD COLUMN "integration_files" character varying NULL;
-- Modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "credential_set" jsonb NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "last_used_at" timestamptz NULL, ADD COLUMN "expires_at" timestamptz NULL;
-- Modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "credential_set" jsonb NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "last_used_at" timestamptz NULL, ADD COLUMN "expires_at" timestamptz NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "integration_type" character varying NULL, ADD COLUMN "metadata" jsonb NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "integration_type" character varying NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "file_integrations" character varying NULL;
-- Create "file_secrets" table
CREATE TABLE "file_secrets" ("file_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("file_id", "hush_id"));
-- Modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_integrations_files" FOREIGN KEY ("integration_files") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_files_integrations" FOREIGN KEY ("file_integrations") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "file_secrets" table
ALTER TABLE "file_secrets" ADD CONSTRAINT "file_secrets_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
