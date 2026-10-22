-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "externally_visible" boolean NULL DEFAULT false, ADD COLUMN "catalog_entity_key" character varying NULL, ADD COLUMN "catalog_entity_id" character varying NULL, ADD CONSTRAINT "entities_entities_adopted_entities" FOREIGN KEY ("catalog_entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "entity_catalog_entity_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_catalog_entity_id_owner_id" ON "entities" ("catalog_entity_id", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "entity_catalog_entity_id_owner_id" to table: "entities"
DROP INDEX "entity_catalog_entity_id_owner_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_entities_adopted_entities", DROP COLUMN "catalog_entity_id", DROP COLUMN "catalog_entity_key", DROP COLUMN "externally_visible";
