-- Modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "first_seen_at" timestamptz NULL, ADD COLUMN "last_seen_at" timestamptz NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL, ADD COLUMN "metadata" jsonb NULL;
-- Modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "first_seen_at" timestamptz NULL, ADD COLUMN "last_seen_at" timestamptz NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL, ADD COLUMN "metadata" jsonb NULL;
-- Modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL;
-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- Modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ALTER COLUMN "identity_holder_type" SET DEFAULT 'UNSPECIFIED';
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "installation_metadata" jsonb NULL;
-- Modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "destinations" jsonb NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
