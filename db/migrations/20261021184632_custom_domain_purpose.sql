-- Modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "purpose" character varying NULL;
-- Drop index "entity_external_id_owner_id" from table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- Create index "entity_external_id_owner_id" to table: "entities"
CREATE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_id IS NOT NULL));
