-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "catalog_entity_id" character varying NULL, ADD COLUMN "externally_visible" boolean NULL DEFAULT false, ADD COLUMN "catalog_entity_key" character varying NULL;
