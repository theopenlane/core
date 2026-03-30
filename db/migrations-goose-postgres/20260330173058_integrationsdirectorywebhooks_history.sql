-- +goose Up
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "first_seen_at" timestamptz NULL, ADD COLUMN "last_seen_at" timestamptz NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL, ADD COLUMN "metadata" jsonb NULL;
-- modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "first_seen_at" timestamptz NULL, ADD COLUMN "last_seen_at" timestamptz NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL, ADD COLUMN "metadata" jsonb NULL;
-- modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "directory_instance_id" character varying NULL, ADD COLUMN "added_at" timestamptz NULL, ADD COLUMN "removed_at" timestamptz NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "installation_metadata" jsonb NULL;
-- modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "destinations" jsonb NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "external_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "observed_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "observed_at", DROP COLUMN "integration_id", DROP COLUMN "external_id";
-- reverse: modify "notification_template_history" table
ALTER TABLE "notification_template_history" DROP COLUMN "destinations";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "installation_metadata";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "observed_at", DROP COLUMN "external_id";
-- reverse: modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" DROP COLUMN "removed_at", DROP COLUMN "added_at", DROP COLUMN "directory_instance_id";
-- reverse: modify "directory_group_history" table
ALTER TABLE "directory_group_history" DROP COLUMN "metadata", DROP COLUMN "removed_at", DROP COLUMN "added_at", DROP COLUMN "last_seen_at", DROP COLUMN "first_seen_at", DROP COLUMN "directory_instance_id";
-- reverse: modify "directory_account_history" table
ALTER TABLE "directory_account_history" DROP COLUMN "metadata", DROP COLUMN "removed_at", DROP COLUMN "added_at", DROP COLUMN "last_seen_at", DROP COLUMN "first_seen_at", DROP COLUMN "directory_instance_id";
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "observed_at", DROP COLUMN "integration_id", DROP COLUMN "external_id";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "observed_at", DROP COLUMN "integration_id";
