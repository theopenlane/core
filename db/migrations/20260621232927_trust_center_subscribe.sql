-- Modify "notes" table
ALTER TABLE "notes" ADD COLUMN "notify_subscribers" boolean NULL DEFAULT false, ADD COLUMN "notified_at" timestamptz NULL;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "notify_subscribers_on_subprocessor_change" boolean NULL DEFAULT false, ADD COLUMN "subprocessors_notified_at" timestamptz NULL;
-- Drop index "subscriber_email_owner_id" from table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- Modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "contact_id" character varying NULL, ADD COLUMN "trust_center_id" character varying NULL, ADD COLUMN "user_id" character varying NULL, ADD CONSTRAINT "subscribers_contacts_subscribers" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_trust_centers_subscribers" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_users_subscribers" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NULL));
-- Create index "subscriber_email_trust_center_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_trust_center_id" ON "subscribers" ("email", "trust_center_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NOT NULL));
-- Modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD COLUMN "subscriber_id" character varying NULL, ADD CONSTRAINT "campaign_targets_subscribers_campaign_targets" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "campaigntarget_subscriber_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_subscriber_id" ON "campaign_targets" ("subscriber_id");
-- Modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "campaigns_trust_centers_campaigns" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "email_templates_trust_centers_email_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
