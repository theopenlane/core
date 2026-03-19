-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "logo_file_id" character varying NULL;

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "logo_file_id";
