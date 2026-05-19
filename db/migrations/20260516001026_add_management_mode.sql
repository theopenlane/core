-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
