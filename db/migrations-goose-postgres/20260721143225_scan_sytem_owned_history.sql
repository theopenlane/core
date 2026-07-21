-- +goose Up
-- modify "scan_history" table
ALTER TABLE "scan_history" ALTER COLUMN "status" SET DEFAULT 'PENDING', ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;

-- +goose Down
-- reverse: modify "scan_history" table
ALTER TABLE "scan_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned", ALTER COLUMN "status" SET DEFAULT 'PROCESSING';
