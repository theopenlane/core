-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "allowed_email_domains" jsonb NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "allowed_email_domains" jsonb NULL;
