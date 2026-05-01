-- +goose Up
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "integration_id" character varying NULL;
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ALTER COLUMN "format" DROP NOT NULL, ALTER COLUMN "template_context" DROP NOT NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "campaign_email" boolean NOT NULL DEFAULT false;
-- modify "notification_template_history" table
ALTER TABLE "notification_template_history" ALTER COLUMN "channel" DROP NOT NULL;

-- +goose Down
-- reverse: modify "notification_template_history" table
ALTER TABLE "notification_template_history" ALTER COLUMN "channel" SET NOT NULL;
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "campaign_email";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" ALTER COLUMN "template_context" SET NOT NULL, ALTER COLUMN "format" SET NOT NULL;
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "integration_id";
