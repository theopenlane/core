-- Modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'OPEN', DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "category" character varying NULL, ADD COLUMN "score" bigint NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- Modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ALTER COLUMN "status" SET DEFAULT 'DRAFT';
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ALTER COLUMN "status" SET DEFAULT 'DRAFT';
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "version", ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "version", ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- Modify "standards" table
ALTER TABLE "standards" ALTER COLUMN "status" SET DEFAULT 'ACTIVE', ALTER COLUMN "revision" SET DEFAULT 'v0.0.1', ADD COLUMN "governing_body_logo_url" character varying NULL;
-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "status" character varying NULL DEFAULT 'READY';
-- Modify "standard_history" table
ALTER TABLE "standard_history" ALTER COLUMN "status" SET DEFAULT 'ACTIVE', ALTER COLUMN "revision" SET DEFAULT 'v0.0.1', ADD COLUMN "governing_body_logo_url" character varying NULL;
-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "status" character varying NULL DEFAULT 'READY';
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', ALTER COLUMN "details" TYPE text, ADD COLUMN "action_plan_type" character varying NULL, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_due" timestamptz NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "owner_id" character varying NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', ALTER COLUMN "details" TYPE text, ADD COLUMN "action_plan_type" character varying NULL, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_due" timestamptz NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "action_plan_approver" character varying NULL, ADD COLUMN "action_plan_delegate" character varying NULL, ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("action_plan_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("action_plan_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_organizations_action_plans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "internal_policy_approver" character varying NULL, ADD COLUMN "internal_policy_delegate" character varying NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("internal_policy_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("internal_policy_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "procedure_approver" character varying NULL, ADD COLUMN "procedure_delegate" character varying NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("procedure_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("procedure_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "risks" table
ALTER TABLE "risks" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'OPEN', DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "category" character varying NULL, ADD COLUMN "score" bigint NULL, ADD COLUMN "risk_stakeholder" character varying NULL, ADD COLUMN "risk_delegate" character varying NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("risk_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("risk_stakeholder") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
