-- +goose Up
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "purpose" character varying NOT NULL DEFAULT 'TRUSTCENTER';
-- create index "customdomain_purpose_domain_type_owner_id" to table: "custom_domains"
CREATE UNIQUE INDEX "customdomain_purpose_domain_type_owner_id" ON "custom_domains" ("purpose", "domain_type", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "entity_external_id_owner_id" from table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- create index "entity_external_id_owner_id" to table: "entities"
CREATE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_id IS NOT NULL));

-- +goose Down
-- reverse: create index "entity_external_id_owner_id" to table: "entities"
DROP INDEX "entity_external_id_owner_id";
-- reverse: drop index "entity_external_id_owner_id" from table: "entities"
CREATE UNIQUE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_id IS NOT NULL));
-- reverse: create index "customdomain_purpose_domain_type_owner_id" to table: "custom_domains"
DROP INDEX "customdomain_purpose_domain_type_owner_id";
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "purpose";
