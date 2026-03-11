-- Modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
-- Modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
