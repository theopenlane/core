-- +goose Up
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL;
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL;
-- modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD COLUMN "business_owner_identity_holder_id" character varying NULL, ADD COLUMN "technical_owner_identity_holder_id" character varying NULL, ADD COLUMN "security_owner_identity_holder_id" character varying NULL;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL;

-- +goose Down
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "assigned_to_identity_holder_id", DROP COLUMN "reviewed_by_identity_holder_id";
-- reverse: modify "platform_history" table
ALTER TABLE "platform_history" DROP COLUMN "security_owner_identity_holder_id", DROP COLUMN "technical_owner_identity_holder_id", DROP COLUMN "business_owner_identity_holder_id", DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "assigned_to_identity_holder_id", DROP COLUMN "reviewed_by_identity_holder_id";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "reviewed_by_identity_holder_id", DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "internal_owner_identity_holder_id";
