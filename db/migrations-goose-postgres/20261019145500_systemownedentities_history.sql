-- +goose Up
-- rename catalog pointer columns on "entity_history" table
ALTER TABLE "entity_history" RENAME COLUMN "catalog_entity_id" TO "catalog_id";
ALTER TABLE "entity_history" RENAME COLUMN "catalog_entity_key" TO "catalog_key";
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "raw_ingest_payload" jsonb NULL;

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "raw_ingest_payload";
-- reverse: rename catalog pointer columns on "entity_history" table
ALTER TABLE "entity_history" RENAME COLUMN "catalog_key" TO "catalog_entity_key";
ALTER TABLE "entity_history" RENAME COLUMN "catalog_id" TO "catalog_entity_id";
