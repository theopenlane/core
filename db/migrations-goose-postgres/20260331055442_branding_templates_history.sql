-- +goose Up
-- modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "email_branding_id", ALTER COLUMN "template_context" SET NOT NULL;

-- +goose Down
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" ALTER COLUMN "template_context" DROP NOT NULL, ADD COLUMN "email_branding_id" character varying NULL;
