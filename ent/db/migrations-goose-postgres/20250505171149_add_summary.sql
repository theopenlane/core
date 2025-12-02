-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "summary" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "summary" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "summary" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "summary" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "summary" character varying NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "summary" character varying NULL;

-- +goose Down
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "summary";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "summary";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "summary";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "summary";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "summary";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "summary";
