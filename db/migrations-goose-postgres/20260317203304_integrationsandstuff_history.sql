-- +goose Up
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "destinations" jsonb NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "observed_at", DROP COLUMN "integration_id", DROP COLUMN "external_id";
-- reverse: modify "notification_template_history" table
ALTER TABLE "notification_template_history" DROP COLUMN "destinations";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "observed_at", DROP COLUMN "external_id";
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "observed_at", DROP COLUMN "integration_id", DROP COLUMN "external_id";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "observed_at", DROP COLUMN "integration_id";
