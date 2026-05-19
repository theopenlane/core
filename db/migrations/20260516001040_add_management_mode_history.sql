-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
