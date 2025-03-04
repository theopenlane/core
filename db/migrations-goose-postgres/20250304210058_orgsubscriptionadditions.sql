-- +goose Up
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "trial_expires_at" timestamptz NULL, ADD COLUMN "days_until_due" character varying NULL, ADD COLUMN "payment_method_added" boolean NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "trial_expires_at" timestamptz NULL, ADD COLUMN "days_until_due" character varying NULL, ADD COLUMN "payment_method_added" boolean NULL;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "org_subscription_events" character varying NULL, ADD CONSTRAINT "events_org_subscriptions_events" FOREIGN KEY ("org_subscription_events") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "events_org_subscriptions_events", DROP COLUMN "org_subscription_events";
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "revoked_at", DROP COLUMN "revoked_by", DROP COLUMN "revoked_reason", DROP COLUMN "is_active";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "payment_method_added", DROP COLUMN "days_until_due", DROP COLUMN "trial_expires_at";
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "payment_method_added", DROP COLUMN "days_until_due", DROP COLUMN "trial_expires_at";
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "revoked_at", DROP COLUMN "revoked_by", DROP COLUMN "revoked_reason", DROP COLUMN "is_active";
