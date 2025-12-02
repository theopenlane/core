-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "internal_policy_controls";
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "control_internal_policies", DROP COLUMN "subcontrol_internal_policies";
-- create "internal_policy_controls" table
CREATE TABLE "internal_policy_controls" ("internal_policy_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_id"), CONSTRAINT "internal_policy_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_controls_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "internal_policy_narratives", DROP COLUMN "procedure_narratives";
-- create "internal_policy_narratives" table
CREATE TABLE "internal_policy_narratives" ("internal_policy_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "narrative_id"), CONSTRAINT "internal_policy_narratives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "subcontrol_risks";
-- create "internal_policy_risks" table
CREATE TABLE "internal_policy_risks" ("internal_policy_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "risk_id"), CONSTRAINT "internal_policy_risks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "internal_policy_subcontrols" table
CREATE TABLE "internal_policy_subcontrols" ("internal_policy_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "subcontrol_id"), CONSTRAINT "internal_policy_subcontrols_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "subcontrol_procedures";
-- create "procedure_narratives" table
CREATE TABLE "procedure_narratives" ("procedure_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "narrative_id"), CONSTRAINT "procedure_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "procedure_narratives_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "subcontrol_procedures" table
CREATE TABLE "subcontrol_procedures" ("subcontrol_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "procedure_id"), CONSTRAINT "subcontrol_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_procedures_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "subcontrol_risks" table
CREATE TABLE "subcontrol_risks" ("subcontrol_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "risk_id"), CONSTRAINT "subcontrol_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_risks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "subcontrol_risks" table
DROP TABLE "subcontrol_risks";
-- reverse: create "subcontrol_procedures" table
DROP TABLE "subcontrol_procedures";
-- reverse: create "procedure_narratives" table
DROP TABLE "procedure_narratives";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "subcontrol_procedures" character varying NULL;
-- reverse: create "internal_policy_subcontrols" table
DROP TABLE "internal_policy_subcontrols";
-- reverse: create "internal_policy_risks" table
DROP TABLE "internal_policy_risks";
-- reverse: modify "risks" table
ALTER TABLE "risks" ADD COLUMN "subcontrol_risks" character varying NULL;
-- reverse: create "internal_policy_narratives" table
DROP TABLE "internal_policy_narratives";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "procedure_narratives" character varying NULL, ADD COLUMN "internal_policy_narratives" character varying NULL;
-- reverse: create "internal_policy_controls" table
DROP TABLE "internal_policy_controls";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "subcontrol_internal_policies" character varying NULL, ADD COLUMN "control_internal_policies" character varying NULL;
-- reverse: modify "controls" table
ALTER TABLE "controls" ADD COLUMN "internal_policy_controls" character varying NULL;
