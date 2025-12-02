-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "identity_provider_auth_tested" boolean NOT NULL DEFAULT false;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "identity_provider_auth_tested" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "identity_provider_auth_tested";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "identity_provider_auth_tested";
