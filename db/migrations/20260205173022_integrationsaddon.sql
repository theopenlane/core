-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "operation_config" jsonb NULL, ADD COLUMN "mapping_version" character varying NULL;
-- Modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD COLUMN "provider" character varying NOT NULL, ADD COLUMN "external_event_id" character varying NULL;
-- Create index "integrationwebhook_owner_id_provider_external_event_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_owner_id_provider_external_event_id" ON "integration_webhooks" ("owner_id", "provider", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "config" jsonb NULL, ADD COLUMN "provider_state" jsonb NULL;
