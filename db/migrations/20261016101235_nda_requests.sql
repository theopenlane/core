-- Modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD COLUMN "auto_approved" boolean NULL DEFAULT false;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "enable_auto_approval" boolean NULL, ADD COLUMN "auto_approval_rules" jsonb NULL;
