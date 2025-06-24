-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ALTER COLUMN "compliance_webhook_token" DROP NOT NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ALTER COLUMN "compliance_webhook_token" DROP NOT NULL;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" ALTER COLUMN "compliance_webhook_token" SET NOT NULL;
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ALTER COLUMN "compliance_webhook_token" SET NOT NULL;
