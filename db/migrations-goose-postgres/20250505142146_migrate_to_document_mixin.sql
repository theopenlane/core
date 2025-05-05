-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "summary" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "summary" character varying NULL;

-- +goose Down
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "summary";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "summary";
