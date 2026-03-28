-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
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
