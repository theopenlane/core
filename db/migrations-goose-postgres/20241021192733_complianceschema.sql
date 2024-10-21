-- +goose Up
-- create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "assigned" character varying NULL, "due_date" character varying NULL, "priority" character varying NULL, "source" character varying NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "action_plans_mapping_id_key" to table: "action_plans"
CREATE UNIQUE INDEX "action_plans_mapping_id_key" ON "action_plans" ("mapping_id");
-- create "action_plan_history" table
CREATE TABLE "action_plan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "assigned" character varying NULL, "due_date" character varying NULL, "priority" character varying NULL, "source" character varying NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "actionplanhistory_history_time" to table: "action_plan_history"
CREATE INDEX "actionplanhistory_history_time" ON "action_plan_history" ("history_time");
-- create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "satisfies" text NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, "control_objective_controls" character varying NULL, "internal_policy_controls" character varying NULL, PRIMARY KEY ("id"));
-- create index "controls_mapping_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_mapping_id_key" ON "controls" ("mapping_id");
-- create "control_history" table
CREATE TABLE "control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "satisfies" text NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "controlhistory_history_time" to table: "control_history"
CREATE INDEX "controlhistory_history_time" ON "control_history" ("history_time");
-- create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_objective_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, "control_controlobjectives" character varying NULL, PRIMARY KEY ("id"));
-- create index "control_objectives_mapping_id_key" to table: "control_objectives"
CREATE UNIQUE INDEX "control_objectives_mapping_id_key" ON "control_objectives" ("mapping_id");
-- create "control_objective_history" table
CREATE TABLE "control_objective_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_objective_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "controlobjectivehistory_history_time" to table: "control_objective_history"
CREATE INDEX "controlobjectivehistory_history_time" ON "control_objective_history" ("history_time");
-- create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NOT NULL, "status" character varying NULL, "policy_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "internal_policies_mapping_id_key" to table: "internal_policies"
CREATE UNIQUE INDEX "internal_policies_mapping_id_key" ON "internal_policies" ("mapping_id");
-- create "internal_policy_history" table
CREATE TABLE "internal_policy_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NOT NULL, "status" character varying NULL, "policy_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
CREATE INDEX "internalpolicyhistory_history_time" ON "internal_policy_history" ("history_time");
-- create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "narratives_mapping_id_key" to table: "narratives"
CREATE UNIQUE INDEX "narratives_mapping_id_key" ON "narratives" ("mapping_id");
-- create "narrative_history" table
CREATE TABLE "narrative_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "narrativehistory_history_time" to table: "narrative_history"
CREATE INDEX "narrativehistory_history_time" ON "narrative_history" ("history_time");
-- create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "procedure_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, "control_objective_procedures" character varying NULL, "standard_procedures" character varying NULL, PRIMARY KEY ("id"));
-- create index "procedures_mapping_id_key" to table: "procedures"
CREATE UNIQUE INDEX "procedures_mapping_id_key" ON "procedures" ("mapping_id");
-- create "procedure_history" table
CREATE TABLE "procedure_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "procedure_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "procedurehistory_history_time" to table: "procedure_history"
CREATE INDEX "procedurehistory_history_time" ON "procedure_history" ("history_time");
-- create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "risk_type" character varying NULL, "business_costs" text NULL, "impact" text NULL, "likelihood" text NULL, "mitigation" text NULL, "satisfies" text NULL, "severity" text NULL, "jsonschema" jsonb NULL, "control_objective_risks" character varying NULL, PRIMARY KEY ("id"));
-- create index "risks_mapping_id_key" to table: "risks"
CREATE UNIQUE INDEX "risks_mapping_id_key" ON "risks" ("mapping_id");
-- create "risk_history" table
CREATE TABLE "risk_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "risk_type" character varying NULL, "business_costs" text NULL, "impact" text NULL, "likelihood" text NULL, "mitigation" text NULL, "satisfies" text NULL, "severity" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "riskhistory_history_time" to table: "risk_history"
CREATE INDEX "riskhistory_history_time" ON "risk_history" ("history_time");
-- create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "family" character varying NULL, "status" character varying NULL, "standard_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "standards_mapping_id_key" to table: "standards"
CREATE UNIQUE INDEX "standards_mapping_id_key" ON "standards" ("mapping_id");
-- create "standard_history" table
CREATE TABLE "standard_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "family" character varying NULL, "status" character varying NULL, "standard_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "standardhistory_history_time" to table: "standard_history"
CREATE INDEX "standardhistory_history_time" ON "standard_history" ("history_time");
-- create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "subcontrol_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "subcontrol_number" character varying NULL, "subcontrol_family" text NULL, "subcontrol_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "assigned_to" character varying NULL, "implementation_status" character varying NULL, "implementation_notes" character varying NULL, "implementation_date" character varying NULL, "implementation_evidence" character varying NULL, "implementation_verification" character varying NULL, "implementation_verification_date" character varying NULL, "jsonschema" jsonb NULL, "control_objective_subcontrols" character varying NULL, PRIMARY KEY ("id"));
-- create index "subcontrols_mapping_id_key" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrols_mapping_id_key" ON "subcontrols" ("mapping_id");
-- create "subcontrol_history" table
CREATE TABLE "subcontrol_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "subcontrol_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "subcontrol_number" character varying NULL, "subcontrol_family" text NULL, "subcontrol_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "assigned_to" character varying NULL, "implementation_status" character varying NULL, "implementation_notes" character varying NULL, "implementation_date" character varying NULL, "implementation_evidence" character varying NULL, "implementation_verification" character varying NULL, "implementation_verification_date" character varying NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "subcontrolhistory_history_time" to table: "subcontrol_history"
CREATE INDEX "subcontrolhistory_history_time" ON "subcontrol_history" ("history_time");
-- create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- create "control_subcontrols" table
CREATE TABLE "control_subcontrols" ("control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("control_id", "subcontrol_id"));
-- create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- create "control_objective_narratives" table
CREATE TABLE "control_objective_narratives" ("control_objective_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "narrative_id"));
-- create "internal_policy_controlobjectives" table
CREATE TABLE "internal_policy_controlobjectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- create "internal_policy_narratives" table
CREATE TABLE "internal_policy_narratives" ("internal_policy_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "narrative_id"));
-- create "procedure_narratives" table
CREATE TABLE "procedure_narratives" ("procedure_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "narrative_id"));
-- create "procedure_risks" table
CREATE TABLE "procedure_risks" ("procedure_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "risk_id"));
-- create "risk_actionplans" table
CREATE TABLE "risk_actionplans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- create "standard_controlobjectives" table
CREATE TABLE "standard_controlobjectives" ("standard_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "control_objective_id"));
-- create "standard_controls" table
CREATE TABLE "standard_controls" ("standard_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "control_id"));
-- create "standard_actionplans" table
CREATE TABLE "standard_actionplans" ("standard_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "action_plan_id"));
-- modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_control_objectives_controls" FOREIGN KEY ("control_objective_controls") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_internal_policies_controls" FOREIGN KEY ("internal_policy_controls") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_controls_controlobjectives" FOREIGN KEY ("control_controlobjectives") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_standards_procedures" FOREIGN KEY ("standard_procedures") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_control_objectives_subcontrols" FOREIGN KEY ("control_objective_subcontrols") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_procedures" table
ALTER TABLE "control_procedures" ADD CONSTRAINT "control_procedures_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_subcontrols" table
ALTER TABLE "control_subcontrols" ADD CONSTRAINT "control_subcontrols_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_narratives" table
ALTER TABLE "control_narratives" ADD CONSTRAINT "control_narratives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_risks" table
ALTER TABLE "control_risks" ADD CONSTRAINT "control_risks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_narratives" table
ALTER TABLE "control_objective_narratives" ADD CONSTRAINT "control_objective_narratives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_controlobjectives" table
ALTER TABLE "internal_policy_controlobjectives" ADD CONSTRAINT "internal_policy_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_controlobjectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" ADD CONSTRAINT "internal_policy_narratives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_narratives" table
ALTER TABLE "procedure_narratives" ADD CONSTRAINT "procedure_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_narratives_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_risks" table
ALTER TABLE "procedure_risks" ADD CONSTRAINT "procedure_risks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_actionplans" table
ALTER TABLE "risk_actionplans" ADD CONSTRAINT "risk_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_actionplans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_controlobjectives" table
ALTER TABLE "standard_controlobjectives" ADD CONSTRAINT "standard_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_controlobjectives_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_controls" table
ALTER TABLE "standard_controls" ADD CONSTRAINT "standard_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_controls_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_actionplans" table
ALTER TABLE "standard_actionplans" ADD CONSTRAINT "standard_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_actionplans_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "standard_actionplans" table
ALTER TABLE "standard_actionplans" DROP CONSTRAINT "standard_actionplans_standard_id", DROP CONSTRAINT "standard_actionplans_action_plan_id";
-- reverse: modify "standard_controls" table
ALTER TABLE "standard_controls" DROP CONSTRAINT "standard_controls_standard_id", DROP CONSTRAINT "standard_controls_control_id";
-- reverse: modify "standard_controlobjectives" table
ALTER TABLE "standard_controlobjectives" DROP CONSTRAINT "standard_controlobjectives_standard_id", DROP CONSTRAINT "standard_controlobjectives_control_objective_id";
-- reverse: modify "risk_actionplans" table
ALTER TABLE "risk_actionplans" DROP CONSTRAINT "risk_actionplans_risk_id", DROP CONSTRAINT "risk_actionplans_action_plan_id";
-- reverse: modify "procedure_risks" table
ALTER TABLE "procedure_risks" DROP CONSTRAINT "procedure_risks_risk_id", DROP CONSTRAINT "procedure_risks_procedure_id";
-- reverse: modify "procedure_narratives" table
ALTER TABLE "procedure_narratives" DROP CONSTRAINT "procedure_narratives_procedure_id", DROP CONSTRAINT "procedure_narratives_narrative_id";
-- reverse: modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" DROP CONSTRAINT "internal_policy_narratives_narrative_id", DROP CONSTRAINT "internal_policy_narratives_internal_policy_id";
-- reverse: modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" DROP CONSTRAINT "internal_policy_procedures_procedure_id", DROP CONSTRAINT "internal_policy_procedures_internal_policy_id";
-- reverse: modify "internal_policy_controlobjectives" table
ALTER TABLE "internal_policy_controlobjectives" DROP CONSTRAINT "internal_policy_controlobjectives_internal_policy_id", DROP CONSTRAINT "internal_policy_controlobjectives_control_objective_id";
-- reverse: modify "control_objective_narratives" table
ALTER TABLE "control_objective_narratives" DROP CONSTRAINT "control_objective_narratives_narrative_id", DROP CONSTRAINT "control_objective_narratives_control_objective_id";
-- reverse: modify "control_risks" table
ALTER TABLE "control_risks" DROP CONSTRAINT "control_risks_risk_id", DROP CONSTRAINT "control_risks_control_id";
-- reverse: modify "control_narratives" table
ALTER TABLE "control_narratives" DROP CONSTRAINT "control_narratives_narrative_id", DROP CONSTRAINT "control_narratives_control_id";
-- reverse: modify "control_subcontrols" table
ALTER TABLE "control_subcontrols" DROP CONSTRAINT "control_subcontrols_subcontrol_id", DROP CONSTRAINT "control_subcontrols_control_id";
-- reverse: modify "control_procedures" table
ALTER TABLE "control_procedures" DROP CONSTRAINT "control_procedures_procedure_id", DROP CONSTRAINT "control_procedures_control_id";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_control_objectives_subcontrols";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_control_objectives_risks";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_standards_procedures", DROP CONSTRAINT "procedures_control_objectives_procedures";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_controls_controlobjectives";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_internal_policies_controls", DROP CONSTRAINT "controls_control_objectives_controls";
-- reverse: create "standard_actionplans" table
DROP TABLE "standard_actionplans";
-- reverse: create "standard_controls" table
DROP TABLE "standard_controls";
-- reverse: create "standard_controlobjectives" table
DROP TABLE "standard_controlobjectives";
-- reverse: create "risk_actionplans" table
DROP TABLE "risk_actionplans";
-- reverse: create "procedure_risks" table
DROP TABLE "procedure_risks";
-- reverse: create "procedure_narratives" table
DROP TABLE "procedure_narratives";
-- reverse: create "internal_policy_narratives" table
DROP TABLE "internal_policy_narratives";
-- reverse: create "internal_policy_procedures" table
DROP TABLE "internal_policy_procedures";
-- reverse: create "internal_policy_controlobjectives" table
DROP TABLE "internal_policy_controlobjectives";
-- reverse: create "control_objective_narratives" table
DROP TABLE "control_objective_narratives";
-- reverse: create "control_risks" table
DROP TABLE "control_risks";
-- reverse: create "control_narratives" table
DROP TABLE "control_narratives";
-- reverse: create "control_subcontrols" table
DROP TABLE "control_subcontrols";
-- reverse: create "control_procedures" table
DROP TABLE "control_procedures";
-- reverse: create index "subcontrolhistory_history_time" to table: "subcontrol_history"
DROP INDEX "subcontrolhistory_history_time";
-- reverse: create "subcontrol_history" table
DROP TABLE "subcontrol_history";
-- reverse: create index "subcontrols_mapping_id_key" to table: "subcontrols"
DROP INDEX "subcontrols_mapping_id_key";
-- reverse: create "subcontrols" table
DROP TABLE "subcontrols";
-- reverse: create index "standardhistory_history_time" to table: "standard_history"
DROP INDEX "standardhistory_history_time";
-- reverse: create "standard_history" table
DROP TABLE "standard_history";
-- reverse: create index "standards_mapping_id_key" to table: "standards"
DROP INDEX "standards_mapping_id_key";
-- reverse: create "standards" table
DROP TABLE "standards";
-- reverse: create index "riskhistory_history_time" to table: "risk_history"
DROP INDEX "riskhistory_history_time";
-- reverse: create "risk_history" table
DROP TABLE "risk_history";
-- reverse: create index "risks_mapping_id_key" to table: "risks"
DROP INDEX "risks_mapping_id_key";
-- reverse: create "risks" table
DROP TABLE "risks";
-- reverse: create index "procedurehistory_history_time" to table: "procedure_history"
DROP INDEX "procedurehistory_history_time";
-- reverse: create "procedure_history" table
DROP TABLE "procedure_history";
-- reverse: create index "procedures_mapping_id_key" to table: "procedures"
DROP INDEX "procedures_mapping_id_key";
-- reverse: create "procedures" table
DROP TABLE "procedures";
-- reverse: create index "narrativehistory_history_time" to table: "narrative_history"
DROP INDEX "narrativehistory_history_time";
-- reverse: create "narrative_history" table
DROP TABLE "narrative_history";
-- reverse: create index "narratives_mapping_id_key" to table: "narratives"
DROP INDEX "narratives_mapping_id_key";
-- reverse: create "narratives" table
DROP TABLE "narratives";
-- reverse: create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
DROP INDEX "internalpolicyhistory_history_time";
-- reverse: create "internal_policy_history" table
DROP TABLE "internal_policy_history";
-- reverse: create index "internal_policies_mapping_id_key" to table: "internal_policies"
DROP INDEX "internal_policies_mapping_id_key";
-- reverse: create "internal_policies" table
DROP TABLE "internal_policies";
-- reverse: create index "controlobjectivehistory_history_time" to table: "control_objective_history"
DROP INDEX "controlobjectivehistory_history_time";
-- reverse: create "control_objective_history" table
DROP TABLE "control_objective_history";
-- reverse: create index "control_objectives_mapping_id_key" to table: "control_objectives"
DROP INDEX "control_objectives_mapping_id_key";
-- reverse: create "control_objectives" table
DROP TABLE "control_objectives";
-- reverse: create index "controlhistory_history_time" to table: "control_history"
DROP INDEX "controlhistory_history_time";
-- reverse: create "control_history" table
DROP TABLE "control_history";
-- reverse: create index "controls_mapping_id_key" to table: "controls"
DROP INDEX "controls_mapping_id_key";
-- reverse: create "controls" table
DROP TABLE "controls";
-- reverse: create index "actionplanhistory_history_time" to table: "action_plan_history"
DROP INDEX "actionplanhistory_history_time";
-- reverse: create "action_plan_history" table
DROP TABLE "action_plan_history";
-- reverse: create index "action_plans_mapping_id_key" to table: "action_plans"
DROP INDEX "action_plans_mapping_id_key";
-- reverse: create "action_plans" table
DROP TABLE "action_plans";
