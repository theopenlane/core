-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "compliance_webhook_token";

-- +goose Down
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "compliance_webhook_token" character varying NULL;
