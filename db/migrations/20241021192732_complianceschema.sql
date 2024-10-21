-- Create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "assigned" character varying NULL, "due_date" character varying NULL, "priority" character varying NULL, "source" character varying NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "action_plans_mapping_id_key" to table: "action_plans"
CREATE UNIQUE INDEX "action_plans_mapping_id_key" ON "action_plans" ("mapping_id");
-- Create "action_plan_history" table
CREATE TABLE "action_plan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "assigned" character varying NULL, "due_date" character varying NULL, "priority" character varying NULL, "source" character varying NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "actionplanhistory_history_time" to table: "action_plan_history"
CREATE INDEX "actionplanhistory_history_time" ON "action_plan_history" ("history_time");
-- Create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "satisfies" text NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, "control_objective_controls" character varying NULL, "internal_policy_controls" character varying NULL, PRIMARY KEY ("id"));
-- Create index "controls_mapping_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_mapping_id_key" ON "controls" ("mapping_id");
-- Create "control_history" table
CREATE TABLE "control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "satisfies" text NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "controlhistory_history_time" to table: "control_history"
CREATE INDEX "controlhistory_history_time" ON "control_history" ("history_time");
-- Create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_objective_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, "control_controlobjectives" character varying NULL, PRIMARY KEY ("id"));
-- Create index "control_objectives_mapping_id_key" to table: "control_objectives"
CREATE UNIQUE INDEX "control_objectives_mapping_id_key" ON "control_objectives" ("mapping_id");
-- Create "control_objective_history" table
CREATE TABLE "control_objective_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_objective_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "control_number" character varying NULL, "control_family" text NULL, "control_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "controlobjectivehistory_history_time" to table: "control_objective_history"
CREATE INDEX "controlobjectivehistory_history_time" ON "control_objective_history" ("history_time");
-- Create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NOT NULL, "status" character varying NULL, "policy_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "internal_policies_mapping_id_key" to table: "internal_policies"
CREATE UNIQUE INDEX "internal_policies_mapping_id_key" ON "internal_policies" ("mapping_id");
-- Create "internal_policy_history" table
CREATE TABLE "internal_policy_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NOT NULL, "status" character varying NULL, "policy_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
CREATE INDEX "internalpolicyhistory_history_time" ON "internal_policy_history" ("history_time");
-- Create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "narratives_mapping_id_key" to table: "narratives"
CREATE UNIQUE INDEX "narratives_mapping_id_key" ON "narratives" ("mapping_id");
-- Create "narrative_history" table
CREATE TABLE "narrative_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "narrativehistory_history_time" to table: "narrative_history"
CREATE INDEX "narrativehistory_history_time" ON "narrative_history" ("history_time");
-- Create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "procedure_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, "control_objective_procedures" character varying NULL, "standard_procedures" character varying NULL, PRIMARY KEY ("id"));
-- Create index "procedures_mapping_id_key" to table: "procedures"
CREATE UNIQUE INDEX "procedures_mapping_id_key" ON "procedures" ("mapping_id");
-- Create "procedure_history" table
CREATE TABLE "procedure_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "procedure_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "procedurehistory_history_time" to table: "procedure_history"
CREATE INDEX "procedurehistory_history_time" ON "procedure_history" ("history_time");
-- Create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "risk_type" character varying NULL, "business_costs" text NULL, "impact" text NULL, "likelihood" text NULL, "mitigation" text NULL, "satisfies" text NULL, "severity" text NULL, "jsonschema" jsonb NULL, "control_objective_risks" character varying NULL, PRIMARY KEY ("id"));
-- Create index "risks_mapping_id_key" to table: "risks"
CREATE UNIQUE INDEX "risks_mapping_id_key" ON "risks" ("mapping_id");
-- Create "risk_history" table
CREATE TABLE "risk_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "risk_type" character varying NULL, "business_costs" text NULL, "impact" text NULL, "likelihood" text NULL, "mitigation" text NULL, "satisfies" text NULL, "severity" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "riskhistory_history_time" to table: "risk_history"
CREATE INDEX "riskhistory_history_time" ON "risk_history" ("history_time");
-- Create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "family" character varying NULL, "status" character varying NULL, "standard_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "standards_mapping_id_key" to table: "standards"
CREATE UNIQUE INDEX "standards_mapping_id_key" ON "standards" ("mapping_id");
-- Create "standard_history" table
CREATE TABLE "standard_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "family" character varying NULL, "status" character varying NULL, "standard_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "standardhistory_history_time" to table: "standard_history"
CREATE INDEX "standardhistory_history_time" ON "standard_history" ("history_time");
-- Create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "subcontrol_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "subcontrol_number" character varying NULL, "subcontrol_family" text NULL, "subcontrol_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "assigned_to" character varying NULL, "implementation_status" character varying NULL, "implementation_notes" character varying NULL, "implementation_date" character varying NULL, "implementation_evidence" character varying NULL, "implementation_verification" character varying NULL, "implementation_verification_date" character varying NULL, "jsonschema" jsonb NULL, "control_objective_subcontrols" character varying NULL, PRIMARY KEY ("id"));
-- Create index "subcontrols_mapping_id_key" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrols_mapping_id_key" ON "subcontrols" ("mapping_id");
-- Create "subcontrol_history" table
CREATE TABLE "subcontrol_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "subcontrol_type" character varying NULL, "version" character varying NULL, "owner" character varying NULL, "subcontrol_number" character varying NULL, "subcontrol_family" text NULL, "subcontrol_class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "assigned_to" character varying NULL, "implementation_status" character varying NULL, "implementation_notes" character varying NULL, "implementation_date" character varying NULL, "implementation_evidence" character varying NULL, "implementation_verification" character varying NULL, "implementation_verification_date" character varying NULL, "jsonschema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "subcontrolhistory_history_time" to table: "subcontrol_history"
CREATE INDEX "subcontrolhistory_history_time" ON "subcontrol_history" ("history_time");
-- Create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- Create "control_subcontrols" table
CREATE TABLE "control_subcontrols" ("control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("control_id", "subcontrol_id"));
-- Create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- Create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- Create "control_objective_narratives" table
CREATE TABLE "control_objective_narratives" ("control_objective_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "narrative_id"));
-- Create "internal_policy_controlobjectives" table
CREATE TABLE "internal_policy_controlobjectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- Create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- Create "internal_policy_narratives" table
CREATE TABLE "internal_policy_narratives" ("internal_policy_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "narrative_id"));
-- Create "procedure_narratives" table
CREATE TABLE "procedure_narratives" ("procedure_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "narrative_id"));
-- Create "procedure_risks" table
CREATE TABLE "procedure_risks" ("procedure_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "risk_id"));
-- Create "risk_actionplans" table
CREATE TABLE "risk_actionplans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- Create "standard_controlobjectives" table
CREATE TABLE "standard_controlobjectives" ("standard_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "control_objective_id"));
-- Create "standard_controls" table
CREATE TABLE "standard_controls" ("standard_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "control_id"));
-- Create "standard_actionplans" table
CREATE TABLE "standard_actionplans" ("standard_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "action_plan_id"));
-- Modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_control_objectives_controls" FOREIGN KEY ("control_objective_controls") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_internal_policies_controls" FOREIGN KEY ("internal_policy_controls") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_controls_controlobjectives" FOREIGN KEY ("control_controlobjectives") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_standards_procedures" FOREIGN KEY ("standard_procedures") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_control_objectives_subcontrols" FOREIGN KEY ("control_objective_subcontrols") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_procedures" table
ALTER TABLE "control_procedures" ADD CONSTRAINT "control_procedures_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_subcontrols" table
ALTER TABLE "control_subcontrols" ADD CONSTRAINT "control_subcontrols_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_narratives" table
ALTER TABLE "control_narratives" ADD CONSTRAINT "control_narratives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_risks" table
ALTER TABLE "control_risks" ADD CONSTRAINT "control_risks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_objective_narratives" table
ALTER TABLE "control_objective_narratives" ADD CONSTRAINT "control_objective_narratives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_controlobjectives" table
ALTER TABLE "internal_policy_controlobjectives" ADD CONSTRAINT "internal_policy_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_controlobjectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" ADD CONSTRAINT "internal_policy_narratives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_narratives" table
ALTER TABLE "procedure_narratives" ADD CONSTRAINT "procedure_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_narratives_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_risks" table
ALTER TABLE "procedure_risks" ADD CONSTRAINT "procedure_risks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_actionplans" table
ALTER TABLE "risk_actionplans" ADD CONSTRAINT "risk_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_actionplans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "standard_controlobjectives" table
ALTER TABLE "standard_controlobjectives" ADD CONSTRAINT "standard_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_controlobjectives_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "standard_controls" table
ALTER TABLE "standard_controls" ADD CONSTRAINT "standard_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_controls_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "standard_actionplans" table
ALTER TABLE "standard_actionplans" ADD CONSTRAINT "standard_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_actionplans_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
