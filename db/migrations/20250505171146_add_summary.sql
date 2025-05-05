-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "summary" character varying NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "summary" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "summary" character varying NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "summary" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "summary" character varying NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "summary" character varying NULL;
