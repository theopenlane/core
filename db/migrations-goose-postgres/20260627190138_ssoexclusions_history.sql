-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "check_result_history" table
ALTER TABLE "check_result_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "discussion_history" table
ALTER TABLE "discussion_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "document_data_history" table
ALTER TABLE "document_data_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "entity_type_history" table
ALTER TABLE "entity_type_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "finding_control_history" table
ALTER TABLE "finding_control_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "group_membership_history" table
ALTER TABLE "group_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "group_setting_history" table
ALTER TABLE "group_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "job_template_history" table
ALTER TABLE "job_template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "mappable_domain_history" table
ALTER TABLE "mappable_domain_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "notification_preference_history" table
ALTER TABLE "notification_preference_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt" boolean NULL DEFAULT false, ADD COLUMN "sso_exempt_reason" character varying NULL, ADD COLUMN "sso_exempt_granted_by" character varying NULL, ADD COLUMN "sso_exempt_granted_at" timestamptz NULL;
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "slug_name" character varying NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "identity_provider_jit_provisioning" boolean NOT NULL DEFAULT true, ADD COLUMN "jit_allowed_email_domains" jsonb NULL, ADD COLUMN "sso_exempt_domains" jsonb NULL, ADD COLUMN "allow_support_access" boolean NULL DEFAULT false;
-- modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "program_membership_history" table
ALTER TABLE "program_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "review_history" table
ALTER TABLE "review_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "scan_history" table
ALTER TABLE "scan_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "sla_definition_history" table
ALTER TABLE "sla_definition_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "system_detail_history" table
ALTER TABLE "system_detail_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_compliance_history" table
ALTER TABLE "trust_center_compliance_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_entity_history" table
ALTER TABLE "trust_center_entity_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_faqs_history" table
ALTER TABLE "trust_center_faqs_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "user_setting_history" table
ALTER TABLE "user_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "vendor_risk_score_history" table
ALTER TABLE "vendor_risk_score_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "vendor_scoring_config_history" table
ALTER TABLE "vendor_scoring_config_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_assignment_target_history" table
ALTER TABLE "workflow_assignment_target_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_event_history" table
ALTER TABLE "workflow_event_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "updated_by_impersonator" character varying NULL;

-- +goose Down
-- reverse: modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_event_history" table
ALTER TABLE "workflow_event_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_assignment_target_history" table
ALTER TABLE "workflow_assignment_target_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "vendor_scoring_config_history" table
ALTER TABLE "vendor_scoring_config_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "vendor_risk_score_history" table
ALTER TABLE "vendor_risk_score_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_history" table
ALTER TABLE "trust_center_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_faqs_history" table
ALTER TABLE "trust_center_faqs_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_entity_history" table
ALTER TABLE "trust_center_entity_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_compliance_history" table
ALTER TABLE "trust_center_compliance_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "system_detail_history" table
ALTER TABLE "system_detail_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "sla_definition_history" table
ALTER TABLE "sla_definition_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "scan_history" table
ALTER TABLE "scan_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "review_history" table
ALTER TABLE "review_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "remediation_history" table
ALTER TABLE "remediation_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "platform_history" table
ALTER TABLE "platform_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "allow_support_access", DROP COLUMN "sso_exempt_domains", DROP COLUMN "jit_allowed_email_domains", DROP COLUMN "identity_provider_jit_provisioning", DROP COLUMN "updated_by_impersonator";
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "slug_name", DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "sso_exempt_granted_at", DROP COLUMN "sso_exempt_granted_by", DROP COLUMN "sso_exempt_reason", DROP COLUMN "sso_exempt", DROP COLUMN "updated_by_impersonator";
-- reverse: modify "notification_template_history" table
ALTER TABLE "notification_template_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "notification_preference_history" table
ALTER TABLE "notification_preference_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "mappable_domain_history" table
ALTER TABLE "mappable_domain_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "job_template_history" table
ALTER TABLE "job_template_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "finding_control_history" table
ALTER TABLE "finding_control_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "discussion_history" table
ALTER TABLE "discussion_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_group_history" table
ALTER TABLE "directory_group_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_account_history" table
ALTER TABLE "directory_account_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "check_result_history" table
ALTER TABLE "check_result_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "updated_by_impersonator";
