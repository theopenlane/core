-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "details_json" jsonb NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "details_json" jsonb NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "desired_outcome_json" jsonb NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "description_json" jsonb NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "details_json" jsonb NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "text_json" jsonb NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "details_json" jsonb NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "mitigation_json" jsonb NULL, ADD COLUMN "details_json" jsonb NULL, ADD COLUMN "business_costs_json" jsonb NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "description_json" jsonb NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "details_json" jsonb NULL;

-- +goose Down
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "details_json";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "description_json";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "business_costs_json", DROP COLUMN "details_json", DROP COLUMN "mitigation_json";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "details_json";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "text_json";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "details_json";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "description_json";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "desired_outcome_json";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP COLUMN "details_json";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "details_json";
