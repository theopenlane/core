-- Modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" ADD COLUMN "auto_approved" boolean NULL DEFAULT false;
-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "enable_auto_approval" boolean NULL, ADD COLUMN "auto_approval_rules" jsonb NULL;
