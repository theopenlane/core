-- +goose Up
-- modify "exports" table
ALTER TABLE "exports" ADD COLUMN "mode" character varying NOT NULL DEFAULT 'FLAT', ADD COLUMN "export_metadata" jsonb NULL;

-- +goose Down
-- reverse: modify "exports" table
ALTER TABLE "exports" DROP COLUMN "export_metadata", DROP COLUMN "mode";
