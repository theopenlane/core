-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "external_id" character varying NULL;
-- modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "assessment_policy_history" table
ALTER TABLE "assessment_policy_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "external_id" character varying NULL;
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "external_id" character varying NULL;
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "external_id" character varying NULL;
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "review_history" table
ALTER TABLE "review_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "system_detail_history" table
ALTER TABLE "system_detail_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;

-- +goose Down
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "system_detail_history" table
ALTER TABLE "system_detail_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "review_history" table
ALTER TABLE "review_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "remediation_history" table
ALTER TABLE "remediation_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "external_id";
-- reverse: modify "platform_history" table
ALTER TABLE "platform_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "external_id";
-- reverse: modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "external_id";
-- reverse: modify "assessment_policy_history" table
ALTER TABLE "assessment_policy_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "external_id";
