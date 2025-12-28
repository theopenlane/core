-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "controls" table
ALTER TABLE "controls" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ADD COLUMN "approval_fields" jsonb NULL, ADD COLUMN "approval_edges" jsonb NULL, ADD COLUMN "approval_submission_mode" character varying NULL DEFAULT 'MANUAL_SUBMIT';
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "subcontrol_id" character varying NULL, ADD COLUMN "action_plan_id" character varying NULL, ADD COLUMN "procedure_id" character varying NULL, ADD CONSTRAINT "workflow_object_refs_action_plans_action_plan" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_procedures_procedure" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_subcontrols_subcontrol" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "workflowobjectref_workflow_instance_id_action_plan_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_action_plan_id" ON "workflow_object_refs" ("workflow_instance_id", "action_plan_id");
-- Create index "workflowobjectref_workflow_instance_id_procedure_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_procedure_id" ON "workflow_object_refs" ("workflow_instance_id", "procedure_id");
-- Create index "workflowobjectref_workflow_instance_id_subcontrol_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_subcontrol_id" ON "workflow_object_refs" ("workflow_instance_id", "subcontrol_id");
-- Create "workflow_proposals" table
CREATE TABLE "workflow_proposals" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "domain_key" character varying NOT NULL, "state" character varying NOT NULL DEFAULT 'DRAFT', "revision" bigint NOT NULL DEFAULT 1, "changes" jsonb NULL, "proposed_hash" character varying NULL, "approved_hash" character varying NULL, "submitted_at" timestamptz NULL, "owner_id" character varying NULL, "workflow_object_ref_id" character varying NOT NULL, "submitted_by_user_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_proposals_organizations_workflow_proposals" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_proposals_users_user" FOREIGN KEY ("submitted_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_proposals_workflow_object_refs_workflow_object_ref" FOREIGN KEY ("workflow_object_ref_id") REFERENCES "workflow_object_refs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY ((ARRAY['DRAFT'::character varying, 'SUBMITTED'::character varying])::text[]));
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "current_action_index" bigint NOT NULL DEFAULT 0, ADD COLUMN "subcontrol_id" character varying NULL, ADD COLUMN "action_plan_id" character varying NULL, ADD COLUMN "procedure_id" character varying NULL, ADD COLUMN "workflow_proposal_id" character varying NULL, ADD CONSTRAINT "workflow_instances_action_plans_action_plan" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_procedures_procedure" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_subcontrols_subcontrol" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_workflow_proposals_workflow_proposal" FOREIGN KEY ("workflow_proposal_id") REFERENCES "workflow_proposals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
