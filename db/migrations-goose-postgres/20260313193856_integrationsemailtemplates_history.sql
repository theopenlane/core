-- +goose Up
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "definition_id" character varying NULL, ADD COLUMN "definition_version" character varying NULL, ADD COLUMN "definition_slug" character varying NULL, ADD COLUMN "family" character varying NULL, ADD COLUMN "status" character varying NOT NULL DEFAULT 'PENDING', ADD COLUMN "provider_metadata_snapshot" jsonb NULL;
-- modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;

-- +goose Down
-- reverse: modify "notification_template_history" table
ALTER TABLE "notification_template_history" DROP COLUMN "defaults", DROP COLUMN "template_context", DROP COLUMN "revision";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "provider_metadata_snapshot", DROP COLUMN "status", DROP COLUMN "family", DROP COLUMN "definition_slug", DROP COLUMN "definition_version", DROP COLUMN "definition_id";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "defaults", DROP COLUMN "template_context", DROP COLUMN "revision";
