-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "reference_framework" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "reference_framework" character varying NULL;
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ALTER COLUMN "mapping_type" SET NOT NULL, ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "confidence" bigint NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL';
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "reference_framework" character varying NULL;
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ALTER COLUMN "mapping_type" SET NOT NULL, ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "confidence" bigint NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL', ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "mapped_controls_organizations_mapped_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "mapped_control_from_controls" table
CREATE TABLE "mapped_control_from_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"), CONSTRAINT "mapped_control_from_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "mapped_control_from_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "reference_framework" character varying NULL;
-- create "mapped_control_from_subcontrols" table
CREATE TABLE "mapped_control_from_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"), CONSTRAINT "mapped_control_from_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "mapped_control_from_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "mapped_control_to_controls" table
CREATE TABLE "mapped_control_to_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"), CONSTRAINT "mapped_control_to_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "mapped_control_to_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "mapped_control_to_subcontrols" table
CREATE TABLE "mapped_control_to_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"), CONSTRAINT "mapped_control_to_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "mapped_control_to_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "mapped_control_to_subcontrols" table
DROP TABLE "mapped_control_to_subcontrols";
-- reverse: create "mapped_control_to_controls" table
DROP TABLE "mapped_control_to_controls";
-- reverse: create "mapped_control_from_subcontrols" table
DROP TABLE "mapped_control_from_subcontrols";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "reference_framework";
-- reverse: create "mapped_control_from_controls" table
DROP TABLE "mapped_control_from_controls";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP CONSTRAINT "mapped_controls_organizations_mapped_controls", DROP COLUMN "owner_id", DROP COLUMN "source", DROP COLUMN "confidence", ALTER COLUMN "mapping_type" DROP NOT NULL, ALTER COLUMN "mapping_type" DROP DEFAULT;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "reference_framework";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "source", DROP COLUMN "confidence", DROP COLUMN "owner_id", ALTER COLUMN "mapping_type" DROP NOT NULL, ALTER COLUMN "mapping_type" DROP DEFAULT;
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "reference_framework";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "reference_framework";
