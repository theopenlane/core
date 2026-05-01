-- Drop index "campaigntarget_campaign_id_email" from table: "campaign_targets"
DROP INDEX "campaigntarget_campaign_id_email";
-- Create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
CREATE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
-- Drop index "emailtemplate_key" from table: "email_templates"
DROP INDEX "emailtemplate_key";
-- Drop index "emailtemplate_owner_id_key" from table: "email_templates"
DROP INDEX "emailtemplate_owner_id_key";
-- Modify "email_templates" table
ALTER TABLE "email_templates" ALTER COLUMN "format" DROP NOT NULL, ALTER COLUMN "template_context" DROP NOT NULL;
-- Create index "emailtemplate_owner_id_key" to table: "email_templates"
CREATE INDEX "emailtemplate_owner_id_key" ON "email_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "email_branding_blocked_groups", DROP COLUMN "email_branding_editors", DROP COLUMN "email_branding_viewers";
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "campaign_email" boolean NOT NULL DEFAULT false;
-- Drop index "notificationpreference_owner_id_user_id_channel" from table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id_user_id_channel";
-- Create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
CREATE INDEX "notificationpreference_owner_id_user_id_channel" ON "notification_preferences" ("owner_id", "user_id", "channel") WHERE (deleted_at IS NULL);
-- Drop index "notificationtemplate_key" from table: "notification_templates"
DROP INDEX "notificationtemplate_key";
-- Drop index "notificationtemplate_owner_id_channel_locale_topic_pattern" from table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern";
-- Drop index "notificationtemplate_owner_id_key" from table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_key";
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" ALTER COLUMN "channel" DROP NOT NULL;
-- Create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- Create index "notificationtemplate_owner_id_key" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_key" ON "notification_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- Modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_email_brandings_campaigns", ADD COLUMN "integration_id" character varying NULL, ADD CONSTRAINT "campaigns_integrations_campaigns" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
