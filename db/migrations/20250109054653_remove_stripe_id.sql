-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "stripe_id";
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "stripe_id";
