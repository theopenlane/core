-- +goose Up
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ALTER COLUMN "mapping_type" SET NOT NULL, ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "confidence" character varying NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL';
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ALTER COLUMN "mapping_type" SET NOT NULL, ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "confidence" character varying NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL';
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "mapped_control_from_controls" character varying NULL, ADD COLUMN "mapped_control_to_controls" character varying NULL, ADD CONSTRAINT "controls_mapped_controls_from_controls" FOREIGN KEY ("mapped_control_from_controls") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_mapped_controls_to_controls" FOREIGN KEY ("mapped_control_to_controls") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "mapped_control_from_subcontrols" character varying NULL, ADD COLUMN "mapped_control_to_subcontrols" character varying NULL, ADD CONSTRAINT "subcontrols_mapped_controls_from_subcontrols" FOREIGN KEY ("mapped_control_from_subcontrols") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_mapped_controls_to_subcontrols" FOREIGN KEY ("mapped_control_to_subcontrols") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_mapped_controls_to_subcontrols", DROP CONSTRAINT "subcontrols_mapped_controls_from_subcontrols", DROP COLUMN "mapped_control_to_subcontrols", DROP COLUMN "mapped_control_from_subcontrols";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_mapped_controls_to_controls", DROP CONSTRAINT "controls_mapped_controls_from_controls", DROP COLUMN "mapped_control_to_controls", DROP COLUMN "mapped_control_from_controls";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "source", DROP COLUMN "confidence", ALTER COLUMN "mapping_type" DROP NOT NULL, ALTER COLUMN "mapping_type" DROP DEFAULT;
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "source", DROP COLUMN "confidence", ALTER COLUMN "mapping_type" DROP NOT NULL, ALTER COLUMN "mapping_type" DROP DEFAULT;
