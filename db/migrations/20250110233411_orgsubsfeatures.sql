-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "feature_lookup_keys" jsonb NULL;
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "feature_lookup_keys" jsonb NULL;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "billing_notifications_enabled" boolean NOT NULL DEFAULT true;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "billing_notifications_enabled" boolean NOT NULL DEFAULT true;
