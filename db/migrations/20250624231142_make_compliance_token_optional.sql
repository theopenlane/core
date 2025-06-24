-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ALTER COLUMN "compliance_webhook_token" DROP NOT NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ALTER COLUMN "compliance_webhook_token" DROP NOT NULL;
