-- Modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "catalog_entity_id", DROP COLUMN "catalog_entity_key", ADD COLUMN "catalog_id" character varying NULL, ADD COLUMN "catalog_key" character varying NULL, ADD COLUMN "raw_ingest_payload" jsonb NULL;
