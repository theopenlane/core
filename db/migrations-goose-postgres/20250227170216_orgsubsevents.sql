-- +goose Up
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- modify "event_history" table
ALTER TABLE "event_history" ALTER COLUMN "event_id" SET NOT NULL, ALTER COLUMN "event_type" DROP NOT NULL, ADD COLUMN "source" character varying NULL, ADD COLUMN "additional_processing_required" boolean NULL DEFAULT false, ADD COLUMN "additional_processing_details" character varying NULL, ADD COLUMN "processed_by" character varying NULL, ADD COLUMN "processed_at" timestamptz NULL;
-- modify "events" table
ALTER TABLE "events" ALTER COLUMN "event_id" SET NOT NULL, ALTER COLUMN "event_type" DROP NOT NULL, ADD COLUMN "source" character varying NULL, ADD COLUMN "additional_processing_required" boolean NULL DEFAULT false, ADD COLUMN "additional_processing_details" character varying NULL, ADD COLUMN "processed_by" character varying NULL, ADD COLUMN "processed_at" timestamptz NULL;
-- create index "events_event_id_key" to table: "events"
CREATE UNIQUE INDEX "events_event_id_key" ON "events" ("event_id");
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "trial_expires_at" timestamptz NULL, ADD COLUMN "days_until_due" character varying NULL, ADD COLUMN "payment_method_added" boolean NULL;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "trial_expires_at" timestamptz NULL, ADD COLUMN "days_until_due" character varying NULL, ADD COLUMN "payment_method_added" boolean NULL;
-- create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"), CONSTRAINT "org_subscription_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "org_subscription_events_org_subscription_id" FOREIGN KEY ("org_subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "org_subscription_events" table
DROP TABLE "org_subscription_events";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "payment_method_added", DROP COLUMN "days_until_due", DROP COLUMN "trial_expires_at";
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "revoked_at", DROP COLUMN "revoked_by", DROP COLUMN "revoked_reason", DROP COLUMN "is_active";
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "payment_method_added", DROP COLUMN "days_until_due", DROP COLUMN "trial_expires_at";
-- reverse: create index "events_event_id_key" to table: "events"
DROP INDEX "events_event_id_key";
-- reverse: modify "events" table
ALTER TABLE "events" DROP COLUMN "processed_at", DROP COLUMN "processed_by", DROP COLUMN "additional_processing_details", DROP COLUMN "additional_processing_required", DROP COLUMN "source", ALTER COLUMN "event_type" SET NOT NULL, ALTER COLUMN "event_id" DROP NOT NULL;
-- reverse: modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "processed_at", DROP COLUMN "processed_by", DROP COLUMN "additional_processing_details", DROP COLUMN "additional_processing_required", DROP COLUMN "source", ALTER COLUMN "event_type" SET NOT NULL, ALTER COLUMN "event_id" DROP NOT NULL;
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "revoked_at", DROP COLUMN "revoked_by", DROP COLUMN "revoked_reason", DROP COLUMN "is_active";
