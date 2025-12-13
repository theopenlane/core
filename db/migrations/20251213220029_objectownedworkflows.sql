-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL;
-- Modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "control_id" character varying NULL, ADD COLUMN "internal_policy_id" character varying NULL, ADD COLUMN "evidence_id" character varying NULL;
-- Modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "evidence_id" character varying NULL;
-- Modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "control_id" character varying NULL, ADD COLUMN "internal_policy_id" character varying NULL, ADD COLUMN "evidence_id" character varying NULL, ADD CONSTRAINT "workflow_instances_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "evidence_id" character varying NULL, ADD CONSTRAINT "workflow_object_refs_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "workflowobjectref_workflow_instance_id_evidence_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_evidence_id" ON "workflow_object_refs" ("workflow_instance_id", "evidence_id");
