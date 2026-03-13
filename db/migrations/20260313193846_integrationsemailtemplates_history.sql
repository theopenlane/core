-- Modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "definition_id" character varying NULL, ADD COLUMN "definition_version" character varying NULL, ADD COLUMN "definition_slug" character varying NULL, ADD COLUMN "family" character varying NULL, ADD COLUMN "status" character varying NOT NULL DEFAULT 'PENDING', ADD COLUMN "provider_metadata_snapshot" jsonb NULL;
-- Modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
