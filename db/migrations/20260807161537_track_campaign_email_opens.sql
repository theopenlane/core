-- Modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD COLUMN "email_opened_at" timestamptz NULL, ADD COLUMN "email_clicked_at" timestamptz NULL, ADD COLUMN "email_open_count" bigint NULL DEFAULT 0, ADD COLUMN "email_click_count" bigint NULL DEFAULT 0;
