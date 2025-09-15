-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "identity_provider_auth_tested" boolean NOT NULL DEFAULT false;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "identity_provider_auth_tested" boolean NOT NULL DEFAULT false;
