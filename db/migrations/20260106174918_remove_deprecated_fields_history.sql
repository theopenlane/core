-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "action_plan_type";
-- Modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "control_type";
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "policy_type";
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "procedure_type";
-- Modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "program_type";
-- Modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "risk_type", DROP COLUMN "category";
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "control_type";
-- Modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "category";
