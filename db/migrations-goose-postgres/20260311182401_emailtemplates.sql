-- +goose Up
-- modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "email_template_files" character varying NULL, ADD CONSTRAINT "files_email_templates_files" FOREIGN KEY ("email_template_files") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_email_templates_files", DROP COLUMN "email_template_files";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" DROP COLUMN "template_context", DROP COLUMN "revision";
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" DROP COLUMN "template_context", DROP COLUMN "revision";
