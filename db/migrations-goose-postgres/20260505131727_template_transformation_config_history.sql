-- +goose Up
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "transform_configuration" jsonb NULL;

-- +goose Down
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "transform_configuration";
