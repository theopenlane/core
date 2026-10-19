-- +goose Up
-- drop index "entity_external_id_owner_id" from table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- drop index "entity_name_owner_id" from table: "entities"
DROP INDEX "entity_name_owner_id";
-- modify "entities" table
ALTER TABLE "entities" DROP COLUMN "catalog_entity_key", DROP COLUMN "catalog_entity_id", ADD COLUMN "catalog_key" character varying NULL, ADD COLUMN "raw_ingest_payload" jsonb NULL, ADD COLUMN "catalog_id" character varying NULL, ADD CONSTRAINT "entities_entities_adopted" FOREIGN KEY ("catalog_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "entity_external_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_id IS NOT NULL));
-- create index "entity_name_owner_id" to table: "entities"
CREATE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "entity_catalog_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_catalog_id_owner_id" ON "entities" ("catalog_id", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "entity_catalog_id_owner_id" to table: "entities"
DROP INDEX "entity_catalog_id_owner_id";
-- reverse: create index "entity_name_owner_id" to table: "entities"
DROP INDEX "entity_name_owner_id";
-- reverse: create index "entity_external_id_owner_id" to table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_entities_adopted", DROP COLUMN "catalog_id", DROP COLUMN "raw_ingest_payload", DROP COLUMN "catalog_key", ADD COLUMN "catalog_entity_id" character varying NULL, ADD COLUMN "catalog_entity_key" character varying NULL;
-- reverse: drop index "entity_name_owner_id" from table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- reverse: drop index "entity_external_id_owner_id" from table: "entities"
CREATE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
