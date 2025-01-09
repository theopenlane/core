-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "billing_address";
ALTER TABLE "organization_setting_history" ADD COLUMN "billing_address" jsonb;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "billing_address";
ALTER TABLE "organization_settings" ADD COLUMN "billing_address" jsonb;
