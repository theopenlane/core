-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "stakeholder_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- Modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "status" SET DEFAULT 'NULL', ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- Modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "owner_id" character varying NULL;
-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "status" SET DEFAULT 'NULL', ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_groups_approver", DROP CONSTRAINT "action_plans_groups_delegate", DROP COLUMN "action_plan_approver", DROP COLUMN "action_plan_delegate", ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "control_implementations_organizations_control_implementations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_groups_control_owner", DROP CONSTRAINT "controls_groups_delegate", ALTER COLUMN "status" SET DEFAULT 'NULL', DROP COLUMN "control_control_owner", DROP COLUMN "control_delegate", ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_groups_approver", DROP CONSTRAINT "internal_policies_groups_delegate", DROP COLUMN "internal_policy_approver", DROP COLUMN "internal_policy_delegate", ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_groups_approver", DROP CONSTRAINT "procedures_groups_delegate", DROP COLUMN "procedure_approver", DROP COLUMN "procedure_delegate", ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_groups_delegate", DROP CONSTRAINT "risks_groups_stakeholder", DROP COLUMN "risk_stakeholder", DROP COLUMN "risk_delegate", ADD COLUMN "stakeholder_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("stakeholder_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_groups_control_owner", DROP CONSTRAINT "subcontrols_groups_delegate", ALTER COLUMN "status" SET DEFAULT 'NULL', DROP COLUMN "subcontrol_control_owner", DROP COLUMN "subcontrol_delegate", ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "subcontrol_control_implementations" table
CREATE TABLE "subcontrol_control_implementations" ("subcontrol_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_implementation_id"), CONSTRAINT "subcontrol_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_control_implementations_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
