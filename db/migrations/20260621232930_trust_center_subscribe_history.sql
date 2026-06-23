-- Modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "trust_center_id" character varying NULL;
-- Modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ADD COLUMN "subscriber_id" character varying NULL;
-- Modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "trust_center_id" character varying NULL;
-- Modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "notify_subscribers" boolean NULL DEFAULT false, ADD COLUMN "notified_at" timestamptz NULL;
-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "notify_subscribers_on_subprocessor_change" boolean NULL DEFAULT false, ADD COLUMN "subprocessors_notified_at" timestamptz NULL;
