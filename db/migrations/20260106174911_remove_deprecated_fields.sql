-- Modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "action_plan_type";
-- Modify "controls" table
ALTER TABLE "controls" DROP COLUMN "control_type";
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "policy_type";
-- Modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "procedure_type";
-- Modify "programs" table
ALTER TABLE "programs" DROP COLUMN "program_type";
-- Modify "risks" table
ALTER TABLE "risks" DROP COLUMN "risk_type", DROP COLUMN "category";
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "control_type";
-- Modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "category";
