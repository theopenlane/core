-- +goose Up
-- modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "compliance_webhook_token";

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "compliance_webhook_token" character varying NULL;
