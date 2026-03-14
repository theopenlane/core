-- +goose Up
-- modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
-- drop index "integration_owner_id_kind" from table: "integrations"
DROP INDEX "integration_owner_id_kind";
-- drop index "integration_platform_id" from table: "integrations"
DROP INDEX "integration_platform_id";
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "definition_id" character varying NULL, ADD COLUMN "definition_version" character varying NULL, ADD COLUMN "definition_slug" character varying NULL, ADD COLUMN "family" character varying NULL, ADD COLUMN "status" character varying NOT NULL DEFAULT 'PENDING', ADD COLUMN "provider_metadata_snapshot" jsonb NULL;
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL, ADD COLUMN "defaults" jsonb NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "email_template_files" character varying NULL, ADD CONSTRAINT "files_email_templates_files" FOREIGN KEY ("email_template_files") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_email_templates_files", DROP COLUMN "email_template_files";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" DROP COLUMN "defaults", DROP COLUMN "template_context", DROP COLUMN "revision";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "provider_metadata_snapshot", DROP COLUMN "status", DROP COLUMN "family", DROP COLUMN "definition_slug", DROP COLUMN "definition_version", DROP COLUMN "definition_id";
-- reverse: drop index "integration_platform_id" from table: "integrations"
CREATE INDEX "integration_platform_id" ON "integrations" ("platform_id");
-- reverse: drop index "integration_owner_id_kind" from table: "integrations"
CREATE INDEX "integration_owner_id_kind" ON "integrations" ("owner_id", "kind") WHERE (deleted_at IS NULL);
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" DROP COLUMN "defaults", DROP COLUMN "template_context", DROP COLUMN "revision";
