-- +goose Up
-- drop index "campaigntarget_campaign_id_email" from table: "campaign_targets"
DROP INDEX "campaigntarget_campaign_id_email";
-- create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
CREATE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
-- drop index "emailtemplate_key" from table: "email_templates"
DROP INDEX "emailtemplate_key";
-- drop index "emailtemplate_owner_id_key" from table: "email_templates"
DROP INDEX "emailtemplate_owner_id_key";
-- modify "email_templates" table
ALTER TABLE "email_templates" ALTER COLUMN "format" DROP NOT NULL, ALTER COLUMN "template_context" DROP NOT NULL;
-- create index "emailtemplate_owner_id_key" to table: "email_templates"
CREATE INDEX "emailtemplate_owner_id_key" ON "email_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "email_branding_blocked_groups", DROP COLUMN "email_branding_editors", DROP COLUMN "email_branding_viewers";
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
-- drop index "notificationtemplate_owner_id_key" from table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_key";
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ALTER COLUMN "channel" DROP NOT NULL;
-- create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- create index "notificationtemplate_owner_id_key" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_key" ON "notification_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_email_brandings_campaigns", ADD COLUMN "integration_id" character varying NULL, ADD CONSTRAINT "campaigns_integrations_campaigns" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_integrations_campaigns", DROP COLUMN "integration_id", ADD CONSTRAINT "campaigns_email_brandings_campaigns" FOREIGN KEY ("email_branding_id") REFERENCES "email_brandings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: create index "notificationtemplate_owner_id_key" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_key";
-- reverse: create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" ALTER COLUMN "channel" SET NOT NULL;
-- reverse: drop index "notificationtemplate_owner_id_key" from table: "notification_templates"
CREATE UNIQUE INDEX "notificationtemplate_owner_id_key" ON "notification_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
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
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "email_branding_viewers" character varying NULL, ADD COLUMN "email_branding_editors" character varying NULL, ADD COLUMN "email_branding_blocked_groups" character varying NULL;
-- reverse: create index "emailtemplate_owner_id_key" to table: "email_templates"
DROP INDEX "emailtemplate_owner_id_key";
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" ALTER COLUMN "template_context" SET NOT NULL, ALTER COLUMN "format" SET NOT NULL;
-- reverse: drop index "emailtemplate_owner_id_key" from table: "email_templates"
CREATE UNIQUE INDEX "emailtemplate_owner_id_key" ON "email_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- reverse: drop index "emailtemplate_key" from table: "email_templates"
CREATE UNIQUE INDEX "emailtemplate_key" ON "email_templates" ("key") WHERE ((deleted_at IS NULL) AND (system_owned = true));
-- reverse: create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
DROP INDEX "campaigntarget_campaign_id_email";
-- reverse: drop index "campaigntarget_campaign_id_email" from table: "campaign_targets"
CREATE UNIQUE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
