-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "details_json" jsonb NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "details_json" jsonb NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "description_json" jsonb NULL;
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "details_json" jsonb NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "details_json" jsonb NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "description_json" jsonb NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" DROP DEFAULT;
-- modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" DROP DEFAULT;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "details_json" jsonb NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "details_json" jsonb NULL;
-- modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "text_json" jsonb NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "text_json" jsonb NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "details_json" jsonb NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "details_json" jsonb NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "description_json" jsonb NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "description_json" jsonb NULL;
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "details_json" jsonb NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "details_json" jsonb NULL;

-- +goose Down
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "details_json";
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "details_json";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "description_json";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "description_json";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "business_costs_json", DROP COLUMN "details_json", DROP COLUMN "mitigation_json";
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "business_costs_json", DROP COLUMN "details_json", DROP COLUMN "mitigation_json";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "details_json";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "details_json";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "text_json";
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "text_json";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "details_json";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "details_json";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "description_json";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "desired_outcome_json";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "desired_outcome_json";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP COLUMN "details_json";
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" DROP COLUMN "details_json";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "description_json";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "details_json";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "details_json";
