-- Modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "integration_id" character varying NULL;
-- Modify "email_template_history" table
ALTER TABLE "email_template_history" ALTER COLUMN "format" DROP NOT NULL, ALTER COLUMN "template_context" DROP NOT NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "campaign_email" boolean NOT NULL DEFAULT false;
-- Modify "notification_template_history" table
ALTER TABLE "notification_template_history" ALTER COLUMN "channel" DROP NOT NULL;
