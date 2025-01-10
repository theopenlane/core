-- +goose Up
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "feature_lookup_keys" jsonb NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "feature_lookup_keys" jsonb NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "email_notifications_enabled" boolean NOT NULL DEFAULT false;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "email_notifications_enabled" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "email_notifications_enabled";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "email_notifications_enabled";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "feature_lookup_keys";
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "feature_lookup_keys";
