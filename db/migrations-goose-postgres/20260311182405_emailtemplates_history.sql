-- +goose Up
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;
-- modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "template_context" character varying NULL;

-- +goose Down
-- reverse: modify "notification_template_history" table
ALTER TABLE "notification_template_history" DROP COLUMN "template_context", DROP COLUMN "revision";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "template_context", DROP COLUMN "revision";
