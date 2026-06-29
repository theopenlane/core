-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "campaign_history" table
ALTER TABLE "campaign_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "check_result_history" table
ALTER TABLE "check_result_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "control_implementation_history" table
ALTER TABLE "control_implementation_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "discussion_history" table
ALTER TABLE "discussion_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "dns_verification_history" table
ALTER TABLE "dns_verification_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "document_data_history" table
ALTER TABLE "document_data_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "email_template_history" table
ALTER TABLE "email_template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "entity_type_history" table
ALTER TABLE "entity_type_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "finding_control_history" table
ALTER TABLE "finding_control_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "group_membership_history" table
ALTER TABLE "group_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "group_setting_history" table
ALTER TABLE "group_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "job_template_history" table
ALTER TABLE "job_template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "mappable_domain_history" table
ALTER TABLE "mappable_domain_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "notification_preference_history" table
ALTER TABLE "notification_preference_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "notification_template_history" table
ALTER TABLE "notification_template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt" boolean NULL DEFAULT false, ADD COLUMN "sso_exempt_reason" character varying NULL, ADD COLUMN "sso_exempt_granted_by" character varying NULL, ADD COLUMN "sso_exempt_granted_at" timestamptz NULL;
-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "slug_name" character varying NULL;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "identity_provider_jit_provisioning" boolean NOT NULL DEFAULT true, ADD COLUMN "jit_allowed_email_domains" jsonb NULL, ADD COLUMN "sso_exempt_domains" jsonb NULL, ADD COLUMN "allow_support_access" boolean NULL DEFAULT false;
-- Modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "program_membership_history" table
ALTER TABLE "program_membership_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "review_history" table
ALTER TABLE "review_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "scan_history" table
ALTER TABLE "scan_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "sla_definition_history" table
ALTER TABLE "sla_definition_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "system_detail_history" table
ALTER TABLE "system_detail_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_compliance_history" table
ALTER TABLE "trust_center_compliance_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_entity_history" table
ALTER TABLE "trust_center_entity_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_faqs_history" table
ALTER TABLE "trust_center_faqs_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "user_setting_history" table
ALTER TABLE "user_setting_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "vendor_risk_score_history" table
ALTER TABLE "vendor_risk_score_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "vendor_scoring_config_history" table
ALTER TABLE "vendor_scoring_config_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_assignment_target_history" table
ALTER TABLE "workflow_assignment_target_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_event_history" table
ALTER TABLE "workflow_event_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "updated_by_impersonator" character varying NULL;
