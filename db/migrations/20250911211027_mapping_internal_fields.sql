-- Modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "internal_id" character varying NULL;
-- Modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "internal_id" character varying NULL;
