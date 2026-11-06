-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "source_key" character varying NULL;
-- modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "source_key" character varying NULL;
-- modify "assessment_policy_history" table
ALTER TABLE "assessment_policy_history" ADD COLUMN "source_key" character varying NULL;
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "source_key" character varying NULL;
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "source_key" character varying NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "source_key" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "source_key" character varying NULL;
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "source_key" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "source_key" character varying NULL;
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "source_key" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "source_key" character varying NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "source_key" character varying NULL;
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "source_key" character varying NULL;
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "source_key" character varying NULL;
-- modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "source_key" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "source_key" character varying NULL;
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "source_key" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "source_key" character varying NULL;
-- modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "source_key" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "source_key" character varying NULL;
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "source_key" character varying NULL;
-- modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "source_key" character varying NULL;
-- modify "review_history" table
ALTER TABLE "review_history" ADD COLUMN "source_key" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "source_key" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "source_key" character varying NULL;
-- modify "system_detail_history" table
ALTER TABLE "system_detail_history" ADD COLUMN "source_key" character varying NULL;
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "source_key" character varying NULL;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "source_key" character varying NULL;

-- +goose Down
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "source_key";
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "source_key";
-- reverse: modify "system_detail_history" table
ALTER TABLE "system_detail_history" DROP COLUMN "source_key";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "source_key";
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "source_key";
-- reverse: modify "review_history" table
ALTER TABLE "review_history" DROP COLUMN "source_key";
-- reverse: modify "remediation_history" table
ALTER TABLE "remediation_history" DROP COLUMN "source_key";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "source_key";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "source_key";
-- reverse: modify "platform_history" table
ALTER TABLE "platform_history" DROP COLUMN "source_key";
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "source_key";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "source_key";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "source_key";
-- reverse: modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" DROP COLUMN "source_key";
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "source_key";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "source_key";
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "source_key";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "source_key";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "source_key";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "source_key";
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" DROP COLUMN "source_key";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "source_key";
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "source_key";
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "source_key";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "source_key";
-- reverse: modify "assessment_policy_history" table
ALTER TABLE "assessment_policy_history" DROP COLUMN "source_key";
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "source_key";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "source_key";
