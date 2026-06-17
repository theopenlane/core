-- +goose Up
-- modify "org_memberships" table
ALTER TABLE "org_memberships" ADD COLUMN "sso_exempt" boolean NULL DEFAULT false;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "identity_provider_exempt_domains" jsonb NULL;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "identity_provider_exempt_domains";
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "sso_exempt";
