-- +goose Up
-- modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ADD COLUMN "email_opened_at" timestamptz NULL, ADD COLUMN "email_clicked_at" timestamptz NULL, ADD COLUMN "email_open_count" bigint NULL DEFAULT 0, ADD COLUMN "email_click_count" bigint NULL DEFAULT 0;

-- +goose Down
-- reverse: modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" DROP COLUMN "email_click_count", DROP COLUMN "email_open_count", DROP COLUMN "email_clicked_at", DROP COLUMN "email_opened_at";
