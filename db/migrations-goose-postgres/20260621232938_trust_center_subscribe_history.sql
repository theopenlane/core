-- +goose Up
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "trust_center_id" character varying NULL;
-- modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ADD COLUMN "subscriber_id" character varying NULL;
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "trust_center_id" character varying NULL;
-- modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "notify_subscribers" boolean NULL DEFAULT false, ADD COLUMN "notified_at" timestamptz NULL;
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "notify_subscribers_on_subprocessor_change" boolean NULL DEFAULT false, ADD COLUMN "subprocessors_notified_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "subprocessors_notified_at", DROP COLUMN "notify_subscribers_on_subprocessor_change";
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "notified_at", DROP COLUMN "notify_subscribers";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "trust_center_id";
-- reverse: modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" DROP COLUMN "subscriber_id";
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "trust_center_id";
