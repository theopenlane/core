-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "aliases" jsonb NULL;

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "aliases";
