-- +goose Up
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "email_branding_id" character varying NULL, ADD COLUMN "email_template_id" character varying NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "provider_metadata" jsonb NULL;
-- modify "user_setting_history" table
ALTER TABLE "user_setting_history" ADD COLUMN "delegate_user_id" character varying NULL, ADD COLUMN "delegate_start_at" timestamptz NULL, ADD COLUMN "delegate_end_at" timestamptz NULL;
-- modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "due_at" timestamptz NULL;
-- create "email_branding_history" table
CREATE TABLE "email_branding_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "brand_name" character varying NULL, "logo_remote_url" character varying NULL, "primary_color" character varying NULL, "secondary_color" character varying NULL, "background_color" character varying NULL, "text_color" character varying NULL, "button_color" character varying NULL, "button_text_color" character varying NULL, "link_color" character varying NULL, "font_family" character varying NULL, "is_default" boolean NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "emailbrandinghistory_history_time" to table: "email_branding_history"
CREATE INDEX "emailbrandinghistory_history_time" ON "email_branding_history" ("history_time");
-- create "email_template_history" table
CREATE TABLE "email_template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "format" character varying NOT NULL DEFAULT 'HTML', "locale" character varying NOT NULL DEFAULT 'en-US', "subject_template" character varying NULL, "preheader_template" character varying NULL, "body_template" text NULL, "text_template" text NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "email_branding_id" character varying NULL, "integration_id" character varying NULL, "workflow_definition_id" character varying NULL, "workflow_instance_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "emailtemplatehistory_history_time" to table: "email_template_history"
CREATE INDEX "emailtemplatehistory_history_time" ON "email_template_history" ("history_time");
-- create "notification_preference_history" table
CREATE TABLE "notification_preference_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "user_id" character varying NOT NULL, "channel" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'ENABLED', "provider" character varying NULL, "destination" character varying NULL, "config" jsonb NULL, "enabled" boolean NOT NULL DEFAULT true, "cadence" character varying NOT NULL DEFAULT 'IMMEDIATE', "priority" character varying NULL, "topic_patterns" jsonb NULL, "topic_overrides" jsonb NULL, "template_id" character varying NULL, "mute_until" timestamptz NULL, "quiet_hours_start" character varying NULL, "quiet_hours_end" character varying NULL, "timezone" character varying NULL, "is_default" boolean NOT NULL DEFAULT false, "verified_at" timestamptz NULL, "last_used_at" timestamptz NULL, "last_error" text NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "notificationpreferencehistory_history_time" to table: "notification_preference_history"
CREATE INDEX "notificationpreferencehistory_history_time" ON "notification_preference_history" ("history_time");
-- create "notification_template_history" table
CREATE TABLE "notification_template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "channel" character varying NOT NULL, "format" character varying NOT NULL DEFAULT 'MARKDOWN', "locale" character varying NOT NULL DEFAULT 'en-US', "topic_pattern" character varying NOT NULL, "integration_id" character varying NULL, "workflow_definition_id" character varying NULL, "email_template_id" character varying NULL, "title_template" character varying NULL, "subject_template" character varying NULL, "body_template" text NULL, "blocks" jsonb NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, PRIMARY KEY ("id"));
-- create index "notificationtemplatehistory_history_time" to table: "notification_template_history"
CREATE INDEX "notificationtemplatehistory_history_time" ON "notification_template_history" ("history_time");

-- +goose Down
-- reverse: create index "notificationtemplatehistory_history_time" to table: "notification_template_history"
DROP INDEX "notificationtemplatehistory_history_time";
-- reverse: create "notification_template_history" table
DROP TABLE "notification_template_history";
-- reverse: create index "notificationpreferencehistory_history_time" to table: "notification_preference_history"
DROP INDEX "notificationpreferencehistory_history_time";
-- reverse: create "notification_preference_history" table
DROP TABLE "notification_preference_history";
-- reverse: create index "emailtemplatehistory_history_time" to table: "email_template_history"
DROP INDEX "emailtemplatehistory_history_time";
-- reverse: create "email_template_history" table
DROP TABLE "email_template_history";
-- reverse: create index "emailbrandinghistory_history_time" to table: "email_branding_history"
DROP INDEX "emailbrandinghistory_history_time";
-- reverse: create "email_branding_history" table
DROP TABLE "email_branding_history";
-- reverse: modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" DROP COLUMN "due_at";
-- reverse: modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "delegate_end_at", DROP COLUMN "delegate_start_at", DROP COLUMN "delegate_user_id";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "provider_metadata";
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "email_template_id", DROP COLUMN "email_branding_id";
