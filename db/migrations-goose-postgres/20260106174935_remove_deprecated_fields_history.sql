-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "action_plan_type";
-- modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "control_type";
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "policy_type";
-- modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "procedure_type";
-- modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "program_type";
-- modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "risk_type", DROP COLUMN "category";
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "control_type";
-- modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "category";

-- +goose Down
-- reverse: modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "category" character varying NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "control_type" character varying NULL DEFAULT 'PREVENTATIVE';
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "category" character varying NULL, ADD COLUMN "risk_type" character varying NULL;
-- reverse: modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "program_type" character varying NOT NULL DEFAULT 'FRAMEWORK';
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "procedure_type" character varying NULL;
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "policy_type" character varying NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "control_type" character varying NULL DEFAULT 'PREVENTATIVE';
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "action_plan_type" character varying NULL;
