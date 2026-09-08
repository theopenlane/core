-- Modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- Modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL;
-- Modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL;
-- Modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- Modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD COLUMN "business_owner_identity_holder_id" character varying NULL, ADD COLUMN "technical_owner_identity_holder_id" character varying NULL, ADD COLUMN "security_owner_identity_holder_id" character varying NULL;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL;
