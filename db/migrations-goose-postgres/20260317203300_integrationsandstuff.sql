-- +goose Up
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- drop index "integrationwebhook_owner_id_provider_external_event_id" from table: "integration_webhooks"
DROP INDEX "integrationwebhook_owner_id_provider_external_event_id";
-- create index "integrationwebhook_integration_id_name_external_event_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_integration_id_name_external_event_id" ON "integration_webhooks" ("integration_id", "name", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "destinations" jsonb NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "observed_at", DROP COLUMN "integration_id", DROP COLUMN "external_id";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" DROP COLUMN "destinations";
-- reverse: create index "integrationwebhook_integration_id_name_external_event_id" to table: "integration_webhooks"
DROP INDEX "integrationwebhook_integration_id_name_external_event_id";
-- reverse: drop index "integrationwebhook_owner_id_provider_external_event_id" from table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_owner_id_provider_external_event_id" ON "integration_webhooks" ("owner_id", "provider", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "observed_at", DROP COLUMN "external_id";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "observed_at", DROP COLUMN "integration_id", DROP COLUMN "external_id";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP COLUMN "observed_at", DROP COLUMN "integration_id";
