-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "description_json" jsonb NULL;
-- Modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "description_json" jsonb NULL;
-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" DROP DEFAULT;
-- Modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" DROP DEFAULT;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "details_json" jsonb NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "text_json" jsonb NULL;
-- Modify "notes" table
ALTER TABLE "notes" ADD COLUMN "text_json" jsonb NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "details_json" jsonb NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "description_json" jsonb NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "description_json" jsonb NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "details_json" jsonb NULL;
