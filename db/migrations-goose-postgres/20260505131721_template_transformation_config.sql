-- +goose Up
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "transform_configuration" jsonb NULL;

-- +goose Down
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "transform_configuration";
