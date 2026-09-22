-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "catalog_entity_id" character varying NULL, ADD CONSTRAINT "entities_entities_adopted_entities" FOREIGN KEY ("catalog_entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "entity_catalog_entity_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_catalog_entity_id_owner_id" ON "entities" ("catalog_entity_id", "owner_id") WHERE (deleted_at IS NULL);
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: create index "entity_catalog_entity_id_owner_id" to table: "entities"
DROP INDEX "entity_catalog_entity_id_owner_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_entities_adopted_entities", DROP COLUMN "catalog_entity_id";
