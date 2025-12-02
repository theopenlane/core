-- +goose Up
-- modify "exports" table
ALTER TABLE "exports" ADD COLUMN "fields" jsonb NULL;

-- +goose Down
-- reverse: modify "exports" table
ALTER TABLE "exports" DROP COLUMN "fields";
