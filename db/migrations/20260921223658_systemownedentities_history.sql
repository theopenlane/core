-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "catalog_entity_id" character varying NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
