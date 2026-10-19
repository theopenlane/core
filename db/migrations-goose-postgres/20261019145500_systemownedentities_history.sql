-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "catalog_entity_id", DROP COLUMN "catalog_entity_key", ADD COLUMN "catalog_id" character varying NULL, ADD COLUMN "catalog_key" character varying NULL, ADD COLUMN "raw_ingest_payload" jsonb NULL;

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "raw_ingest_payload", DROP COLUMN "catalog_key", DROP COLUMN "catalog_id", ADD COLUMN "catalog_entity_key" character varying NULL, ADD COLUMN "catalog_entity_id" character varying NULL;
