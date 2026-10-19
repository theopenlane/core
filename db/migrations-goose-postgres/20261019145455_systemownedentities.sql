-- +goose Up
-- drop index "entity_external_id_owner_id" from table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- drop index "entity_name_owner_id" from table: "entities"
DROP INDEX "entity_name_owner_id";
-- rename catalog pointer columns on "entities" table, keeping adopted rows linked
ALTER TABLE "entities" RENAME COLUMN "catalog_entity_key" TO "catalog_key";
ALTER TABLE "entities" RENAME COLUMN "catalog_entity_id" TO "catalog_id";
ALTER TABLE "entities" RENAME CONSTRAINT "entities_entities_adopted_entities" TO "entities_entities_adopted";
ALTER INDEX "entity_catalog_entity_id_owner_id" RENAME TO "entity_catalog_id_owner_id";
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "raw_ingest_payload" jsonb NULL;
-- create index "entity_external_id_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_id IS NOT NULL));
-- create index "entity_name_owner_id" to table: "entities"
CREATE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "entity_name_owner_id" to table: "entities"
DROP INDEX "entity_name_owner_id";
-- reverse: create index "entity_external_id_owner_id" to table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "raw_ingest_payload";
-- reverse: rename catalog pointer columns on "entities" table
ALTER INDEX "entity_catalog_id_owner_id" RENAME TO "entity_catalog_entity_id_owner_id";
ALTER TABLE "entities" RENAME CONSTRAINT "entities_entities_adopted" TO "entities_entities_adopted_entities";
ALTER TABLE "entities" RENAME COLUMN "catalog_id" TO "catalog_entity_id";
ALTER TABLE "entities" RENAME COLUMN "catalog_key" TO "catalog_entity_key";
-- reverse: drop index "entity_name_owner_id" from table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- reverse: drop index "entity_external_id_owner_id" from table: "entities"
CREATE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
