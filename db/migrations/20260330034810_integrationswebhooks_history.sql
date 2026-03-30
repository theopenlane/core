-- Modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "installation_metadata" jsonb NULL;
-- Modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "destinations" jsonb NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
