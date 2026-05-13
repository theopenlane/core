-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "logo_remote_url" character varying NULL;

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "logo_remote_url";
