-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "entity_type_history" table
ALTER TABLE "entity_type_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "job_runners" table
ALTER TABLE "job_runners" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "job_template_history" table
ALTER TABLE "job_template_history" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "job_templates" table
ALTER TABLE "job_templates" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "subprocessors" table
ALTER TABLE "subprocessors" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "owner_id" character varying NULL;
-- modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_organizations_trust_center_docs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "trustcenterdoc_owner_id" to table: "trust_center_docs"
CREATE INDEX "trustcenterdoc_owner_id" ON "trust_center_docs" ("owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "trustcenterdoc_owner_id" to table: "trust_center_docs"
DROP INDEX "trustcenterdoc_owner_id";
-- reverse: modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" DROP CONSTRAINT "trust_center_docs_organizations_trust_center_docs", DROP COLUMN "owner_id";
-- reverse: modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" DROP COLUMN "owner_id";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "subprocessors" table
ALTER TABLE "subprocessors" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "job_templates" table
ALTER TABLE "job_templates" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "job_template_history" table
ALTER TABLE "job_template_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "job_runners" table
ALTER TABLE "job_runners" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
