-- +goose Up
-- drop index "campaigntarget_campaign_id_email" from table: "campaign_targets"
DROP INDEX "campaigntarget_campaign_id_email";
-- create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
CREATE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
-- drop index "emailtemplate_key" from table: "email_templates"
DROP INDEX "emailtemplate_key";
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "campaign_email" boolean NOT NULL DEFAULT false;
-- drop index "notificationpreference_owner_id_user_id_channel" from table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id_user_id_channel";
-- create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
CREATE INDEX "notificationpreference_owner_id_user_id_channel" ON "notification_preferences" ("owner_id", "user_id", "channel") WHERE (deleted_at IS NULL);
-- drop index "notificationtemplate_key" from table: "notification_templates"
DROP INDEX "notificationtemplate_key";
-- drop index "notificationtemplate_owner_id_channel_locale_topic_pattern" from table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern";
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ALTER COLUMN "channel" DROP NOT NULL;
-- create index "notificationtemplate_key" to table: "notification_templates"
CREATE INDEX "notificationtemplate_key" ON "notification_templates" ("key") WHERE ((deleted_at IS NULL) AND (system_owned = true));
-- create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "integration_id" character varying NULL, ADD CONSTRAINT "campaigns_integrations_campaigns" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_integrations_campaigns", DROP COLUMN "integration_id";
-- reverse: create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern";
-- reverse: create index "notificationtemplate_key" to table: "notification_templates"
DROP INDEX "notificationtemplate_key";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" ALTER COLUMN "channel" SET NOT NULL;
-- reverse: drop index "notificationtemplate_owner_id_channel_locale_topic_pattern" from table: "notification_templates"
CREATE UNIQUE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- reverse: drop index "notificationtemplate_key" from table: "notification_templates"
CREATE UNIQUE INDEX "notificationtemplate_key" ON "notification_templates" ("key") WHERE ((deleted_at IS NULL) AND (system_owned = true));
-- reverse: create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id_user_id_channel";
-- reverse: drop index "notificationpreference_owner_id_user_id_channel" from table: "notification_preferences"
CREATE UNIQUE INDEX "notificationpreference_owner_id_user_id_channel" ON "notification_preferences" ("owner_id", "user_id", "channel") WHERE (deleted_at IS NULL);
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "campaign_email";
-- reverse: drop index "emailtemplate_key" from table: "email_templates"
CREATE UNIQUE INDEX "emailtemplate_key" ON "email_templates" ("key") WHERE ((deleted_at IS NULL) AND (system_owned = true));
-- reverse: create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
DROP INDEX "campaigntarget_campaign_id_email";
-- reverse: drop index "campaigntarget_campaign_id_email" from table: "campaign_targets"
CREATE UNIQUE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
