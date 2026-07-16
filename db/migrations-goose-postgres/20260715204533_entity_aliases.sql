-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "aliases" jsonb NULL;

-- +goose Down
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "aliases";
