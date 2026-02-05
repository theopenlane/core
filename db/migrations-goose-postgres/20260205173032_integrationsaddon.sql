-- +goose Up
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "operation_config" jsonb NULL, ADD COLUMN "mapping_version" character varying NULL;
-- modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD COLUMN "provider" character varying NOT NULL, ADD COLUMN "external_event_id" character varying NULL;
-- create index "integrationwebhook_owner_id_provider_external_event_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_owner_id_provider_external_event_id" ON "integration_webhooks" ("owner_id", "provider", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "config" jsonb NULL, ADD COLUMN "provider_state" jsonb NULL;

-- +goose Down
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "provider_state", DROP COLUMN "config";
-- reverse: create index "integrationwebhook_owner_id_provider_external_event_id" to table: "integration_webhooks"
DROP INDEX "integrationwebhook_owner_id_provider_external_event_id";
-- reverse: modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" DROP COLUMN "external_event_id", DROP COLUMN "provider";
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP COLUMN "mapping_version", DROP COLUMN "operation_config";
