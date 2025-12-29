-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "details_json" jsonb NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "description_json" jsonb NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "details_json" jsonb NULL;
-- Modify "notes" table
ALTER TABLE "notes" ADD COLUMN "text_json" jsonb NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "details_json" jsonb NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "description_json" jsonb NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "details_json" jsonb NULL;
