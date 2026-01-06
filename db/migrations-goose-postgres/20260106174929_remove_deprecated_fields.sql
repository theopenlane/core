-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "action_plan_type";
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "control_type";
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "policy_type";
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "procedure_type";
-- modify "programs" table
ALTER TABLE "programs" DROP COLUMN "program_type";
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "risk_type", DROP COLUMN "category";
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "control_type";
-- modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "category";

-- +goose Down
-- reverse: modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "category" character varying NULL;
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "control_type" character varying NULL DEFAULT 'PREVENTATIVE';
-- reverse: modify "risks" table
ALTER TABLE "risks" ADD COLUMN "category" character varying NULL, ADD COLUMN "risk_type" character varying NULL;
-- reverse: modify "programs" table
ALTER TABLE "programs" ADD COLUMN "program_type" character varying NOT NULL DEFAULT 'FRAMEWORK';
-- reverse: modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "procedure_type" character varying NULL;
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "policy_type" character varying NULL;
-- reverse: modify "controls" table
ALTER TABLE "controls" ADD COLUMN "control_type" character varying NULL DEFAULT 'PREVENTATIVE';
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "action_plan_type" character varying NULL;
