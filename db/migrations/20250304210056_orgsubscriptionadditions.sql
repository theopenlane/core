-- Modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "trial_expires_at" timestamptz NULL, ADD COLUMN "days_until_due" character varying NULL, ADD COLUMN "payment_method_added" boolean NULL;
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "trial_expires_at" timestamptz NULL, ADD COLUMN "days_until_due" character varying NULL, ADD COLUMN "payment_method_added" boolean NULL;
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- Modify "events" table
ALTER TABLE "events" ADD COLUMN "org_subscription_events" character varying NULL, ADD CONSTRAINT "events_org_subscriptions_events" FOREIGN KEY ("org_subscription_events") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
