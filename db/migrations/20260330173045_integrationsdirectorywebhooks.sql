-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "first_seen_at" timestamptz NULL, ADD COLUMN "last_seen_at" timestamptz NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL, ADD COLUMN "metadata" jsonb NULL;
-- Create index "directoryaccount_directory_instance_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_instance_id_canonical_email" ON "directory_accounts" ("directory_instance_id", "canonical_email");
-- Create index "directoryaccount_directory_instance_id_external_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_instance_id_external_id" ON "directory_accounts" ("directory_instance_id", "external_id");
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "first_seen_at" timestamptz NULL, ADD COLUMN "last_seen_at" timestamptz NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL, ADD COLUMN "metadata" jsonb NULL;
-- Create index "directorygroup_directory_instance_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_instance_id_email" ON "directory_groups" ("directory_instance_id", "email");
-- Create index "directorygroup_directory_instance_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_instance_id_external_id" ON "directory_groups" ("directory_instance_id", "external_id");
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL;
-- Create index "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" to table: "directory_memberships"
CREATE INDEX "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" ON "directory_memberships" ("directory_instance_id", "directory_account_id", "directory_group_id");
-- Modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD COLUMN "directory_instance_id" character varying NULL;
-- Create index "directorysyncrun_directory_instance_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_directory_instance_id_started_at" ON "directory_sync_runs" ("directory_instance_id", "started_at");
-- Modify "entities" table
ALTER TABLE "entities" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Drop index "integrationwebhook_owner_id_provider_external_event_id" from table: "integration_webhooks"
DROP INDEX "integrationwebhook_owner_id_provider_external_event_id";
-- Modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD COLUMN "endpoint_id" character varying NULL;
-- Create index "integrationwebhook_endpoint_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_endpoint_id" ON "integration_webhooks" ("endpoint_id") WHERE ((deleted_at IS NULL) AND (endpoint_id IS NOT NULL));
-- Create index "integrationwebhook_integration_id_name_external_event_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_integration_id_name_external_event_id" ON "integration_webhooks" ("integration_id", "name", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "installation_metadata" jsonb NULL;
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "destinations" jsonb NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
