-- +goose Up
-- modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "sso_exempt" boolean NULL DEFAULT false;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "identity_provider_exempt_domains" jsonb NULL;

-- +goose Down
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "identity_provider_exempt_domains";
-- reverse: modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "sso_exempt";
