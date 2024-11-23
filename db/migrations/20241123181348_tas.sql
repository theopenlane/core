-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ALTER COLUMN "description" SET NOT NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ALTER COLUMN "description" SET NOT NULL;
