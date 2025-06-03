-- +goose Up
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ALTER COLUMN "mapping_type" SET NOT NULL, ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "confidence" bigint NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL';
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "reference_framework" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "reference_framework" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_control_implementation_creators" character varying NULL, ADD COLUMN "organization_evidence_creators" character varying NULL, ADD COLUMN "organization_mapped_control_creators" character varying NULL, ADD COLUMN "organization_scheduled_job_creators" character varying NULL, ADD COLUMN "organization_standard_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_control_implementation_creators" FOREIGN KEY ("organization_control_implementation_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_evidence_creators" FOREIGN KEY ("organization_evidence_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_mapped_control_creators" FOREIGN KEY ("organization_mapped_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scheduled_job_creators" FOREIGN KEY ("organization_scheduled_job_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_standard_creators" FOREIGN KEY ("organization_standard_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "control_implementation_blocked_groups" table
CREATE TABLE "control_implementation_blocked_groups" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"), CONSTRAINT "control_implementation_blocked_groups_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_implementation_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_implementation_editors" table
CREATE TABLE "control_implementation_editors" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"), CONSTRAINT "control_implementation_editors_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_implementation_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_implementation_viewers" table
CREATE TABLE "control_implementation_viewers" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"), CONSTRAINT "control_implementation_viewers_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_implementation_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ALTER COLUMN "mapping_type" SET NOT NULL, ALTER COLUMN "mapping_type" SET DEFAULT 'EQUAL', ADD COLUMN "confidence" bigint NULL, ADD COLUMN "source" character varying NULL DEFAULT 'MANUAL', ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "mapped_controls_organizations_mapped_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "mapped_control_blocked_groups" table
CREATE TABLE "mapped_control_blocked_groups" ("mapped_control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "group_id"), CONSTRAINT "mapped_control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "mapped_control_blocked_groups_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "mapped_control_editors" table
CREATE TABLE "mapped_control_editors" ("mapped_control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "group_id"), CONSTRAINT "mapped_control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "mapped_control_editors_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "reference_framework" character varying NULL;
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
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "reference_framework";
-- reverse: create "mapped_control_editors" table
DROP TABLE "mapped_control_editors";
-- reverse: create "mapped_control_blocked_groups" table
DROP TABLE "mapped_control_blocked_groups";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP CONSTRAINT "mapped_controls_organizations_mapped_controls", DROP COLUMN "owner_id", DROP COLUMN "source", DROP COLUMN "confidence", ALTER COLUMN "mapping_type" DROP NOT NULL, ALTER COLUMN "mapping_type" DROP DEFAULT;
-- reverse: create "control_implementation_viewers" table
DROP TABLE "control_implementation_viewers";
-- reverse: create "control_implementation_editors" table
DROP TABLE "control_implementation_editors";
-- reverse: create "control_implementation_blocked_groups" table
DROP TABLE "control_implementation_blocked_groups";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_organizations_standard_creators", DROP CONSTRAINT "groups_organizations_scheduled_job_creators", DROP CONSTRAINT "groups_organizations_mapped_control_creators", DROP CONSTRAINT "groups_organizations_evidence_creators", DROP CONSTRAINT "groups_organizations_control_implementation_creators", DROP COLUMN "organization_standard_creators", DROP COLUMN "organization_scheduled_job_creators", DROP COLUMN "organization_mapped_control_creators", DROP COLUMN "organization_evidence_creators", DROP COLUMN "organization_control_implementation_creators";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "reference_framework";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "reference_framework";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "source", DROP COLUMN "confidence", DROP COLUMN "owner_id", ALTER COLUMN "mapping_type" DROP NOT NULL, ALTER COLUMN "mapping_type" DROP DEFAULT;
