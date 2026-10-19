-- Rename catalog pointer columns on "entity_history" table
ALTER TABLE "entity_history" RENAME COLUMN "catalog_entity_id" TO "catalog_id";
ALTER TABLE "entity_history" RENAME COLUMN "catalog_entity_key" TO "catalog_key";
-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "raw_ingest_payload" jsonb NULL;
