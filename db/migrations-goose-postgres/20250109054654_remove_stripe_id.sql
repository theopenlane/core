-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "stripe_id";
-- modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "stripe_id";

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "stripe_id" character varying NULL;
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "stripe_id" character varying NULL;
