-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "catalog_entity_id" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;

-- +goose Down
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "catalog_entity_id";
