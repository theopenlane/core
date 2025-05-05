-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "summary" character varying NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "summary" character varying NULL;
