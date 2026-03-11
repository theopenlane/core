-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "email_template_files" character varying NULL, ADD CONSTRAINT "files_email_templates_files" FOREIGN KEY ("email_template_files") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
