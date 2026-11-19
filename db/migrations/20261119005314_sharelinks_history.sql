-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "external_sharing_enabled" boolean NULL DEFAULT false;
