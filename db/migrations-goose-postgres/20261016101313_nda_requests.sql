-- +goose Up
-- modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD COLUMN "auto_approved" boolean NULL DEFAULT false;
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "enable_auto_approval" boolean NULL, ADD COLUMN "auto_approval_rules" jsonb NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "auto_approval_rules", DROP COLUMN "enable_auto_approval";
-- reverse: modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" DROP COLUMN "auto_approved";
