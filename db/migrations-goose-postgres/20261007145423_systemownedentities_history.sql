-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "catalog_entity_id" character varying NULL, ADD COLUMN "externally_visible" boolean NULL DEFAULT false, ADD COLUMN "catalog_entity_key" character varying NULL;

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "catalog_entity_key", DROP COLUMN "externally_visible", DROP COLUMN "catalog_entity_id";
