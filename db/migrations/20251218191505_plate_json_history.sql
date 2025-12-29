-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "description_json" jsonb NULL;
-- Modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "text_json" jsonb NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "details_json" jsonb NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "description_json" jsonb NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "details_json" jsonb NULL;
