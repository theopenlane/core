-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'OPEN', DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "category" character varying NULL, ADD COLUMN "score" bigint NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ALTER COLUMN "status" SET DEFAULT 'DRAFT';
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ALTER COLUMN "status" SET DEFAULT 'DRAFT';
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "version", ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "version", ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- modify "standards" table
ALTER TABLE "standards" ALTER COLUMN "status" SET DEFAULT 'ACTIVE', ALTER COLUMN "revision" SET DEFAULT 'v0.0.1', ADD COLUMN "governing_body_logo_url" character varying NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "status" character varying NULL DEFAULT 'READY';
-- modify "standard_history" table
ALTER TABLE "standard_history" ALTER COLUMN "status" SET DEFAULT 'ACTIVE', ALTER COLUMN "revision" SET DEFAULT 'v0.0.1', ADD COLUMN "governing_body_logo_url" character varying NULL;
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "status" character varying NULL DEFAULT 'READY';
-- modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1';
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', ALTER COLUMN "details" TYPE text, ADD COLUMN "action_plan_type" character varying NULL, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_due" timestamptz NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "owner_id" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', ALTER COLUMN "details" TYPE text, ADD COLUMN "action_plan_type" character varying NULL, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_due" timestamptz NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "action_plan_approver" character varying NULL, ADD COLUMN "action_plan_delegate" character varying NULL, ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("action_plan_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("action_plan_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_organizations_action_plans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "internal_policy_approver" character varying NULL, ADD COLUMN "internal_policy_delegate" character varying NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("internal_policy_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("internal_policy_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'DRAFT', DROP COLUMN "version", DROP COLUMN "purpose_and_scope", DROP COLUMN "background", DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "approval_required" boolean NULL DEFAULT true, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "revision" character varying NULL DEFAULT 'v0.0.1', ADD COLUMN "procedure_approver" character varying NULL, ADD COLUMN "procedure_delegate" character varying NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("procedure_approver") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("procedure_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "description", ALTER COLUMN "status" SET DEFAULT 'OPEN', DROP COLUMN "satisfies", ALTER COLUMN "details" TYPE text, ADD COLUMN "category" character varying NULL, ADD COLUMN "score" bigint NULL, ADD COLUMN "risk_stakeholder" character varying NULL, ADD COLUMN "risk_delegate" character varying NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("risk_delegate") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("risk_stakeholder") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_groups_stakeholder", DROP CONSTRAINT "risks_groups_delegate", DROP COLUMN "risk_delegate", DROP COLUMN "risk_stakeholder", DROP COLUMN "score", DROP COLUMN "category", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "satisfies" text NULL, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_groups_delegate", DROP CONSTRAINT "procedures_groups_approver", DROP COLUMN "procedure_delegate", DROP COLUMN "procedure_approver", DROP COLUMN "revision", DROP COLUMN "review_frequency", DROP COLUMN "approval_required", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "satisfies" text NULL, ADD COLUMN "background" text NULL, ADD COLUMN "purpose_and_scope" text NULL, ADD COLUMN "version" character varying NULL, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_groups_delegate", DROP CONSTRAINT "internal_policies_groups_approver", DROP COLUMN "internal_policy_delegate", DROP COLUMN "internal_policy_approver", DROP COLUMN "revision", DROP COLUMN "review_frequency", DROP COLUMN "approval_required", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "background" text NULL, ADD COLUMN "purpose_and_scope" text NULL, ADD COLUMN "version" character varying NULL, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_organizations_action_plans", DROP CONSTRAINT "action_plans_groups_delegate", DROP CONSTRAINT "action_plans_groups_approver", DROP COLUMN "owner_id", DROP COLUMN "action_plan_delegate", DROP COLUMN "action_plan_approver", DROP COLUMN "revision", DROP COLUMN "review_frequency", DROP COLUMN "review_due", DROP COLUMN "approval_required", DROP COLUMN "action_plan_type", ALTER COLUMN "details" TYPE jsonb, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "owner_id", DROP COLUMN "revision", DROP COLUMN "review_frequency", DROP COLUMN "review_due", DROP COLUMN "approval_required", DROP COLUMN "action_plan_type", ALTER COLUMN "details" TYPE jsonb, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "revision", DROP COLUMN "review_frequency", DROP COLUMN "approval_required", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "satisfies" text NULL, ADD COLUMN "background" text NULL, ADD COLUMN "purpose_and_scope" text NULL, ADD COLUMN "version" character varying NULL, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "status";
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "governing_body_logo_url", ALTER COLUMN "revision" DROP DEFAULT, ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "status";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "governing_body_logo_url", ALTER COLUMN "revision" DROP DEFAULT, ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "revision", ADD COLUMN "version" character varying NULL;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "revision", ADD COLUMN "version" character varying NULL;
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "revision", DROP COLUMN "review_frequency", DROP COLUMN "approval_required", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "background" text NULL, ADD COLUMN "purpose_and_scope" text NULL, ADD COLUMN "version" character varying NULL, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "score", DROP COLUMN "category", ALTER COLUMN "details" TYPE jsonb, ADD COLUMN "satisfies" text NULL, ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "description" text NULL;
