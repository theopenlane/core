-- +goose Up
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "notify_subscribers" boolean NULL DEFAULT false, ADD COLUMN "notified_at" timestamptz NULL;
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "notify_subscribers_on_subprocessor_change" boolean NULL DEFAULT false, ADD COLUMN "subprocessors_notified_at" timestamptz NULL;
-- drop index "subscriber_email_owner_id" from table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "contact_id" character varying NULL, ADD COLUMN "trust_center_id" character varying NULL, ADD COLUMN "user_id" character varying NULL, ADD CONSTRAINT "subscribers_contacts_subscribers" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_trust_centers_subscribers" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_users_subscribers" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NULL));
-- create index "subscriber_email_trust_center_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_trust_center_id" ON "subscribers" ("email", "trust_center_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NOT NULL));
-- modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD COLUMN "subscriber_id" character varying NULL, ADD CONSTRAINT "campaign_targets_subscribers_campaign_targets" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "campaigntarget_subscriber_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_subscriber_id" ON "campaign_targets" ("subscriber_id");
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "campaigns_trust_centers_campaigns" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "email_templates_trust_centers_email_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" DROP CONSTRAINT "email_templates_trust_centers_email_templates", DROP COLUMN "trust_center_id";
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_trust_centers_campaigns", DROP COLUMN "trust_center_id";
-- reverse: create index "campaigntarget_subscriber_id" to table: "campaign_targets"
DROP INDEX "campaigntarget_subscriber_id";
-- reverse: modify "campaign_targets" table
ALTER TABLE "campaign_targets" DROP CONSTRAINT "campaign_targets_subscribers_campaign_targets", DROP COLUMN "subscriber_id";
-- reverse: create index "subscriber_email_trust_center_id" to table: "subscribers"
DROP INDEX "subscriber_email_trust_center_id";
-- reverse: create index "subscriber_email_owner_id" to table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP CONSTRAINT "subscribers_users_subscribers", DROP CONSTRAINT "subscribers_trust_centers_subscribers", DROP CONSTRAINT "subscribers_contacts_subscribers", DROP COLUMN "user_id", DROP COLUMN "trust_center_id", DROP COLUMN "contact_id";
-- reverse: drop index "subscriber_email_owner_id" from table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false));
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "subprocessors_notified_at", DROP COLUMN "notify_subscribers_on_subprocessor_change";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "notified_at", DROP COLUMN "notify_subscribers";
