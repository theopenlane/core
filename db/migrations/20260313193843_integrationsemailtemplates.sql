-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
-- Drop index "integration_owner_id_kind" from table: "integrations"
DROP INDEX "integration_owner_id_kind";
-- Drop index "integration_platform_id" from table: "integrations"
DROP INDEX "integration_platform_id";
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "definition_id" character varying NULL, ADD COLUMN "definition_version" character varying NULL, ADD COLUMN "definition_slug" character varying NULL, ADD COLUMN "family" character varying NULL, ADD COLUMN "status" character varying NOT NULL DEFAULT 'PENDING', ADD COLUMN "provider_metadata_snapshot" jsonb NULL;
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "email_template_files" character varying NULL, ADD CONSTRAINT "files_email_templates_files" FOREIGN KEY ("email_template_files") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
