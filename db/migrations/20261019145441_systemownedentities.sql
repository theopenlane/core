-- Drop index "entity_external_id_owner_id" from table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- Drop index "entity_name_owner_id" from table: "entities"
DROP INDEX "entity_name_owner_id";
-- Modify "entities" table
ALTER TABLE "entities" DROP COLUMN "catalog_entity_key", DROP COLUMN "catalog_entity_id", ADD COLUMN "catalog_key" character varying NULL, ADD COLUMN "raw_ingest_payload" jsonb NULL, ADD COLUMN "catalog_id" character varying NULL, ADD CONSTRAINT "entities_entities_adopted" FOREIGN KEY ("catalog_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "entity_external_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_id IS NOT NULL));
-- Create index "entity_name_owner_id" to table: "entities"
CREATE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "entity_catalog_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_catalog_id_owner_id" ON "entities" ("catalog_id", "owner_id") WHERE (deleted_at IS NULL);
