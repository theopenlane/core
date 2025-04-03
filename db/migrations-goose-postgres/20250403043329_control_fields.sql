-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "stakeholder_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "status" SET DEFAULT 'NULL', ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "owner_id" character varying NULL;
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "status" SET DEFAULT 'NULL', ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_groups_approver", DROP CONSTRAINT "action_plans_groups_delegate", DROP COLUMN "action_plan_approver", DROP COLUMN "action_plan_delegate", ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "control_implementations_organizations_control_implementations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_groups_control_owner", DROP CONSTRAINT "controls_groups_delegate", ALTER COLUMN "status" SET DEFAULT 'NULL', DROP COLUMN "control_control_owner", DROP COLUMN "control_delegate", ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_groups_approver", DROP CONSTRAINT "internal_policies_groups_delegate", DROP COLUMN "internal_policy_approver", DROP COLUMN "internal_policy_delegate", ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_groups_approver", DROP CONSTRAINT "procedures_groups_delegate", DROP COLUMN "procedure_approver", DROP COLUMN "procedure_delegate", ADD COLUMN "approver_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_groups_delegate", DROP CONSTRAINT "risks_groups_stakeholder", DROP COLUMN "risk_stakeholder", DROP COLUMN "risk_delegate", ADD COLUMN "stakeholder_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("stakeholder_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_groups_control_owner", DROP CONSTRAINT "subcontrols_groups_delegate", ALTER COLUMN "status" SET DEFAULT 'NULL', DROP COLUMN "subcontrol_control_owner", DROP COLUMN "subcontrol_delegate", ADD COLUMN "control_owner_id" character varying NULL, ADD COLUMN "delegate_id" character varying NULL, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "subcontrol_control_implementations" table
CREATE TABLE "subcontrol_control_implementations" ("subcontrol_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_implementation_id"), CONSTRAINT "subcontrol_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_control_implementations_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "subcontrol_control_implementations" table
DROP TABLE "subcontrol_control_implementations";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_groups_delegate", DROP CONSTRAINT "subcontrols_groups_control_owner", DROP COLUMN "delegate_id", DROP COLUMN "control_owner_id", ADD COLUMN "subcontrol_delegate" character varying NULL, ADD COLUMN "subcontrol_control_owner" character varying NULL, ALTER COLUMN "status" DROP DEFAULT, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("subcontrol_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("subcontrol_control_owner") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_groups_stakeholder", DROP CONSTRAINT "risks_groups_delegate", DROP COLUMN "delegate_id", DROP COLUMN "stakeholder_id", ADD COLUMN "risk_delegate" character varying NULL, ADD COLUMN "risk_stakeholder" character varying NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("risk_stakeholder") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("risk_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_groups_delegate", DROP CONSTRAINT "procedures_groups_approver", DROP COLUMN "delegate_id", DROP COLUMN "approver_id", ADD COLUMN "procedure_delegate" character varying NULL, ADD COLUMN "procedure_approver" character varying NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("procedure_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("procedure_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_groups_delegate", DROP CONSTRAINT "internal_policies_groups_approver", DROP COLUMN "delegate_id", DROP COLUMN "approver_id", ADD COLUMN "internal_policy_delegate" character varying NULL, ADD COLUMN "internal_policy_approver" character varying NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("internal_policy_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("internal_policy_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_groups_delegate", DROP CONSTRAINT "controls_groups_control_owner", DROP COLUMN "delegate_id", DROP COLUMN "control_owner_id", ADD COLUMN "control_delegate" character varying NULL, ADD COLUMN "control_control_owner" character varying NULL, ALTER COLUMN "status" DROP DEFAULT, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("control_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_control_owner") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP CONSTRAINT "control_implementations_organizations_control_implementations", DROP COLUMN "owner_id";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_groups_delegate", DROP CONSTRAINT "action_plans_groups_approver", DROP COLUMN "delegate_id", DROP COLUMN "approver_id", ADD COLUMN "action_plan_delegate" character varying NULL, ADD COLUMN "action_plan_approver" character varying NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("action_plan_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("action_plan_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "delegate_id", DROP COLUMN "control_owner_id", ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "delegate_id", DROP COLUMN "approver_id";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "delegate_id", DROP COLUMN "approver_id";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "delegate_id", DROP COLUMN "approver_id";
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" DROP COLUMN "owner_id";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "delegate_id", DROP COLUMN "control_owner_id", ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "delegate_id", DROP COLUMN "stakeholder_id";
