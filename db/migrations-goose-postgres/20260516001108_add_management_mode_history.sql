-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';

-- +goose Down
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "management_mode";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "management_mode";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "management_mode";
