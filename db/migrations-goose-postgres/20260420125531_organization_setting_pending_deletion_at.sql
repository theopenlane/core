-- +goose Up
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "pending_deletion_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "pending_deletion_at";
