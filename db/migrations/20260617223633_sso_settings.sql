-- Modify "org_memberships" table
ALTER TABLE "org_memberships" ADD COLUMN "sso_exempt" boolean NULL DEFAULT false;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "identity_provider_exempt_domains" jsonb NULL;
