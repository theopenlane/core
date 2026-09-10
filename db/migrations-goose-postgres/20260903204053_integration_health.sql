-- +goose Up
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "health" jsonb NULL;

-- +goose Down
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "health";
