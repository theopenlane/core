-- Create index "action_plan_blocked_groups_group_id_idx" to table: "action_plan_blocked_groups"
CREATE INDEX "action_plan_blocked_groups_group_id_idx" ON "action_plan_blocked_groups" ("group_id");
-- Create index "action_plan_editors_group_id_idx" to table: "action_plan_editors"
CREATE INDEX "action_plan_editors_group_id_idx" ON "action_plan_editors" ("group_id");
-- Create index "action_plan_tasks_task_id_idx" to table: "action_plan_tasks"
CREATE INDEX "action_plan_tasks_task_id_idx" ON "action_plan_tasks" ("task_id");
-- Create index "action_plan_viewers_group_id_idx" to table: "action_plan_viewers"
CREATE INDEX "action_plan_viewers_group_id_idx" ON "action_plan_viewers" ("group_id");
-- Drop index "actionplan_owner_id" from table: "action_plans"
DROP INDEX "actionplan_owner_id";
-- Create index "action_plan_file_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_file_id_idx" ON "action_plans" ("file_id");
-- Create index "action_plan_owner_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_owner_id_idx" ON "action_plans" ("owner_id");
-- Drop index "apitoken_owner_id" from table: "api_tokens"
DROP INDEX "apitoken_owner_id";
-- Create index "api_token_owner_id_idx" to table: "api_tokens"
CREATE INDEX "api_token_owner_id_idx" ON "api_tokens" ("owner_id");
-- Drop index "assessmentresponse_owner_id" from table: "assessment_responses"
DROP INDEX "assessmentresponse_owner_id";
-- Create index "assessment_response_document_data_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_document_data_id_idx" ON "assessment_responses" ("document_data_id");
-- Create index "assessment_response_owner_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_owner_id_idx" ON "assessment_responses" ("owner_id");
-- Drop index "assessment_owner_id" from table: "assessments"
DROP INDEX "assessment_owner_id";
-- Create index "assessment_owner_id_idx" to table: "assessments"
CREATE INDEX "assessment_owner_id_idx" ON "assessments" ("owner_id");
-- Create index "assessment_template_id_idx" to table: "assessments"
CREATE INDEX "assessment_template_id_idx" ON "assessments" ("template_id");
-- Create index "asset_connected_assets_connected_from_id_idx" to table: "asset_connected_assets"
CREATE INDEX "asset_connected_assets_connected_from_id_idx" ON "asset_connected_assets" ("connected_from_id");
-- Drop index "asset_owner_id" from table: "assets"
DROP INDEX "asset_owner_id";
-- Create index "asset_integration_id_idx" to table: "assets"
CREATE INDEX "asset_integration_id_idx" ON "assets" ("integration_id");
-- Create index "asset_owner_id_idx" to table: "assets"
CREATE INDEX "asset_owner_id_idx" ON "assets" ("owner_id");
-- Create index "asset_source_platform_id_idx" to table: "assets"
CREATE INDEX "asset_source_platform_id_idx" ON "assets" ("source_platform_id");
-- Create index "campaign_blocked_groups_group_id_idx" to table: "campaign_blocked_groups"
CREATE INDEX "campaign_blocked_groups_group_id_idx" ON "campaign_blocked_groups" ("group_id");
-- Create index "campaign_contacts_contact_id_idx" to table: "campaign_contacts"
CREATE INDEX "campaign_contacts_contact_id_idx" ON "campaign_contacts" ("contact_id");
-- Create index "campaign_editors_group_id_idx" to table: "campaign_editors"
CREATE INDEX "campaign_editors_group_id_idx" ON "campaign_editors" ("group_id");
-- Create index "campaign_groups_group_id_idx" to table: "campaign_groups"
CREATE INDEX "campaign_groups_group_id_idx" ON "campaign_groups" ("group_id");
-- Create index "campaign_identity_holders_identity_holder_id_idx" to table: "campaign_identity_holders"
CREATE INDEX "campaign_identity_holders_identity_holder_id_idx" ON "campaign_identity_holders" ("identity_holder_id");
-- Drop index "campaigntarget_owner_id" from table: "campaign_targets"
DROP INDEX "campaigntarget_owner_id";
-- Create index "campaign_target_owner_id_idx" to table: "campaign_targets"
CREATE INDEX "campaign_target_owner_id_idx" ON "campaign_targets" ("owner_id");
-- Create index "campaign_users_user_id_idx" to table: "campaign_users"
CREATE INDEX "campaign_users_user_id_idx" ON "campaign_users" ("user_id");
-- Create index "campaign_viewers_group_id_idx" to table: "campaign_viewers"
CREATE INDEX "campaign_viewers_group_id_idx" ON "campaign_viewers" ("group_id");
-- Drop index "campaign_owner_id" from table: "campaigns"
DROP INDEX "campaign_owner_id";
-- Create index "campaign_assessment_id_idx" to table: "campaigns"
CREATE INDEX "campaign_assessment_id_idx" ON "campaigns" ("assessment_id");
-- Create index "campaign_email_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_email_template_id_idx" ON "campaigns" ("email_template_id");
-- Create index "campaign_integration_id_idx" to table: "campaigns"
CREATE INDEX "campaign_integration_id_idx" ON "campaigns" ("integration_id");
-- Create index "campaign_owner_id_idx" to table: "campaigns"
CREATE INDEX "campaign_owner_id_idx" ON "campaigns" ("owner_id");
-- Create index "campaign_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_template_id_idx" ON "campaigns" ("template_id");
-- Create index "campaign_trust_center_id_idx" to table: "campaigns"
CREATE INDEX "campaign_trust_center_id_idx" ON "campaigns" ("trust_center_id");
-- Create index "check_result_controls_control_id_idx" to table: "check_result_controls"
CREATE INDEX "check_result_controls_control_id_idx" ON "check_result_controls" ("control_id");
-- Create index "check_result_integration_id_idx" to table: "check_results"
CREATE INDEX "check_result_integration_id_idx" ON "check_results" ("integration_id");
-- Create index "contact_files_file_id_idx" to table: "contact_files"
CREATE INDEX "contact_files_file_id_idx" ON "contact_files" ("file_id");
-- Drop index "contact_owner_id" from table: "contacts"
DROP INDEX "contact_owner_id";
-- Create index "contact_owner_id_idx" to table: "contacts"
CREATE INDEX "contact_owner_id_idx" ON "contacts" ("owner_id");
-- Create index "control_action_plans_action_plan_id_idx" to table: "control_action_plans"
CREATE INDEX "control_action_plans_action_plan_id_idx" ON "control_action_plans" ("action_plan_id");
-- Create index "control_assets_asset_id_idx" to table: "control_assets"
CREATE INDEX "control_assets_asset_id_idx" ON "control_assets" ("asset_id");
-- Create index "control_blocked_groups_group_id_idx" to table: "control_blocked_groups"
CREATE INDEX "control_blocked_groups_group_id_idx" ON "control_blocked_groups" ("group_id");
-- Create index "control_campaigns_campaign_id_idx" to table: "control_campaigns"
CREATE INDEX "control_campaigns_campaign_id_idx" ON "control_campaigns" ("campaign_id");
-- Create index "control_control_implementations_control_implementation_id_idx" to table: "control_control_implementations"
CREATE INDEX "control_control_implementations_control_implementation_id_idx" ON "control_control_implementations" ("control_implementation_id");
-- Create index "control_control_objectives_control_objective_id_idx" to table: "control_control_objectives"
CREATE INDEX "control_control_objectives_control_objective_id_idx" ON "control_control_objectives" ("control_objective_id");
-- Create index "control_editors_group_id_idx" to table: "control_editors"
CREATE INDEX "control_editors_group_id_idx" ON "control_editors" ("group_id");
-- Create index "control_entities_entity_id_idx" to table: "control_entities"
CREATE INDEX "control_entities_entity_id_idx" ON "control_entities" ("entity_id");
-- Create index "control_identity_holders_identity_holder_id_idx" to table: "control_identity_holders"
CREATE INDEX "control_identity_holders_identity_holder_id_idx" ON "control_identity_holders" ("identity_holder_id");
-- Create index "control_implementation_blocked_groups_group_id_idx" to table: "control_implementation_blocked_groups"
CREATE INDEX "control_implementation_blocked_groups_group_id_idx" ON "control_implementation_blocked_groups" ("group_id");
-- Create index "control_implementation_editors_group_id_idx" to table: "control_implementation_editors"
CREATE INDEX "control_implementation_editors_group_id_idx" ON "control_implementation_editors" ("group_id");
-- Create index "control_implementation_tasks_task_id_idx" to table: "control_implementation_tasks"
CREATE INDEX "control_implementation_tasks_task_id_idx" ON "control_implementation_tasks" ("task_id");
-- Create index "control_implementation_viewers_group_id_idx" to table: "control_implementation_viewers"
CREATE INDEX "control_implementation_viewers_group_id_idx" ON "control_implementation_viewers" ("group_id");
-- Drop index "controlimplementation_owner_id" from table: "control_implementations"
DROP INDEX "controlimplementation_owner_id";
-- Create index "control_implementation_owner_id_idx" to table: "control_implementations"
CREATE INDEX "control_implementation_owner_id_idx" ON "control_implementations" ("owner_id");
-- Create index "control_narratives_narrative_id_idx" to table: "control_narratives"
CREATE INDEX "control_narratives_narrative_id_idx" ON "control_narratives" ("narrative_id");
-- Create index "control_objective_blocked_groups_group_id_idx" to table: "control_objective_blocked_groups"
CREATE INDEX "control_objective_blocked_groups_group_id_idx" ON "control_objective_blocked_groups" ("group_id");
-- Create index "control_objective_editors_group_id_idx" to table: "control_objective_editors"
CREATE INDEX "control_objective_editors_group_id_idx" ON "control_objective_editors" ("group_id");
-- Create index "control_objective_tasks_task_id_idx" to table: "control_objective_tasks"
CREATE INDEX "control_objective_tasks_task_id_idx" ON "control_objective_tasks" ("task_id");
-- Create index "control_objective_viewers_group_id_idx" to table: "control_objective_viewers"
CREATE INDEX "control_objective_viewers_group_id_idx" ON "control_objective_viewers" ("group_id");
-- Drop index "controlobjective_owner_id" from table: "control_objectives"
DROP INDEX "controlobjective_owner_id";
-- Create index "control_objective_owner_id_idx" to table: "control_objectives"
CREATE INDEX "control_objective_owner_id_idx" ON "control_objectives" ("owner_id");
-- Create index "control_procedures_procedure_id_idx" to table: "control_procedures"
CREATE INDEX "control_procedures_procedure_id_idx" ON "control_procedures" ("procedure_id");
-- Create index "control_risks_risk_id_idx" to table: "control_risks"
CREATE INDEX "control_risks_risk_id_idx" ON "control_risks" ("risk_id");
-- Create index "control_scans_scan_id_idx" to table: "control_scans"
CREATE INDEX "control_scans_scan_id_idx" ON "control_scans" ("scan_id");
-- Create index "control_tasks_task_id_idx" to table: "control_tasks"
CREATE INDEX "control_tasks_task_id_idx" ON "control_tasks" ("task_id");
-- Drop index "control_owner_id" from table: "controls"
DROP INDEX "control_owner_id";
-- Create index "control_owner_id_idx" to table: "controls"
CREATE INDEX "control_owner_id_idx" ON "controls" ("owner_id");
-- Drop index "customdomain_owner_id" from table: "custom_domains"
DROP INDEX "customdomain_owner_id";
-- Create index "custom_domain_dns_verification_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_dns_verification_id_idx" ON "custom_domains" ("dns_verification_id");
-- Create index "custom_domain_mappable_domain_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_mappable_domain_id_idx" ON "custom_domains" ("mappable_domain_id");
-- Create index "custom_domain_owner_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_owner_id_idx" ON "custom_domains" ("owner_id");
-- Drop index "customtypeenum_owner_id" from table: "custom_type_enums"
DROP INDEX "customtypeenum_owner_id";
-- Create index "custom_type_enum_owner_id_idx" to table: "custom_type_enums"
CREATE INDEX "custom_type_enum_owner_id_idx" ON "custom_type_enums" ("owner_id");
-- Create index "directory_account_avatar_local_file_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_avatar_local_file_id_idx" ON "directory_accounts" ("avatar_local_file_id");
-- Create index "directory_account_owner_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_owner_id_idx" ON "directory_accounts" ("owner_id");
-- Create index "directory_group_owner_id_idx" to table: "directory_groups"
CREATE INDEX "directory_group_owner_id_idx" ON "directory_groups" ("owner_id");
-- Create index "directory_membership_directory_group_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_directory_group_id_idx" ON "directory_memberships" ("directory_group_id");
-- Create index "directory_membership_owner_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_owner_id_idx" ON "directory_memberships" ("owner_id");
-- Create index "directory_sync_run_owner_id_idx" to table: "directory_sync_runs"
CREATE INDEX "directory_sync_run_owner_id_idx" ON "directory_sync_runs" ("owner_id");
-- Drop index "discussion_owner_id" from table: "discussions"
DROP INDEX "discussion_owner_id";
-- Create index "discussion_owner_id_idx" to table: "discussions"
CREATE INDEX "discussion_owner_id_idx" ON "discussions" ("owner_id");
-- Drop index "dnsverification_owner_id" from table: "dns_verifications"
DROP INDEX "dnsverification_owner_id";
-- Create index "dns_verification_owner_id_idx" to table: "dns_verifications"
CREATE INDEX "dns_verification_owner_id_idx" ON "dns_verifications" ("owner_id");
-- Drop index "documentdata_owner_id" from table: "document_data"
DROP INDEX "documentdata_owner_id";
-- Create index "document_owner_id_idx" to table: "document_data"
CREATE INDEX "document_owner_id_idx" ON "document_data" ("owner_id");
-- Create index "document_template_id_idx" to table: "document_data"
CREATE INDEX "document_template_id_idx" ON "document_data" ("template_id");
-- Create index "document_data_files_file_id_idx" to table: "document_data_files"
CREATE INDEX "document_data_files_file_id_idx" ON "document_data_files" ("file_id");
-- Drop index "emailtemplate_owner_id" from table: "email_templates"
DROP INDEX "emailtemplate_owner_id";
-- Create index "email_template_integration_id_idx" to table: "email_templates"
CREATE INDEX "email_template_integration_id_idx" ON "email_templates" ("integration_id");
-- Create index "email_template_owner_id_idx" to table: "email_templates"
CREATE INDEX "email_template_owner_id_idx" ON "email_templates" ("owner_id");
-- Create index "email_template_trust_center_id_idx" to table: "email_templates"
CREATE INDEX "email_template_trust_center_id_idx" ON "email_templates" ("trust_center_id");
-- Create index "email_template_workflow_definition_id_idx" to table: "email_templates"
CREATE INDEX "email_template_workflow_definition_id_idx" ON "email_templates" ("workflow_definition_id");
-- Create index "email_template_workflow_instance_id_idx" to table: "email_templates"
CREATE INDEX "email_template_workflow_instance_id_idx" ON "email_templates" ("workflow_instance_id");
-- Create index "email_verification_tokens_owner_id_fk" to table: "email_verification_tokens"
CREATE INDEX "email_verification_tokens_owner_id_fk" ON "email_verification_tokens" ("owner_id");
-- Drop index "entity_owner_id" from table: "entities"
DROP INDEX "entity_owner_id";
-- Create index "entity_entity_type_id_idx" to table: "entities"
CREATE INDEX "entity_entity_type_id_idx" ON "entities" ("entity_type_id");
-- Create index "entity_logo_file_id_idx" to table: "entities"
CREATE INDEX "entity_logo_file_id_idx" ON "entities" ("logo_file_id");
-- Create index "entity_owner_id_idx" to table: "entities"
CREATE INDEX "entity_owner_id_idx" ON "entities" ("owner_id");
-- Create index "entity_assets_asset_id_idx" to table: "entity_assets"
CREATE INDEX "entity_assets_asset_id_idx" ON "entity_assets" ("asset_id");
-- Create index "entity_blocked_groups_group_id_idx" to table: "entity_blocked_groups"
CREATE INDEX "entity_blocked_groups_group_id_idx" ON "entity_blocked_groups" ("group_id");
-- Create index "entity_contacts_contact_id_idx" to table: "entity_contacts"
CREATE INDEX "entity_contacts_contact_id_idx" ON "entity_contacts" ("contact_id");
-- Create index "entity_documents_document_data_id_idx" to table: "entity_documents"
CREATE INDEX "entity_documents_document_data_id_idx" ON "entity_documents" ("document_data_id");
-- Create index "entity_editors_group_id_idx" to table: "entity_editors"
CREATE INDEX "entity_editors_group_id_idx" ON "entity_editors" ("group_id");
-- Create index "entity_files_file_id_idx" to table: "entity_files"
CREATE INDEX "entity_files_file_id_idx" ON "entity_files" ("file_id");
-- Create index "entity_integrations_integration_id_idx" to table: "entity_integrations"
CREATE INDEX "entity_integrations_integration_id_idx" ON "entity_integrations" ("integration_id");
-- Create index "entity_subprocessors_subprocessor_id_idx" to table: "entity_subprocessors"
CREATE INDEX "entity_subprocessors_subprocessor_id_idx" ON "entity_subprocessors" ("subprocessor_id");
-- Create index "entity_system_details_system_detail_id_idx" to table: "entity_system_details"
CREATE INDEX "entity_system_details_system_detail_id_idx" ON "entity_system_details" ("system_detail_id");
-- Drop index "entitytype_owner_id" from table: "entity_types"
DROP INDEX "entitytype_owner_id";
-- Create index "entity_type_owner_id_idx" to table: "entity_types"
CREATE INDEX "entity_type_owner_id_idx" ON "entity_types" ("owner_id");
-- Create index "evidence_control_objectives_control_objective_id_idx" to table: "evidence_control_objectives"
CREATE INDEX "evidence_control_objectives_control_objective_id_idx" ON "evidence_control_objectives" ("control_objective_id");
-- Create index "evidence_controls_control_id_idx" to table: "evidence_controls"
CREATE INDEX "evidence_controls_control_id_idx" ON "evidence_controls" ("control_id");
-- Create index "evidence_files_file_id_idx" to table: "evidence_files"
CREATE INDEX "evidence_files_file_id_idx" ON "evidence_files" ("file_id");
-- Create index "evidence_subcontrols_subcontrol_id_idx" to table: "evidence_subcontrols"
CREATE INDEX "evidence_subcontrols_subcontrol_id_idx" ON "evidence_subcontrols" ("subcontrol_id");
-- Drop index "evidence_owner_id" from table: "evidences"
DROP INDEX "evidence_owner_id";
-- Create index "evidence_owner_id_idx" to table: "evidences"
CREATE INDEX "evidence_owner_id_idx" ON "evidences" ("owner_id");
-- Drop index "export_owner_id" from table: "exports"
DROP INDEX "export_owner_id";
-- Create index "export_owner_id_idx" to table: "exports"
CREATE INDEX "export_owner_id_idx" ON "exports" ("owner_id");
-- Create index "file_download_tokens_owner_id_fk" to table: "file_download_tokens"
CREATE INDEX "file_download_tokens_owner_id_fk" ON "file_download_tokens" ("owner_id");
-- Create index "file_events_event_id_idx" to table: "file_events"
CREATE INDEX "file_events_event_id_idx" ON "file_events" ("event_id");
-- Create index "file_secrets_hush_id_idx" to table: "file_secrets"
CREATE INDEX "file_secrets_hush_id_idx" ON "file_secrets" ("hush_id");
-- Create index "finding_action_plans_action_plan_id_idx" to table: "finding_action_plans"
CREATE INDEX "finding_action_plans_action_plan_id_idx" ON "finding_action_plans" ("action_plan_id");
-- Create index "finding_assets_asset_id_idx" to table: "finding_assets"
CREATE INDEX "finding_assets_asset_id_idx" ON "finding_assets" ("asset_id");
-- Create index "finding_blocked_groups_group_id_idx" to table: "finding_blocked_groups"
CREATE INDEX "finding_blocked_groups_group_id_idx" ON "finding_blocked_groups" ("group_id");
-- Create index "finding_check_results_check_result_id_idx" to table: "finding_check_results"
CREATE INDEX "finding_check_results_check_result_id_idx" ON "finding_check_results" ("check_result_id");
-- Create index "finding_control_control_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_control_id_idx" ON "finding_controls" ("control_id");
-- Create index "finding_control_owner_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_owner_id_idx" ON "finding_controls" ("owner_id");
-- Create index "finding_control_standard_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_standard_id_idx" ON "finding_controls" ("standard_id");
-- Create index "finding_directory_accounts_directory_account_id_idx" to table: "finding_directory_accounts"
CREATE INDEX "finding_directory_accounts_directory_account_id_idx" ON "finding_directory_accounts" ("directory_account_id");
-- Create index "finding_editors_group_id_idx" to table: "finding_editors"
CREATE INDEX "finding_editors_group_id_idx" ON "finding_editors" ("group_id");
-- Create index "finding_entities_entity_id_idx" to table: "finding_entities"
CREATE INDEX "finding_entities_entity_id_idx" ON "finding_entities" ("entity_id");
-- Create index "finding_identity_holders_identity_holder_id_idx" to table: "finding_identity_holders"
CREATE INDEX "finding_identity_holders_identity_holder_id_idx" ON "finding_identity_holders" ("identity_holder_id");
-- Create index "finding_programs_program_id_idx" to table: "finding_programs"
CREATE INDEX "finding_programs_program_id_idx" ON "finding_programs" ("program_id");
-- Create index "finding_risks_risk_id_idx" to table: "finding_risks"
CREATE INDEX "finding_risks_risk_id_idx" ON "finding_risks" ("risk_id");
-- Create index "finding_scans_scan_id_idx" to table: "finding_scans"
CREATE INDEX "finding_scans_scan_id_idx" ON "finding_scans" ("scan_id");
-- Create index "finding_subcontrols_subcontrol_id_idx" to table: "finding_subcontrols"
CREATE INDEX "finding_subcontrols_subcontrol_id_idx" ON "finding_subcontrols" ("subcontrol_id");
-- Create index "finding_tasks_task_id_idx" to table: "finding_tasks"
CREATE INDEX "finding_tasks_task_id_idx" ON "finding_tasks" ("task_id");
-- Create index "finding_vulnerabilities_vulnerability_id_idx" to table: "finding_vulnerabilities"
CREATE INDEX "finding_vulnerabilities_vulnerability_id_idx" ON "finding_vulnerabilities" ("vulnerability_id");
-- Drop index "finding_owner_id" from table: "findings"
DROP INDEX "finding_owner_id";
-- Create index "finding_owner_id_idx" to table: "findings"
CREATE INDEX "finding_owner_id_idx" ON "findings" ("owner_id");
-- Create index "group_events_event_id_idx" to table: "group_events"
CREATE INDEX "group_events_event_id_idx" ON "group_events" ("event_id");
-- Create index "group_files_file_id_idx" to table: "group_files"
CREATE INDEX "group_files_file_id_idx" ON "group_files" ("file_id");
-- Create index "group_membership_events_event_id_idx" to table: "group_membership_events"
CREATE INDEX "group_membership_events_event_id_idx" ON "group_membership_events" ("event_id");
-- Create index "group_membership_group_id_idx" to table: "group_memberships"
CREATE INDEX "group_membership_group_id_idx" ON "group_memberships" ("group_id");
-- Create index "group_setting_group_id_idx" to table: "group_settings"
CREATE INDEX "group_setting_group_id_idx" ON "group_settings" ("group_id");
-- Create index "group_tasks_task_id_idx" to table: "group_tasks"
CREATE INDEX "group_tasks_task_id_idx" ON "group_tasks" ("task_id");
-- Drop index "group_owner_id" from table: "groups"
DROP INDEX "group_owner_id";
-- Create index "group_avatar_local_file_id_idx" to table: "groups"
CREATE INDEX "group_avatar_local_file_id_idx" ON "groups" ("avatar_local_file_id");
-- Create index "group_owner_id_idx" to table: "groups"
CREATE INDEX "group_owner_id_idx" ON "groups" ("owner_id");
-- Create index "hush_events_event_id_idx" to table: "hush_events"
CREATE INDEX "hush_events_event_id_idx" ON "hush_events" ("event_id");
-- Drop index "hush_owner_id" from table: "hushes"
DROP INDEX "hush_owner_id";
-- Create index "secret_owner_id_idx" to table: "hushes"
CREATE INDEX "secret_owner_id_idx" ON "hushes" ("owner_id");
-- Create index "identity_holder_assessments_assessment_id_idx" to table: "identity_holder_assessments"
CREATE INDEX "identity_holder_assessments_assessment_id_idx" ON "identity_holder_assessments" ("assessment_id");
-- Create index "identity_holder_assets_asset_id_idx" to table: "identity_holder_assets"
CREATE INDEX "identity_holder_assets_asset_id_idx" ON "identity_holder_assets" ("asset_id");
-- Create index "identity_holder_entities_entity_id_idx" to table: "identity_holder_entities"
CREATE INDEX "identity_holder_entities_entity_id_idx" ON "identity_holder_entities" ("entity_id");
-- Create index "identity_holder_files_file_id_idx" to table: "identity_holder_files"
CREATE INDEX "identity_holder_files_file_id_idx" ON "identity_holder_files" ("file_id");
-- Create index "identity_holder_tasks_task_id_idx" to table: "identity_holder_tasks"
CREATE INDEX "identity_holder_tasks_task_id_idx" ON "identity_holder_tasks" ("task_id");
-- Create index "identity_holder_templates_template_id_idx" to table: "identity_holder_templates"
CREATE INDEX "identity_holder_templates_template_id_idx" ON "identity_holder_templates" ("template_id");
-- Drop index "identityholder_owner_id" from table: "identity_holders"
DROP INDEX "identityholder_owner_id";
-- Create index "identity_holder_employer_entity_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_employer_entity_id_idx" ON "identity_holders" ("employer_entity_id");
-- Create index "identity_holder_owner_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_owner_id_idx" ON "identity_holders" ("owner_id");
-- Create index "impersonation_event_organization_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_organization_id_idx" ON "impersonation_events" ("organization_id");
-- Create index "impersonation_event_target_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_target_user_id_idx" ON "impersonation_events" ("target_user_id");
-- Create index "impersonation_event_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_user_id_idx" ON "impersonation_events" ("user_id");
-- Create index "integration_action_plans_action_plan_id_idx" to table: "integration_action_plans"
CREATE INDEX "integration_action_plans_action_plan_id_idx" ON "integration_action_plans" ("action_plan_id");
-- Create index "integration_events_event_id_idx" to table: "integration_events"
CREATE INDEX "integration_events_event_id_idx" ON "integration_events" ("event_id");
-- Create index "integration_findings_finding_id_idx" to table: "integration_findings"
CREATE INDEX "integration_findings_finding_id_idx" ON "integration_findings" ("finding_id");
-- Create index "integration_internal_policies_internal_policy_id_idx" to table: "integration_internal_policies"
CREATE INDEX "integration_internal_policies_internal_policy_id_idx" ON "integration_internal_policies" ("internal_policy_id");
-- Create index "integration_remediations_remediation_id_idx" to table: "integration_remediations"
CREATE INDEX "integration_remediations_remediation_id_idx" ON "integration_remediations" ("remediation_id");
-- Create index "integration_reviews_review_id_idx" to table: "integration_reviews"
CREATE INDEX "integration_reviews_review_id_idx" ON "integration_reviews" ("review_id");
-- Drop index "integrationrun_owner_id" from table: "integration_runs"
DROP INDEX "integrationrun_owner_id";
-- Create index "integration_run_event_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_event_id_idx" ON "integration_runs" ("event_id");
-- Create index "integration_run_owner_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_owner_id_idx" ON "integration_runs" ("owner_id");
-- Create index "integration_run_request_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_request_file_id_idx" ON "integration_runs" ("request_file_id");
-- Create index "integration_run_response_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_response_file_id_idx" ON "integration_runs" ("response_file_id");
-- Create index "integration_secrets_hush_id_idx" to table: "integration_secrets"
CREATE INDEX "integration_secrets_hush_id_idx" ON "integration_secrets" ("hush_id");
-- Create index "integration_vulnerabilities_vulnerability_id_idx" to table: "integration_vulnerabilities"
CREATE INDEX "integration_vulnerabilities_vulnerability_id_idx" ON "integration_vulnerabilities" ("vulnerability_id");
-- Drop index "integrationwebhook_owner_id" from table: "integration_webhooks"
DROP INDEX "integrationwebhook_owner_id";
-- Create index "integration_webhook_owner_id_idx" to table: "integration_webhooks"
CREATE INDEX "integration_webhook_owner_id_idx" ON "integration_webhooks" ("owner_id");
-- Drop index "integration_owner_id" from table: "integrations"
DROP INDEX "integration_owner_id";
-- Create index "integration_owner_id_idx" to table: "integrations"
CREATE INDEX "integration_owner_id_idx" ON "integrations" ("owner_id");
-- Create index "integration_platform_id_idx" to table: "integrations"
CREATE INDEX "integration_platform_id_idx" ON "integrations" ("platform_id");
-- Drop index "internalpolicy_owner_id" from table: "internal_policies"
DROP INDEX "internalpolicy_owner_id";
-- Create index "internal_policy_file_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_file_id_idx" ON "internal_policies" ("file_id");
-- Create index "internal_policy_owner_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_owner_id_idx" ON "internal_policies" ("owner_id");
-- Create index "internal_policy_assets_asset_id_idx" to table: "internal_policy_assets"
CREATE INDEX "internal_policy_assets_asset_id_idx" ON "internal_policy_assets" ("asset_id");
-- Create index "internal_policy_blocked_groups_group_id_idx" to table: "internal_policy_blocked_groups"
CREATE INDEX "internal_policy_blocked_groups_group_id_idx" ON "internal_policy_blocked_groups" ("group_id");
-- Create index "internal_policy_control_objectives_control_objective_id_idx" to table: "internal_policy_control_objectives"
CREATE INDEX "internal_policy_control_objectives_control_objective_id_idx" ON "internal_policy_control_objectives" ("control_objective_id");
-- Create index "internal_policy_controls_control_id_idx" to table: "internal_policy_controls"
CREATE INDEX "internal_policy_controls_control_id_idx" ON "internal_policy_controls" ("control_id");
-- Create index "internal_policy_editors_group_id_idx" to table: "internal_policy_editors"
CREATE INDEX "internal_policy_editors_group_id_idx" ON "internal_policy_editors" ("group_id");
-- Create index "internal_policy_entities_entity_id_idx" to table: "internal_policy_entities"
CREATE INDEX "internal_policy_entities_entity_id_idx" ON "internal_policy_entities" ("entity_id");
-- Create index "internal_policy_identity_holders_identity_holder_id_idx" to table: "internal_policy_identity_holders"
CREATE INDEX "internal_policy_identity_holders_identity_holder_id_idx" ON "internal_policy_identity_holders" ("identity_holder_id");
-- Create index "internal_policy_narratives_narrative_id_idx" to table: "internal_policy_narratives"
CREATE INDEX "internal_policy_narratives_narrative_id_idx" ON "internal_policy_narratives" ("narrative_id");
-- Create index "internal_policy_procedures_procedure_id_idx" to table: "internal_policy_procedures"
CREATE INDEX "internal_policy_procedures_procedure_id_idx" ON "internal_policy_procedures" ("procedure_id");
-- Create index "internal_policy_risks_risk_id_idx" to table: "internal_policy_risks"
CREATE INDEX "internal_policy_risks_risk_id_idx" ON "internal_policy_risks" ("risk_id");
-- Create index "internal_policy_subcontrols_subcontrol_id_idx" to table: "internal_policy_subcontrols"
CREATE INDEX "internal_policy_subcontrols_subcontrol_id_idx" ON "internal_policy_subcontrols" ("subcontrol_id");
-- Create index "internal_policy_tasks_task_id_idx" to table: "internal_policy_tasks"
CREATE INDEX "internal_policy_tasks_task_id_idx" ON "internal_policy_tasks" ("task_id");
-- Create index "invite_events_event_id_idx" to table: "invite_events"
CREATE INDEX "invite_events_event_id_idx" ON "invite_events" ("event_id");
-- Create index "invite_groups_group_id_idx" to table: "invite_groups"
CREATE INDEX "invite_groups_group_id_idx" ON "invite_groups" ("group_id");
-- Drop index "invite_owner_id" from table: "invites"
DROP INDEX "invite_owner_id";
-- Create index "invite_owner_id_idx" to table: "invites"
CREATE INDEX "invite_owner_id_idx" ON "invites" ("owner_id");
-- Drop index "jobresult_owner_id" from table: "job_results"
DROP INDEX "jobresult_owner_id";
-- Create index "job_result_file_id_idx" to table: "job_results"
CREATE INDEX "job_result_file_id_idx" ON "job_results" ("file_id");
-- Create index "job_result_owner_id_idx" to table: "job_results"
CREATE INDEX "job_result_owner_id_idx" ON "job_results" ("owner_id");
-- Create index "job_result_scheduled_job_id_idx" to table: "job_results"
CREATE INDEX "job_result_scheduled_job_id_idx" ON "job_results" ("scheduled_job_id");
-- Create index "job_runner_job_runner_tokens_job_runner_token_id_idx" to table: "job_runner_job_runner_tokens"
CREATE INDEX "job_runner_job_runner_tokens_job_runner_token_id_idx" ON "job_runner_job_runner_tokens" ("job_runner_token_id");
-- Drop index "jobrunnerregistrationtoken_owner_id" from table: "job_runner_registration_tokens"
DROP INDEX "jobrunnerregistrationtoken_owner_id";
-- Create index "job_runner_registration_token_job_runner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_job_runner_id_idx" ON "job_runner_registration_tokens" ("job_runner_id");
-- Create index "job_runner_registration_token_owner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_owner_id_idx" ON "job_runner_registration_tokens" ("owner_id");
-- Drop index "jobrunnertoken_owner_id" from table: "job_runner_tokens"
DROP INDEX "jobrunnertoken_owner_id";
-- Create index "job_runner_token_owner_id_idx" to table: "job_runner_tokens"
CREATE INDEX "job_runner_token_owner_id_idx" ON "job_runner_tokens" ("owner_id");
-- Drop index "jobrunner_owner_id" from table: "job_runners"
DROP INDEX "jobrunner_owner_id";
-- Create index "job_runner_owner_id_idx" to table: "job_runners"
CREATE INDEX "job_runner_owner_id_idx" ON "job_runners" ("owner_id");
-- Drop index "jobtemplate_owner_id" from table: "job_templates"
DROP INDEX "jobtemplate_owner_id";
-- Create index "job_template_owner_id_idx" to table: "job_templates"
CREATE INDEX "job_template_owner_id_idx" ON "job_templates" ("owner_id");
-- Create index "mapped_control_blocked_groups_group_id_idx" to table: "mapped_control_blocked_groups"
CREATE INDEX "mapped_control_blocked_groups_group_id_idx" ON "mapped_control_blocked_groups" ("group_id");
-- Create index "mapped_control_editors_group_id_idx" to table: "mapped_control_editors"
CREATE INDEX "mapped_control_editors_group_id_idx" ON "mapped_control_editors" ("group_id");
-- Create index "mapped_control_from_controls_control_id_idx" to table: "mapped_control_from_controls"
CREATE INDEX "mapped_control_from_controls_control_id_idx" ON "mapped_control_from_controls" ("control_id");
-- Create index "mapped_control_from_subcontrols_subcontrol_id_idx" to table: "mapped_control_from_subcontrols"
CREATE INDEX "mapped_control_from_subcontrols_subcontrol_id_idx" ON "mapped_control_from_subcontrols" ("subcontrol_id");
-- Create index "mapped_control_to_controls_control_id_idx" to table: "mapped_control_to_controls"
CREATE INDEX "mapped_control_to_controls_control_id_idx" ON "mapped_control_to_controls" ("control_id");
-- Create index "mapped_control_to_subcontrols_subcontrol_id_idx" to table: "mapped_control_to_subcontrols"
CREATE INDEX "mapped_control_to_subcontrols_subcontrol_id_idx" ON "mapped_control_to_subcontrols" ("subcontrol_id");
-- Drop index "mappedcontrol_owner_id" from table: "mapped_controls"
DROP INDEX "mappedcontrol_owner_id";
-- Create index "mapped_control_owner_id_idx" to table: "mapped_controls"
CREATE INDEX "mapped_control_owner_id_idx" ON "mapped_controls" ("owner_id");
-- Create index "narrative_blocked_groups_group_id_idx" to table: "narrative_blocked_groups"
CREATE INDEX "narrative_blocked_groups_group_id_idx" ON "narrative_blocked_groups" ("group_id");
-- Create index "narrative_editors_group_id_idx" to table: "narrative_editors"
CREATE INDEX "narrative_editors_group_id_idx" ON "narrative_editors" ("group_id");
-- Create index "narrative_viewers_group_id_idx" to table: "narrative_viewers"
CREATE INDEX "narrative_viewers_group_id_idx" ON "narrative_viewers" ("group_id");
-- Drop index "narrative_owner_id" from table: "narratives"
DROP INDEX "narrative_owner_id";
-- Create index "narrative_owner_id_idx" to table: "narratives"
CREATE INDEX "narrative_owner_id_idx" ON "narratives" ("owner_id");
-- Drop index "note_owner_id" from table: "notes"
DROP INDEX "note_owner_id";
-- Create index "note_discussion_id_idx" to table: "notes"
CREATE INDEX "note_discussion_id_idx" ON "notes" ("discussion_id");
-- Create index "note_owner_id_idx" to table: "notes"
CREATE INDEX "note_owner_id_idx" ON "notes" ("owner_id");
-- Create index "note_trust_center_id_idx" to table: "notes"
CREATE INDEX "note_trust_center_id_idx" ON "notes" ("trust_center_id");
-- Drop index "notificationpreference_owner_id" from table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id";
-- Create index "notification_preference_owner_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_owner_id_idx" ON "notification_preferences" ("owner_id");
-- Create index "notification_preference_template_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_template_id_idx" ON "notification_preferences" ("template_id");
-- Create index "notification_preference_user_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_user_id_idx" ON "notification_preferences" ("user_id");
-- Drop index "notificationtemplate_owner_id" from table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id";
-- Create index "notification_template_email_template_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_email_template_id_idx" ON "notification_templates" ("email_template_id");
-- Create index "notification_template_integration_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_integration_id_idx" ON "notification_templates" ("integration_id");
-- Create index "notification_template_owner_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_owner_id_idx" ON "notification_templates" ("owner_id");
-- Create index "notification_template_workflow_definition_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_workflow_definition_id_idx" ON "notification_templates" ("workflow_definition_id");
-- Create index "notification_owner_id_idx" to table: "notifications"
CREATE INDEX "notification_owner_id_idx" ON "notifications" ("owner_id");
-- Create index "notification_template_id_idx" to table: "notifications"
CREATE INDEX "notification_template_id_idx" ON "notifications" ("template_id");
-- Create index "org_membership_events_event_id_idx" to table: "org_membership_events"
CREATE INDEX "org_membership_events_event_id_idx" ON "org_membership_events" ("event_id");
-- Create index "org_membership_organization_id_idx" to table: "org_memberships"
CREATE INDEX "org_membership_organization_id_idx" ON "org_memberships" ("organization_id");
-- Create index "org_module_org_prices_org_price_id_idx" to table: "org_module_org_prices"
CREATE INDEX "org_module_org_prices_org_price_id_idx" ON "org_module_org_prices" ("org_price_id");
-- Drop index "orgmodule_owner_id" from table: "org_modules"
DROP INDEX "orgmodule_owner_id";
-- Create index "org_module_owner_id_idx" to table: "org_modules"
CREATE INDEX "org_module_owner_id_idx" ON "org_modules" ("owner_id");
-- Create index "org_module_subscription_id_idx" to table: "org_modules"
CREATE INDEX "org_module_subscription_id_idx" ON "org_modules" ("subscription_id");
-- Drop index "orgprice_owner_id" from table: "org_prices"
DROP INDEX "orgprice_owner_id";
-- Create index "org_price_owner_id_idx" to table: "org_prices"
CREATE INDEX "org_price_owner_id_idx" ON "org_prices" ("owner_id");
-- Create index "org_price_subscription_id_idx" to table: "org_prices"
CREATE INDEX "org_price_subscription_id_idx" ON "org_prices" ("subscription_id");
-- Create index "org_product_org_prices_org_price_id_idx" to table: "org_product_org_prices"
CREATE INDEX "org_product_org_prices_org_price_id_idx" ON "org_product_org_prices" ("org_price_id");
-- Drop index "orgproduct_owner_id" from table: "org_products"
DROP INDEX "orgproduct_owner_id";
-- Create index "org_product_owner_id_idx" to table: "org_products"
CREATE INDEX "org_product_owner_id_idx" ON "org_products" ("owner_id");
-- Create index "org_product_subscription_id_idx" to table: "org_products"
CREATE INDEX "org_product_subscription_id_idx" ON "org_products" ("subscription_id");
-- Create index "org_subscription_events_event_id_idx" to table: "org_subscription_events"
CREATE INDEX "org_subscription_events_event_id_idx" ON "org_subscription_events" ("event_id");
-- Drop index "orgsubscription_owner_id" from table: "org_subscriptions"
DROP INDEX "orgsubscription_owner_id";
-- Create index "org_subscription_owner_id_idx" to table: "org_subscriptions"
CREATE INDEX "org_subscription_owner_id_idx" ON "org_subscriptions" ("owner_id");
-- Create index "organization_events_event_id_idx" to table: "organization_events"
CREATE INDEX "organization_events_event_id_idx" ON "organization_events" ("event_id");
-- Create index "organization_files_file_id_idx" to table: "organization_files"
CREATE INDEX "organization_files_file_id_idx" ON "organization_files" ("file_id");
-- Create index "organization_personal_access_tokens_personal_access_token_id_id" to table: "organization_personal_access_tokens"
CREATE INDEX "organization_personal_access_tokens_personal_access_token_id_id" ON "organization_personal_access_tokens" ("personal_access_token_id");
-- Create index "organization_setting_files_file_id_idx" to table: "organization_setting_files"
CREATE INDEX "organization_setting_files_file_id_idx" ON "organization_setting_files" ("file_id");
-- Create index "organization_setting_organization_id_idx" to table: "organization_settings"
CREATE INDEX "organization_setting_organization_id_idx" ON "organization_settings" ("organization_id");
-- Create index "organization_avatar_local_file_id_idx" to table: "organizations"
CREATE INDEX "organization_avatar_local_file_id_idx" ON "organizations" ("avatar_local_file_id");
-- Create index "organization_parent_organization_id_idx" to table: "organizations"
CREATE INDEX "organization_parent_organization_id_idx" ON "organizations" ("parent_organization_id");
-- Create index "password_reset_tokens_owner_id_fk" to table: "password_reset_tokens"
CREATE INDEX "password_reset_tokens_owner_id_fk" ON "password_reset_tokens" ("owner_id");
-- Create index "personal_access_token_events_event_id_idx" to table: "personal_access_token_events"
CREATE INDEX "personal_access_token_events_event_id_idx" ON "personal_access_token_events" ("event_id");
-- Create index "personal_access_tokens_owner_id_fk" to table: "personal_access_tokens"
CREATE INDEX "personal_access_tokens_owner_id_fk" ON "personal_access_tokens" ("owner_id");
-- Create index "platform_applicable_frameworks_standard_id_idx" to table: "platform_applicable_frameworks"
CREATE INDEX "platform_applicable_frameworks_standard_id_idx" ON "platform_applicable_frameworks" ("standard_id");
-- Create index "platform_assessments_assessment_id_idx" to table: "platform_assessments"
CREATE INDEX "platform_assessments_assessment_id_idx" ON "platform_assessments" ("assessment_id");
-- Create index "platform_assets_asset_id_idx" to table: "platform_assets"
CREATE INDEX "platform_assets_asset_id_idx" ON "platform_assets" ("asset_id");
-- Create index "platform_blocked_groups_group_id_idx" to table: "platform_blocked_groups"
CREATE INDEX "platform_blocked_groups_group_id_idx" ON "platform_blocked_groups" ("group_id");
-- Create index "platform_controls_control_id_idx" to table: "platform_controls"
CREATE INDEX "platform_controls_control_id_idx" ON "platform_controls" ("control_id");
-- Create index "platform_editors_group_id_idx" to table: "platform_editors"
CREATE INDEX "platform_editors_group_id_idx" ON "platform_editors" ("group_id");
-- Create index "platform_entities_entity_id_idx" to table: "platform_entities"
CREATE INDEX "platform_entities_entity_id_idx" ON "platform_entities" ("entity_id");
-- Create index "platform_evidence_evidence_id_idx" to table: "platform_evidence"
CREATE INDEX "platform_evidence_evidence_id_idx" ON "platform_evidence" ("evidence_id");
-- Create index "platform_files_file_id_idx" to table: "platform_files"
CREATE INDEX "platform_files_file_id_idx" ON "platform_files" ("file_id");
-- Create index "platform_identity_holders_identity_holder_id_idx" to table: "platform_identity_holders"
CREATE INDEX "platform_identity_holders_identity_holder_id_idx" ON "platform_identity_holders" ("identity_holder_id");
-- Create index "platform_out_of_scope_assets_asset_id_idx" to table: "platform_out_of_scope_assets"
CREATE INDEX "platform_out_of_scope_assets_asset_id_idx" ON "platform_out_of_scope_assets" ("asset_id");
-- Create index "platform_out_of_scope_vendors_entity_id_idx" to table: "platform_out_of_scope_vendors"
CREATE INDEX "platform_out_of_scope_vendors_entity_id_idx" ON "platform_out_of_scope_vendors" ("entity_id");
-- Create index "platform_risks_risk_id_idx" to table: "platform_risks"
CREATE INDEX "platform_risks_risk_id_idx" ON "platform_risks" ("risk_id");
-- Create index "platform_scans_scan_id_idx" to table: "platform_scans"
CREATE INDEX "platform_scans_scan_id_idx" ON "platform_scans" ("scan_id");
-- Create index "platform_source_entities_entity_id_idx" to table: "platform_source_entities"
CREATE INDEX "platform_source_entities_entity_id_idx" ON "platform_source_entities" ("entity_id");
-- Create index "platform_system_details_system_detail_id_idx" to table: "platform_system_details"
CREATE INDEX "platform_system_details_system_detail_id_idx" ON "platform_system_details" ("system_detail_id");
-- Create index "platform_tasks_task_id_idx" to table: "platform_tasks"
CREATE INDEX "platform_tasks_task_id_idx" ON "platform_tasks" ("task_id");
-- Create index "platform_viewers_group_id_idx" to table: "platform_viewers"
CREATE INDEX "platform_viewers_group_id_idx" ON "platform_viewers" ("group_id");
-- Drop index "platform_owner_id" from table: "platforms"
DROP INDEX "platform_owner_id";
-- Create index "platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_owner_id_idx" ON "platforms" ("owner_id");
-- Create index "platform_platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_platform_owner_id_idx" ON "platforms" ("platform_owner_id");
-- Create index "procedure_blocked_groups_group_id_idx" to table: "procedure_blocked_groups"
CREATE INDEX "procedure_blocked_groups_group_id_idx" ON "procedure_blocked_groups" ("group_id");
-- Create index "procedure_editors_group_id_idx" to table: "procedure_editors"
CREATE INDEX "procedure_editors_group_id_idx" ON "procedure_editors" ("group_id");
-- Create index "procedure_narratives_narrative_id_idx" to table: "procedure_narratives"
CREATE INDEX "procedure_narratives_narrative_id_idx" ON "procedure_narratives" ("narrative_id");
-- Create index "procedure_risks_risk_id_idx" to table: "procedure_risks"
CREATE INDEX "procedure_risks_risk_id_idx" ON "procedure_risks" ("risk_id");
-- Create index "procedure_tasks_task_id_idx" to table: "procedure_tasks"
CREATE INDEX "procedure_tasks_task_id_idx" ON "procedure_tasks" ("task_id");
-- Drop index "procedure_owner_id" from table: "procedures"
DROP INDEX "procedure_owner_id";
-- Create index "procedure_file_id_idx" to table: "procedures"
CREATE INDEX "procedure_file_id_idx" ON "procedures" ("file_id");
-- Create index "procedure_owner_id_idx" to table: "procedures"
CREATE INDEX "procedure_owner_id_idx" ON "procedures" ("owner_id");
-- Create index "program_action_plans_action_plan_id_idx" to table: "program_action_plans"
CREATE INDEX "program_action_plans_action_plan_id_idx" ON "program_action_plans" ("action_plan_id");
-- Create index "program_blocked_groups_group_id_idx" to table: "program_blocked_groups"
CREATE INDEX "program_blocked_groups_group_id_idx" ON "program_blocked_groups" ("group_id");
-- Create index "program_control_objectives_control_objective_id_idx" to table: "program_control_objectives"
CREATE INDEX "program_control_objectives_control_objective_id_idx" ON "program_control_objectives" ("control_objective_id");
-- Create index "program_controls_control_id_idx" to table: "program_controls"
CREATE INDEX "program_controls_control_id_idx" ON "program_controls" ("control_id");
-- Create index "program_editors_group_id_idx" to table: "program_editors"
CREATE INDEX "program_editors_group_id_idx" ON "program_editors" ("group_id");
-- Create index "program_evidence_evidence_id_idx" to table: "program_evidence"
CREATE INDEX "program_evidence_evidence_id_idx" ON "program_evidence" ("evidence_id");
-- Create index "program_files_file_id_idx" to table: "program_files"
CREATE INDEX "program_files_file_id_idx" ON "program_files" ("file_id");
-- Create index "program_internal_policies_internal_policy_id_idx" to table: "program_internal_policies"
CREATE INDEX "program_internal_policies_internal_policy_id_idx" ON "program_internal_policies" ("internal_policy_id");
-- Create index "program_membership_program_id_idx" to table: "program_memberships"
CREATE INDEX "program_membership_program_id_idx" ON "program_memberships" ("program_id");
-- Create index "program_narratives_narrative_id_idx" to table: "program_narratives"
CREATE INDEX "program_narratives_narrative_id_idx" ON "program_narratives" ("narrative_id");
-- Create index "program_procedures_procedure_id_idx" to table: "program_procedures"
CREATE INDEX "program_procedures_procedure_id_idx" ON "program_procedures" ("procedure_id");
-- Create index "program_risks_risk_id_idx" to table: "program_risks"
CREATE INDEX "program_risks_risk_id_idx" ON "program_risks" ("risk_id");
-- Create index "program_system_details_system_detail_id_idx" to table: "program_system_details"
CREATE INDEX "program_system_details_system_detail_id_idx" ON "program_system_details" ("system_detail_id");
-- Create index "program_tasks_task_id_idx" to table: "program_tasks"
CREATE INDEX "program_tasks_task_id_idx" ON "program_tasks" ("task_id");
-- Create index "program_viewers_group_id_idx" to table: "program_viewers"
CREATE INDEX "program_viewers_group_id_idx" ON "program_viewers" ("group_id");
-- Drop index "program_owner_id" from table: "programs"
DROP INDEX "program_owner_id";
-- Create index "program_owner_id_idx" to table: "programs"
CREATE INDEX "program_owner_id_idx" ON "programs" ("owner_id");
-- Create index "program_program_owner_id_idx" to table: "programs"
CREATE INDEX "program_program_owner_id_idx" ON "programs" ("program_owner_id");
-- Create index "remediation_action_plans_action_plan_id_idx" to table: "remediation_action_plans"
CREATE INDEX "remediation_action_plans_action_plan_id_idx" ON "remediation_action_plans" ("action_plan_id");
-- Create index "remediation_assets_asset_id_idx" to table: "remediation_assets"
CREATE INDEX "remediation_assets_asset_id_idx" ON "remediation_assets" ("asset_id");
-- Create index "remediation_blocked_groups_group_id_idx" to table: "remediation_blocked_groups"
CREATE INDEX "remediation_blocked_groups_group_id_idx" ON "remediation_blocked_groups" ("group_id");
-- Create index "remediation_controls_control_id_idx" to table: "remediation_controls"
CREATE INDEX "remediation_controls_control_id_idx" ON "remediation_controls" ("control_id");
-- Create index "remediation_editors_group_id_idx" to table: "remediation_editors"
CREATE INDEX "remediation_editors_group_id_idx" ON "remediation_editors" ("group_id");
-- Create index "remediation_entities_entity_id_idx" to table: "remediation_entities"
CREATE INDEX "remediation_entities_entity_id_idx" ON "remediation_entities" ("entity_id");
-- Create index "remediation_findings_finding_id_idx" to table: "remediation_findings"
CREATE INDEX "remediation_findings_finding_id_idx" ON "remediation_findings" ("finding_id");
-- Create index "remediation_programs_program_id_idx" to table: "remediation_programs"
CREATE INDEX "remediation_programs_program_id_idx" ON "remediation_programs" ("program_id");
-- Create index "remediation_risks_risk_id_idx" to table: "remediation_risks"
CREATE INDEX "remediation_risks_risk_id_idx" ON "remediation_risks" ("risk_id");
-- Create index "remediation_subcontrols_subcontrol_id_idx" to table: "remediation_subcontrols"
CREATE INDEX "remediation_subcontrols_subcontrol_id_idx" ON "remediation_subcontrols" ("subcontrol_id");
-- Create index "remediation_vulnerabilities_vulnerability_id_idx" to table: "remediation_vulnerabilities"
CREATE INDEX "remediation_vulnerabilities_vulnerability_id_idx" ON "remediation_vulnerabilities" ("vulnerability_id");
-- Drop index "remediation_owner_id" from table: "remediations"
DROP INDEX "remediation_owner_id";
-- Create index "remediation_owner_id_idx" to table: "remediations"
CREATE INDEX "remediation_owner_id_idx" ON "remediations" ("owner_id");
-- Create index "review_action_plans_action_plan_id_idx" to table: "review_action_plans"
CREATE INDEX "review_action_plans_action_plan_id_idx" ON "review_action_plans" ("action_plan_id");
-- Create index "review_assets_asset_id_idx" to table: "review_assets"
CREATE INDEX "review_assets_asset_id_idx" ON "review_assets" ("asset_id");
-- Create index "review_blocked_groups_group_id_idx" to table: "review_blocked_groups"
CREATE INDEX "review_blocked_groups_group_id_idx" ON "review_blocked_groups" ("group_id");
-- Create index "review_controls_control_id_idx" to table: "review_controls"
CREATE INDEX "review_controls_control_id_idx" ON "review_controls" ("control_id");
-- Create index "review_editors_group_id_idx" to table: "review_editors"
CREATE INDEX "review_editors_group_id_idx" ON "review_editors" ("group_id");
-- Create index "review_entities_entity_id_idx" to table: "review_entities"
CREATE INDEX "review_entities_entity_id_idx" ON "review_entities" ("entity_id");
-- Create index "review_findings_finding_id_idx" to table: "review_findings"
CREATE INDEX "review_findings_finding_id_idx" ON "review_findings" ("finding_id");
-- Create index "review_internal_policies_internal_policy_id_idx" to table: "review_internal_policies"
CREATE INDEX "review_internal_policies_internal_policy_id_idx" ON "review_internal_policies" ("internal_policy_id");
-- Create index "review_programs_program_id_idx" to table: "review_programs"
CREATE INDEX "review_programs_program_id_idx" ON "review_programs" ("program_id");
-- Create index "review_remediations_remediation_id_idx" to table: "review_remediations"
CREATE INDEX "review_remediations_remediation_id_idx" ON "review_remediations" ("remediation_id");
-- Create index "review_risks_risk_id_idx" to table: "review_risks"
CREATE INDEX "review_risks_risk_id_idx" ON "review_risks" ("risk_id");
-- Create index "review_subcontrols_subcontrol_id_idx" to table: "review_subcontrols"
CREATE INDEX "review_subcontrols_subcontrol_id_idx" ON "review_subcontrols" ("subcontrol_id");
-- Create index "review_vulnerabilities_vulnerability_id_idx" to table: "review_vulnerabilities"
CREATE INDEX "review_vulnerabilities_vulnerability_id_idx" ON "review_vulnerabilities" ("vulnerability_id");
-- Drop index "review_owner_id" from table: "reviews"
DROP INDEX "review_owner_id";
-- Create index "review_owner_id_idx" to table: "reviews"
CREATE INDEX "review_owner_id_idx" ON "reviews" ("owner_id");
-- Create index "review_reviewer_id_idx" to table: "reviews"
CREATE INDEX "review_reviewer_id_idx" ON "reviews" ("reviewer_id");
-- Create index "risk_action_plans_action_plan_id_idx" to table: "risk_action_plans"
CREATE INDEX "risk_action_plans_action_plan_id_idx" ON "risk_action_plans" ("action_plan_id");
-- Create index "risk_blocked_groups_group_id_idx" to table: "risk_blocked_groups"
CREATE INDEX "risk_blocked_groups_group_id_idx" ON "risk_blocked_groups" ("group_id");
-- Create index "risk_editors_group_id_idx" to table: "risk_editors"
CREATE INDEX "risk_editors_group_id_idx" ON "risk_editors" ("group_id");
-- Create index "risk_tasks_task_id_idx" to table: "risk_tasks"
CREATE INDEX "risk_tasks_task_id_idx" ON "risk_tasks" ("task_id");
-- Create index "risk_viewers_group_id_idx" to table: "risk_viewers"
CREATE INDEX "risk_viewers_group_id_idx" ON "risk_viewers" ("group_id");
-- Drop index "risk_owner_id" from table: "risks"
DROP INDEX "risk_owner_id";
-- Create index "risk_delegate_id_idx" to table: "risks"
CREATE INDEX "risk_delegate_id_idx" ON "risks" ("delegate_id");
-- Create index "risk_owner_id_idx" to table: "risks"
CREATE INDEX "risk_owner_id_idx" ON "risks" ("owner_id");
-- Create index "risk_stakeholder_id_idx" to table: "risks"
CREATE INDEX "risk_stakeholder_id_idx" ON "risks" ("stakeholder_id");
-- Create index "scan_action_plans_action_plan_id_idx" to table: "scan_action_plans"
CREATE INDEX "scan_action_plans_action_plan_id_idx" ON "scan_action_plans" ("action_plan_id");
-- Create index "scan_assets_asset_id_idx" to table: "scan_assets"
CREATE INDEX "scan_assets_asset_id_idx" ON "scan_assets" ("asset_id");
-- Create index "scan_blocked_groups_group_id_idx" to table: "scan_blocked_groups"
CREATE INDEX "scan_blocked_groups_group_id_idx" ON "scan_blocked_groups" ("group_id");
-- Create index "scan_editors_group_id_idx" to table: "scan_editors"
CREATE INDEX "scan_editors_group_id_idx" ON "scan_editors" ("group_id");
-- Create index "scan_entities_entity_id_idx" to table: "scan_entities"
CREATE INDEX "scan_entities_entity_id_idx" ON "scan_entities" ("entity_id");
-- Create index "scan_evidence_evidence_id_idx" to table: "scan_evidence"
CREATE INDEX "scan_evidence_evidence_id_idx" ON "scan_evidence" ("evidence_id");
-- Create index "scan_files_file_id_idx" to table: "scan_files"
CREATE INDEX "scan_files_file_id_idx" ON "scan_files" ("file_id");
-- Create index "scan_remediations_remediation_id_idx" to table: "scan_remediations"
CREATE INDEX "scan_remediations_remediation_id_idx" ON "scan_remediations" ("remediation_id");
-- Create index "scan_tasks_task_id_idx" to table: "scan_tasks"
CREATE INDEX "scan_tasks_task_id_idx" ON "scan_tasks" ("task_id");
-- Drop index "scan_owner_id" from table: "scans"
DROP INDEX "scan_owner_id";
-- Create index "scan_generated_by_platform_id_idx" to table: "scans"
CREATE INDEX "scan_generated_by_platform_id_idx" ON "scans" ("generated_by_platform_id");
-- Create index "scan_owner_id_idx" to table: "scans"
CREATE INDEX "scan_owner_id_idx" ON "scans" ("owner_id");
-- Create index "scan_performed_by_group_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_group_id_idx" ON "scans" ("performed_by_group_id");
-- Create index "scan_performed_by_user_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_user_id_idx" ON "scans" ("performed_by_user_id");
-- Create index "scheduled_job_controls_control_id_idx" to table: "scheduled_job_controls"
CREATE INDEX "scheduled_job_controls_control_id_idx" ON "scheduled_job_controls" ("control_id");
-- Drop index "scheduledjobrun_owner_id" from table: "scheduled_job_runs"
DROP INDEX "scheduledjobrun_owner_id";
-- Create index "scheduled_job_run_job_runner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_job_runner_id_idx" ON "scheduled_job_runs" ("job_runner_id");
-- Create index "scheduled_job_run_owner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_owner_id_idx" ON "scheduled_job_runs" ("owner_id");
-- Create index "scheduled_job_run_scheduled_job_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_scheduled_job_id_idx" ON "scheduled_job_runs" ("scheduled_job_id");
-- Create index "scheduled_job_subcontrols_subcontrol_id_idx" to table: "scheduled_job_subcontrols"
CREATE INDEX "scheduled_job_subcontrols_subcontrol_id_idx" ON "scheduled_job_subcontrols" ("subcontrol_id");
-- Drop index "scheduledjob_owner_id" from table: "scheduled_jobs"
DROP INDEX "scheduledjob_owner_id";
-- Create index "scheduled_job_job_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_id_idx" ON "scheduled_jobs" ("job_id");
-- Create index "scheduled_job_job_runner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_runner_id_idx" ON "scheduled_jobs" ("job_runner_id");
-- Create index "scheduled_job_owner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_owner_id_idx" ON "scheduled_jobs" ("owner_id");
-- Drop index "sladefinition_owner_id" from table: "sla_definitions"
DROP INDEX "sladefinition_owner_id";
-- Create index "sla_definition_owner_id_idx" to table: "sla_definitions"
CREATE INDEX "sla_definition_owner_id_idx" ON "sla_definitions" ("owner_id");
-- Drop index "standard_owner_id" from table: "standards"
DROP INDEX "standard_owner_id";
-- Create index "standard_logo_file_id_idx" to table: "standards"
CREATE INDEX "standard_logo_file_id_idx" ON "standards" ("logo_file_id");
-- Create index "standard_owner_id_idx" to table: "standards"
CREATE INDEX "standard_owner_id_idx" ON "standards" ("owner_id");
-- Create index "subcontrol_assets_asset_id_idx" to table: "subcontrol_assets"
CREATE INDEX "subcontrol_assets_asset_id_idx" ON "subcontrol_assets" ("asset_id");
-- Create index "subcontrol_control_implementations_control_implementation_id_id" to table: "subcontrol_control_implementations"
CREATE INDEX "subcontrol_control_implementations_control_implementation_id_id" ON "subcontrol_control_implementations" ("control_implementation_id");
-- Create index "subcontrol_control_objectives_control_objective_id_idx" to table: "subcontrol_control_objectives"
CREATE INDEX "subcontrol_control_objectives_control_objective_id_idx" ON "subcontrol_control_objectives" ("control_objective_id");
-- Create index "subcontrol_entities_entity_id_idx" to table: "subcontrol_entities"
CREATE INDEX "subcontrol_entities_entity_id_idx" ON "subcontrol_entities" ("entity_id");
-- Create index "subcontrol_identity_holders_identity_holder_id_idx" to table: "subcontrol_identity_holders"
CREATE INDEX "subcontrol_identity_holders_identity_holder_id_idx" ON "subcontrol_identity_holders" ("identity_holder_id");
-- Create index "subcontrol_procedures_procedure_id_idx" to table: "subcontrol_procedures"
CREATE INDEX "subcontrol_procedures_procedure_id_idx" ON "subcontrol_procedures" ("procedure_id");
-- Create index "subcontrol_risks_risk_id_idx" to table: "subcontrol_risks"
CREATE INDEX "subcontrol_risks_risk_id_idx" ON "subcontrol_risks" ("risk_id");
-- Create index "subcontrol_scans_scan_id_idx" to table: "subcontrol_scans"
CREATE INDEX "subcontrol_scans_scan_id_idx" ON "subcontrol_scans" ("scan_id");
-- Create index "subcontrol_tasks_task_id_idx" to table: "subcontrol_tasks"
CREATE INDEX "subcontrol_tasks_task_id_idx" ON "subcontrol_tasks" ("task_id");
-- Drop index "subcontrol_owner_id" from table: "subcontrols"
DROP INDEX "subcontrol_owner_id";
-- Create index "subcontrol_owner_id_idx" to table: "subcontrols"
CREATE INDEX "subcontrol_owner_id_idx" ON "subcontrols" ("owner_id");
-- Drop index "subprocessor_owner_id" from table: "subprocessors"
DROP INDEX "subprocessor_owner_id";
-- Create index "subprocessor_logo_file_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_logo_file_id_idx" ON "subprocessors" ("logo_file_id");
-- Create index "subprocessor_owner_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_owner_id_idx" ON "subprocessors" ("owner_id");
-- Create index "subscriber_events_event_id_idx" to table: "subscriber_events"
CREATE INDEX "subscriber_events_event_id_idx" ON "subscriber_events" ("event_id");
-- Drop index "subscriber_owner_id" from table: "subscribers"
DROP INDEX "subscriber_owner_id";
-- Create index "subscriber_contact_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_contact_id_idx" ON "subscribers" ("contact_id");
-- Create index "subscriber_owner_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_owner_id_idx" ON "subscribers" ("owner_id");
-- Create index "subscriber_trust_center_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_trust_center_id_idx" ON "subscribers" ("trust_center_id");
-- Create index "subscriber_user_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_user_id_idx" ON "subscribers" ("user_id");
-- Create index "system_detail_assets_asset_id_idx" to table: "system_detail_assets"
CREATE INDEX "system_detail_assets_asset_id_idx" ON "system_detail_assets" ("asset_id");
-- Drop index "systemdetail_owner_id" from table: "system_details"
DROP INDEX "systemdetail_owner_id";
-- Create index "system_detail_owner_id_idx" to table: "system_details"
CREATE INDEX "system_detail_owner_id_idx" ON "system_details" ("owner_id");
-- Drop index "tagdefinition_owner_id" from table: "tag_definitions"
DROP INDEX "tagdefinition_owner_id";
-- Create index "tag_definition_owner_id_idx" to table: "tag_definitions"
CREATE INDEX "tag_definition_owner_id_idx" ON "tag_definitions" ("owner_id");
-- Create index "task_evidence_evidence_id_idx" to table: "task_evidence"
CREATE INDEX "task_evidence_evidence_id_idx" ON "task_evidence" ("evidence_id");
-- Drop index "task_owner_id" from table: "tasks"
DROP INDEX "task_owner_id";
-- Create index "task_assignee_id_idx" to table: "tasks"
CREATE INDEX "task_assignee_id_idx" ON "tasks" ("assignee_id");
-- Create index "task_assigner_id_idx" to table: "tasks"
CREATE INDEX "task_assigner_id_idx" ON "tasks" ("assigner_id");
-- Create index "task_owner_id_idx" to table: "tasks"
CREATE INDEX "task_owner_id_idx" ON "tasks" ("owner_id");
-- Create index "task_parent_task_id_idx" to table: "tasks"
CREATE INDEX "task_parent_task_id_idx" ON "tasks" ("parent_task_id");
-- Create index "template_files_file_id_idx" to table: "template_files"
CREATE INDEX "template_files_file_id_idx" ON "template_files" ("file_id");
-- Drop index "template_owner_id" from table: "templates"
DROP INDEX "template_owner_id";
-- Create index "template_owner_id_idx" to table: "templates"
CREATE INDEX "template_owner_id_idx" ON "templates" ("owner_id");
-- Create index "tfa_settings_owner_id_fk" to table: "tfa_settings"
CREATE INDEX "tfa_settings_owner_id_fk" ON "tfa_settings" ("owner_id");
-- Create index "trust_center_compliance_trust_center_id_idx" to table: "trust_center_compliances"
CREATE INDEX "trust_center_compliance_trust_center_id_idx" ON "trust_center_compliances" ("trust_center_id");
-- Create index "trust_center_doc_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_file_id_idx" ON "trust_center_docs" ("file_id");
-- Create index "trust_center_doc_original_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_original_file_id_idx" ON "trust_center_docs" ("original_file_id");
-- Create index "trust_center_doc_standard_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_standard_id_idx" ON "trust_center_docs" ("standard_id");
-- Create index "trust_center_doc_trust_center_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_trust_center_id_idx" ON "trust_center_docs" ("trust_center_id");
-- Create index "trust_center_entity_entity_type_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_entity_type_id_idx" ON "trust_center_entities" ("entity_type_id");
-- Create index "trust_center_entity_logo_file_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_logo_file_id_idx" ON "trust_center_entities" ("logo_file_id");
-- Create index "trust_center_entity_trust_center_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_trust_center_id_idx" ON "trust_center_entities" ("trust_center_id");
-- Create index "trust_center_faq_trust_center_id_idx" to table: "trust_center_faqs"
CREATE INDEX "trust_center_faq_trust_center_id_idx" ON "trust_center_faqs" ("trust_center_id");
-- Create index "trust_center_nda_request_approved_by_user_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_approved_by_user_id_idx" ON "trust_center_nda_requests" ("approved_by_user_id");
-- Create index "trust_center_nda_request_document_data_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_document_data_id_idx" ON "trust_center_nda_requests" ("document_data_id");
-- Create index "trust_center_nda_request_file_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_file_id_idx" ON "trust_center_nda_requests" ("file_id");
-- Create index "trust_center_nda_request_trust_center_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_trust_center_id_idx" ON "trust_center_nda_requests" ("trust_center_id");
-- Create index "trust_center_setting_favicon_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_favicon_local_file_id_idx" ON "trust_center_settings" ("favicon_local_file_id");
-- Create index "trust_center_setting_hero_image_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_hero_image_local_file_id_idx" ON "trust_center_settings" ("hero_image_local_file_id");
-- Create index "trust_center_setting_logo_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_logo_local_file_id_idx" ON "trust_center_settings" ("logo_local_file_id");
-- Create index "trust_center_setting_nda_approver_group_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_nda_approver_group_id_idx" ON "trust_center_settings" ("nda_approver_group_id");
-- Create index "trust_center_subprocessor_trust_center_id_idx" to table: "trust_center_subprocessors"
CREATE INDEX "trust_center_subprocessor_trust_center_id_idx" ON "trust_center_subprocessors" ("trust_center_id");
-- Drop index "trustcenterwatermarkconfig_owner_id" from table: "trust_center_watermark_configs"
DROP INDEX "trustcenterwatermarkconfig_owner_id";
-- Create index "trust_center_watermark_config_logo_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_logo_id_idx" ON "trust_center_watermark_configs" ("logo_id");
-- Create index "trust_center_watermark_config_owner_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_owner_id_idx" ON "trust_center_watermark_configs" ("owner_id");
-- Drop index "trustcenter_owner_id" from table: "trust_centers"
DROP INDEX "trustcenter_owner_id";
-- Create index "trust_center_custom_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_custom_domain_id_idx" ON "trust_centers" ("custom_domain_id");
-- Create index "trust_center_owner_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_owner_id_idx" ON "trust_centers" ("owner_id");
-- Create index "trust_center_preview_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_preview_domain_id_idx" ON "trust_centers" ("preview_domain_id");
-- Create index "user_events_event_id_idx" to table: "user_events"
CREATE INDEX "user_events_event_id_idx" ON "user_events" ("event_id");
-- Create index "user_setting_user_id_idx" to table: "user_settings"
CREATE INDEX "user_setting_user_id_idx" ON "user_settings" ("user_id");
-- Drop index "vendorriskscore_owner_id" from table: "vendor_risk_scores"
DROP INDEX "vendorriskscore_owner_id";
-- Create index "vendor_risk_score_assessment_response_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_assessment_response_id_idx" ON "vendor_risk_scores" ("assessment_response_id");
-- Create index "vendor_risk_score_entity_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_entity_id_idx" ON "vendor_risk_scores" ("entity_id");
-- Create index "vendor_risk_score_owner_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_owner_id_idx" ON "vendor_risk_scores" ("owner_id");
-- Create index "vendor_risk_score_vendor_scoring_config_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_vendor_scoring_config_id_idx" ON "vendor_risk_scores" ("vendor_scoring_config_id");
-- Drop index "vendorscoringconfig_owner_id" from table: "vendor_scoring_configs"
DROP INDEX "vendorscoringconfig_owner_id";
-- Create index "vendor_scoring_config_owner_id_idx" to table: "vendor_scoring_configs"
CREATE INDEX "vendor_scoring_config_owner_id_idx" ON "vendor_scoring_configs" ("owner_id");
-- Drop index "vulnerability_owner_id" from table: "vulnerabilities"
DROP INDEX "vulnerability_owner_id";
-- Create index "vulnerability_owner_id_idx" to table: "vulnerabilities"
CREATE INDEX "vulnerability_owner_id_idx" ON "vulnerabilities" ("owner_id");
-- Create index "vulnerability_action_plans_action_plan_id_idx" to table: "vulnerability_action_plans"
CREATE INDEX "vulnerability_action_plans_action_plan_id_idx" ON "vulnerability_action_plans" ("action_plan_id");
-- Create index "vulnerability_assets_asset_id_idx" to table: "vulnerability_assets"
CREATE INDEX "vulnerability_assets_asset_id_idx" ON "vulnerability_assets" ("asset_id");
-- Create index "vulnerability_controls_control_id_idx" to table: "vulnerability_controls"
CREATE INDEX "vulnerability_controls_control_id_idx" ON "vulnerability_controls" ("control_id");
-- Create index "vulnerability_entities_entity_id_idx" to table: "vulnerability_entities"
CREATE INDEX "vulnerability_entities_entity_id_idx" ON "vulnerability_entities" ("entity_id");
-- Create index "vulnerability_programs_program_id_idx" to table: "vulnerability_programs"
CREATE INDEX "vulnerability_programs_program_id_idx" ON "vulnerability_programs" ("program_id");
-- Create index "vulnerability_risks_risk_id_idx" to table: "vulnerability_risks"
CREATE INDEX "vulnerability_risks_risk_id_idx" ON "vulnerability_risks" ("risk_id");
-- Create index "vulnerability_scans_scan_id_idx" to table: "vulnerability_scans"
CREATE INDEX "vulnerability_scans_scan_id_idx" ON "vulnerability_scans" ("scan_id");
-- Create index "vulnerability_subcontrols_subcontrol_id_idx" to table: "vulnerability_subcontrols"
CREATE INDEX "vulnerability_subcontrols_subcontrol_id_idx" ON "vulnerability_subcontrols" ("subcontrol_id");
-- Create index "vulnerability_tasks_task_id_idx" to table: "vulnerability_tasks"
CREATE INDEX "vulnerability_tasks_task_id_idx" ON "vulnerability_tasks" ("task_id");
-- Create index "webauthns_owner_id_fk" to table: "webauthns"
CREATE INDEX "webauthns_owner_id_fk" ON "webauthns" ("owner_id");
-- Drop index "workflowassignmenttarget_owner_id" from table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_owner_id";
-- Create index "workflow_assignment_target_owner_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_owner_id_idx" ON "workflow_assignment_targets" ("owner_id");
-- Create index "workflow_assignment_target_target_group_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_group_id_idx" ON "workflow_assignment_targets" ("target_group_id");
-- Create index "workflow_assignment_target_target_user_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_user_id_idx" ON "workflow_assignment_targets" ("target_user_id");
-- Drop index "workflowassignment_owner_id" from table: "workflow_assignments"
DROP INDEX "workflowassignment_owner_id";
-- Create index "workflow_assignment_actor_group_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_group_id_idx" ON "workflow_assignments" ("actor_group_id");
-- Create index "workflow_assignment_actor_user_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_user_id_idx" ON "workflow_assignments" ("actor_user_id");
-- Create index "workflow_assignment_owner_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_owner_id_idx" ON "workflow_assignments" ("owner_id");
-- Drop index "workflowdefinition_owner_id" from table: "workflow_definitions"
DROP INDEX "workflowdefinition_owner_id";
-- Create index "workflow_definition_owner_id_idx" to table: "workflow_definitions"
CREATE INDEX "workflow_definition_owner_id_idx" ON "workflow_definitions" ("owner_id");
-- Drop index "workflowevent_owner_id" from table: "workflow_events"
DROP INDEX "workflowevent_owner_id";
-- Create index "workflow_event_owner_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_owner_id_idx" ON "workflow_events" ("owner_id");
-- Create index "workflow_event_workflow_instance_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_workflow_instance_id_idx" ON "workflow_events" ("workflow_instance_id");
-- Drop index "workflowinstance_owner_id" from table: "workflow_instances"
DROP INDEX "workflowinstance_owner_id";
-- Create index "workflow_instance_action_plan_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_action_plan_id_idx" ON "workflow_instances" ("action_plan_id");
-- Create index "workflow_instance_assessment_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_assessment_id_idx" ON "workflow_instances" ("assessment_id");
-- Create index "workflow_instance_assessment_response_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_assessment_response_id_idx" ON "workflow_instances" ("assessment_response_id");
-- Create index "workflow_instance_campaign_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_campaign_id_idx" ON "workflow_instances" ("campaign_id");
-- Create index "workflow_instance_campaign_target_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_campaign_target_id_idx" ON "workflow_instances" ("campaign_target_id");
-- Create index "workflow_instance_control_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_control_id_idx" ON "workflow_instances" ("control_id");
-- Create index "workflow_instance_evidence_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_evidence_id_idx" ON "workflow_instances" ("evidence_id");
-- Create index "workflow_instance_finding_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_finding_id_idx" ON "workflow_instances" ("finding_id");
-- Create index "workflow_instance_identity_holder_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_identity_holder_id_idx" ON "workflow_instances" ("identity_holder_id");
-- Create index "workflow_instance_integration_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_integration_id_idx" ON "workflow_instances" ("integration_id");
-- Create index "workflow_instance_internal_policy_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_internal_policy_id_idx" ON "workflow_instances" ("internal_policy_id");
-- Create index "workflow_instance_owner_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_owner_id_idx" ON "workflow_instances" ("owner_id");
-- Create index "workflow_instance_platform_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_platform_id_idx" ON "workflow_instances" ("platform_id");
-- Create index "workflow_instance_procedure_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_procedure_id_idx" ON "workflow_instances" ("procedure_id");
-- Create index "workflow_instance_remediation_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_remediation_id_idx" ON "workflow_instances" ("remediation_id");
-- Create index "workflow_instance_risk_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_risk_id_idx" ON "workflow_instances" ("risk_id");
-- Create index "workflow_instance_subcontrol_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_subcontrol_id_idx" ON "workflow_instances" ("subcontrol_id");
-- Create index "workflow_instance_task_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_task_id_idx" ON "workflow_instances" ("task_id");
-- Create index "workflow_instance_vulnerability_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_vulnerability_id_idx" ON "workflow_instances" ("vulnerability_id");
-- Create index "workflow_instance_workflow_proposal_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_workflow_proposal_id_idx" ON "workflow_instances" ("workflow_proposal_id");
-- Create index "workflow_object_ref_action_plan_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_action_plan_id_idx" ON "workflow_object_refs" ("action_plan_id");
-- Create index "workflow_object_ref_assessment_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_assessment_id_idx" ON "workflow_object_refs" ("assessment_id");
-- Create index "workflow_object_ref_assessment_response_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_assessment_response_id_idx" ON "workflow_object_refs" ("assessment_response_id");
-- Create index "workflow_object_ref_campaign_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_campaign_id_idx" ON "workflow_object_refs" ("campaign_id");
-- Create index "workflow_object_ref_campaign_target_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_campaign_target_id_idx" ON "workflow_object_refs" ("campaign_target_id");
-- Create index "workflow_object_ref_control_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_control_id_idx" ON "workflow_object_refs" ("control_id");
-- Create index "workflow_object_ref_directory_account_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_directory_account_id_idx" ON "workflow_object_refs" ("directory_account_id");
-- Create index "workflow_object_ref_directory_group_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_directory_group_id_idx" ON "workflow_object_refs" ("directory_group_id");
-- Create index "workflow_object_ref_directory_membership_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_directory_membership_id_idx" ON "workflow_object_refs" ("directory_membership_id");
-- Create index "workflow_object_ref_evidence_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_evidence_id_idx" ON "workflow_object_refs" ("evidence_id");
-- Create index "workflow_object_ref_finding_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_finding_id_idx" ON "workflow_object_refs" ("finding_id");
-- Create index "workflow_object_ref_identity_holder_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_identity_holder_id_idx" ON "workflow_object_refs" ("identity_holder_id");
-- Create index "workflow_object_ref_internal_policy_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_internal_policy_id_idx" ON "workflow_object_refs" ("internal_policy_id");
-- Create index "workflow_object_ref_owner_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_owner_id_idx" ON "workflow_object_refs" ("owner_id");
-- Create index "workflow_object_ref_platform_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_platform_id_idx" ON "workflow_object_refs" ("platform_id");
-- Create index "workflow_object_ref_procedure_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_procedure_id_idx" ON "workflow_object_refs" ("procedure_id");
-- Create index "workflow_object_ref_remediation_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_remediation_id_idx" ON "workflow_object_refs" ("remediation_id");
-- Create index "workflow_object_ref_risk_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_risk_id_idx" ON "workflow_object_refs" ("risk_id");
-- Create index "workflow_object_ref_subcontrol_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_subcontrol_id_idx" ON "workflow_object_refs" ("subcontrol_id");
-- Create index "workflow_object_ref_task_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_task_id_idx" ON "workflow_object_refs" ("task_id");
-- Create index "workflow_object_ref_vulnerability_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_vulnerability_id_idx" ON "workflow_object_refs" ("vulnerability_id");
-- Create index "workflow_proposal_owner_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_owner_id_idx" ON "workflow_proposals" ("owner_id");
-- Create index "workflow_proposal_submitted_by_user_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_submitted_by_user_id_idx" ON "workflow_proposals" ("submitted_by_user_id");
