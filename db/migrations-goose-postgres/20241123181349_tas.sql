-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ALTER COLUMN "description" SET NOT NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ALTER COLUMN "description" SET NOT NULL;

-- +goose Down
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" ALTER COLUMN "description" DROP NOT NULL;
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" ALTER COLUMN "description" DROP NOT NULL;
