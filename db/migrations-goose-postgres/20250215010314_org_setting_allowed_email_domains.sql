-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "allowed_email_domains" jsonb NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "allowed_email_domains" jsonb NULL;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "allowed_email_domains";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "allowed_email_domains";
