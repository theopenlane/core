-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "pending_deletion_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "pending_deletion_at";
