-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "pending_deletion_at" timestamptz NULL;
