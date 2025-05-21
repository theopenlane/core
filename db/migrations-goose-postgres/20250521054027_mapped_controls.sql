-- +goose Up
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "confidence" character varying NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL';
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "confidence" character varying NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL';
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "mapped_control_from_control" character varying NULL, ADD COLUMN "mapped_control_to_control" character varying NULL, ADD CONSTRAINT "controls_mapped_controls_from_control" FOREIGN KEY ("mapped_control_from_control") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_mapped_controls_to_control" FOREIGN KEY ("mapped_control_to_control") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "mapped_control_from_subcontrol" character varying NULL, ADD COLUMN "mapped_control_to_subcontrol" character varying NULL, ADD CONSTRAINT "subcontrols_mapped_controls_from_subcontrol" FOREIGN KEY ("mapped_control_from_subcontrol") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_mapped_controls_to_subcontrol" FOREIGN KEY ("mapped_control_to_subcontrol") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_mapped_controls_to_subcontrol", DROP CONSTRAINT "subcontrols_mapped_controls_from_subcontrol", DROP COLUMN "mapped_control_to_subcontrol", DROP COLUMN "mapped_control_from_subcontrol";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_mapped_controls_to_control", DROP CONSTRAINT "controls_mapped_controls_from_control", DROP COLUMN "mapped_control_to_control", DROP COLUMN "mapped_control_from_control";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "source", DROP COLUMN "confidence", ALTER COLUMN "mapping_type" DROP DEFAULT;
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "source", DROP COLUMN "confidence", ALTER COLUMN "mapping_type" DROP DEFAULT;
