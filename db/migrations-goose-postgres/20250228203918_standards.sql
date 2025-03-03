-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "subcontrol_action_plans" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "name", DROP COLUMN "version", DROP COLUMN "control_number", DROP COLUMN "family", DROP COLUMN "class", DROP COLUMN "satisfies", DROP COLUMN "mapped_frameworks", DROP COLUMN "details", DROP COLUMN "example_evidence", ADD COLUMN "example_evidence" jsonb, ADD COLUMN "ref_code" character varying NOT NULL, ADD COLUMN "category" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD COLUMN "subcategory" character varying NULL, ADD COLUMN "mapped_categories" jsonb NULL, ADD COLUMN "assessment_objectives" jsonb NULL, ADD COLUMN "assessment_methods" jsonb NULL, ADD COLUMN "control_questions" jsonb NULL, ADD COLUMN "implementation_guidance" jsonb NULL, ADD COLUMN "references" jsonb NULL, ADD COLUMN "standard_id" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "description", DROP COLUMN "control_number", DROP COLUMN "family", DROP COLUMN "class", DROP COLUMN "mapped_frameworks", DROP COLUMN "details", DROP COLUMN "example_evidence", ADD COLUMN "desired_outcome" text NULL, ADD COLUMN "category" character varying NULL, ADD COLUMN "subcategory" character varying NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "description", DROP COLUMN "control_number", DROP COLUMN "family", DROP COLUMN "class", DROP COLUMN "mapped_frameworks", DROP COLUMN "details", DROP COLUMN "control_control_objectives", DROP COLUMN "example_evidence", ADD COLUMN "desired_outcome" text NULL, ADD COLUMN "category" character varying NULL, ADD COLUMN "subcategory" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "name", DROP COLUMN "version", DROP COLUMN "control_number", DROP COLUMN "family", DROP COLUMN "class", DROP COLUMN "satisfies", DROP COLUMN "mapped_frameworks", DROP COLUMN "details", DROP COLUMN "control_objective_controls", DROP COLUMN "example_evidence", ADD COLUMN "example_evidence" jsonb, ADD COLUMN "ref_code" character varying NOT NULL, ADD COLUMN "category" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD COLUMN "subcategory" character varying NULL, ADD COLUMN "mapped_categories" jsonb NULL, ADD COLUMN "assessment_objectives" jsonb NULL, ADD COLUMN "assessment_methods" jsonb NULL, ADD COLUMN "control_questions" jsonb NULL, ADD COLUMN "implementation_guidance" jsonb NULL, ADD COLUMN "references" jsonb NULL, ADD COLUMN "control_implementation" character varying NULL, ADD COLUMN "control_mapped_controls" character varying NULL, ADD COLUMN "control_control_owner" character varying NULL, ADD COLUMN "control_delegate" character varying NULL, ADD COLUMN "standard_id" character varying NULL, ADD COLUMN "subcontrol_mapped_controls" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "control_internal_policies" character varying NULL, ADD COLUMN "subcontrol_internal_policies" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text;
-- modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "control_objective_narratives" character varying NULL, ADD COLUMN "internal_policy_narratives" character varying NULL, ADD COLUMN "procedure_narratives" character varying NULL, ADD COLUMN "subcontrol_narratives" character varying NULL;
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "standard_procedures", ADD COLUMN "subcontrol_procedures" character varying NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "subcontrol_risks" character varying NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" ALTER COLUMN "description" TYPE character varying, DROP COLUMN "family", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", DROP COLUMN "satisfies", DROP COLUMN "details", ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "short_name" character varying NULL, ADD COLUMN "framework" text NULL, ADD COLUMN "governing_body" character varying NULL, ADD COLUMN "domains" jsonb NULL, ADD COLUMN "link" character varying NULL, ADD COLUMN "is_public" boolean NULL DEFAULT false, ADD COLUMN "free_to_use" boolean NULL DEFAULT false, ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "revision" character varying NULL;
-- modify "standards" table
ALTER TABLE "standards" ALTER COLUMN "description" TYPE character varying, DROP COLUMN "family", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", DROP COLUMN "satisfies", DROP COLUMN "details", ADD COLUMN "short_name" character varying NULL, ADD COLUMN "framework" text NULL, ADD COLUMN "governing_body" character varying NULL, ADD COLUMN "domains" jsonb NULL, ADD COLUMN "link" character varying NULL, ADD COLUMN "is_public" boolean NULL DEFAULT false, ADD COLUMN "free_to_use" boolean NULL DEFAULT false, ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "revision" character varying NULL, ADD COLUMN "owner_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "name", DROP COLUMN "subcontrol_type", DROP COLUMN "version", DROP COLUMN "subcontrol_number", DROP COLUMN "family", DROP COLUMN "class", DROP COLUMN "mapped_frameworks", DROP COLUMN "implementation_evidence", DROP COLUMN "implementation_status", DROP COLUMN "implementation_date", DROP COLUMN "implementation_verification", DROP COLUMN "implementation_verification_date", DROP COLUMN "details", DROP COLUMN "example_evidence", ADD COLUMN "example_evidence" jsonb, ADD COLUMN "ref_code" character varying NOT NULL, ADD COLUMN "control_type" character varying NULL, ADD COLUMN "category" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD COLUMN "subcategory" character varying NULL, ADD COLUMN "mapped_categories" jsonb NULL, ADD COLUMN "assessment_objectives" jsonb NULL, ADD COLUMN "assessment_methods" jsonb NULL, ADD COLUMN "control_questions" jsonb NULL, ADD COLUMN "implementation_guidance" jsonb NULL, ADD COLUMN "references" jsonb NULL, ADD COLUMN "control_id" character varying NOT NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "name", DROP COLUMN "subcontrol_type", DROP COLUMN "version", DROP COLUMN "subcontrol_number", DROP COLUMN "family", DROP COLUMN "class", DROP COLUMN "mapped_frameworks", DROP COLUMN "implementation_evidence", DROP COLUMN "implementation_status", DROP COLUMN "implementation_date", DROP COLUMN "implementation_verification", DROP COLUMN "implementation_verification_date", DROP COLUMN "details", DROP COLUMN "control_objective_subcontrols", DROP COLUMN "example_evidence", ADD COLUMN "example_evidence" jsonb, ADD COLUMN "ref_code" character varying NOT NULL, ADD COLUMN "control_type" character varying NULL, ADD COLUMN "category" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD COLUMN "subcategory" character varying NULL, ADD COLUMN "mapped_categories" jsonb NULL, ADD COLUMN "assessment_objectives" jsonb NULL, ADD COLUMN "assessment_methods" jsonb NULL, ADD COLUMN "control_questions" jsonb NULL, ADD COLUMN "implementation_guidance" jsonb NULL, ADD COLUMN "references" jsonb NULL, ADD COLUMN "control_id" character varying NOT NULL, ADD COLUMN "program_subcontrols" character varying NULL, ADD COLUMN "subcontrol_control_owner" character varying NULL, ADD COLUMN "subcontrol_delegate" character varying NULL;
-- create "control_implementations" table
CREATE TABLE "control_implementations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "control_id" character varying NOT NULL, "status" character varying NULL, "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, PRIMARY KEY ("id"));
-- create "control_implementation_history" table
CREATE TABLE "control_implementation_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "control_id" character varying NOT NULL, "status" character varying NULL, "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, PRIMARY KEY ("id"));
-- create index "controlimplementationhistory_history_time" to table: "control_implementation_history"
CREATE INDEX "controlimplementationhistory_history_time" ON "control_implementation_history" ("history_time");
-- create "mapped_controls" table
CREATE TABLE "mapped_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "mapping_type" character varying NULL, "relation" character varying NULL, "control_id" character varying NOT NULL, "mapped_control_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "mappedcontrol_control_id_mapped_control_id" to table: "mapped_controls"
CREATE UNIQUE INDEX "mappedcontrol_control_id_mapped_control_id" ON "mapped_controls" ("control_id", "mapped_control_id");
-- create "mapped_control_history" table
CREATE TABLE "mapped_control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "control_id" character varying NOT NULL, "mapped_control_id" character varying NOT NULL, "mapping_type" character varying NULL, "relation" character varying NULL, PRIMARY KEY ("id"));
-- create index "mappedcontrolhistory_history_time" to table: "mapped_control_history"
CREATE INDEX "mappedcontrolhistory_history_time" ON "mapped_control_history" ("history_time");
-- create "control_control_objectives" table
CREATE TABLE "control_control_objectives" ("control_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_objective_id"));
-- create "subcontrol_control_objectives" table
CREATE TABLE "subcontrol_control_objectives" ("subcontrol_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_objective_id"));
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD CONSTRAINT "action_plans_subcontrols_action_plans" FOREIGN KEY ("subcontrol_action_plans") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_control_implementations_implementation" FOREIGN KEY ("control_implementation") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_mapped_controls_mapped_controls" FOREIGN KEY ("control_mapped_controls") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_standards_controls" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_subcontrols_mapped_controls" FOREIGN KEY ("subcontrol_mapped_controls") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_users_control_owner" FOREIGN KEY ("control_control_owner") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_users_delegate" FOREIGN KEY ("control_delegate") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_controls_internal_policies" FOREIGN KEY ("control_internal_policies") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_subcontrols_internal_policies" FOREIGN KEY ("subcontrol_internal_policies") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD CONSTRAINT "narratives_control_objectives_narratives" FOREIGN KEY ("control_objective_narratives") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_internal_policies_narratives" FOREIGN KEY ("internal_policy_narratives") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_procedures_narratives" FOREIGN KEY ("procedure_narratives") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_subcontrols_narratives" FOREIGN KEY ("subcontrol_narratives") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_subcontrols_procedures" FOREIGN KEY ("subcontrol_procedures") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_subcontrols_risks" FOREIGN KEY ("subcontrol_risks") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD CONSTRAINT "standards_organizations_standards" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_controls_subcontrols" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "subcontrols_programs_subcontrols" FOREIGN KEY ("program_subcontrols") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_control_owner" FOREIGN KEY ("subcontrol_control_owner") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_delegate" FOREIGN KEY ("subcontrol_delegate") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD CONSTRAINT "mapped_controls_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "mapped_controls_controls_mapped_control" FOREIGN KEY ("mapped_control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "control_control_objectives" table
ALTER TABLE "control_control_objectives" ADD CONSTRAINT "control_control_objectives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" ADD CONSTRAINT "subcontrol_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_objectives_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" DROP CONSTRAINT "subcontrol_control_objectives_subcontrol_id", DROP CONSTRAINT "subcontrol_control_objectives_control_objective_id";
-- reverse: modify "control_control_objectives" table
ALTER TABLE "control_control_objectives" DROP CONSTRAINT "control_control_objectives_control_objective_id", DROP CONSTRAINT "control_control_objectives_control_id";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP CONSTRAINT "mapped_controls_controls_mapped_control", DROP CONSTRAINT "mapped_controls_controls_control";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_users_delegate", DROP CONSTRAINT "subcontrols_users_control_owner", DROP CONSTRAINT "subcontrols_programs_subcontrols", DROP CONSTRAINT "subcontrols_controls_subcontrols";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP CONSTRAINT "standards_organizations_standards";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_subcontrols_risks";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_subcontrols_procedures";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_subcontrols_narratives", DROP CONSTRAINT "narratives_procedures_narratives", DROP CONSTRAINT "narratives_internal_policies_narratives", DROP CONSTRAINT "narratives_control_objectives_narratives";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_subcontrols_internal_policies", DROP CONSTRAINT "internal_policies_controls_internal_policies";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_users_delegate", DROP CONSTRAINT "controls_users_control_owner", DROP CONSTRAINT "controls_subcontrols_mapped_controls", DROP CONSTRAINT "controls_standards_controls", DROP CONSTRAINT "controls_mapped_controls_mapped_controls", DROP CONSTRAINT "controls_control_implementations_implementation";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_subcontrols_action_plans";
-- reverse: create "subcontrol_control_objectives" table
DROP TABLE "subcontrol_control_objectives";
-- reverse: create "control_control_objectives" table
DROP TABLE "control_control_objectives";
-- reverse: create index "mappedcontrolhistory_history_time" to table: "mapped_control_history"
DROP INDEX "mappedcontrolhistory_history_time";
-- reverse: create "mapped_control_history" table
DROP TABLE "mapped_control_history";
-- reverse: create index "mappedcontrol_control_id_mapped_control_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_control_id_mapped_control_id";
-- reverse: create "mapped_controls" table
DROP TABLE "mapped_controls";
-- reverse: create index "controlimplementationhistory_history_time" to table: "control_implementation_history"
DROP INDEX "controlimplementationhistory_history_time";
-- reverse: create "control_implementation_history" table
DROP TABLE "control_implementation_history";
-- reverse: create "control_implementations" table
DROP TABLE "control_implementations";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "subcontrol_delegate", DROP COLUMN "subcontrol_control_owner", DROP COLUMN "program_subcontrols", DROP COLUMN "control_id", DROP COLUMN "references", DROP COLUMN "implementation_guidance", DROP COLUMN "control_questions", DROP COLUMN "assessment_methods", DROP COLUMN "assessment_objectives", DROP COLUMN "mapped_categories", DROP COLUMN "subcategory", DROP COLUMN "category_id", DROP COLUMN "category", DROP COLUMN "control_type", DROP COLUMN "ref_code", ALTER COLUMN "example_evidence" TYPE text, ADD COLUMN "control_objective_subcontrols" character varying NULL, ADD COLUMN "details" jsonb NULL, ADD COLUMN "implementation_verification_date" timestamptz NULL, ADD COLUMN "implementation_verification" character varying NULL, ADD COLUMN "implementation_date" timestamptz NULL, ADD COLUMN "implementation_status" character varying NULL, ADD COLUMN "implementation_evidence" character varying NULL, ADD COLUMN "mapped_frameworks" text NULL, ADD COLUMN "class" character varying NULL, ADD COLUMN "family" text NULL, ADD COLUMN "subcontrol_number" character varying NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "subcontrol_type" character varying NULL, ADD COLUMN "name" character varying NOT NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "control_id", DROP COLUMN "references", DROP COLUMN "implementation_guidance", DROP COLUMN "control_questions", DROP COLUMN "assessment_methods", DROP COLUMN "assessment_objectives", DROP COLUMN "mapped_categories", DROP COLUMN "subcategory", DROP COLUMN "category_id", DROP COLUMN "category", DROP COLUMN "control_type", DROP COLUMN "ref_code", ALTER COLUMN "example_evidence" TYPE text, ADD COLUMN "details" jsonb NULL, ADD COLUMN "implementation_verification_date" timestamptz NULL, ADD COLUMN "implementation_verification" character varying NULL, ADD COLUMN "implementation_date" timestamptz NULL, ADD COLUMN "implementation_status" character varying NULL, ADD COLUMN "implementation_evidence" character varying NULL, ADD COLUMN "mapped_frameworks" text NULL, ADD COLUMN "class" character varying NULL, ADD COLUMN "family" text NULL, ADD COLUMN "subcontrol_number" character varying NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "subcontrol_type" character varying NULL, ADD COLUMN "name" character varying NOT NULL;
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "owner_id", DROP COLUMN "revision", DROP COLUMN "system_owned", DROP COLUMN "free_to_use", DROP COLUMN "is_public", DROP COLUMN "link", DROP COLUMN "domains", DROP COLUMN "governing_body", DROP COLUMN "framework", DROP COLUMN "short_name", ADD COLUMN "details" jsonb NULL, ADD COLUMN "satisfies" text NULL, ADD COLUMN "background" text NULL, ADD COLUMN "purpose_and_scope" text NULL, ADD COLUMN "family" character varying NULL, ALTER COLUMN "description" TYPE text;
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "revision", DROP COLUMN "system_owned", DROP COLUMN "free_to_use", DROP COLUMN "is_public", DROP COLUMN "link", DROP COLUMN "domains", DROP COLUMN "governing_body", DROP COLUMN "framework", DROP COLUMN "short_name", DROP COLUMN "owner_id", ADD COLUMN "details" jsonb NULL, ADD COLUMN "satisfies" text NULL, ADD COLUMN "background" text NULL, ADD COLUMN "purpose_and_scope" text NULL, ADD COLUMN "family" character varying NULL, ALTER COLUMN "description" TYPE text;
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "subcontrol_risks";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "subcontrol_procedures", ADD COLUMN "standard_procedures" character varying NULL;
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "subcontrol_narratives", DROP COLUMN "procedure_narratives", DROP COLUMN "internal_policy_narratives", DROP COLUMN "control_objective_narratives", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "satisfies" text NULL;
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "satisfies" text NULL;
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "subcontrol_internal_policies", DROP COLUMN "control_internal_policies";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "subcontrol_mapped_controls", DROP COLUMN "standard_id", DROP COLUMN "control_delegate", DROP COLUMN "control_control_owner", DROP COLUMN "control_mapped_controls", DROP COLUMN "control_implementation", DROP COLUMN "references", DROP COLUMN "implementation_guidance", DROP COLUMN "control_questions", DROP COLUMN "assessment_methods", DROP COLUMN "assessment_objectives", DROP COLUMN "mapped_categories", DROP COLUMN "subcategory", DROP COLUMN "category_id", DROP COLUMN "category", DROP COLUMN "ref_code", ALTER COLUMN "example_evidence" TYPE text, ADD COLUMN "control_objective_controls" character varying NULL, ADD COLUMN "details" jsonb NULL, ADD COLUMN "mapped_frameworks" text NULL, ADD COLUMN "satisfies" text NULL, ADD COLUMN "class" character varying NULL, ADD COLUMN "family" text NULL, ADD COLUMN "control_number" character varying NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "name" character varying NOT NULL;
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "subcategory", DROP COLUMN "category", DROP COLUMN "desired_outcome", ADD COLUMN "example_evidence" text NULL, ADD COLUMN "control_control_objectives" character varying NULL, ADD COLUMN "details" jsonb NULL, ADD COLUMN "mapped_frameworks" text NULL, ADD COLUMN "class" character varying NULL, ADD COLUMN "family" text NULL, ADD COLUMN "control_number" character varying NULL, ADD COLUMN "description" text NULL;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "subcategory", DROP COLUMN "category", DROP COLUMN "desired_outcome", ADD COLUMN "example_evidence" text NULL, ADD COLUMN "details" jsonb NULL, ADD COLUMN "mapped_frameworks" text NULL, ADD COLUMN "class" character varying NULL, ADD COLUMN "family" text NULL, ADD COLUMN "control_number" character varying NULL, ADD COLUMN "description" text NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "standard_id", DROP COLUMN "references", DROP COLUMN "implementation_guidance", DROP COLUMN "control_questions", DROP COLUMN "assessment_methods", DROP COLUMN "assessment_objectives", DROP COLUMN "mapped_categories", DROP COLUMN "subcategory", DROP COLUMN "category_id", DROP COLUMN "category", DROP COLUMN "ref_code", ALTER COLUMN "example_evidence" TYPE text, ADD COLUMN "details" jsonb NULL, ADD COLUMN "mapped_frameworks" text NULL, ADD COLUMN "satisfies" text NULL, ADD COLUMN "class" character varying NULL, ADD COLUMN "family" text NULL, ADD COLUMN "control_number" character varying NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "name" character varying NOT NULL;
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "subcontrol_action_plans";
