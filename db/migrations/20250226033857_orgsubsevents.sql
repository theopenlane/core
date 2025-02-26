-- Modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- Modify "event_history" table
ALTER TABLE "event_history" ALTER COLUMN "event_id" SET NOT NULL, ALTER COLUMN "event_type" DROP NOT NULL, ADD COLUMN "source" character varying NULL, ADD COLUMN "additional_processing_required" boolean NULL DEFAULT false, ADD COLUMN "additional_processing_details" character varying NULL, ADD COLUMN "processed_by" character varying NULL, ADD COLUMN "processed_at" timestamptz NULL;
-- Modify "events" table
ALTER TABLE "events" ALTER COLUMN "event_id" SET NOT NULL, ALTER COLUMN "event_type" DROP NOT NULL, ADD COLUMN "source" character varying NULL, ADD COLUMN "additional_processing_required" boolean NULL DEFAULT false, ADD COLUMN "additional_processing_details" character varying NULL, ADD COLUMN "processed_by" character varying NULL, ADD COLUMN "processed_at" timestamptz NULL;
-- Create index "events_event_id_key" to table: "events"
CREATE UNIQUE INDEX "events_event_id_key" ON "events" ("event_id");
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "is_active" boolean NULL DEFAULT true, ADD COLUMN "revoked_reason" character varying NULL, ADD COLUMN "revoked_by" character varying NULL, ADD COLUMN "revoked_at" timestamptz NULL;
-- Create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"), CONSTRAINT "org_subscription_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "org_subscription_events_org_subscription_id" FOREIGN KEY ("org_subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
