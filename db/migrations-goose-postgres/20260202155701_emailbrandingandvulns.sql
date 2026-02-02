-- +goose Up
-- modify "user_settings" table
ALTER TABLE "user_settings" ADD COLUMN "delegate_user_id" character varying NULL, ADD COLUMN "delegate_start_at" timestamptz NULL, ADD COLUMN "delegate_end_at" timestamptz NULL;
-- modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "due_at" timestamptz NULL;
-- create "email_brandings" table
CREATE TABLE "email_brandings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "brand_name" character varying NULL, "logo_remote_url" character varying NULL, "primary_color" character varying NULL, "secondary_color" character varying NULL, "background_color" character varying NULL, "text_color" character varying NULL, "button_color" character varying NULL, "button_text_color" character varying NULL, "link_color" character varying NULL, "font_family" character varying NULL, "is_default" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "email_brandings_organizations_email_brandings" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "emailbranding_owner_id" to table: "email_brandings"
CREATE INDEX "emailbranding_owner_id" ON "email_brandings" ("owner_id") WHERE (deleted_at IS NULL);
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "provider_metadata" jsonb NULL;
-- create index "integration_owner_id_kind" to table: "integrations"
CREATE UNIQUE INDEX "integration_owner_id_kind" ON "integrations" ("owner_id", "kind") WHERE (deleted_at IS NULL);
-- create "email_templates" table
CREATE TABLE "email_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "format" character varying NOT NULL DEFAULT 'HTML', "locale" character varying NOT NULL DEFAULT 'en-US', "subject_template" character varying NULL, "preheader_template" character varying NULL, "body_template" text NULL, "text_template" text NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "email_branding_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "workflow_definition_id" character varying NULL, "workflow_instance_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "email_templates_email_brandings_email_templates" FOREIGN KEY ("email_branding_id") REFERENCES "email_brandings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "email_templates_integrations_email_templates" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "email_templates_organizations_email_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "email_templates_workflow_definitions_email_templates" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "email_templates_workflow_instances_email_templates" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "emailtemplate_key" to table: "email_templates"
CREATE UNIQUE INDEX "emailtemplate_key" ON "email_templates" ("key") WHERE ((deleted_at IS NULL) AND (system_owned = true));
-- create index "emailtemplate_owner_id" to table: "email_templates"
CREATE INDEX "emailtemplate_owner_id" ON "email_templates" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "emailtemplate_owner_id_key" to table: "email_templates"
CREATE UNIQUE INDEX "emailtemplate_owner_id_key" ON "email_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "email_branding_id" character varying NULL, ADD COLUMN "email_template_id" character varying NULL, ADD CONSTRAINT "campaigns_email_brandings_campaigns" FOREIGN KEY ("email_branding_id") REFERENCES "email_brandings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_email_templates_campaigns" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "integration_runs" table
CREATE TABLE "integration_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "operation_name" character varying NULL, "operation_kind" character varying NULL, "run_type" character varying NULL, "status" character varying NULL, "started_at" timestamptz NOT NULL, "finished_at" timestamptz NULL, "duration_ms" bigint NULL, "summary" character varying NULL, "error" text NULL, "metrics" jsonb NULL, "integration_id" character varying NULL, "request_file_id" character varying NULL, "response_file_id" character varying NULL, "event_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "integration_runs_events_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "integration_runs_files_request_file" FOREIGN KEY ("request_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "integration_runs_files_response_file" FOREIGN KEY ("response_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "integration_runs_integrations_integration_runs" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "integration_runs_organizations_integration_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "integrationrun_integration_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_integration_id_started_at" ON "integration_runs" ("integration_id", "started_at") WHERE (deleted_at IS NULL);
-- create index "integrationrun_owner_id" to table: "integration_runs"
CREATE INDEX "integrationrun_owner_id" ON "integration_runs" ("owner_id") WHERE (deleted_at IS NULL);
-- create "integration_webhooks" table
CREATE TABLE "integration_webhooks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NULL, "status" character varying NULL, "endpoint_url" character varying NULL, "secret_token" character varying NULL, "allowed_events" jsonb NULL, "last_delivery_id" character varying NULL, "last_delivery_at" timestamptz NULL, "last_delivery_status" character varying NULL, "last_delivery_error" text NULL, "metadata" jsonb NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "integration_webhooks_integrations_integration_webhooks" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "integration_webhooks_organizations_integration_webhooks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "integrationwebhook_owner_id" to table: "integration_webhooks"
CREATE INDEX "integrationwebhook_owner_id" ON "integration_webhooks" ("owner_id") WHERE (deleted_at IS NULL);
-- create "notification_templates" table
CREATE TABLE "notification_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "channel" character varying NOT NULL, "format" character varying NOT NULL DEFAULT 'MARKDOWN', "locale" character varying NOT NULL DEFAULT 'en-US', "topic_pattern" character varying NOT NULL, "title_template" character varying NULL, "subject_template" character varying NULL, "body_template" text NULL, "blocks" jsonb NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "email_template_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "workflow_definition_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "notification_templates_email_templates_notification_templates" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notification_templates_integrations_notification_templates" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notification_templates_organizations_notification_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notification_templates_workflo_439a17f2830fbf868eeb61d3d3fdac37" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "notificationtemplate_key" to table: "notification_templates"
CREATE UNIQUE INDEX "notificationtemplate_key" ON "notification_templates" ("key") WHERE ((deleted_at IS NULL) AND (system_owned = true));
-- create index "notificationtemplate_owner_id" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id" ON "notification_templates" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
CREATE UNIQUE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- create index "notificationtemplate_owner_id_key" to table: "notification_templates"
CREATE UNIQUE INDEX "notificationtemplate_owner_id_key" ON "notification_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- create "notification_preferences" table
CREATE TABLE "notification_preferences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "channel" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'ENABLED', "provider" character varying NULL, "destination" character varying NULL, "config" jsonb NULL, "enabled" boolean NOT NULL DEFAULT true, "cadence" character varying NOT NULL DEFAULT 'IMMEDIATE', "priority" character varying NULL, "topic_patterns" jsonb NULL, "topic_overrides" jsonb NULL, "mute_until" timestamptz NULL, "quiet_hours_start" character varying NULL, "quiet_hours_end" character varying NULL, "timezone" character varying NULL, "is_default" boolean NOT NULL DEFAULT false, "verified_at" timestamptz NULL, "last_used_at" timestamptz NULL, "last_error" text NULL, "metadata" jsonb NULL, "user_id" character varying NOT NULL, "template_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "notification_preferences_notif_aabd0a3ca9e335110ce7c2348e4f4cf0" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notification_preferences_organizations_notification_preferences" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notification_preferences_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "notificationpreference_owner_id" to table: "notification_preferences"
CREATE INDEX "notificationpreference_owner_id" ON "notification_preferences" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
CREATE UNIQUE INDEX "notificationpreference_owner_id_user_id_channel" ON "notification_preferences" ("owner_id", "user_id", "channel") WHERE (deleted_at IS NULL);
-- modify "notifications" table
ALTER TABLE "notifications" ADD COLUMN "template_id" character varying NULL, ADD COLUMN "notification_template_notifications" character varying NULL, ADD CONSTRAINT "notifications_notification_templates_notifications" FOREIGN KEY ("notification_template_notifications") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "notifications" table
ALTER TABLE "notifications" DROP CONSTRAINT "notifications_notification_templates_notifications", DROP COLUMN "notification_template_notifications", DROP COLUMN "template_id";
-- reverse: create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id_user_id_channel";
-- reverse: create index "notificationpreference_owner_id" to table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id";
-- reverse: create "notification_preferences" table
DROP TABLE "notification_preferences";
-- reverse: create index "notificationtemplate_owner_id_key" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_key";
-- reverse: create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern";
-- reverse: create index "notificationtemplate_owner_id" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id";
-- reverse: create index "notificationtemplate_key" to table: "notification_templates"
DROP INDEX "notificationtemplate_key";
-- reverse: create "notification_templates" table
DROP TABLE "notification_templates";
-- reverse: create index "integrationwebhook_owner_id" to table: "integration_webhooks"
DROP INDEX "integrationwebhook_owner_id";
-- reverse: create "integration_webhooks" table
DROP TABLE "integration_webhooks";
-- reverse: create index "integrationrun_owner_id" to table: "integration_runs"
DROP INDEX "integrationrun_owner_id";
-- reverse: create index "integrationrun_integration_id_started_at" to table: "integration_runs"
DROP INDEX "integrationrun_integration_id_started_at";
-- reverse: create "integration_runs" table
DROP TABLE "integration_runs";
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_email_templates_campaigns", DROP CONSTRAINT "campaigns_email_brandings_campaigns", DROP COLUMN "email_template_id", DROP COLUMN "email_branding_id";
-- reverse: create index "emailtemplate_owner_id_key" to table: "email_templates"
DROP INDEX "emailtemplate_owner_id_key";
-- reverse: create index "emailtemplate_owner_id" to table: "email_templates"
DROP INDEX "emailtemplate_owner_id";
-- reverse: create index "emailtemplate_key" to table: "email_templates"
DROP INDEX "emailtemplate_key";
-- reverse: create "email_templates" table
DROP TABLE "email_templates";
-- reverse: create index "integration_owner_id_kind" to table: "integrations"
DROP INDEX "integration_owner_id_kind";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "provider_metadata";
-- reverse: create index "emailbranding_owner_id" to table: "email_brandings"
DROP INDEX "emailbranding_owner_id";
-- reverse: create "email_brandings" table
DROP TABLE "email_brandings";
-- reverse: modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" DROP COLUMN "due_at";
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "delegate_end_at", DROP COLUMN "delegate_start_at", DROP COLUMN "delegate_user_id";
