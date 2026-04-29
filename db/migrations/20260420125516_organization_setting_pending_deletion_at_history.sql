-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "pending_deletion_at" timestamptz NULL;
