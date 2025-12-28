-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "proposed_changes", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_at", ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" ADD COLUMN "approval_fields" jsonb NULL, ADD COLUMN "approval_edges" jsonb NULL, ADD COLUMN "approval_submission_mode" character varying NULL DEFAULT 'MANUAL_SUBMIT';
-- modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "workflow_proposal_id" character varying NULL, ADD COLUMN "current_action_index" bigint NOT NULL DEFAULT 0, ADD COLUMN "subcontrol_id" character varying NULL, ADD COLUMN "action_plan_id" character varying NULL, ADD COLUMN "procedure_id" character varying NULL;
-- modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "subcontrol_id" character varying NULL, ADD COLUMN "action_plan_id" character varying NULL, ADD COLUMN "procedure_id" character varying NULL;

-- +goose Down
-- reverse: modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" DROP COLUMN "procedure_id", DROP COLUMN "action_plan_id", DROP COLUMN "subcontrol_id";
-- reverse: modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" DROP COLUMN "procedure_id", DROP COLUMN "action_plan_id", DROP COLUMN "subcontrol_id", DROP COLUMN "current_action_index", DROP COLUMN "workflow_proposal_id";
-- reverse: modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" DROP COLUMN "approval_submission_mode", DROP COLUMN "approval_edges", DROP COLUMN "approval_fields";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "workflow_eligible_marker", ADD COLUMN "proposed_at" timestamptz NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_changes" jsonb NULL;
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "workflow_eligible_marker", ADD COLUMN "proposed_at" timestamptz NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_changes" jsonb NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "workflow_eligible_marker", ADD COLUMN "proposed_at" timestamptz NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_changes" jsonb NULL;
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "workflow_eligible_marker";
