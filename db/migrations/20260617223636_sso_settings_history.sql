-- Modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "sso_exempt" boolean NULL DEFAULT false;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "identity_provider_exempt_domains" jsonb NULL;
