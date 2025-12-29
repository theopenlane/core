-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" ADD COLUMN "approval_fields" jsonb NULL, ADD COLUMN "approval_edges" jsonb NULL, ADD COLUMN "approval_submission_mode" character varying NULL DEFAULT 'MANUAL_SUBMIT';
-- Modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "workflow_proposal_id" character varying NULL, ADD COLUMN "current_action_index" bigint NOT NULL DEFAULT 0, ADD COLUMN "subcontrol_id" character varying NULL, ADD COLUMN "action_plan_id" character varying NULL, ADD COLUMN "procedure_id" character varying NULL;
-- Modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "subcontrol_id" character varying NULL, ADD COLUMN "action_plan_id" character varying NULL, ADD COLUMN "procedure_id" character varying NULL;
