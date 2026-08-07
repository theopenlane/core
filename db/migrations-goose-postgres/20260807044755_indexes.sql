-- +goose Up
-- create index "action_plan_blocked_groups_group_id_idx" to table: "action_plan_blocked_groups"
CREATE INDEX "action_plan_blocked_groups_group_id_idx" ON "action_plan_blocked_groups" ("group_id");
-- create index "action_plan_editors_group_id_idx" to table: "action_plan_editors"
CREATE INDEX "action_plan_editors_group_id_idx" ON "action_plan_editors" ("group_id");
-- create index "action_plan_tasks_task_id_idx" to table: "action_plan_tasks"
CREATE INDEX "action_plan_tasks_task_id_idx" ON "action_plan_tasks" ("task_id");
-- create index "action_plan_viewers_group_id_idx" to table: "action_plan_viewers"
CREATE INDEX "action_plan_viewers_group_id_idx" ON "action_plan_viewers" ("group_id");
-- drop index "actionplan_owner_id" from table: "action_plans"
DROP INDEX "actionplan_owner_id";
-- create index "action_plan_file_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_file_id_idx" ON "action_plans" ("file_id");
-- create index "action_plan_owner_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_owner_id_idx" ON "action_plans" ("owner_id");
-- drop index "apitoken_owner_id" from table: "api_tokens"
DROP INDEX "apitoken_owner_id";
-- create index "api_token_owner_id_idx" to table: "api_tokens"
CREATE INDEX "api_token_owner_id_idx" ON "api_tokens" ("owner_id");
-- drop index "assessmentresponse_owner_id" from table: "assessment_responses"
DROP INDEX "assessmentresponse_owner_id";
-- create index "assessment_response_document_data_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_document_data_id_idx" ON "assessment_responses" ("document_data_id");
-- create index "assessment_response_owner_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_owner_id_idx" ON "assessment_responses" ("owner_id");
-- drop index "assessment_owner_id" from table: "assessments"
DROP INDEX "assessment_owner_id";
-- create index "assessment_owner_id_idx" to table: "assessments"
CREATE INDEX "assessment_owner_id_idx" ON "assessments" ("owner_id");
-- create index "assessment_template_id_idx" to table: "assessments"
CREATE INDEX "assessment_template_id_idx" ON "assessments" ("template_id");
-- create index "asset_connected_assets_connected_from_id_idx" to table: "asset_connected_assets"
CREATE INDEX "asset_connected_assets_connected_from_id_idx" ON "asset_connected_assets" ("connected_from_id");
-- drop index "asset_owner_id" from table: "assets"
DROP INDEX "asset_owner_id";
-- create index "asset_integration_id_idx" to table: "assets"
CREATE INDEX "asset_integration_id_idx" ON "assets" ("integration_id");
-- create index "asset_owner_id_idx" to table: "assets"
CREATE INDEX "asset_owner_id_idx" ON "assets" ("owner_id");
-- create index "asset_source_platform_id_idx" to table: "assets"
CREATE INDEX "asset_source_platform_id_idx" ON "assets" ("source_platform_id");
-- create index "campaign_blocked_groups_group_id_idx" to table: "campaign_blocked_groups"
CREATE INDEX "campaign_blocked_groups_group_id_idx" ON "campaign_blocked_groups" ("group_id");
-- create index "campaign_contacts_contact_id_idx" to table: "campaign_contacts"
CREATE INDEX "campaign_contacts_contact_id_idx" ON "campaign_contacts" ("contact_id");
-- create index "campaign_editors_group_id_idx" to table: "campaign_editors"
CREATE INDEX "campaign_editors_group_id_idx" ON "campaign_editors" ("group_id");
-- create index "campaign_groups_group_id_idx" to table: "campaign_groups"
CREATE INDEX "campaign_groups_group_id_idx" ON "campaign_groups" ("group_id");
-- create index "campaign_identity_holders_identity_holder_id_idx" to table: "campaign_identity_holders"
CREATE INDEX "campaign_identity_holders_identity_holder_id_idx" ON "campaign_identity_holders" ("identity_holder_id");
-- drop index "campaigntarget_owner_id" from table: "campaign_targets"
DROP INDEX "campaigntarget_owner_id";
-- create index "campaign_target_owner_id_idx" to table: "campaign_targets"
CREATE INDEX "campaign_target_owner_id_idx" ON "campaign_targets" ("owner_id");
-- create index "campaign_users_user_id_idx" to table: "campaign_users"
CREATE INDEX "campaign_users_user_id_idx" ON "campaign_users" ("user_id");
-- create index "campaign_viewers_group_id_idx" to table: "campaign_viewers"
CREATE INDEX "campaign_viewers_group_id_idx" ON "campaign_viewers" ("group_id");
-- drop index "campaign_owner_id" from table: "campaigns"
DROP INDEX "campaign_owner_id";
-- create index "campaign_assessment_id_idx" to table: "campaigns"
CREATE INDEX "campaign_assessment_id_idx" ON "campaigns" ("assessment_id");
-- create index "campaign_email_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_email_template_id_idx" ON "campaigns" ("email_template_id");
-- create index "campaign_integration_id_idx" to table: "campaigns"
CREATE INDEX "campaign_integration_id_idx" ON "campaigns" ("integration_id");
-- create index "campaign_owner_id_idx" to table: "campaigns"
CREATE INDEX "campaign_owner_id_idx" ON "campaigns" ("owner_id");
-- create index "campaign_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_template_id_idx" ON "campaigns" ("template_id");
-- create index "campaign_trust_center_id_idx" to table: "campaigns"
CREATE INDEX "campaign_trust_center_id_idx" ON "campaigns" ("trust_center_id");
-- create index "check_result_controls_control_id_idx" to table: "check_result_controls"
CREATE INDEX "check_result_controls_control_id_idx" ON "check_result_controls" ("control_id");
-- create index "check_result_integration_id_idx" to table: "check_results"
CREATE INDEX "check_result_integration_id_idx" ON "check_results" ("integration_id");
-- create index "contact_files_file_id_idx" to table: "contact_files"
CREATE INDEX "contact_files_file_id_idx" ON "contact_files" ("file_id");
-- drop index "contact_owner_id" from table: "contacts"
DROP INDEX "contact_owner_id";
-- create index "contact_owner_id_idx" to table: "contacts"
CREATE INDEX "contact_owner_id_idx" ON "contacts" ("owner_id");
-- create index "control_action_plans_action_plan_id_idx" to table: "control_action_plans"
CREATE INDEX "control_action_plans_action_plan_id_idx" ON "control_action_plans" ("action_plan_id");
-- create index "control_assets_asset_id_idx" to table: "control_assets"
CREATE INDEX "control_assets_asset_id_idx" ON "control_assets" ("asset_id");
-- create index "control_blocked_groups_group_id_idx" to table: "control_blocked_groups"
CREATE INDEX "control_blocked_groups_group_id_idx" ON "control_blocked_groups" ("group_id");
-- create index "control_campaigns_campaign_id_idx" to table: "control_campaigns"
CREATE INDEX "control_campaigns_campaign_id_idx" ON "control_campaigns" ("campaign_id");
-- create index "control_control_implementations_control_implementation_id_idx" to table: "control_control_implementations"
CREATE INDEX "control_control_implementations_control_implementation_id_idx" ON "control_control_implementations" ("control_implementation_id");
-- create index "control_control_objectives_control_objective_id_idx" to table: "control_control_objectives"
CREATE INDEX "control_control_objectives_control_objective_id_idx" ON "control_control_objectives" ("control_objective_id");
-- create index "control_editors_group_id_idx" to table: "control_editors"
CREATE INDEX "control_editors_group_id_idx" ON "control_editors" ("group_id");
-- create index "control_entities_entity_id_idx" to table: "control_entities"
CREATE INDEX "control_entities_entity_id_idx" ON "control_entities" ("entity_id");
-- create index "control_identity_holders_identity_holder_id_idx" to table: "control_identity_holders"
CREATE INDEX "control_identity_holders_identity_holder_id_idx" ON "control_identity_holders" ("identity_holder_id");
-- create index "control_implementation_blocked_groups_group_id_idx" to table: "control_implementation_blocked_groups"
CREATE INDEX "control_implementation_blocked_groups_group_id_idx" ON "control_implementation_blocked_groups" ("group_id");
-- create index "control_implementation_editors_group_id_idx" to table: "control_implementation_editors"
CREATE INDEX "control_implementation_editors_group_id_idx" ON "control_implementation_editors" ("group_id");
-- create index "control_implementation_tasks_task_id_idx" to table: "control_implementation_tasks"
CREATE INDEX "control_implementation_tasks_task_id_idx" ON "control_implementation_tasks" ("task_id");
-- create index "control_implementation_viewers_group_id_idx" to table: "control_implementation_viewers"
CREATE INDEX "control_implementation_viewers_group_id_idx" ON "control_implementation_viewers" ("group_id");
-- drop index "controlimplementation_owner_id" from table: "control_implementations"
DROP INDEX "controlimplementation_owner_id";
-- create index "control_implementation_owner_id_idx" to table: "control_implementations"
CREATE INDEX "control_implementation_owner_id_idx" ON "control_implementations" ("owner_id");
-- create index "control_narratives_narrative_id_idx" to table: "control_narratives"
CREATE INDEX "control_narratives_narrative_id_idx" ON "control_narratives" ("narrative_id");
-- create index "control_objective_blocked_groups_group_id_idx" to table: "control_objective_blocked_groups"
CREATE INDEX "control_objective_blocked_groups_group_id_idx" ON "control_objective_blocked_groups" ("group_id");
-- create index "control_objective_editors_group_id_idx" to table: "control_objective_editors"
CREATE INDEX "control_objective_editors_group_id_idx" ON "control_objective_editors" ("group_id");
-- create index "control_objective_tasks_task_id_idx" to table: "control_objective_tasks"
CREATE INDEX "control_objective_tasks_task_id_idx" ON "control_objective_tasks" ("task_id");
-- create index "control_objective_viewers_group_id_idx" to table: "control_objective_viewers"
CREATE INDEX "control_objective_viewers_group_id_idx" ON "control_objective_viewers" ("group_id");
-- drop index "controlobjective_owner_id" from table: "control_objectives"
DROP INDEX "controlobjective_owner_id";
-- create index "control_objective_owner_id_idx" to table: "control_objectives"
CREATE INDEX "control_objective_owner_id_idx" ON "control_objectives" ("owner_id");
-- create index "control_procedures_procedure_id_idx" to table: "control_procedures"
CREATE INDEX "control_procedures_procedure_id_idx" ON "control_procedures" ("procedure_id");
-- create index "control_risks_risk_id_idx" to table: "control_risks"
CREATE INDEX "control_risks_risk_id_idx" ON "control_risks" ("risk_id");
-- create index "control_scans_scan_id_idx" to table: "control_scans"
CREATE INDEX "control_scans_scan_id_idx" ON "control_scans" ("scan_id");
-- create index "control_tasks_task_id_idx" to table: "control_tasks"
CREATE INDEX "control_tasks_task_id_idx" ON "control_tasks" ("task_id");
-- drop index "control_owner_id" from table: "controls"
DROP INDEX "control_owner_id";
-- create index "control_owner_id_idx" to table: "controls"
CREATE INDEX "control_owner_id_idx" ON "controls" ("owner_id");
-- drop index "customdomain_owner_id" from table: "custom_domains"
DROP INDEX "customdomain_owner_id";
-- create index "custom_domain_dns_verification_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_dns_verification_id_idx" ON "custom_domains" ("dns_verification_id");
-- create index "custom_domain_mappable_domain_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_mappable_domain_id_idx" ON "custom_domains" ("mappable_domain_id");
-- create index "custom_domain_owner_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_owner_id_idx" ON "custom_domains" ("owner_id");
-- drop index "customtypeenum_owner_id" from table: "custom_type_enums"
DROP INDEX "customtypeenum_owner_id";
-- create index "custom_type_enum_owner_id_idx" to table: "custom_type_enums"
CREATE INDEX "custom_type_enum_owner_id_idx" ON "custom_type_enums" ("owner_id");
-- create index "directory_account_avatar_local_file_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_avatar_local_file_id_idx" ON "directory_accounts" ("avatar_local_file_id");
-- create index "directory_account_owner_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_owner_id_idx" ON "directory_accounts" ("owner_id");
-- create index "directory_group_owner_id_idx" to table: "directory_groups"
CREATE INDEX "directory_group_owner_id_idx" ON "directory_groups" ("owner_id");
-- create index "directory_membership_directory_group_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_directory_group_id_idx" ON "directory_memberships" ("directory_group_id");
-- create index "directory_membership_owner_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_owner_id_idx" ON "directory_memberships" ("owner_id");
-- create index "directory_sync_run_owner_id_idx" to table: "directory_sync_runs"
CREATE INDEX "directory_sync_run_owner_id_idx" ON "directory_sync_runs" ("owner_id");
-- drop index "discussion_owner_id" from table: "discussions"
DROP INDEX "discussion_owner_id";
-- create index "discussion_owner_id_idx" to table: "discussions"
CREATE INDEX "discussion_owner_id_idx" ON "discussions" ("owner_id");
-- drop index "dnsverification_owner_id" from table: "dns_verifications"
DROP INDEX "dnsverification_owner_id";
-- create index "dns_verification_owner_id_idx" to table: "dns_verifications"
CREATE INDEX "dns_verification_owner_id_idx" ON "dns_verifications" ("owner_id");
-- drop index "documentdata_owner_id" from table: "document_data"
DROP INDEX "documentdata_owner_id";
-- create index "document_owner_id_idx" to table: "document_data"
CREATE INDEX "document_owner_id_idx" ON "document_data" ("owner_id");
-- create index "document_template_id_idx" to table: "document_data"
CREATE INDEX "document_template_id_idx" ON "document_data" ("template_id");
-- create index "document_data_files_file_id_idx" to table: "document_data_files"
CREATE INDEX "document_data_files_file_id_idx" ON "document_data_files" ("file_id");
-- drop index "emailtemplate_owner_id" from table: "email_templates"
DROP INDEX "emailtemplate_owner_id";
-- create index "email_template_integration_id_idx" to table: "email_templates"
CREATE INDEX "email_template_integration_id_idx" ON "email_templates" ("integration_id");
-- create index "email_template_owner_id_idx" to table: "email_templates"
CREATE INDEX "email_template_owner_id_idx" ON "email_templates" ("owner_id");
-- create index "email_template_trust_center_id_idx" to table: "email_templates"
CREATE INDEX "email_template_trust_center_id_idx" ON "email_templates" ("trust_center_id");
-- create index "email_template_workflow_definition_id_idx" to table: "email_templates"
CREATE INDEX "email_template_workflow_definition_id_idx" ON "email_templates" ("workflow_definition_id");
-- create index "email_template_workflow_instance_id_idx" to table: "email_templates"
CREATE INDEX "email_template_workflow_instance_id_idx" ON "email_templates" ("workflow_instance_id");
-- create index "email_verification_tokens_owner_id_fk" to table: "email_verification_tokens"
CREATE INDEX "email_verification_tokens_owner_id_fk" ON "email_verification_tokens" ("owner_id");
-- drop index "entity_owner_id" from table: "entities"
DROP INDEX "entity_owner_id";
-- create index "entity_entity_type_id_idx" to table: "entities"
CREATE INDEX "entity_entity_type_id_idx" ON "entities" ("entity_type_id");
-- create index "entity_logo_file_id_idx" to table: "entities"
CREATE INDEX "entity_logo_file_id_idx" ON "entities" ("logo_file_id");
-- create index "entity_owner_id_idx" to table: "entities"
CREATE INDEX "entity_owner_id_idx" ON "entities" ("owner_id");
-- create index "entity_assets_asset_id_idx" to table: "entity_assets"
CREATE INDEX "entity_assets_asset_id_idx" ON "entity_assets" ("asset_id");
-- create index "entity_blocked_groups_group_id_idx" to table: "entity_blocked_groups"
CREATE INDEX "entity_blocked_groups_group_id_idx" ON "entity_blocked_groups" ("group_id");
-- create index "entity_contacts_contact_id_idx" to table: "entity_contacts"
CREATE INDEX "entity_contacts_contact_id_idx" ON "entity_contacts" ("contact_id");
-- create index "entity_documents_document_data_id_idx" to table: "entity_documents"
CREATE INDEX "entity_documents_document_data_id_idx" ON "entity_documents" ("document_data_id");
-- create index "entity_editors_group_id_idx" to table: "entity_editors"
CREATE INDEX "entity_editors_group_id_idx" ON "entity_editors" ("group_id");
-- create index "entity_files_file_id_idx" to table: "entity_files"
CREATE INDEX "entity_files_file_id_idx" ON "entity_files" ("file_id");
-- create index "entity_integrations_integration_id_idx" to table: "entity_integrations"
CREATE INDEX "entity_integrations_integration_id_idx" ON "entity_integrations" ("integration_id");
-- create index "entity_subprocessors_subprocessor_id_idx" to table: "entity_subprocessors"
CREATE INDEX "entity_subprocessors_subprocessor_id_idx" ON "entity_subprocessors" ("subprocessor_id");
-- create index "entity_system_details_system_detail_id_idx" to table: "entity_system_details"
CREATE INDEX "entity_system_details_system_detail_id_idx" ON "entity_system_details" ("system_detail_id");
-- drop index "entitytype_owner_id" from table: "entity_types"
DROP INDEX "entitytype_owner_id";
-- create index "entity_type_owner_id_idx" to table: "entity_types"
CREATE INDEX "entity_type_owner_id_idx" ON "entity_types" ("owner_id");
-- create index "evidence_control_objectives_control_objective_id_idx" to table: "evidence_control_objectives"
CREATE INDEX "evidence_control_objectives_control_objective_id_idx" ON "evidence_control_objectives" ("control_objective_id");
-- create index "evidence_controls_control_id_idx" to table: "evidence_controls"
CREATE INDEX "evidence_controls_control_id_idx" ON "evidence_controls" ("control_id");
-- create index "evidence_files_file_id_idx" to table: "evidence_files"
CREATE INDEX "evidence_files_file_id_idx" ON "evidence_files" ("file_id");
-- create index "evidence_subcontrols_subcontrol_id_idx" to table: "evidence_subcontrols"
CREATE INDEX "evidence_subcontrols_subcontrol_id_idx" ON "evidence_subcontrols" ("subcontrol_id");
-- drop index "evidence_owner_id" from table: "evidences"
DROP INDEX "evidence_owner_id";
-- create index "evidence_owner_id_idx" to table: "evidences"
CREATE INDEX "evidence_owner_id_idx" ON "evidences" ("owner_id");
-- drop index "export_owner_id" from table: "exports"
DROP INDEX "export_owner_id";
-- create index "export_owner_id_idx" to table: "exports"
CREATE INDEX "export_owner_id_idx" ON "exports" ("owner_id");
-- create index "file_download_tokens_owner_id_fk" to table: "file_download_tokens"
CREATE INDEX "file_download_tokens_owner_id_fk" ON "file_download_tokens" ("owner_id");
-- create index "file_events_event_id_idx" to table: "file_events"
CREATE INDEX "file_events_event_id_idx" ON "file_events" ("event_id");
-- create index "file_secrets_hush_id_idx" to table: "file_secrets"
CREATE INDEX "file_secrets_hush_id_idx" ON "file_secrets" ("hush_id");
-- create index "finding_action_plans_action_plan_id_idx" to table: "finding_action_plans"
CREATE INDEX "finding_action_plans_action_plan_id_idx" ON "finding_action_plans" ("action_plan_id");
-- create index "finding_assets_asset_id_idx" to table: "finding_assets"
CREATE INDEX "finding_assets_asset_id_idx" ON "finding_assets" ("asset_id");
-- create index "finding_blocked_groups_group_id_idx" to table: "finding_blocked_groups"
CREATE INDEX "finding_blocked_groups_group_id_idx" ON "finding_blocked_groups" ("group_id");
-- create index "finding_check_results_check_result_id_idx" to table: "finding_check_results"
CREATE INDEX "finding_check_results_check_result_id_idx" ON "finding_check_results" ("check_result_id");
-- create index "finding_control_control_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_control_id_idx" ON "finding_controls" ("control_id");
-- create index "finding_control_owner_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_owner_id_idx" ON "finding_controls" ("owner_id");
-- create index "finding_control_standard_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_standard_id_idx" ON "finding_controls" ("standard_id");
-- create index "finding_directory_accounts_directory_account_id_idx" to table: "finding_directory_accounts"
CREATE INDEX "finding_directory_accounts_directory_account_id_idx" ON "finding_directory_accounts" ("directory_account_id");
-- create index "finding_editors_group_id_idx" to table: "finding_editors"
CREATE INDEX "finding_editors_group_id_idx" ON "finding_editors" ("group_id");
-- create index "finding_entities_entity_id_idx" to table: "finding_entities"
CREATE INDEX "finding_entities_entity_id_idx" ON "finding_entities" ("entity_id");
-- create index "finding_identity_holders_identity_holder_id_idx" to table: "finding_identity_holders"
CREATE INDEX "finding_identity_holders_identity_holder_id_idx" ON "finding_identity_holders" ("identity_holder_id");
-- create index "finding_programs_program_id_idx" to table: "finding_programs"
CREATE INDEX "finding_programs_program_id_idx" ON "finding_programs" ("program_id");
-- create index "finding_risks_risk_id_idx" to table: "finding_risks"
CREATE INDEX "finding_risks_risk_id_idx" ON "finding_risks" ("risk_id");
-- create index "finding_scans_scan_id_idx" to table: "finding_scans"
CREATE INDEX "finding_scans_scan_id_idx" ON "finding_scans" ("scan_id");
-- create index "finding_subcontrols_subcontrol_id_idx" to table: "finding_subcontrols"
CREATE INDEX "finding_subcontrols_subcontrol_id_idx" ON "finding_subcontrols" ("subcontrol_id");
-- create index "finding_tasks_task_id_idx" to table: "finding_tasks"
CREATE INDEX "finding_tasks_task_id_idx" ON "finding_tasks" ("task_id");
-- create index "finding_vulnerabilities_vulnerability_id_idx" to table: "finding_vulnerabilities"
CREATE INDEX "finding_vulnerabilities_vulnerability_id_idx" ON "finding_vulnerabilities" ("vulnerability_id");
-- drop index "finding_owner_id" from table: "findings"
DROP INDEX "finding_owner_id";
-- create index "finding_owner_id_idx" to table: "findings"
CREATE INDEX "finding_owner_id_idx" ON "findings" ("owner_id");
-- create index "group_events_event_id_idx" to table: "group_events"
CREATE INDEX "group_events_event_id_idx" ON "group_events" ("event_id");
-- create index "group_files_file_id_idx" to table: "group_files"
CREATE INDEX "group_files_file_id_idx" ON "group_files" ("file_id");
-- create index "group_membership_events_event_id_idx" to table: "group_membership_events"
CREATE INDEX "group_membership_events_event_id_idx" ON "group_membership_events" ("event_id");
-- create index "group_membership_group_id_idx" to table: "group_memberships"
CREATE INDEX "group_membership_group_id_idx" ON "group_memberships" ("group_id");
-- create index "group_setting_group_id_idx" to table: "group_settings"
CREATE INDEX "group_setting_group_id_idx" ON "group_settings" ("group_id");
-- create index "group_tasks_task_id_idx" to table: "group_tasks"
CREATE INDEX "group_tasks_task_id_idx" ON "group_tasks" ("task_id");
-- drop index "group_owner_id" from table: "groups"
DROP INDEX "group_owner_id";
-- create index "group_avatar_local_file_id_idx" to table: "groups"
CREATE INDEX "group_avatar_local_file_id_idx" ON "groups" ("avatar_local_file_id");
-- create index "group_owner_id_idx" to table: "groups"
CREATE INDEX "group_owner_id_idx" ON "groups" ("owner_id");
-- create index "hush_events_event_id_idx" to table: "hush_events"
CREATE INDEX "hush_events_event_id_idx" ON "hush_events" ("event_id");
-- drop index "hush_owner_id" from table: "hushes"
DROP INDEX "hush_owner_id";
-- create index "secret_owner_id_idx" to table: "hushes"
CREATE INDEX "secret_owner_id_idx" ON "hushes" ("owner_id");
-- create index "identity_holder_assessments_assessment_id_idx" to table: "identity_holder_assessments"
CREATE INDEX "identity_holder_assessments_assessment_id_idx" ON "identity_holder_assessments" ("assessment_id");
-- create index "identity_holder_assets_asset_id_idx" to table: "identity_holder_assets"
CREATE INDEX "identity_holder_assets_asset_id_idx" ON "identity_holder_assets" ("asset_id");
-- create index "identity_holder_entities_entity_id_idx" to table: "identity_holder_entities"
CREATE INDEX "identity_holder_entities_entity_id_idx" ON "identity_holder_entities" ("entity_id");
-- create index "identity_holder_files_file_id_idx" to table: "identity_holder_files"
CREATE INDEX "identity_holder_files_file_id_idx" ON "identity_holder_files" ("file_id");
-- create index "identity_holder_tasks_task_id_idx" to table: "identity_holder_tasks"
CREATE INDEX "identity_holder_tasks_task_id_idx" ON "identity_holder_tasks" ("task_id");
-- create index "identity_holder_templates_template_id_idx" to table: "identity_holder_templates"
CREATE INDEX "identity_holder_templates_template_id_idx" ON "identity_holder_templates" ("template_id");
-- drop index "identityholder_owner_id" from table: "identity_holders"
DROP INDEX "identityholder_owner_id";
-- create index "identity_holder_employer_entity_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_employer_entity_id_idx" ON "identity_holders" ("employer_entity_id");
-- create index "identity_holder_owner_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_owner_id_idx" ON "identity_holders" ("owner_id");
-- create index "impersonation_event_organization_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_organization_id_idx" ON "impersonation_events" ("organization_id");
-- create index "impersonation_event_target_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_target_user_id_idx" ON "impersonation_events" ("target_user_id");
-- create index "impersonation_event_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_user_id_idx" ON "impersonation_events" ("user_id");
-- create index "integration_action_plans_action_plan_id_idx" to table: "integration_action_plans"
CREATE INDEX "integration_action_plans_action_plan_id_idx" ON "integration_action_plans" ("action_plan_id");
-- create index "integration_events_event_id_idx" to table: "integration_events"
CREATE INDEX "integration_events_event_id_idx" ON "integration_events" ("event_id");
-- create index "integration_findings_finding_id_idx" to table: "integration_findings"
CREATE INDEX "integration_findings_finding_id_idx" ON "integration_findings" ("finding_id");
-- create index "integration_internal_policies_internal_policy_id_idx" to table: "integration_internal_policies"
CREATE INDEX "integration_internal_policies_internal_policy_id_idx" ON "integration_internal_policies" ("internal_policy_id");
-- create index "integration_remediations_remediation_id_idx" to table: "integration_remediations"
CREATE INDEX "integration_remediations_remediation_id_idx" ON "integration_remediations" ("remediation_id");
-- create index "integration_reviews_review_id_idx" to table: "integration_reviews"
CREATE INDEX "integration_reviews_review_id_idx" ON "integration_reviews" ("review_id");
-- drop index "integrationrun_owner_id" from table: "integration_runs"
DROP INDEX "integrationrun_owner_id";
-- create index "integration_run_event_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_event_id_idx" ON "integration_runs" ("event_id");
-- create index "integration_run_owner_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_owner_id_idx" ON "integration_runs" ("owner_id");
-- create index "integration_run_request_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_request_file_id_idx" ON "integration_runs" ("request_file_id");
-- create index "integration_run_response_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_response_file_id_idx" ON "integration_runs" ("response_file_id");
-- create index "integration_secrets_hush_id_idx" to table: "integration_secrets"
CREATE INDEX "integration_secrets_hush_id_idx" ON "integration_secrets" ("hush_id");
-- create index "integration_vulnerabilities_vulnerability_id_idx" to table: "integration_vulnerabilities"
CREATE INDEX "integration_vulnerabilities_vulnerability_id_idx" ON "integration_vulnerabilities" ("vulnerability_id");
-- drop index "integrationwebhook_owner_id" from table: "integration_webhooks"
DROP INDEX "integrationwebhook_owner_id";
-- create index "integration_webhook_owner_id_idx" to table: "integration_webhooks"
CREATE INDEX "integration_webhook_owner_id_idx" ON "integration_webhooks" ("owner_id");
-- drop index "integration_owner_id" from table: "integrations"
DROP INDEX "integration_owner_id";
-- create index "integration_owner_id_idx" to table: "integrations"
CREATE INDEX "integration_owner_id_idx" ON "integrations" ("owner_id");
-- create index "integration_platform_id_idx" to table: "integrations"
CREATE INDEX "integration_platform_id_idx" ON "integrations" ("platform_id");
-- drop index "internalpolicy_owner_id" from table: "internal_policies"
DROP INDEX "internalpolicy_owner_id";
-- create index "internal_policy_file_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_file_id_idx" ON "internal_policies" ("file_id");
-- create index "internal_policy_owner_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_owner_id_idx" ON "internal_policies" ("owner_id");
-- create index "internal_policy_assets_asset_id_idx" to table: "internal_policy_assets"
CREATE INDEX "internal_policy_assets_asset_id_idx" ON "internal_policy_assets" ("asset_id");
-- create index "internal_policy_blocked_groups_group_id_idx" to table: "internal_policy_blocked_groups"
CREATE INDEX "internal_policy_blocked_groups_group_id_idx" ON "internal_policy_blocked_groups" ("group_id");
-- create index "internal_policy_control_objectives_control_objective_id_idx" to table: "internal_policy_control_objectives"
CREATE INDEX "internal_policy_control_objectives_control_objective_id_idx" ON "internal_policy_control_objectives" ("control_objective_id");
-- create index "internal_policy_controls_control_id_idx" to table: "internal_policy_controls"
CREATE INDEX "internal_policy_controls_control_id_idx" ON "internal_policy_controls" ("control_id");
-- create index "internal_policy_editors_group_id_idx" to table: "internal_policy_editors"
CREATE INDEX "internal_policy_editors_group_id_idx" ON "internal_policy_editors" ("group_id");
-- create index "internal_policy_entities_entity_id_idx" to table: "internal_policy_entities"
CREATE INDEX "internal_policy_entities_entity_id_idx" ON "internal_policy_entities" ("entity_id");
-- create index "internal_policy_identity_holders_identity_holder_id_idx" to table: "internal_policy_identity_holders"
CREATE INDEX "internal_policy_identity_holders_identity_holder_id_idx" ON "internal_policy_identity_holders" ("identity_holder_id");
-- create index "internal_policy_narratives_narrative_id_idx" to table: "internal_policy_narratives"
CREATE INDEX "internal_policy_narratives_narrative_id_idx" ON "internal_policy_narratives" ("narrative_id");
-- create index "internal_policy_procedures_procedure_id_idx" to table: "internal_policy_procedures"
CREATE INDEX "internal_policy_procedures_procedure_id_idx" ON "internal_policy_procedures" ("procedure_id");
-- create index "internal_policy_risks_risk_id_idx" to table: "internal_policy_risks"
CREATE INDEX "internal_policy_risks_risk_id_idx" ON "internal_policy_risks" ("risk_id");
-- create index "internal_policy_subcontrols_subcontrol_id_idx" to table: "internal_policy_subcontrols"
CREATE INDEX "internal_policy_subcontrols_subcontrol_id_idx" ON "internal_policy_subcontrols" ("subcontrol_id");
-- create index "internal_policy_tasks_task_id_idx" to table: "internal_policy_tasks"
CREATE INDEX "internal_policy_tasks_task_id_idx" ON "internal_policy_tasks" ("task_id");
-- create index "invite_events_event_id_idx" to table: "invite_events"
CREATE INDEX "invite_events_event_id_idx" ON "invite_events" ("event_id");
-- create index "invite_groups_group_id_idx" to table: "invite_groups"
CREATE INDEX "invite_groups_group_id_idx" ON "invite_groups" ("group_id");
-- drop index "invite_owner_id" from table: "invites"
DROP INDEX "invite_owner_id";
-- create index "invite_owner_id_idx" to table: "invites"
CREATE INDEX "invite_owner_id_idx" ON "invites" ("owner_id");
-- drop index "jobresult_owner_id" from table: "job_results"
DROP INDEX "jobresult_owner_id";
-- create index "job_result_file_id_idx" to table: "job_results"
CREATE INDEX "job_result_file_id_idx" ON "job_results" ("file_id");
-- create index "job_result_owner_id_idx" to table: "job_results"
CREATE INDEX "job_result_owner_id_idx" ON "job_results" ("owner_id");
-- create index "job_result_scheduled_job_id_idx" to table: "job_results"
CREATE INDEX "job_result_scheduled_job_id_idx" ON "job_results" ("scheduled_job_id");
-- create index "job_runner_job_runner_tokens_job_runner_token_id_idx" to table: "job_runner_job_runner_tokens"
CREATE INDEX "job_runner_job_runner_tokens_job_runner_token_id_idx" ON "job_runner_job_runner_tokens" ("job_runner_token_id");
-- drop index "jobrunnerregistrationtoken_owner_id" from table: "job_runner_registration_tokens"
DROP INDEX "jobrunnerregistrationtoken_owner_id";
-- create index "job_runner_registration_token_job_runner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_job_runner_id_idx" ON "job_runner_registration_tokens" ("job_runner_id");
-- create index "job_runner_registration_token_owner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_owner_id_idx" ON "job_runner_registration_tokens" ("owner_id");
-- drop index "jobrunnertoken_owner_id" from table: "job_runner_tokens"
DROP INDEX "jobrunnertoken_owner_id";
-- create index "job_runner_token_owner_id_idx" to table: "job_runner_tokens"
CREATE INDEX "job_runner_token_owner_id_idx" ON "job_runner_tokens" ("owner_id");
-- drop index "jobrunner_owner_id" from table: "job_runners"
DROP INDEX "jobrunner_owner_id";
-- create index "job_runner_owner_id_idx" to table: "job_runners"
CREATE INDEX "job_runner_owner_id_idx" ON "job_runners" ("owner_id");
-- drop index "jobtemplate_owner_id" from table: "job_templates"
DROP INDEX "jobtemplate_owner_id";
-- create index "job_template_owner_id_idx" to table: "job_templates"
CREATE INDEX "job_template_owner_id_idx" ON "job_templates" ("owner_id");
-- create index "mapped_control_blocked_groups_group_id_idx" to table: "mapped_control_blocked_groups"
CREATE INDEX "mapped_control_blocked_groups_group_id_idx" ON "mapped_control_blocked_groups" ("group_id");
-- create index "mapped_control_editors_group_id_idx" to table: "mapped_control_editors"
CREATE INDEX "mapped_control_editors_group_id_idx" ON "mapped_control_editors" ("group_id");
-- create index "mapped_control_from_controls_control_id_idx" to table: "mapped_control_from_controls"
CREATE INDEX "mapped_control_from_controls_control_id_idx" ON "mapped_control_from_controls" ("control_id");
-- create index "mapped_control_from_subcontrols_subcontrol_id_idx" to table: "mapped_control_from_subcontrols"
CREATE INDEX "mapped_control_from_subcontrols_subcontrol_id_idx" ON "mapped_control_from_subcontrols" ("subcontrol_id");
-- create index "mapped_control_to_controls_control_id_idx" to table: "mapped_control_to_controls"
CREATE INDEX "mapped_control_to_controls_control_id_idx" ON "mapped_control_to_controls" ("control_id");
-- create index "mapped_control_to_subcontrols_subcontrol_id_idx" to table: "mapped_control_to_subcontrols"
CREATE INDEX "mapped_control_to_subcontrols_subcontrol_id_idx" ON "mapped_control_to_subcontrols" ("subcontrol_id");
-- drop index "mappedcontrol_owner_id" from table: "mapped_controls"
DROP INDEX "mappedcontrol_owner_id";
-- create index "mapped_control_owner_id_idx" to table: "mapped_controls"
CREATE INDEX "mapped_control_owner_id_idx" ON "mapped_controls" ("owner_id");
-- create index "narrative_blocked_groups_group_id_idx" to table: "narrative_blocked_groups"
CREATE INDEX "narrative_blocked_groups_group_id_idx" ON "narrative_blocked_groups" ("group_id");
-- create index "narrative_editors_group_id_idx" to table: "narrative_editors"
CREATE INDEX "narrative_editors_group_id_idx" ON "narrative_editors" ("group_id");
-- create index "narrative_viewers_group_id_idx" to table: "narrative_viewers"
CREATE INDEX "narrative_viewers_group_id_idx" ON "narrative_viewers" ("group_id");
-- drop index "narrative_owner_id" from table: "narratives"
DROP INDEX "narrative_owner_id";
-- create index "narrative_owner_id_idx" to table: "narratives"
CREATE INDEX "narrative_owner_id_idx" ON "narratives" ("owner_id");
-- drop index "note_owner_id" from table: "notes"
DROP INDEX "note_owner_id";
-- create index "note_discussion_id_idx" to table: "notes"
CREATE INDEX "note_discussion_id_idx" ON "notes" ("discussion_id");
-- create index "note_owner_id_idx" to table: "notes"
CREATE INDEX "note_owner_id_idx" ON "notes" ("owner_id");
-- create index "note_trust_center_id_idx" to table: "notes"
CREATE INDEX "note_trust_center_id_idx" ON "notes" ("trust_center_id");
-- drop index "notificationpreference_owner_id" from table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id";
-- create index "notification_preference_owner_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_owner_id_idx" ON "notification_preferences" ("owner_id");
-- create index "notification_preference_template_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_template_id_idx" ON "notification_preferences" ("template_id");
-- create index "notification_preference_user_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_user_id_idx" ON "notification_preferences" ("user_id");
-- drop index "notificationtemplate_owner_id" from table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id";
-- create index "notification_template_email_template_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_email_template_id_idx" ON "notification_templates" ("email_template_id");
-- create index "notification_template_integration_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_integration_id_idx" ON "notification_templates" ("integration_id");
-- create index "notification_template_owner_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_owner_id_idx" ON "notification_templates" ("owner_id");
-- create index "notification_template_workflow_definition_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_workflow_definition_id_idx" ON "notification_templates" ("workflow_definition_id");
-- create index "notification_owner_id_idx" to table: "notifications"
CREATE INDEX "notification_owner_id_idx" ON "notifications" ("owner_id");
-- create index "notification_template_id_idx" to table: "notifications"
CREATE INDEX "notification_template_id_idx" ON "notifications" ("template_id");
-- create index "org_membership_events_event_id_idx" to table: "org_membership_events"
CREATE INDEX "org_membership_events_event_id_idx" ON "org_membership_events" ("event_id");
-- create index "org_membership_organization_id_idx" to table: "org_memberships"
CREATE INDEX "org_membership_organization_id_idx" ON "org_memberships" ("organization_id");
-- create index "org_module_org_prices_org_price_id_idx" to table: "org_module_org_prices"
CREATE INDEX "org_module_org_prices_org_price_id_idx" ON "org_module_org_prices" ("org_price_id");
-- drop index "orgmodule_owner_id" from table: "org_modules"
DROP INDEX "orgmodule_owner_id";
-- create index "org_module_owner_id_idx" to table: "org_modules"
CREATE INDEX "org_module_owner_id_idx" ON "org_modules" ("owner_id");
-- create index "org_module_subscription_id_idx" to table: "org_modules"
CREATE INDEX "org_module_subscription_id_idx" ON "org_modules" ("subscription_id");
-- drop index "orgprice_owner_id" from table: "org_prices"
DROP INDEX "orgprice_owner_id";
-- create index "org_price_owner_id_idx" to table: "org_prices"
CREATE INDEX "org_price_owner_id_idx" ON "org_prices" ("owner_id");
-- create index "org_price_subscription_id_idx" to table: "org_prices"
CREATE INDEX "org_price_subscription_id_idx" ON "org_prices" ("subscription_id");
-- create index "org_product_org_prices_org_price_id_idx" to table: "org_product_org_prices"
CREATE INDEX "org_product_org_prices_org_price_id_idx" ON "org_product_org_prices" ("org_price_id");
-- drop index "orgproduct_owner_id" from table: "org_products"
DROP INDEX "orgproduct_owner_id";
-- create index "org_product_owner_id_idx" to table: "org_products"
CREATE INDEX "org_product_owner_id_idx" ON "org_products" ("owner_id");
-- create index "org_product_subscription_id_idx" to table: "org_products"
CREATE INDEX "org_product_subscription_id_idx" ON "org_products" ("subscription_id");
-- create index "org_subscription_events_event_id_idx" to table: "org_subscription_events"
CREATE INDEX "org_subscription_events_event_id_idx" ON "org_subscription_events" ("event_id");
-- drop index "orgsubscription_owner_id" from table: "org_subscriptions"
DROP INDEX "orgsubscription_owner_id";
-- create index "org_subscription_owner_id_idx" to table: "org_subscriptions"
CREATE INDEX "org_subscription_owner_id_idx" ON "org_subscriptions" ("owner_id");
-- create index "organization_events_event_id_idx" to table: "organization_events"
CREATE INDEX "organization_events_event_id_idx" ON "organization_events" ("event_id");
-- create index "organization_files_file_id_idx" to table: "organization_files"
CREATE INDEX "organization_files_file_id_idx" ON "organization_files" ("file_id");
-- create index "organization_personal_access_tokens_personal_access_token_id_id" to table: "organization_personal_access_tokens"
CREATE INDEX "organization_personal_access_tokens_personal_access_token_id_id" ON "organization_personal_access_tokens" ("personal_access_token_id");
-- create index "organization_setting_files_file_id_idx" to table: "organization_setting_files"
CREATE INDEX "organization_setting_files_file_id_idx" ON "organization_setting_files" ("file_id");
-- create index "organization_setting_organization_id_idx" to table: "organization_settings"
CREATE INDEX "organization_setting_organization_id_idx" ON "organization_settings" ("organization_id");
-- create index "organization_avatar_local_file_id_idx" to table: "organizations"
CREATE INDEX "organization_avatar_local_file_id_idx" ON "organizations" ("avatar_local_file_id");
-- create index "organization_parent_organization_id_idx" to table: "organizations"
CREATE INDEX "organization_parent_organization_id_idx" ON "organizations" ("parent_organization_id");
-- create index "password_reset_tokens_owner_id_fk" to table: "password_reset_tokens"
CREATE INDEX "password_reset_tokens_owner_id_fk" ON "password_reset_tokens" ("owner_id");
-- create index "personal_access_token_events_event_id_idx" to table: "personal_access_token_events"
CREATE INDEX "personal_access_token_events_event_id_idx" ON "personal_access_token_events" ("event_id");
-- create index "personal_access_tokens_owner_id_fk" to table: "personal_access_tokens"
CREATE INDEX "personal_access_tokens_owner_id_fk" ON "personal_access_tokens" ("owner_id");
-- create index "platform_applicable_frameworks_standard_id_idx" to table: "platform_applicable_frameworks"
CREATE INDEX "platform_applicable_frameworks_standard_id_idx" ON "platform_applicable_frameworks" ("standard_id");
-- create index "platform_assessments_assessment_id_idx" to table: "platform_assessments"
CREATE INDEX "platform_assessments_assessment_id_idx" ON "platform_assessments" ("assessment_id");
-- create index "platform_assets_asset_id_idx" to table: "platform_assets"
CREATE INDEX "platform_assets_asset_id_idx" ON "platform_assets" ("asset_id");
-- create index "platform_blocked_groups_group_id_idx" to table: "platform_blocked_groups"
CREATE INDEX "platform_blocked_groups_group_id_idx" ON "platform_blocked_groups" ("group_id");
-- create index "platform_controls_control_id_idx" to table: "platform_controls"
CREATE INDEX "platform_controls_control_id_idx" ON "platform_controls" ("control_id");
-- create index "platform_editors_group_id_idx" to table: "platform_editors"
CREATE INDEX "platform_editors_group_id_idx" ON "platform_editors" ("group_id");
-- create index "platform_entities_entity_id_idx" to table: "platform_entities"
CREATE INDEX "platform_entities_entity_id_idx" ON "platform_entities" ("entity_id");
-- create index "platform_evidence_evidence_id_idx" to table: "platform_evidence"
CREATE INDEX "platform_evidence_evidence_id_idx" ON "platform_evidence" ("evidence_id");
-- create index "platform_files_file_id_idx" to table: "platform_files"
CREATE INDEX "platform_files_file_id_idx" ON "platform_files" ("file_id");
-- create index "platform_identity_holders_identity_holder_id_idx" to table: "platform_identity_holders"
CREATE INDEX "platform_identity_holders_identity_holder_id_idx" ON "platform_identity_holders" ("identity_holder_id");
-- create index "platform_out_of_scope_assets_asset_id_idx" to table: "platform_out_of_scope_assets"
CREATE INDEX "platform_out_of_scope_assets_asset_id_idx" ON "platform_out_of_scope_assets" ("asset_id");
-- create index "platform_out_of_scope_vendors_entity_id_idx" to table: "platform_out_of_scope_vendors"
CREATE INDEX "platform_out_of_scope_vendors_entity_id_idx" ON "platform_out_of_scope_vendors" ("entity_id");
-- create index "platform_risks_risk_id_idx" to table: "platform_risks"
CREATE INDEX "platform_risks_risk_id_idx" ON "platform_risks" ("risk_id");
-- create index "platform_scans_scan_id_idx" to table: "platform_scans"
CREATE INDEX "platform_scans_scan_id_idx" ON "platform_scans" ("scan_id");
-- create index "platform_source_entities_entity_id_idx" to table: "platform_source_entities"
CREATE INDEX "platform_source_entities_entity_id_idx" ON "platform_source_entities" ("entity_id");
-- create index "platform_system_details_system_detail_id_idx" to table: "platform_system_details"
CREATE INDEX "platform_system_details_system_detail_id_idx" ON "platform_system_details" ("system_detail_id");
-- create index "platform_tasks_task_id_idx" to table: "platform_tasks"
CREATE INDEX "platform_tasks_task_id_idx" ON "platform_tasks" ("task_id");
-- create index "platform_viewers_group_id_idx" to table: "platform_viewers"
CREATE INDEX "platform_viewers_group_id_idx" ON "platform_viewers" ("group_id");
-- drop index "platform_owner_id" from table: "platforms"
DROP INDEX "platform_owner_id";
-- create index "platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_owner_id_idx" ON "platforms" ("owner_id");
-- create index "platform_platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_platform_owner_id_idx" ON "platforms" ("platform_owner_id");
-- create index "procedure_blocked_groups_group_id_idx" to table: "procedure_blocked_groups"
CREATE INDEX "procedure_blocked_groups_group_id_idx" ON "procedure_blocked_groups" ("group_id");
-- create index "procedure_editors_group_id_idx" to table: "procedure_editors"
CREATE INDEX "procedure_editors_group_id_idx" ON "procedure_editors" ("group_id");
-- create index "procedure_narratives_narrative_id_idx" to table: "procedure_narratives"
CREATE INDEX "procedure_narratives_narrative_id_idx" ON "procedure_narratives" ("narrative_id");
-- create index "procedure_risks_risk_id_idx" to table: "procedure_risks"
CREATE INDEX "procedure_risks_risk_id_idx" ON "procedure_risks" ("risk_id");
-- create index "procedure_tasks_task_id_idx" to table: "procedure_tasks"
CREATE INDEX "procedure_tasks_task_id_idx" ON "procedure_tasks" ("task_id");
-- drop index "procedure_owner_id" from table: "procedures"
DROP INDEX "procedure_owner_id";
-- create index "procedure_file_id_idx" to table: "procedures"
CREATE INDEX "procedure_file_id_idx" ON "procedures" ("file_id");
-- create index "procedure_owner_id_idx" to table: "procedures"
CREATE INDEX "procedure_owner_id_idx" ON "procedures" ("owner_id");
-- create index "program_action_plans_action_plan_id_idx" to table: "program_action_plans"
CREATE INDEX "program_action_plans_action_plan_id_idx" ON "program_action_plans" ("action_plan_id");
-- create index "program_blocked_groups_group_id_idx" to table: "program_blocked_groups"
CREATE INDEX "program_blocked_groups_group_id_idx" ON "program_blocked_groups" ("group_id");
-- create index "program_control_objectives_control_objective_id_idx" to table: "program_control_objectives"
CREATE INDEX "program_control_objectives_control_objective_id_idx" ON "program_control_objectives" ("control_objective_id");
-- create index "program_controls_control_id_idx" to table: "program_controls"
CREATE INDEX "program_controls_control_id_idx" ON "program_controls" ("control_id");
-- create index "program_editors_group_id_idx" to table: "program_editors"
CREATE INDEX "program_editors_group_id_idx" ON "program_editors" ("group_id");
-- create index "program_evidence_evidence_id_idx" to table: "program_evidence"
CREATE INDEX "program_evidence_evidence_id_idx" ON "program_evidence" ("evidence_id");
-- create index "program_files_file_id_idx" to table: "program_files"
CREATE INDEX "program_files_file_id_idx" ON "program_files" ("file_id");
-- create index "program_internal_policies_internal_policy_id_idx" to table: "program_internal_policies"
CREATE INDEX "program_internal_policies_internal_policy_id_idx" ON "program_internal_policies" ("internal_policy_id");
-- create index "program_membership_program_id_idx" to table: "program_memberships"
CREATE INDEX "program_membership_program_id_idx" ON "program_memberships" ("program_id");
-- create index "program_narratives_narrative_id_idx" to table: "program_narratives"
CREATE INDEX "program_narratives_narrative_id_idx" ON "program_narratives" ("narrative_id");
-- create index "program_procedures_procedure_id_idx" to table: "program_procedures"
CREATE INDEX "program_procedures_procedure_id_idx" ON "program_procedures" ("procedure_id");
-- create index "program_risks_risk_id_idx" to table: "program_risks"
CREATE INDEX "program_risks_risk_id_idx" ON "program_risks" ("risk_id");
-- create index "program_system_details_system_detail_id_idx" to table: "program_system_details"
CREATE INDEX "program_system_details_system_detail_id_idx" ON "program_system_details" ("system_detail_id");
-- create index "program_tasks_task_id_idx" to table: "program_tasks"
CREATE INDEX "program_tasks_task_id_idx" ON "program_tasks" ("task_id");
-- create index "program_viewers_group_id_idx" to table: "program_viewers"
CREATE INDEX "program_viewers_group_id_idx" ON "program_viewers" ("group_id");
-- drop index "program_owner_id" from table: "programs"
DROP INDEX "program_owner_id";
-- create index "program_owner_id_idx" to table: "programs"
CREATE INDEX "program_owner_id_idx" ON "programs" ("owner_id");
-- create index "program_program_owner_id_idx" to table: "programs"
CREATE INDEX "program_program_owner_id_idx" ON "programs" ("program_owner_id");
-- create index "remediation_action_plans_action_plan_id_idx" to table: "remediation_action_plans"
CREATE INDEX "remediation_action_plans_action_plan_id_idx" ON "remediation_action_plans" ("action_plan_id");
-- create index "remediation_assets_asset_id_idx" to table: "remediation_assets"
CREATE INDEX "remediation_assets_asset_id_idx" ON "remediation_assets" ("asset_id");
-- create index "remediation_blocked_groups_group_id_idx" to table: "remediation_blocked_groups"
CREATE INDEX "remediation_blocked_groups_group_id_idx" ON "remediation_blocked_groups" ("group_id");
-- create index "remediation_controls_control_id_idx" to table: "remediation_controls"
CREATE INDEX "remediation_controls_control_id_idx" ON "remediation_controls" ("control_id");
-- create index "remediation_editors_group_id_idx" to table: "remediation_editors"
CREATE INDEX "remediation_editors_group_id_idx" ON "remediation_editors" ("group_id");
-- create index "remediation_entities_entity_id_idx" to table: "remediation_entities"
CREATE INDEX "remediation_entities_entity_id_idx" ON "remediation_entities" ("entity_id");
-- create index "remediation_findings_finding_id_idx" to table: "remediation_findings"
CREATE INDEX "remediation_findings_finding_id_idx" ON "remediation_findings" ("finding_id");
-- create index "remediation_programs_program_id_idx" to table: "remediation_programs"
CREATE INDEX "remediation_programs_program_id_idx" ON "remediation_programs" ("program_id");
-- create index "remediation_risks_risk_id_idx" to table: "remediation_risks"
CREATE INDEX "remediation_risks_risk_id_idx" ON "remediation_risks" ("risk_id");
-- create index "remediation_subcontrols_subcontrol_id_idx" to table: "remediation_subcontrols"
CREATE INDEX "remediation_subcontrols_subcontrol_id_idx" ON "remediation_subcontrols" ("subcontrol_id");
-- create index "remediation_vulnerabilities_vulnerability_id_idx" to table: "remediation_vulnerabilities"
CREATE INDEX "remediation_vulnerabilities_vulnerability_id_idx" ON "remediation_vulnerabilities" ("vulnerability_id");
-- drop index "remediation_owner_id" from table: "remediations"
DROP INDEX "remediation_owner_id";
-- create index "remediation_owner_id_idx" to table: "remediations"
CREATE INDEX "remediation_owner_id_idx" ON "remediations" ("owner_id");
-- create index "review_action_plans_action_plan_id_idx" to table: "review_action_plans"
CREATE INDEX "review_action_plans_action_plan_id_idx" ON "review_action_plans" ("action_plan_id");
-- create index "review_assets_asset_id_idx" to table: "review_assets"
CREATE INDEX "review_assets_asset_id_idx" ON "review_assets" ("asset_id");
-- create index "review_blocked_groups_group_id_idx" to table: "review_blocked_groups"
CREATE INDEX "review_blocked_groups_group_id_idx" ON "review_blocked_groups" ("group_id");
-- create index "review_controls_control_id_idx" to table: "review_controls"
CREATE INDEX "review_controls_control_id_idx" ON "review_controls" ("control_id");
-- create index "review_editors_group_id_idx" to table: "review_editors"
CREATE INDEX "review_editors_group_id_idx" ON "review_editors" ("group_id");
-- create index "review_entities_entity_id_idx" to table: "review_entities"
CREATE INDEX "review_entities_entity_id_idx" ON "review_entities" ("entity_id");
-- create index "review_findings_finding_id_idx" to table: "review_findings"
CREATE INDEX "review_findings_finding_id_idx" ON "review_findings" ("finding_id");
-- create index "review_internal_policies_internal_policy_id_idx" to table: "review_internal_policies"
CREATE INDEX "review_internal_policies_internal_policy_id_idx" ON "review_internal_policies" ("internal_policy_id");
-- create index "review_programs_program_id_idx" to table: "review_programs"
CREATE INDEX "review_programs_program_id_idx" ON "review_programs" ("program_id");
-- create index "review_remediations_remediation_id_idx" to table: "review_remediations"
CREATE INDEX "review_remediations_remediation_id_idx" ON "review_remediations" ("remediation_id");
-- create index "review_risks_risk_id_idx" to table: "review_risks"
CREATE INDEX "review_risks_risk_id_idx" ON "review_risks" ("risk_id");
-- create index "review_subcontrols_subcontrol_id_idx" to table: "review_subcontrols"
CREATE INDEX "review_subcontrols_subcontrol_id_idx" ON "review_subcontrols" ("subcontrol_id");
-- create index "review_vulnerabilities_vulnerability_id_idx" to table: "review_vulnerabilities"
CREATE INDEX "review_vulnerabilities_vulnerability_id_idx" ON "review_vulnerabilities" ("vulnerability_id");
-- drop index "review_owner_id" from table: "reviews"
DROP INDEX "review_owner_id";
-- create index "review_owner_id_idx" to table: "reviews"
CREATE INDEX "review_owner_id_idx" ON "reviews" ("owner_id");
-- create index "review_reviewer_id_idx" to table: "reviews"
CREATE INDEX "review_reviewer_id_idx" ON "reviews" ("reviewer_id");
-- create index "risk_action_plans_action_plan_id_idx" to table: "risk_action_plans"
CREATE INDEX "risk_action_plans_action_plan_id_idx" ON "risk_action_plans" ("action_plan_id");
-- create index "risk_blocked_groups_group_id_idx" to table: "risk_blocked_groups"
CREATE INDEX "risk_blocked_groups_group_id_idx" ON "risk_blocked_groups" ("group_id");
-- create index "risk_editors_group_id_idx" to table: "risk_editors"
CREATE INDEX "risk_editors_group_id_idx" ON "risk_editors" ("group_id");
-- create index "risk_tasks_task_id_idx" to table: "risk_tasks"
CREATE INDEX "risk_tasks_task_id_idx" ON "risk_tasks" ("task_id");
-- create index "risk_viewers_group_id_idx" to table: "risk_viewers"
CREATE INDEX "risk_viewers_group_id_idx" ON "risk_viewers" ("group_id");
-- drop index "risk_owner_id" from table: "risks"
DROP INDEX "risk_owner_id";
-- create index "risk_delegate_id_idx" to table: "risks"
CREATE INDEX "risk_delegate_id_idx" ON "risks" ("delegate_id");
-- create index "risk_owner_id_idx" to table: "risks"
CREATE INDEX "risk_owner_id_idx" ON "risks" ("owner_id");
-- create index "risk_stakeholder_id_idx" to table: "risks"
CREATE INDEX "risk_stakeholder_id_idx" ON "risks" ("stakeholder_id");
-- create index "scan_action_plans_action_plan_id_idx" to table: "scan_action_plans"
CREATE INDEX "scan_action_plans_action_plan_id_idx" ON "scan_action_plans" ("action_plan_id");
-- create index "scan_assets_asset_id_idx" to table: "scan_assets"
CREATE INDEX "scan_assets_asset_id_idx" ON "scan_assets" ("asset_id");
-- create index "scan_blocked_groups_group_id_idx" to table: "scan_blocked_groups"
CREATE INDEX "scan_blocked_groups_group_id_idx" ON "scan_blocked_groups" ("group_id");
-- create index "scan_editors_group_id_idx" to table: "scan_editors"
CREATE INDEX "scan_editors_group_id_idx" ON "scan_editors" ("group_id");
-- create index "scan_entities_entity_id_idx" to table: "scan_entities"
CREATE INDEX "scan_entities_entity_id_idx" ON "scan_entities" ("entity_id");
-- create index "scan_evidence_evidence_id_idx" to table: "scan_evidence"
CREATE INDEX "scan_evidence_evidence_id_idx" ON "scan_evidence" ("evidence_id");
-- create index "scan_files_file_id_idx" to table: "scan_files"
CREATE INDEX "scan_files_file_id_idx" ON "scan_files" ("file_id");
-- create index "scan_remediations_remediation_id_idx" to table: "scan_remediations"
CREATE INDEX "scan_remediations_remediation_id_idx" ON "scan_remediations" ("remediation_id");
-- create index "scan_tasks_task_id_idx" to table: "scan_tasks"
CREATE INDEX "scan_tasks_task_id_idx" ON "scan_tasks" ("task_id");
-- drop index "scan_owner_id" from table: "scans"
DROP INDEX "scan_owner_id";
-- create index "scan_generated_by_platform_id_idx" to table: "scans"
CREATE INDEX "scan_generated_by_platform_id_idx" ON "scans" ("generated_by_platform_id");
-- create index "scan_owner_id_idx" to table: "scans"
CREATE INDEX "scan_owner_id_idx" ON "scans" ("owner_id");
-- create index "scan_performed_by_group_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_group_id_idx" ON "scans" ("performed_by_group_id");
-- create index "scan_performed_by_user_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_user_id_idx" ON "scans" ("performed_by_user_id");
-- create index "scheduled_job_controls_control_id_idx" to table: "scheduled_job_controls"
CREATE INDEX "scheduled_job_controls_control_id_idx" ON "scheduled_job_controls" ("control_id");
-- drop index "scheduledjobrun_owner_id" from table: "scheduled_job_runs"
DROP INDEX "scheduledjobrun_owner_id";
-- create index "scheduled_job_run_job_runner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_job_runner_id_idx" ON "scheduled_job_runs" ("job_runner_id");
-- create index "scheduled_job_run_owner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_owner_id_idx" ON "scheduled_job_runs" ("owner_id");
-- create index "scheduled_job_run_scheduled_job_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_scheduled_job_id_idx" ON "scheduled_job_runs" ("scheduled_job_id");
-- create index "scheduled_job_subcontrols_subcontrol_id_idx" to table: "scheduled_job_subcontrols"
CREATE INDEX "scheduled_job_subcontrols_subcontrol_id_idx" ON "scheduled_job_subcontrols" ("subcontrol_id");
-- drop index "scheduledjob_owner_id" from table: "scheduled_jobs"
DROP INDEX "scheduledjob_owner_id";
-- create index "scheduled_job_job_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_id_idx" ON "scheduled_jobs" ("job_id");
-- create index "scheduled_job_job_runner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_runner_id_idx" ON "scheduled_jobs" ("job_runner_id");
-- create index "scheduled_job_owner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_owner_id_idx" ON "scheduled_jobs" ("owner_id");
-- drop index "sladefinition_owner_id" from table: "sla_definitions"
DROP INDEX "sladefinition_owner_id";
-- create index "sla_definition_owner_id_idx" to table: "sla_definitions"
CREATE INDEX "sla_definition_owner_id_idx" ON "sla_definitions" ("owner_id");
-- drop index "standard_owner_id" from table: "standards"
DROP INDEX "standard_owner_id";
-- create index "standard_logo_file_id_idx" to table: "standards"
CREATE INDEX "standard_logo_file_id_idx" ON "standards" ("logo_file_id");
-- create index "standard_owner_id_idx" to table: "standards"
CREATE INDEX "standard_owner_id_idx" ON "standards" ("owner_id");
-- create index "subcontrol_assets_asset_id_idx" to table: "subcontrol_assets"
CREATE INDEX "subcontrol_assets_asset_id_idx" ON "subcontrol_assets" ("asset_id");
-- create index "subcontrol_control_implementations_control_implementation_id_id" to table: "subcontrol_control_implementations"
CREATE INDEX "subcontrol_control_implementations_control_implementation_id_id" ON "subcontrol_control_implementations" ("control_implementation_id");
-- create index "subcontrol_control_objectives_control_objective_id_idx" to table: "subcontrol_control_objectives"
CREATE INDEX "subcontrol_control_objectives_control_objective_id_idx" ON "subcontrol_control_objectives" ("control_objective_id");
-- create index "subcontrol_entities_entity_id_idx" to table: "subcontrol_entities"
CREATE INDEX "subcontrol_entities_entity_id_idx" ON "subcontrol_entities" ("entity_id");
-- create index "subcontrol_identity_holders_identity_holder_id_idx" to table: "subcontrol_identity_holders"
CREATE INDEX "subcontrol_identity_holders_identity_holder_id_idx" ON "subcontrol_identity_holders" ("identity_holder_id");
-- create index "subcontrol_procedures_procedure_id_idx" to table: "subcontrol_procedures"
CREATE INDEX "subcontrol_procedures_procedure_id_idx" ON "subcontrol_procedures" ("procedure_id");
-- create index "subcontrol_risks_risk_id_idx" to table: "subcontrol_risks"
CREATE INDEX "subcontrol_risks_risk_id_idx" ON "subcontrol_risks" ("risk_id");
-- create index "subcontrol_scans_scan_id_idx" to table: "subcontrol_scans"
CREATE INDEX "subcontrol_scans_scan_id_idx" ON "subcontrol_scans" ("scan_id");
-- create index "subcontrol_tasks_task_id_idx" to table: "subcontrol_tasks"
CREATE INDEX "subcontrol_tasks_task_id_idx" ON "subcontrol_tasks" ("task_id");
-- drop index "subcontrol_owner_id" from table: "subcontrols"
DROP INDEX "subcontrol_owner_id";
-- create index "subcontrol_owner_id_idx" to table: "subcontrols"
CREATE INDEX "subcontrol_owner_id_idx" ON "subcontrols" ("owner_id");
-- drop index "subprocessor_owner_id" from table: "subprocessors"
DROP INDEX "subprocessor_owner_id";
-- create index "subprocessor_logo_file_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_logo_file_id_idx" ON "subprocessors" ("logo_file_id");
-- create index "subprocessor_owner_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_owner_id_idx" ON "subprocessors" ("owner_id");
-- create index "subscriber_events_event_id_idx" to table: "subscriber_events"
CREATE INDEX "subscriber_events_event_id_idx" ON "subscriber_events" ("event_id");
-- drop index "subscriber_owner_id" from table: "subscribers"
DROP INDEX "subscriber_owner_id";
-- create index "subscriber_contact_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_contact_id_idx" ON "subscribers" ("contact_id");
-- create index "subscriber_owner_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_owner_id_idx" ON "subscribers" ("owner_id");
-- create index "subscriber_trust_center_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_trust_center_id_idx" ON "subscribers" ("trust_center_id");
-- create index "subscriber_user_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_user_id_idx" ON "subscribers" ("user_id");
-- create index "system_detail_assets_asset_id_idx" to table: "system_detail_assets"
CREATE INDEX "system_detail_assets_asset_id_idx" ON "system_detail_assets" ("asset_id");
-- drop index "systemdetail_owner_id" from table: "system_details"
DROP INDEX "systemdetail_owner_id";
-- create index "system_detail_owner_id_idx" to table: "system_details"
CREATE INDEX "system_detail_owner_id_idx" ON "system_details" ("owner_id");
-- drop index "tagdefinition_owner_id" from table: "tag_definitions"
DROP INDEX "tagdefinition_owner_id";
-- create index "tag_definition_owner_id_idx" to table: "tag_definitions"
CREATE INDEX "tag_definition_owner_id_idx" ON "tag_definitions" ("owner_id");
-- create index "task_evidence_evidence_id_idx" to table: "task_evidence"
CREATE INDEX "task_evidence_evidence_id_idx" ON "task_evidence" ("evidence_id");
-- drop index "task_owner_id" from table: "tasks"
DROP INDEX "task_owner_id";
-- create index "task_assignee_id_idx" to table: "tasks"
CREATE INDEX "task_assignee_id_idx" ON "tasks" ("assignee_id");
-- create index "task_assigner_id_idx" to table: "tasks"
CREATE INDEX "task_assigner_id_idx" ON "tasks" ("assigner_id");
-- create index "task_owner_id_idx" to table: "tasks"
CREATE INDEX "task_owner_id_idx" ON "tasks" ("owner_id");
-- create index "task_parent_task_id_idx" to table: "tasks"
CREATE INDEX "task_parent_task_id_idx" ON "tasks" ("parent_task_id");
-- create index "template_files_file_id_idx" to table: "template_files"
CREATE INDEX "template_files_file_id_idx" ON "template_files" ("file_id");
-- drop index "template_owner_id" from table: "templates"
DROP INDEX "template_owner_id";
-- create index "template_owner_id_idx" to table: "templates"
CREATE INDEX "template_owner_id_idx" ON "templates" ("owner_id");
-- create index "tfa_settings_owner_id_fk" to table: "tfa_settings"
CREATE INDEX "tfa_settings_owner_id_fk" ON "tfa_settings" ("owner_id");
-- create index "trust_center_compliance_trust_center_id_idx" to table: "trust_center_compliances"
CREATE INDEX "trust_center_compliance_trust_center_id_idx" ON "trust_center_compliances" ("trust_center_id");
-- create index "trust_center_doc_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_file_id_idx" ON "trust_center_docs" ("file_id");
-- create index "trust_center_doc_original_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_original_file_id_idx" ON "trust_center_docs" ("original_file_id");
-- create index "trust_center_doc_standard_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_standard_id_idx" ON "trust_center_docs" ("standard_id");
-- create index "trust_center_doc_trust_center_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_trust_center_id_idx" ON "trust_center_docs" ("trust_center_id");
-- create index "trust_center_entity_entity_type_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_entity_type_id_idx" ON "trust_center_entities" ("entity_type_id");
-- create index "trust_center_entity_logo_file_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_logo_file_id_idx" ON "trust_center_entities" ("logo_file_id");
-- create index "trust_center_entity_trust_center_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_trust_center_id_idx" ON "trust_center_entities" ("trust_center_id");
-- create index "trust_center_faq_trust_center_id_idx" to table: "trust_center_faqs"
CREATE INDEX "trust_center_faq_trust_center_id_idx" ON "trust_center_faqs" ("trust_center_id");
-- create index "trust_center_nda_request_approved_by_user_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_approved_by_user_id_idx" ON "trust_center_nda_requests" ("approved_by_user_id");
-- create index "trust_center_nda_request_document_data_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_document_data_id_idx" ON "trust_center_nda_requests" ("document_data_id");
-- create index "trust_center_nda_request_file_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_file_id_idx" ON "trust_center_nda_requests" ("file_id");
-- create index "trust_center_nda_request_trust_center_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_trust_center_id_idx" ON "trust_center_nda_requests" ("trust_center_id");
-- create index "trust_center_setting_favicon_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_favicon_local_file_id_idx" ON "trust_center_settings" ("favicon_local_file_id");
-- create index "trust_center_setting_hero_image_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_hero_image_local_file_id_idx" ON "trust_center_settings" ("hero_image_local_file_id");
-- create index "trust_center_setting_logo_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_logo_local_file_id_idx" ON "trust_center_settings" ("logo_local_file_id");
-- create index "trust_center_setting_nda_approver_group_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_nda_approver_group_id_idx" ON "trust_center_settings" ("nda_approver_group_id");
-- create index "trust_center_subprocessor_trust_center_id_idx" to table: "trust_center_subprocessors"
CREATE INDEX "trust_center_subprocessor_trust_center_id_idx" ON "trust_center_subprocessors" ("trust_center_id");
-- drop index "trustcenterwatermarkconfig_owner_id" from table: "trust_center_watermark_configs"
DROP INDEX "trustcenterwatermarkconfig_owner_id";
-- create index "trust_center_watermark_config_logo_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_logo_id_idx" ON "trust_center_watermark_configs" ("logo_id");
-- create index "trust_center_watermark_config_owner_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_owner_id_idx" ON "trust_center_watermark_configs" ("owner_id");
-- drop index "trustcenter_owner_id" from table: "trust_centers"
DROP INDEX "trustcenter_owner_id";
-- create index "trust_center_custom_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_custom_domain_id_idx" ON "trust_centers" ("custom_domain_id");
-- create index "trust_center_owner_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_owner_id_idx" ON "trust_centers" ("owner_id");
-- create index "trust_center_preview_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_preview_domain_id_idx" ON "trust_centers" ("preview_domain_id");
-- create index "user_events_event_id_idx" to table: "user_events"
CREATE INDEX "user_events_event_id_idx" ON "user_events" ("event_id");
-- create index "user_setting_user_id_idx" to table: "user_settings"
CREATE INDEX "user_setting_user_id_idx" ON "user_settings" ("user_id");
-- drop index "vendorriskscore_owner_id" from table: "vendor_risk_scores"
DROP INDEX "vendorriskscore_owner_id";
-- create index "vendor_risk_score_assessment_response_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_assessment_response_id_idx" ON "vendor_risk_scores" ("assessment_response_id");
-- create index "vendor_risk_score_entity_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_entity_id_idx" ON "vendor_risk_scores" ("entity_id");
-- create index "vendor_risk_score_owner_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_owner_id_idx" ON "vendor_risk_scores" ("owner_id");
-- create index "vendor_risk_score_vendor_scoring_config_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_vendor_scoring_config_id_idx" ON "vendor_risk_scores" ("vendor_scoring_config_id");
-- drop index "vendorscoringconfig_owner_id" from table: "vendor_scoring_configs"
DROP INDEX "vendorscoringconfig_owner_id";
-- create index "vendor_scoring_config_owner_id_idx" to table: "vendor_scoring_configs"
CREATE INDEX "vendor_scoring_config_owner_id_idx" ON "vendor_scoring_configs" ("owner_id");
-- drop index "vulnerability_owner_id" from table: "vulnerabilities"
DROP INDEX "vulnerability_owner_id";
-- create index "vulnerability_owner_id_idx" to table: "vulnerabilities"
CREATE INDEX "vulnerability_owner_id_idx" ON "vulnerabilities" ("owner_id");
-- create index "vulnerability_action_plans_action_plan_id_idx" to table: "vulnerability_action_plans"
CREATE INDEX "vulnerability_action_plans_action_plan_id_idx" ON "vulnerability_action_plans" ("action_plan_id");
-- create index "vulnerability_assets_asset_id_idx" to table: "vulnerability_assets"
CREATE INDEX "vulnerability_assets_asset_id_idx" ON "vulnerability_assets" ("asset_id");
-- create index "vulnerability_controls_control_id_idx" to table: "vulnerability_controls"
CREATE INDEX "vulnerability_controls_control_id_idx" ON "vulnerability_controls" ("control_id");
-- create index "vulnerability_entities_entity_id_idx" to table: "vulnerability_entities"
CREATE INDEX "vulnerability_entities_entity_id_idx" ON "vulnerability_entities" ("entity_id");
-- create index "vulnerability_programs_program_id_idx" to table: "vulnerability_programs"
CREATE INDEX "vulnerability_programs_program_id_idx" ON "vulnerability_programs" ("program_id");
-- create index "vulnerability_risks_risk_id_idx" to table: "vulnerability_risks"
CREATE INDEX "vulnerability_risks_risk_id_idx" ON "vulnerability_risks" ("risk_id");
-- create index "vulnerability_scans_scan_id_idx" to table: "vulnerability_scans"
CREATE INDEX "vulnerability_scans_scan_id_idx" ON "vulnerability_scans" ("scan_id");
-- create index "vulnerability_subcontrols_subcontrol_id_idx" to table: "vulnerability_subcontrols"
CREATE INDEX "vulnerability_subcontrols_subcontrol_id_idx" ON "vulnerability_subcontrols" ("subcontrol_id");
-- create index "vulnerability_tasks_task_id_idx" to table: "vulnerability_tasks"
CREATE INDEX "vulnerability_tasks_task_id_idx" ON "vulnerability_tasks" ("task_id");
-- create index "webauthns_owner_id_fk" to table: "webauthns"
CREATE INDEX "webauthns_owner_id_fk" ON "webauthns" ("owner_id");
-- drop index "workflowassignmenttarget_owner_id" from table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_owner_id";
-- create index "workflow_assignment_target_owner_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_owner_id_idx" ON "workflow_assignment_targets" ("owner_id");
-- create index "workflow_assignment_target_target_group_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_group_id_idx" ON "workflow_assignment_targets" ("target_group_id");
-- create index "workflow_assignment_target_target_user_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_user_id_idx" ON "workflow_assignment_targets" ("target_user_id");
-- drop index "workflowassignment_owner_id" from table: "workflow_assignments"
DROP INDEX "workflowassignment_owner_id";
-- create index "workflow_assignment_actor_group_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_group_id_idx" ON "workflow_assignments" ("actor_group_id");
-- create index "workflow_assignment_actor_user_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_user_id_idx" ON "workflow_assignments" ("actor_user_id");
-- create index "workflow_assignment_owner_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_owner_id_idx" ON "workflow_assignments" ("owner_id");
-- drop index "workflowdefinition_owner_id" from table: "workflow_definitions"
DROP INDEX "workflowdefinition_owner_id";
-- create index "workflow_definition_owner_id_idx" to table: "workflow_definitions"
CREATE INDEX "workflow_definition_owner_id_idx" ON "workflow_definitions" ("owner_id");
-- drop index "workflowevent_owner_id" from table: "workflow_events"
DROP INDEX "workflowevent_owner_id";
-- create index "workflow_event_owner_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_owner_id_idx" ON "workflow_events" ("owner_id");
-- create index "workflow_event_workflow_instance_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_workflow_instance_id_idx" ON "workflow_events" ("workflow_instance_id");
-- drop index "workflowinstance_owner_id" from table: "workflow_instances"
DROP INDEX "workflowinstance_owner_id";
-- create index "workflow_instance_action_plan_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_action_plan_id_idx" ON "workflow_instances" ("action_plan_id");
-- create index "workflow_instance_assessment_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_assessment_id_idx" ON "workflow_instances" ("assessment_id");
-- create index "workflow_instance_assessment_response_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_assessment_response_id_idx" ON "workflow_instances" ("assessment_response_id");
-- create index "workflow_instance_campaign_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_campaign_id_idx" ON "workflow_instances" ("campaign_id");
-- create index "workflow_instance_campaign_target_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_campaign_target_id_idx" ON "workflow_instances" ("campaign_target_id");
-- create index "workflow_instance_control_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_control_id_idx" ON "workflow_instances" ("control_id");
-- create index "workflow_instance_evidence_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_evidence_id_idx" ON "workflow_instances" ("evidence_id");
-- create index "workflow_instance_finding_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_finding_id_idx" ON "workflow_instances" ("finding_id");
-- create index "workflow_instance_identity_holder_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_identity_holder_id_idx" ON "workflow_instances" ("identity_holder_id");
-- create index "workflow_instance_integration_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_integration_id_idx" ON "workflow_instances" ("integration_id");
-- create index "workflow_instance_internal_policy_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_internal_policy_id_idx" ON "workflow_instances" ("internal_policy_id");
-- create index "workflow_instance_owner_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_owner_id_idx" ON "workflow_instances" ("owner_id");
-- create index "workflow_instance_platform_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_platform_id_idx" ON "workflow_instances" ("platform_id");
-- create index "workflow_instance_procedure_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_procedure_id_idx" ON "workflow_instances" ("procedure_id");
-- create index "workflow_instance_remediation_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_remediation_id_idx" ON "workflow_instances" ("remediation_id");
-- create index "workflow_instance_risk_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_risk_id_idx" ON "workflow_instances" ("risk_id");
-- create index "workflow_instance_subcontrol_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_subcontrol_id_idx" ON "workflow_instances" ("subcontrol_id");
-- create index "workflow_instance_task_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_task_id_idx" ON "workflow_instances" ("task_id");
-- create index "workflow_instance_vulnerability_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_vulnerability_id_idx" ON "workflow_instances" ("vulnerability_id");
-- create index "workflow_instance_workflow_proposal_id_idx" to table: "workflow_instances"
CREATE INDEX "workflow_instance_workflow_proposal_id_idx" ON "workflow_instances" ("workflow_proposal_id");
-- create index "workflow_object_ref_action_plan_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_action_plan_id_idx" ON "workflow_object_refs" ("action_plan_id");
-- create index "workflow_object_ref_assessment_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_assessment_id_idx" ON "workflow_object_refs" ("assessment_id");
-- create index "workflow_object_ref_assessment_response_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_assessment_response_id_idx" ON "workflow_object_refs" ("assessment_response_id");
-- create index "workflow_object_ref_campaign_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_campaign_id_idx" ON "workflow_object_refs" ("campaign_id");
-- create index "workflow_object_ref_campaign_target_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_campaign_target_id_idx" ON "workflow_object_refs" ("campaign_target_id");
-- create index "workflow_object_ref_control_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_control_id_idx" ON "workflow_object_refs" ("control_id");
-- create index "workflow_object_ref_directory_account_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_directory_account_id_idx" ON "workflow_object_refs" ("directory_account_id");
-- create index "workflow_object_ref_directory_group_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_directory_group_id_idx" ON "workflow_object_refs" ("directory_group_id");
-- create index "workflow_object_ref_directory_membership_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_directory_membership_id_idx" ON "workflow_object_refs" ("directory_membership_id");
-- create index "workflow_object_ref_evidence_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_evidence_id_idx" ON "workflow_object_refs" ("evidence_id");
-- create index "workflow_object_ref_finding_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_finding_id_idx" ON "workflow_object_refs" ("finding_id");
-- create index "workflow_object_ref_identity_holder_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_identity_holder_id_idx" ON "workflow_object_refs" ("identity_holder_id");
-- create index "workflow_object_ref_internal_policy_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_internal_policy_id_idx" ON "workflow_object_refs" ("internal_policy_id");
-- create index "workflow_object_ref_owner_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_owner_id_idx" ON "workflow_object_refs" ("owner_id");
-- create index "workflow_object_ref_platform_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_platform_id_idx" ON "workflow_object_refs" ("platform_id");
-- create index "workflow_object_ref_procedure_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_procedure_id_idx" ON "workflow_object_refs" ("procedure_id");
-- create index "workflow_object_ref_remediation_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_remediation_id_idx" ON "workflow_object_refs" ("remediation_id");
-- create index "workflow_object_ref_risk_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_risk_id_idx" ON "workflow_object_refs" ("risk_id");
-- create index "workflow_object_ref_subcontrol_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_subcontrol_id_idx" ON "workflow_object_refs" ("subcontrol_id");
-- create index "workflow_object_ref_task_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_task_id_idx" ON "workflow_object_refs" ("task_id");
-- create index "workflow_object_ref_vulnerability_id_idx" to table: "workflow_object_refs"
CREATE INDEX "workflow_object_ref_vulnerability_id_idx" ON "workflow_object_refs" ("vulnerability_id");
-- create index "workflow_proposal_owner_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_owner_id_idx" ON "workflow_proposals" ("owner_id");
-- create index "workflow_proposal_submitted_by_user_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_submitted_by_user_id_idx" ON "workflow_proposals" ("submitted_by_user_id");

-- +goose Down
-- reverse: create index "workflow_proposal_submitted_by_user_id_idx" to table: "workflow_proposals"
DROP INDEX "workflow_proposal_submitted_by_user_id_idx";
-- reverse: create index "workflow_proposal_owner_id_idx" to table: "workflow_proposals"
DROP INDEX "workflow_proposal_owner_id_idx";
-- reverse: create index "workflow_object_ref_vulnerability_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_vulnerability_id_idx";
-- reverse: create index "workflow_object_ref_task_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_task_id_idx";
-- reverse: create index "workflow_object_ref_subcontrol_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_subcontrol_id_idx";
-- reverse: create index "workflow_object_ref_risk_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_risk_id_idx";
-- reverse: create index "workflow_object_ref_remediation_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_remediation_id_idx";
-- reverse: create index "workflow_object_ref_procedure_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_procedure_id_idx";
-- reverse: create index "workflow_object_ref_platform_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_platform_id_idx";
-- reverse: create index "workflow_object_ref_owner_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_owner_id_idx";
-- reverse: create index "workflow_object_ref_internal_policy_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_internal_policy_id_idx";
-- reverse: create index "workflow_object_ref_identity_holder_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_identity_holder_id_idx";
-- reverse: create index "workflow_object_ref_finding_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_finding_id_idx";
-- reverse: create index "workflow_object_ref_evidence_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_evidence_id_idx";
-- reverse: create index "workflow_object_ref_directory_membership_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_directory_membership_id_idx";
-- reverse: create index "workflow_object_ref_directory_group_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_directory_group_id_idx";
-- reverse: create index "workflow_object_ref_directory_account_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_directory_account_id_idx";
-- reverse: create index "workflow_object_ref_control_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_control_id_idx";
-- reverse: create index "workflow_object_ref_campaign_target_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_campaign_target_id_idx";
-- reverse: create index "workflow_object_ref_campaign_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_campaign_id_idx";
-- reverse: create index "workflow_object_ref_assessment_response_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_assessment_response_id_idx";
-- reverse: create index "workflow_object_ref_assessment_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_assessment_id_idx";
-- reverse: create index "workflow_object_ref_action_plan_id_idx" to table: "workflow_object_refs"
DROP INDEX "workflow_object_ref_action_plan_id_idx";
-- reverse: create index "workflow_instance_workflow_proposal_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_workflow_proposal_id_idx";
-- reverse: create index "workflow_instance_vulnerability_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_vulnerability_id_idx";
-- reverse: create index "workflow_instance_task_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_task_id_idx";
-- reverse: create index "workflow_instance_subcontrol_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_subcontrol_id_idx";
-- reverse: create index "workflow_instance_risk_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_risk_id_idx";
-- reverse: create index "workflow_instance_remediation_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_remediation_id_idx";
-- reverse: create index "workflow_instance_procedure_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_procedure_id_idx";
-- reverse: create index "workflow_instance_platform_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_platform_id_idx";
-- reverse: create index "workflow_instance_owner_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_owner_id_idx";
-- reverse: create index "workflow_instance_internal_policy_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_internal_policy_id_idx";
-- reverse: create index "workflow_instance_integration_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_integration_id_idx";
-- reverse: create index "workflow_instance_identity_holder_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_identity_holder_id_idx";
-- reverse: create index "workflow_instance_finding_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_finding_id_idx";
-- reverse: create index "workflow_instance_evidence_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_evidence_id_idx";
-- reverse: create index "workflow_instance_control_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_control_id_idx";
-- reverse: create index "workflow_instance_campaign_target_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_campaign_target_id_idx";
-- reverse: create index "workflow_instance_campaign_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_campaign_id_idx";
-- reverse: create index "workflow_instance_assessment_response_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_assessment_response_id_idx";
-- reverse: create index "workflow_instance_assessment_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_assessment_id_idx";
-- reverse: create index "workflow_instance_action_plan_id_idx" to table: "workflow_instances"
DROP INDEX "workflow_instance_action_plan_id_idx";
-- reverse: drop index "workflowinstance_owner_id" from table: "workflow_instances"
CREATE INDEX "workflowinstance_owner_id" ON "workflow_instances" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "workflow_event_workflow_instance_id_idx" to table: "workflow_events"
DROP INDEX "workflow_event_workflow_instance_id_idx";
-- reverse: create index "workflow_event_owner_id_idx" to table: "workflow_events"
DROP INDEX "workflow_event_owner_id_idx";
-- reverse: drop index "workflowevent_owner_id" from table: "workflow_events"
CREATE INDEX "workflowevent_owner_id" ON "workflow_events" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "workflow_definition_owner_id_idx" to table: "workflow_definitions"
DROP INDEX "workflow_definition_owner_id_idx";
-- reverse: drop index "workflowdefinition_owner_id" from table: "workflow_definitions"
CREATE INDEX "workflowdefinition_owner_id" ON "workflow_definitions" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "workflow_assignment_owner_id_idx" to table: "workflow_assignments"
DROP INDEX "workflow_assignment_owner_id_idx";
-- reverse: create index "workflow_assignment_actor_user_id_idx" to table: "workflow_assignments"
DROP INDEX "workflow_assignment_actor_user_id_idx";
-- reverse: create index "workflow_assignment_actor_group_id_idx" to table: "workflow_assignments"
DROP INDEX "workflow_assignment_actor_group_id_idx";
-- reverse: drop index "workflowassignment_owner_id" from table: "workflow_assignments"
CREATE INDEX "workflowassignment_owner_id" ON "workflow_assignments" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "workflow_assignment_target_target_user_id_idx" to table: "workflow_assignment_targets"
DROP INDEX "workflow_assignment_target_target_user_id_idx";
-- reverse: create index "workflow_assignment_target_target_group_id_idx" to table: "workflow_assignment_targets"
DROP INDEX "workflow_assignment_target_target_group_id_idx";
-- reverse: create index "workflow_assignment_target_owner_id_idx" to table: "workflow_assignment_targets"
DROP INDEX "workflow_assignment_target_owner_id_idx";
-- reverse: drop index "workflowassignmenttarget_owner_id" from table: "workflow_assignment_targets"
CREATE INDEX "workflowassignmenttarget_owner_id" ON "workflow_assignment_targets" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "webauthns_owner_id_fk" to table: "webauthns"
DROP INDEX "webauthns_owner_id_fk";
-- reverse: create index "vulnerability_tasks_task_id_idx" to table: "vulnerability_tasks"
DROP INDEX "vulnerability_tasks_task_id_idx";
-- reverse: create index "vulnerability_subcontrols_subcontrol_id_idx" to table: "vulnerability_subcontrols"
DROP INDEX "vulnerability_subcontrols_subcontrol_id_idx";
-- reverse: create index "vulnerability_scans_scan_id_idx" to table: "vulnerability_scans"
DROP INDEX "vulnerability_scans_scan_id_idx";
-- reverse: create index "vulnerability_risks_risk_id_idx" to table: "vulnerability_risks"
DROP INDEX "vulnerability_risks_risk_id_idx";
-- reverse: create index "vulnerability_programs_program_id_idx" to table: "vulnerability_programs"
DROP INDEX "vulnerability_programs_program_id_idx";
-- reverse: create index "vulnerability_entities_entity_id_idx" to table: "vulnerability_entities"
DROP INDEX "vulnerability_entities_entity_id_idx";
-- reverse: create index "vulnerability_controls_control_id_idx" to table: "vulnerability_controls"
DROP INDEX "vulnerability_controls_control_id_idx";
-- reverse: create index "vulnerability_assets_asset_id_idx" to table: "vulnerability_assets"
DROP INDEX "vulnerability_assets_asset_id_idx";
-- reverse: create index "vulnerability_action_plans_action_plan_id_idx" to table: "vulnerability_action_plans"
DROP INDEX "vulnerability_action_plans_action_plan_id_idx";
-- reverse: create index "vulnerability_owner_id_idx" to table: "vulnerabilities"
DROP INDEX "vulnerability_owner_id_idx";
-- reverse: drop index "vulnerability_owner_id" from table: "vulnerabilities"
CREATE INDEX "vulnerability_owner_id" ON "vulnerabilities" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "vendor_scoring_config_owner_id_idx" to table: "vendor_scoring_configs"
DROP INDEX "vendor_scoring_config_owner_id_idx";
-- reverse: drop index "vendorscoringconfig_owner_id" from table: "vendor_scoring_configs"
CREATE INDEX "vendorscoringconfig_owner_id" ON "vendor_scoring_configs" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "vendor_risk_score_vendor_scoring_config_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_vendor_scoring_config_id_idx";
-- reverse: create index "vendor_risk_score_owner_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_owner_id_idx";
-- reverse: create index "vendor_risk_score_entity_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_entity_id_idx";
-- reverse: create index "vendor_risk_score_assessment_response_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_assessment_response_id_idx";
-- reverse: drop index "vendorriskscore_owner_id" from table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_owner_id" ON "vendor_risk_scores" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "user_setting_user_id_idx" to table: "user_settings"
DROP INDEX "user_setting_user_id_idx";
-- reverse: create index "user_events_event_id_idx" to table: "user_events"
DROP INDEX "user_events_event_id_idx";
-- reverse: create index "trust_center_preview_domain_id_idx" to table: "trust_centers"
DROP INDEX "trust_center_preview_domain_id_idx";
-- reverse: create index "trust_center_owner_id_idx" to table: "trust_centers"
DROP INDEX "trust_center_owner_id_idx";
-- reverse: create index "trust_center_custom_domain_id_idx" to table: "trust_centers"
DROP INDEX "trust_center_custom_domain_id_idx";
-- reverse: drop index "trustcenter_owner_id" from table: "trust_centers"
CREATE INDEX "trustcenter_owner_id" ON "trust_centers" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "trust_center_watermark_config_owner_id_idx" to table: "trust_center_watermark_configs"
DROP INDEX "trust_center_watermark_config_owner_id_idx";
-- reverse: create index "trust_center_watermark_config_logo_id_idx" to table: "trust_center_watermark_configs"
DROP INDEX "trust_center_watermark_config_logo_id_idx";
-- reverse: drop index "trustcenterwatermarkconfig_owner_id" from table: "trust_center_watermark_configs"
CREATE INDEX "trustcenterwatermarkconfig_owner_id" ON "trust_center_watermark_configs" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "trust_center_subprocessor_trust_center_id_idx" to table: "trust_center_subprocessors"
DROP INDEX "trust_center_subprocessor_trust_center_id_idx";
-- reverse: create index "trust_center_setting_nda_approver_group_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_nda_approver_group_id_idx";
-- reverse: create index "trust_center_setting_logo_local_file_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_logo_local_file_id_idx";
-- reverse: create index "trust_center_setting_hero_image_local_file_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_hero_image_local_file_id_idx";
-- reverse: create index "trust_center_setting_favicon_local_file_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_favicon_local_file_id_idx";
-- reverse: create index "trust_center_nda_request_trust_center_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_trust_center_id_idx";
-- reverse: create index "trust_center_nda_request_file_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_file_id_idx";
-- reverse: create index "trust_center_nda_request_document_data_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_document_data_id_idx";
-- reverse: create index "trust_center_nda_request_approved_by_user_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_approved_by_user_id_idx";
-- reverse: create index "trust_center_faq_trust_center_id_idx" to table: "trust_center_faqs"
DROP INDEX "trust_center_faq_trust_center_id_idx";
-- reverse: create index "trust_center_entity_trust_center_id_idx" to table: "trust_center_entities"
DROP INDEX "trust_center_entity_trust_center_id_idx";
-- reverse: create index "trust_center_entity_logo_file_id_idx" to table: "trust_center_entities"
DROP INDEX "trust_center_entity_logo_file_id_idx";
-- reverse: create index "trust_center_entity_entity_type_id_idx" to table: "trust_center_entities"
DROP INDEX "trust_center_entity_entity_type_id_idx";
-- reverse: create index "trust_center_doc_trust_center_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_trust_center_id_idx";
-- reverse: create index "trust_center_doc_standard_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_standard_id_idx";
-- reverse: create index "trust_center_doc_original_file_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_original_file_id_idx";
-- reverse: create index "trust_center_doc_file_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_file_id_idx";
-- reverse: create index "trust_center_compliance_trust_center_id_idx" to table: "trust_center_compliances"
DROP INDEX "trust_center_compliance_trust_center_id_idx";
-- reverse: create index "tfa_settings_owner_id_fk" to table: "tfa_settings"
DROP INDEX "tfa_settings_owner_id_fk";
-- reverse: create index "template_owner_id_idx" to table: "templates"
DROP INDEX "template_owner_id_idx";
-- reverse: drop index "template_owner_id" from table: "templates"
CREATE INDEX "template_owner_id" ON "templates" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "template_files_file_id_idx" to table: "template_files"
DROP INDEX "template_files_file_id_idx";
-- reverse: create index "task_parent_task_id_idx" to table: "tasks"
DROP INDEX "task_parent_task_id_idx";
-- reverse: create index "task_owner_id_idx" to table: "tasks"
DROP INDEX "task_owner_id_idx";
-- reverse: create index "task_assigner_id_idx" to table: "tasks"
DROP INDEX "task_assigner_id_idx";
-- reverse: create index "task_assignee_id_idx" to table: "tasks"
DROP INDEX "task_assignee_id_idx";
-- reverse: drop index "task_owner_id" from table: "tasks"
CREATE INDEX "task_owner_id" ON "tasks" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "task_evidence_evidence_id_idx" to table: "task_evidence"
DROP INDEX "task_evidence_evidence_id_idx";
-- reverse: create index "tag_definition_owner_id_idx" to table: "tag_definitions"
DROP INDEX "tag_definition_owner_id_idx";
-- reverse: drop index "tagdefinition_owner_id" from table: "tag_definitions"
CREATE INDEX "tagdefinition_owner_id" ON "tag_definitions" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "system_detail_owner_id_idx" to table: "system_details"
DROP INDEX "system_detail_owner_id_idx";
-- reverse: drop index "systemdetail_owner_id" from table: "system_details"
CREATE INDEX "systemdetail_owner_id" ON "system_details" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "system_detail_assets_asset_id_idx" to table: "system_detail_assets"
DROP INDEX "system_detail_assets_asset_id_idx";
-- reverse: create index "subscriber_user_id_idx" to table: "subscribers"
DROP INDEX "subscriber_user_id_idx";
-- reverse: create index "subscriber_trust_center_id_idx" to table: "subscribers"
DROP INDEX "subscriber_trust_center_id_idx";
-- reverse: create index "subscriber_owner_id_idx" to table: "subscribers"
DROP INDEX "subscriber_owner_id_idx";
-- reverse: create index "subscriber_contact_id_idx" to table: "subscribers"
DROP INDEX "subscriber_contact_id_idx";
-- reverse: drop index "subscriber_owner_id" from table: "subscribers"
CREATE INDEX "subscriber_owner_id" ON "subscribers" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "subscriber_events_event_id_idx" to table: "subscriber_events"
DROP INDEX "subscriber_events_event_id_idx";
-- reverse: create index "subprocessor_owner_id_idx" to table: "subprocessors"
DROP INDEX "subprocessor_owner_id_idx";
-- reverse: create index "subprocessor_logo_file_id_idx" to table: "subprocessors"
DROP INDEX "subprocessor_logo_file_id_idx";
-- reverse: drop index "subprocessor_owner_id" from table: "subprocessors"
CREATE INDEX "subprocessor_owner_id" ON "subprocessors" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "subcontrol_owner_id_idx" to table: "subcontrols"
DROP INDEX "subcontrol_owner_id_idx";
-- reverse: drop index "subcontrol_owner_id" from table: "subcontrols"
CREATE INDEX "subcontrol_owner_id" ON "subcontrols" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "subcontrol_tasks_task_id_idx" to table: "subcontrol_tasks"
DROP INDEX "subcontrol_tasks_task_id_idx";
-- reverse: create index "subcontrol_scans_scan_id_idx" to table: "subcontrol_scans"
DROP INDEX "subcontrol_scans_scan_id_idx";
-- reverse: create index "subcontrol_risks_risk_id_idx" to table: "subcontrol_risks"
DROP INDEX "subcontrol_risks_risk_id_idx";
-- reverse: create index "subcontrol_procedures_procedure_id_idx" to table: "subcontrol_procedures"
DROP INDEX "subcontrol_procedures_procedure_id_idx";
-- reverse: create index "subcontrol_identity_holders_identity_holder_id_idx" to table: "subcontrol_identity_holders"
DROP INDEX "subcontrol_identity_holders_identity_holder_id_idx";
-- reverse: create index "subcontrol_entities_entity_id_idx" to table: "subcontrol_entities"
DROP INDEX "subcontrol_entities_entity_id_idx";
-- reverse: create index "subcontrol_control_objectives_control_objective_id_idx" to table: "subcontrol_control_objectives"
DROP INDEX "subcontrol_control_objectives_control_objective_id_idx";
-- reverse: create index "subcontrol_control_implementations_control_implementation_id_id" to table: "subcontrol_control_implementations"
DROP INDEX "subcontrol_control_implementations_control_implementation_id_id";
-- reverse: create index "subcontrol_assets_asset_id_idx" to table: "subcontrol_assets"
DROP INDEX "subcontrol_assets_asset_id_idx";
-- reverse: create index "standard_owner_id_idx" to table: "standards"
DROP INDEX "standard_owner_id_idx";
-- reverse: create index "standard_logo_file_id_idx" to table: "standards"
DROP INDEX "standard_logo_file_id_idx";
-- reverse: drop index "standard_owner_id" from table: "standards"
CREATE INDEX "standard_owner_id" ON "standards" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "sla_definition_owner_id_idx" to table: "sla_definitions"
DROP INDEX "sla_definition_owner_id_idx";
-- reverse: drop index "sladefinition_owner_id" from table: "sla_definitions"
CREATE INDEX "sladefinition_owner_id" ON "sla_definitions" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "scheduled_job_owner_id_idx" to table: "scheduled_jobs"
DROP INDEX "scheduled_job_owner_id_idx";
-- reverse: create index "scheduled_job_job_runner_id_idx" to table: "scheduled_jobs"
DROP INDEX "scheduled_job_job_runner_id_idx";
-- reverse: create index "scheduled_job_job_id_idx" to table: "scheduled_jobs"
DROP INDEX "scheduled_job_job_id_idx";
-- reverse: drop index "scheduledjob_owner_id" from table: "scheduled_jobs"
CREATE INDEX "scheduledjob_owner_id" ON "scheduled_jobs" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "scheduled_job_subcontrols_subcontrol_id_idx" to table: "scheduled_job_subcontrols"
DROP INDEX "scheduled_job_subcontrols_subcontrol_id_idx";
-- reverse: create index "scheduled_job_run_scheduled_job_id_idx" to table: "scheduled_job_runs"
DROP INDEX "scheduled_job_run_scheduled_job_id_idx";
-- reverse: create index "scheduled_job_run_owner_id_idx" to table: "scheduled_job_runs"
DROP INDEX "scheduled_job_run_owner_id_idx";
-- reverse: create index "scheduled_job_run_job_runner_id_idx" to table: "scheduled_job_runs"
DROP INDEX "scheduled_job_run_job_runner_id_idx";
-- reverse: drop index "scheduledjobrun_owner_id" from table: "scheduled_job_runs"
CREATE INDEX "scheduledjobrun_owner_id" ON "scheduled_job_runs" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "scheduled_job_controls_control_id_idx" to table: "scheduled_job_controls"
DROP INDEX "scheduled_job_controls_control_id_idx";
-- reverse: create index "scan_performed_by_user_id_idx" to table: "scans"
DROP INDEX "scan_performed_by_user_id_idx";
-- reverse: create index "scan_performed_by_group_id_idx" to table: "scans"
DROP INDEX "scan_performed_by_group_id_idx";
-- reverse: create index "scan_owner_id_idx" to table: "scans"
DROP INDEX "scan_owner_id_idx";
-- reverse: create index "scan_generated_by_platform_id_idx" to table: "scans"
DROP INDEX "scan_generated_by_platform_id_idx";
-- reverse: drop index "scan_owner_id" from table: "scans"
CREATE INDEX "scan_owner_id" ON "scans" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "scan_tasks_task_id_idx" to table: "scan_tasks"
DROP INDEX "scan_tasks_task_id_idx";
-- reverse: create index "scan_remediations_remediation_id_idx" to table: "scan_remediations"
DROP INDEX "scan_remediations_remediation_id_idx";
-- reverse: create index "scan_files_file_id_idx" to table: "scan_files"
DROP INDEX "scan_files_file_id_idx";
-- reverse: create index "scan_evidence_evidence_id_idx" to table: "scan_evidence"
DROP INDEX "scan_evidence_evidence_id_idx";
-- reverse: create index "scan_entities_entity_id_idx" to table: "scan_entities"
DROP INDEX "scan_entities_entity_id_idx";
-- reverse: create index "scan_editors_group_id_idx" to table: "scan_editors"
DROP INDEX "scan_editors_group_id_idx";
-- reverse: create index "scan_blocked_groups_group_id_idx" to table: "scan_blocked_groups"
DROP INDEX "scan_blocked_groups_group_id_idx";
-- reverse: create index "scan_assets_asset_id_idx" to table: "scan_assets"
DROP INDEX "scan_assets_asset_id_idx";
-- reverse: create index "scan_action_plans_action_plan_id_idx" to table: "scan_action_plans"
DROP INDEX "scan_action_plans_action_plan_id_idx";
-- reverse: create index "risk_stakeholder_id_idx" to table: "risks"
DROP INDEX "risk_stakeholder_id_idx";
-- reverse: create index "risk_owner_id_idx" to table: "risks"
DROP INDEX "risk_owner_id_idx";
-- reverse: create index "risk_delegate_id_idx" to table: "risks"
DROP INDEX "risk_delegate_id_idx";
-- reverse: drop index "risk_owner_id" from table: "risks"
CREATE INDEX "risk_owner_id" ON "risks" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "risk_viewers_group_id_idx" to table: "risk_viewers"
DROP INDEX "risk_viewers_group_id_idx";
-- reverse: create index "risk_tasks_task_id_idx" to table: "risk_tasks"
DROP INDEX "risk_tasks_task_id_idx";
-- reverse: create index "risk_editors_group_id_idx" to table: "risk_editors"
DROP INDEX "risk_editors_group_id_idx";
-- reverse: create index "risk_blocked_groups_group_id_idx" to table: "risk_blocked_groups"
DROP INDEX "risk_blocked_groups_group_id_idx";
-- reverse: create index "risk_action_plans_action_plan_id_idx" to table: "risk_action_plans"
DROP INDEX "risk_action_plans_action_plan_id_idx";
-- reverse: create index "review_reviewer_id_idx" to table: "reviews"
DROP INDEX "review_reviewer_id_idx";
-- reverse: create index "review_owner_id_idx" to table: "reviews"
DROP INDEX "review_owner_id_idx";
-- reverse: drop index "review_owner_id" from table: "reviews"
CREATE INDEX "review_owner_id" ON "reviews" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "review_vulnerabilities_vulnerability_id_idx" to table: "review_vulnerabilities"
DROP INDEX "review_vulnerabilities_vulnerability_id_idx";
-- reverse: create index "review_subcontrols_subcontrol_id_idx" to table: "review_subcontrols"
DROP INDEX "review_subcontrols_subcontrol_id_idx";
-- reverse: create index "review_risks_risk_id_idx" to table: "review_risks"
DROP INDEX "review_risks_risk_id_idx";
-- reverse: create index "review_remediations_remediation_id_idx" to table: "review_remediations"
DROP INDEX "review_remediations_remediation_id_idx";
-- reverse: create index "review_programs_program_id_idx" to table: "review_programs"
DROP INDEX "review_programs_program_id_idx";
-- reverse: create index "review_internal_policies_internal_policy_id_idx" to table: "review_internal_policies"
DROP INDEX "review_internal_policies_internal_policy_id_idx";
-- reverse: create index "review_findings_finding_id_idx" to table: "review_findings"
DROP INDEX "review_findings_finding_id_idx";
-- reverse: create index "review_entities_entity_id_idx" to table: "review_entities"
DROP INDEX "review_entities_entity_id_idx";
-- reverse: create index "review_editors_group_id_idx" to table: "review_editors"
DROP INDEX "review_editors_group_id_idx";
-- reverse: create index "review_controls_control_id_idx" to table: "review_controls"
DROP INDEX "review_controls_control_id_idx";
-- reverse: create index "review_blocked_groups_group_id_idx" to table: "review_blocked_groups"
DROP INDEX "review_blocked_groups_group_id_idx";
-- reverse: create index "review_assets_asset_id_idx" to table: "review_assets"
DROP INDEX "review_assets_asset_id_idx";
-- reverse: create index "review_action_plans_action_plan_id_idx" to table: "review_action_plans"
DROP INDEX "review_action_plans_action_plan_id_idx";
-- reverse: create index "remediation_owner_id_idx" to table: "remediations"
DROP INDEX "remediation_owner_id_idx";
-- reverse: drop index "remediation_owner_id" from table: "remediations"
CREATE INDEX "remediation_owner_id" ON "remediations" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "remediation_vulnerabilities_vulnerability_id_idx" to table: "remediation_vulnerabilities"
DROP INDEX "remediation_vulnerabilities_vulnerability_id_idx";
-- reverse: create index "remediation_subcontrols_subcontrol_id_idx" to table: "remediation_subcontrols"
DROP INDEX "remediation_subcontrols_subcontrol_id_idx";
-- reverse: create index "remediation_risks_risk_id_idx" to table: "remediation_risks"
DROP INDEX "remediation_risks_risk_id_idx";
-- reverse: create index "remediation_programs_program_id_idx" to table: "remediation_programs"
DROP INDEX "remediation_programs_program_id_idx";
-- reverse: create index "remediation_findings_finding_id_idx" to table: "remediation_findings"
DROP INDEX "remediation_findings_finding_id_idx";
-- reverse: create index "remediation_entities_entity_id_idx" to table: "remediation_entities"
DROP INDEX "remediation_entities_entity_id_idx";
-- reverse: create index "remediation_editors_group_id_idx" to table: "remediation_editors"
DROP INDEX "remediation_editors_group_id_idx";
-- reverse: create index "remediation_controls_control_id_idx" to table: "remediation_controls"
DROP INDEX "remediation_controls_control_id_idx";
-- reverse: create index "remediation_blocked_groups_group_id_idx" to table: "remediation_blocked_groups"
DROP INDEX "remediation_blocked_groups_group_id_idx";
-- reverse: create index "remediation_assets_asset_id_idx" to table: "remediation_assets"
DROP INDEX "remediation_assets_asset_id_idx";
-- reverse: create index "remediation_action_plans_action_plan_id_idx" to table: "remediation_action_plans"
DROP INDEX "remediation_action_plans_action_plan_id_idx";
-- reverse: create index "program_program_owner_id_idx" to table: "programs"
DROP INDEX "program_program_owner_id_idx";
-- reverse: create index "program_owner_id_idx" to table: "programs"
DROP INDEX "program_owner_id_idx";
-- reverse: drop index "program_owner_id" from table: "programs"
CREATE INDEX "program_owner_id" ON "programs" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "program_viewers_group_id_idx" to table: "program_viewers"
DROP INDEX "program_viewers_group_id_idx";
-- reverse: create index "program_tasks_task_id_idx" to table: "program_tasks"
DROP INDEX "program_tasks_task_id_idx";
-- reverse: create index "program_system_details_system_detail_id_idx" to table: "program_system_details"
DROP INDEX "program_system_details_system_detail_id_idx";
-- reverse: create index "program_risks_risk_id_idx" to table: "program_risks"
DROP INDEX "program_risks_risk_id_idx";
-- reverse: create index "program_procedures_procedure_id_idx" to table: "program_procedures"
DROP INDEX "program_procedures_procedure_id_idx";
-- reverse: create index "program_narratives_narrative_id_idx" to table: "program_narratives"
DROP INDEX "program_narratives_narrative_id_idx";
-- reverse: create index "program_membership_program_id_idx" to table: "program_memberships"
DROP INDEX "program_membership_program_id_idx";
-- reverse: create index "program_internal_policies_internal_policy_id_idx" to table: "program_internal_policies"
DROP INDEX "program_internal_policies_internal_policy_id_idx";
-- reverse: create index "program_files_file_id_idx" to table: "program_files"
DROP INDEX "program_files_file_id_idx";
-- reverse: create index "program_evidence_evidence_id_idx" to table: "program_evidence"
DROP INDEX "program_evidence_evidence_id_idx";
-- reverse: create index "program_editors_group_id_idx" to table: "program_editors"
DROP INDEX "program_editors_group_id_idx";
-- reverse: create index "program_controls_control_id_idx" to table: "program_controls"
DROP INDEX "program_controls_control_id_idx";
-- reverse: create index "program_control_objectives_control_objective_id_idx" to table: "program_control_objectives"
DROP INDEX "program_control_objectives_control_objective_id_idx";
-- reverse: create index "program_blocked_groups_group_id_idx" to table: "program_blocked_groups"
DROP INDEX "program_blocked_groups_group_id_idx";
-- reverse: create index "program_action_plans_action_plan_id_idx" to table: "program_action_plans"
DROP INDEX "program_action_plans_action_plan_id_idx";
-- reverse: create index "procedure_owner_id_idx" to table: "procedures"
DROP INDEX "procedure_owner_id_idx";
-- reverse: create index "procedure_file_id_idx" to table: "procedures"
DROP INDEX "procedure_file_id_idx";
-- reverse: drop index "procedure_owner_id" from table: "procedures"
CREATE INDEX "procedure_owner_id" ON "procedures" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "procedure_tasks_task_id_idx" to table: "procedure_tasks"
DROP INDEX "procedure_tasks_task_id_idx";
-- reverse: create index "procedure_risks_risk_id_idx" to table: "procedure_risks"
DROP INDEX "procedure_risks_risk_id_idx";
-- reverse: create index "procedure_narratives_narrative_id_idx" to table: "procedure_narratives"
DROP INDEX "procedure_narratives_narrative_id_idx";
-- reverse: create index "procedure_editors_group_id_idx" to table: "procedure_editors"
DROP INDEX "procedure_editors_group_id_idx";
-- reverse: create index "procedure_blocked_groups_group_id_idx" to table: "procedure_blocked_groups"
DROP INDEX "procedure_blocked_groups_group_id_idx";
-- reverse: create index "platform_platform_owner_id_idx" to table: "platforms"
DROP INDEX "platform_platform_owner_id_idx";
-- reverse: create index "platform_owner_id_idx" to table: "platforms"
DROP INDEX "platform_owner_id_idx";
-- reverse: drop index "platform_owner_id" from table: "platforms"
CREATE INDEX "platform_owner_id" ON "platforms" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "platform_viewers_group_id_idx" to table: "platform_viewers"
DROP INDEX "platform_viewers_group_id_idx";
-- reverse: create index "platform_tasks_task_id_idx" to table: "platform_tasks"
DROP INDEX "platform_tasks_task_id_idx";
-- reverse: create index "platform_system_details_system_detail_id_idx" to table: "platform_system_details"
DROP INDEX "platform_system_details_system_detail_id_idx";
-- reverse: create index "platform_source_entities_entity_id_idx" to table: "platform_source_entities"
DROP INDEX "platform_source_entities_entity_id_idx";
-- reverse: create index "platform_scans_scan_id_idx" to table: "platform_scans"
DROP INDEX "platform_scans_scan_id_idx";
-- reverse: create index "platform_risks_risk_id_idx" to table: "platform_risks"
DROP INDEX "platform_risks_risk_id_idx";
-- reverse: create index "platform_out_of_scope_vendors_entity_id_idx" to table: "platform_out_of_scope_vendors"
DROP INDEX "platform_out_of_scope_vendors_entity_id_idx";
-- reverse: create index "platform_out_of_scope_assets_asset_id_idx" to table: "platform_out_of_scope_assets"
DROP INDEX "platform_out_of_scope_assets_asset_id_idx";
-- reverse: create index "platform_identity_holders_identity_holder_id_idx" to table: "platform_identity_holders"
DROP INDEX "platform_identity_holders_identity_holder_id_idx";
-- reverse: create index "platform_files_file_id_idx" to table: "platform_files"
DROP INDEX "platform_files_file_id_idx";
-- reverse: create index "platform_evidence_evidence_id_idx" to table: "platform_evidence"
DROP INDEX "platform_evidence_evidence_id_idx";
-- reverse: create index "platform_entities_entity_id_idx" to table: "platform_entities"
DROP INDEX "platform_entities_entity_id_idx";
-- reverse: create index "platform_editors_group_id_idx" to table: "platform_editors"
DROP INDEX "platform_editors_group_id_idx";
-- reverse: create index "platform_controls_control_id_idx" to table: "platform_controls"
DROP INDEX "platform_controls_control_id_idx";
-- reverse: create index "platform_blocked_groups_group_id_idx" to table: "platform_blocked_groups"
DROP INDEX "platform_blocked_groups_group_id_idx";
-- reverse: create index "platform_assets_asset_id_idx" to table: "platform_assets"
DROP INDEX "platform_assets_asset_id_idx";
-- reverse: create index "platform_assessments_assessment_id_idx" to table: "platform_assessments"
DROP INDEX "platform_assessments_assessment_id_idx";
-- reverse: create index "platform_applicable_frameworks_standard_id_idx" to table: "platform_applicable_frameworks"
DROP INDEX "platform_applicable_frameworks_standard_id_idx";
-- reverse: create index "personal_access_tokens_owner_id_fk" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_owner_id_fk";
-- reverse: create index "personal_access_token_events_event_id_idx" to table: "personal_access_token_events"
DROP INDEX "personal_access_token_events_event_id_idx";
-- reverse: create index "password_reset_tokens_owner_id_fk" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_owner_id_fk";
-- reverse: create index "organization_parent_organization_id_idx" to table: "organizations"
DROP INDEX "organization_parent_organization_id_idx";
-- reverse: create index "organization_avatar_local_file_id_idx" to table: "organizations"
DROP INDEX "organization_avatar_local_file_id_idx";
-- reverse: create index "organization_setting_organization_id_idx" to table: "organization_settings"
DROP INDEX "organization_setting_organization_id_idx";
-- reverse: create index "organization_setting_files_file_id_idx" to table: "organization_setting_files"
DROP INDEX "organization_setting_files_file_id_idx";
-- reverse: create index "organization_personal_access_tokens_personal_access_token_id_id" to table: "organization_personal_access_tokens"
DROP INDEX "organization_personal_access_tokens_personal_access_token_id_id";
-- reverse: create index "organization_files_file_id_idx" to table: "organization_files"
DROP INDEX "organization_files_file_id_idx";
-- reverse: create index "organization_events_event_id_idx" to table: "organization_events"
DROP INDEX "organization_events_event_id_idx";
-- reverse: create index "org_subscription_owner_id_idx" to table: "org_subscriptions"
DROP INDEX "org_subscription_owner_id_idx";
-- reverse: drop index "orgsubscription_owner_id" from table: "org_subscriptions"
CREATE INDEX "orgsubscription_owner_id" ON "org_subscriptions" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "org_subscription_events_event_id_idx" to table: "org_subscription_events"
DROP INDEX "org_subscription_events_event_id_idx";
-- reverse: create index "org_product_subscription_id_idx" to table: "org_products"
DROP INDEX "org_product_subscription_id_idx";
-- reverse: create index "org_product_owner_id_idx" to table: "org_products"
DROP INDEX "org_product_owner_id_idx";
-- reverse: drop index "orgproduct_owner_id" from table: "org_products"
CREATE INDEX "orgproduct_owner_id" ON "org_products" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "org_product_org_prices_org_price_id_idx" to table: "org_product_org_prices"
DROP INDEX "org_product_org_prices_org_price_id_idx";
-- reverse: create index "org_price_subscription_id_idx" to table: "org_prices"
DROP INDEX "org_price_subscription_id_idx";
-- reverse: create index "org_price_owner_id_idx" to table: "org_prices"
DROP INDEX "org_price_owner_id_idx";
-- reverse: drop index "orgprice_owner_id" from table: "org_prices"
CREATE INDEX "orgprice_owner_id" ON "org_prices" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "org_module_subscription_id_idx" to table: "org_modules"
DROP INDEX "org_module_subscription_id_idx";
-- reverse: create index "org_module_owner_id_idx" to table: "org_modules"
DROP INDEX "org_module_owner_id_idx";
-- reverse: drop index "orgmodule_owner_id" from table: "org_modules"
CREATE INDEX "orgmodule_owner_id" ON "org_modules" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "org_module_org_prices_org_price_id_idx" to table: "org_module_org_prices"
DROP INDEX "org_module_org_prices_org_price_id_idx";
-- reverse: create index "org_membership_organization_id_idx" to table: "org_memberships"
DROP INDEX "org_membership_organization_id_idx";
-- reverse: create index "org_membership_events_event_id_idx" to table: "org_membership_events"
DROP INDEX "org_membership_events_event_id_idx";
-- reverse: create index "notification_template_id_idx" to table: "notifications"
DROP INDEX "notification_template_id_idx";
-- reverse: create index "notification_owner_id_idx" to table: "notifications"
DROP INDEX "notification_owner_id_idx";
-- reverse: create index "notification_template_workflow_definition_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_workflow_definition_id_idx";
-- reverse: create index "notification_template_owner_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_owner_id_idx";
-- reverse: create index "notification_template_integration_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_integration_id_idx";
-- reverse: create index "notification_template_email_template_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_email_template_id_idx";
-- reverse: drop index "notificationtemplate_owner_id" from table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id" ON "notification_templates" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "notification_preference_user_id_idx" to table: "notification_preferences"
DROP INDEX "notification_preference_user_id_idx";
-- reverse: create index "notification_preference_template_id_idx" to table: "notification_preferences"
DROP INDEX "notification_preference_template_id_idx";
-- reverse: create index "notification_preference_owner_id_idx" to table: "notification_preferences"
DROP INDEX "notification_preference_owner_id_idx";
-- reverse: drop index "notificationpreference_owner_id" from table: "notification_preferences"
CREATE INDEX "notificationpreference_owner_id" ON "notification_preferences" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "note_trust_center_id_idx" to table: "notes"
DROP INDEX "note_trust_center_id_idx";
-- reverse: create index "note_owner_id_idx" to table: "notes"
DROP INDEX "note_owner_id_idx";
-- reverse: create index "note_discussion_id_idx" to table: "notes"
DROP INDEX "note_discussion_id_idx";
-- reverse: drop index "note_owner_id" from table: "notes"
CREATE INDEX "note_owner_id" ON "notes" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "narrative_owner_id_idx" to table: "narratives"
DROP INDEX "narrative_owner_id_idx";
-- reverse: drop index "narrative_owner_id" from table: "narratives"
CREATE INDEX "narrative_owner_id" ON "narratives" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "narrative_viewers_group_id_idx" to table: "narrative_viewers"
DROP INDEX "narrative_viewers_group_id_idx";
-- reverse: create index "narrative_editors_group_id_idx" to table: "narrative_editors"
DROP INDEX "narrative_editors_group_id_idx";
-- reverse: create index "narrative_blocked_groups_group_id_idx" to table: "narrative_blocked_groups"
DROP INDEX "narrative_blocked_groups_group_id_idx";
-- reverse: create index "mapped_control_owner_id_idx" to table: "mapped_controls"
DROP INDEX "mapped_control_owner_id_idx";
-- reverse: drop index "mappedcontrol_owner_id" from table: "mapped_controls"
CREATE INDEX "mappedcontrol_owner_id" ON "mapped_controls" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "mapped_control_to_subcontrols_subcontrol_id_idx" to table: "mapped_control_to_subcontrols"
DROP INDEX "mapped_control_to_subcontrols_subcontrol_id_idx";
-- reverse: create index "mapped_control_to_controls_control_id_idx" to table: "mapped_control_to_controls"
DROP INDEX "mapped_control_to_controls_control_id_idx";
-- reverse: create index "mapped_control_from_subcontrols_subcontrol_id_idx" to table: "mapped_control_from_subcontrols"
DROP INDEX "mapped_control_from_subcontrols_subcontrol_id_idx";
-- reverse: create index "mapped_control_from_controls_control_id_idx" to table: "mapped_control_from_controls"
DROP INDEX "mapped_control_from_controls_control_id_idx";
-- reverse: create index "mapped_control_editors_group_id_idx" to table: "mapped_control_editors"
DROP INDEX "mapped_control_editors_group_id_idx";
-- reverse: create index "mapped_control_blocked_groups_group_id_idx" to table: "mapped_control_blocked_groups"
DROP INDEX "mapped_control_blocked_groups_group_id_idx";
-- reverse: create index "job_template_owner_id_idx" to table: "job_templates"
DROP INDEX "job_template_owner_id_idx";
-- reverse: drop index "jobtemplate_owner_id" from table: "job_templates"
CREATE INDEX "jobtemplate_owner_id" ON "job_templates" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "job_runner_owner_id_idx" to table: "job_runners"
DROP INDEX "job_runner_owner_id_idx";
-- reverse: drop index "jobrunner_owner_id" from table: "job_runners"
CREATE INDEX "jobrunner_owner_id" ON "job_runners" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "job_runner_token_owner_id_idx" to table: "job_runner_tokens"
DROP INDEX "job_runner_token_owner_id_idx";
-- reverse: drop index "jobrunnertoken_owner_id" from table: "job_runner_tokens"
CREATE INDEX "jobrunnertoken_owner_id" ON "job_runner_tokens" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "job_runner_registration_token_owner_id_idx" to table: "job_runner_registration_tokens"
DROP INDEX "job_runner_registration_token_owner_id_idx";
-- reverse: create index "job_runner_registration_token_job_runner_id_idx" to table: "job_runner_registration_tokens"
DROP INDEX "job_runner_registration_token_job_runner_id_idx";
-- reverse: drop index "jobrunnerregistrationtoken_owner_id" from table: "job_runner_registration_tokens"
CREATE INDEX "jobrunnerregistrationtoken_owner_id" ON "job_runner_registration_tokens" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "job_runner_job_runner_tokens_job_runner_token_id_idx" to table: "job_runner_job_runner_tokens"
DROP INDEX "job_runner_job_runner_tokens_job_runner_token_id_idx";
-- reverse: create index "job_result_scheduled_job_id_idx" to table: "job_results"
DROP INDEX "job_result_scheduled_job_id_idx";
-- reverse: create index "job_result_owner_id_idx" to table: "job_results"
DROP INDEX "job_result_owner_id_idx";
-- reverse: create index "job_result_file_id_idx" to table: "job_results"
DROP INDEX "job_result_file_id_idx";
-- reverse: drop index "jobresult_owner_id" from table: "job_results"
CREATE INDEX "jobresult_owner_id" ON "job_results" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "invite_owner_id_idx" to table: "invites"
DROP INDEX "invite_owner_id_idx";
-- reverse: drop index "invite_owner_id" from table: "invites"
CREATE INDEX "invite_owner_id" ON "invites" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "invite_groups_group_id_idx" to table: "invite_groups"
DROP INDEX "invite_groups_group_id_idx";
-- reverse: create index "invite_events_event_id_idx" to table: "invite_events"
DROP INDEX "invite_events_event_id_idx";
-- reverse: create index "internal_policy_tasks_task_id_idx" to table: "internal_policy_tasks"
DROP INDEX "internal_policy_tasks_task_id_idx";
-- reverse: create index "internal_policy_subcontrols_subcontrol_id_idx" to table: "internal_policy_subcontrols"
DROP INDEX "internal_policy_subcontrols_subcontrol_id_idx";
-- reverse: create index "internal_policy_risks_risk_id_idx" to table: "internal_policy_risks"
DROP INDEX "internal_policy_risks_risk_id_idx";
-- reverse: create index "internal_policy_procedures_procedure_id_idx" to table: "internal_policy_procedures"
DROP INDEX "internal_policy_procedures_procedure_id_idx";
-- reverse: create index "internal_policy_narratives_narrative_id_idx" to table: "internal_policy_narratives"
DROP INDEX "internal_policy_narratives_narrative_id_idx";
-- reverse: create index "internal_policy_identity_holders_identity_holder_id_idx" to table: "internal_policy_identity_holders"
DROP INDEX "internal_policy_identity_holders_identity_holder_id_idx";
-- reverse: create index "internal_policy_entities_entity_id_idx" to table: "internal_policy_entities"
DROP INDEX "internal_policy_entities_entity_id_idx";
-- reverse: create index "internal_policy_editors_group_id_idx" to table: "internal_policy_editors"
DROP INDEX "internal_policy_editors_group_id_idx";
-- reverse: create index "internal_policy_controls_control_id_idx" to table: "internal_policy_controls"
DROP INDEX "internal_policy_controls_control_id_idx";
-- reverse: create index "internal_policy_control_objectives_control_objective_id_idx" to table: "internal_policy_control_objectives"
DROP INDEX "internal_policy_control_objectives_control_objective_id_idx";
-- reverse: create index "internal_policy_blocked_groups_group_id_idx" to table: "internal_policy_blocked_groups"
DROP INDEX "internal_policy_blocked_groups_group_id_idx";
-- reverse: create index "internal_policy_assets_asset_id_idx" to table: "internal_policy_assets"
DROP INDEX "internal_policy_assets_asset_id_idx";
-- reverse: create index "internal_policy_owner_id_idx" to table: "internal_policies"
DROP INDEX "internal_policy_owner_id_idx";
-- reverse: create index "internal_policy_file_id_idx" to table: "internal_policies"
DROP INDEX "internal_policy_file_id_idx";
-- reverse: drop index "internalpolicy_owner_id" from table: "internal_policies"
CREATE INDEX "internalpolicy_owner_id" ON "internal_policies" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "integration_platform_id_idx" to table: "integrations"
DROP INDEX "integration_platform_id_idx";
-- reverse: create index "integration_owner_id_idx" to table: "integrations"
DROP INDEX "integration_owner_id_idx";
-- reverse: drop index "integration_owner_id" from table: "integrations"
CREATE INDEX "integration_owner_id" ON "integrations" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "integration_webhook_owner_id_idx" to table: "integration_webhooks"
DROP INDEX "integration_webhook_owner_id_idx";
-- reverse: drop index "integrationwebhook_owner_id" from table: "integration_webhooks"
CREATE INDEX "integrationwebhook_owner_id" ON "integration_webhooks" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "integration_vulnerabilities_vulnerability_id_idx" to table: "integration_vulnerabilities"
DROP INDEX "integration_vulnerabilities_vulnerability_id_idx";
-- reverse: create index "integration_secrets_hush_id_idx" to table: "integration_secrets"
DROP INDEX "integration_secrets_hush_id_idx";
-- reverse: create index "integration_run_response_file_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_response_file_id_idx";
-- reverse: create index "integration_run_request_file_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_request_file_id_idx";
-- reverse: create index "integration_run_owner_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_owner_id_idx";
-- reverse: create index "integration_run_event_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_event_id_idx";
-- reverse: drop index "integrationrun_owner_id" from table: "integration_runs"
CREATE INDEX "integrationrun_owner_id" ON "integration_runs" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "integration_reviews_review_id_idx" to table: "integration_reviews"
DROP INDEX "integration_reviews_review_id_idx";
-- reverse: create index "integration_remediations_remediation_id_idx" to table: "integration_remediations"
DROP INDEX "integration_remediations_remediation_id_idx";
-- reverse: create index "integration_internal_policies_internal_policy_id_idx" to table: "integration_internal_policies"
DROP INDEX "integration_internal_policies_internal_policy_id_idx";
-- reverse: create index "integration_findings_finding_id_idx" to table: "integration_findings"
DROP INDEX "integration_findings_finding_id_idx";
-- reverse: create index "integration_events_event_id_idx" to table: "integration_events"
DROP INDEX "integration_events_event_id_idx";
-- reverse: create index "integration_action_plans_action_plan_id_idx" to table: "integration_action_plans"
DROP INDEX "integration_action_plans_action_plan_id_idx";
-- reverse: create index "impersonation_event_user_id_idx" to table: "impersonation_events"
DROP INDEX "impersonation_event_user_id_idx";
-- reverse: create index "impersonation_event_target_user_id_idx" to table: "impersonation_events"
DROP INDEX "impersonation_event_target_user_id_idx";
-- reverse: create index "impersonation_event_organization_id_idx" to table: "impersonation_events"
DROP INDEX "impersonation_event_organization_id_idx";
-- reverse: create index "identity_holder_owner_id_idx" to table: "identity_holders"
DROP INDEX "identity_holder_owner_id_idx";
-- reverse: create index "identity_holder_employer_entity_id_idx" to table: "identity_holders"
DROP INDEX "identity_holder_employer_entity_id_idx";
-- reverse: drop index "identityholder_owner_id" from table: "identity_holders"
CREATE INDEX "identityholder_owner_id" ON "identity_holders" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "identity_holder_templates_template_id_idx" to table: "identity_holder_templates"
DROP INDEX "identity_holder_templates_template_id_idx";
-- reverse: create index "identity_holder_tasks_task_id_idx" to table: "identity_holder_tasks"
DROP INDEX "identity_holder_tasks_task_id_idx";
-- reverse: create index "identity_holder_files_file_id_idx" to table: "identity_holder_files"
DROP INDEX "identity_holder_files_file_id_idx";
-- reverse: create index "identity_holder_entities_entity_id_idx" to table: "identity_holder_entities"
DROP INDEX "identity_holder_entities_entity_id_idx";
-- reverse: create index "identity_holder_assets_asset_id_idx" to table: "identity_holder_assets"
DROP INDEX "identity_holder_assets_asset_id_idx";
-- reverse: create index "identity_holder_assessments_assessment_id_idx" to table: "identity_holder_assessments"
DROP INDEX "identity_holder_assessments_assessment_id_idx";
-- reverse: create index "secret_owner_id_idx" to table: "hushes"
DROP INDEX "secret_owner_id_idx";
-- reverse: drop index "hush_owner_id" from table: "hushes"
CREATE INDEX "hush_owner_id" ON "hushes" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "hush_events_event_id_idx" to table: "hush_events"
DROP INDEX "hush_events_event_id_idx";
-- reverse: create index "group_owner_id_idx" to table: "groups"
DROP INDEX "group_owner_id_idx";
-- reverse: create index "group_avatar_local_file_id_idx" to table: "groups"
DROP INDEX "group_avatar_local_file_id_idx";
-- reverse: drop index "group_owner_id" from table: "groups"
CREATE INDEX "group_owner_id" ON "groups" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "group_tasks_task_id_idx" to table: "group_tasks"
DROP INDEX "group_tasks_task_id_idx";
-- reverse: create index "group_setting_group_id_idx" to table: "group_settings"
DROP INDEX "group_setting_group_id_idx";
-- reverse: create index "group_membership_group_id_idx" to table: "group_memberships"
DROP INDEX "group_membership_group_id_idx";
-- reverse: create index "group_membership_events_event_id_idx" to table: "group_membership_events"
DROP INDEX "group_membership_events_event_id_idx";
-- reverse: create index "group_files_file_id_idx" to table: "group_files"
DROP INDEX "group_files_file_id_idx";
-- reverse: create index "group_events_event_id_idx" to table: "group_events"
DROP INDEX "group_events_event_id_idx";
-- reverse: create index "finding_owner_id_idx" to table: "findings"
DROP INDEX "finding_owner_id_idx";
-- reverse: drop index "finding_owner_id" from table: "findings"
CREATE INDEX "finding_owner_id" ON "findings" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "finding_vulnerabilities_vulnerability_id_idx" to table: "finding_vulnerabilities"
DROP INDEX "finding_vulnerabilities_vulnerability_id_idx";
-- reverse: create index "finding_tasks_task_id_idx" to table: "finding_tasks"
DROP INDEX "finding_tasks_task_id_idx";
-- reverse: create index "finding_subcontrols_subcontrol_id_idx" to table: "finding_subcontrols"
DROP INDEX "finding_subcontrols_subcontrol_id_idx";
-- reverse: create index "finding_scans_scan_id_idx" to table: "finding_scans"
DROP INDEX "finding_scans_scan_id_idx";
-- reverse: create index "finding_risks_risk_id_idx" to table: "finding_risks"
DROP INDEX "finding_risks_risk_id_idx";
-- reverse: create index "finding_programs_program_id_idx" to table: "finding_programs"
DROP INDEX "finding_programs_program_id_idx";
-- reverse: create index "finding_identity_holders_identity_holder_id_idx" to table: "finding_identity_holders"
DROP INDEX "finding_identity_holders_identity_holder_id_idx";
-- reverse: create index "finding_entities_entity_id_idx" to table: "finding_entities"
DROP INDEX "finding_entities_entity_id_idx";
-- reverse: create index "finding_editors_group_id_idx" to table: "finding_editors"
DROP INDEX "finding_editors_group_id_idx";
-- reverse: create index "finding_directory_accounts_directory_account_id_idx" to table: "finding_directory_accounts"
DROP INDEX "finding_directory_accounts_directory_account_id_idx";
-- reverse: create index "finding_control_standard_id_idx" to table: "finding_controls"
DROP INDEX "finding_control_standard_id_idx";
-- reverse: create index "finding_control_owner_id_idx" to table: "finding_controls"
DROP INDEX "finding_control_owner_id_idx";
-- reverse: create index "finding_control_control_id_idx" to table: "finding_controls"
DROP INDEX "finding_control_control_id_idx";
-- reverse: create index "finding_check_results_check_result_id_idx" to table: "finding_check_results"
DROP INDEX "finding_check_results_check_result_id_idx";
-- reverse: create index "finding_blocked_groups_group_id_idx" to table: "finding_blocked_groups"
DROP INDEX "finding_blocked_groups_group_id_idx";
-- reverse: create index "finding_assets_asset_id_idx" to table: "finding_assets"
DROP INDEX "finding_assets_asset_id_idx";
-- reverse: create index "finding_action_plans_action_plan_id_idx" to table: "finding_action_plans"
DROP INDEX "finding_action_plans_action_plan_id_idx";
-- reverse: create index "file_secrets_hush_id_idx" to table: "file_secrets"
DROP INDEX "file_secrets_hush_id_idx";
-- reverse: create index "file_events_event_id_idx" to table: "file_events"
DROP INDEX "file_events_event_id_idx";
-- reverse: create index "file_download_tokens_owner_id_fk" to table: "file_download_tokens"
DROP INDEX "file_download_tokens_owner_id_fk";
-- reverse: create index "export_owner_id_idx" to table: "exports"
DROP INDEX "export_owner_id_idx";
-- reverse: drop index "export_owner_id" from table: "exports"
CREATE INDEX "export_owner_id" ON "exports" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "evidence_owner_id_idx" to table: "evidences"
DROP INDEX "evidence_owner_id_idx";
-- reverse: drop index "evidence_owner_id" from table: "evidences"
CREATE INDEX "evidence_owner_id" ON "evidences" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "evidence_subcontrols_subcontrol_id_idx" to table: "evidence_subcontrols"
DROP INDEX "evidence_subcontrols_subcontrol_id_idx";
-- reverse: create index "evidence_files_file_id_idx" to table: "evidence_files"
DROP INDEX "evidence_files_file_id_idx";
-- reverse: create index "evidence_controls_control_id_idx" to table: "evidence_controls"
DROP INDEX "evidence_controls_control_id_idx";
-- reverse: create index "evidence_control_objectives_control_objective_id_idx" to table: "evidence_control_objectives"
DROP INDEX "evidence_control_objectives_control_objective_id_idx";
-- reverse: create index "entity_type_owner_id_idx" to table: "entity_types"
DROP INDEX "entity_type_owner_id_idx";
-- reverse: drop index "entitytype_owner_id" from table: "entity_types"
CREATE INDEX "entitytype_owner_id" ON "entity_types" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "entity_system_details_system_detail_id_idx" to table: "entity_system_details"
DROP INDEX "entity_system_details_system_detail_id_idx";
-- reverse: create index "entity_subprocessors_subprocessor_id_idx" to table: "entity_subprocessors"
DROP INDEX "entity_subprocessors_subprocessor_id_idx";
-- reverse: create index "entity_integrations_integration_id_idx" to table: "entity_integrations"
DROP INDEX "entity_integrations_integration_id_idx";
-- reverse: create index "entity_files_file_id_idx" to table: "entity_files"
DROP INDEX "entity_files_file_id_idx";
-- reverse: create index "entity_editors_group_id_idx" to table: "entity_editors"
DROP INDEX "entity_editors_group_id_idx";
-- reverse: create index "entity_documents_document_data_id_idx" to table: "entity_documents"
DROP INDEX "entity_documents_document_data_id_idx";
-- reverse: create index "entity_contacts_contact_id_idx" to table: "entity_contacts"
DROP INDEX "entity_contacts_contact_id_idx";
-- reverse: create index "entity_blocked_groups_group_id_idx" to table: "entity_blocked_groups"
DROP INDEX "entity_blocked_groups_group_id_idx";
-- reverse: create index "entity_assets_asset_id_idx" to table: "entity_assets"
DROP INDEX "entity_assets_asset_id_idx";
-- reverse: create index "entity_owner_id_idx" to table: "entities"
DROP INDEX "entity_owner_id_idx";
-- reverse: create index "entity_logo_file_id_idx" to table: "entities"
DROP INDEX "entity_logo_file_id_idx";
-- reverse: create index "entity_entity_type_id_idx" to table: "entities"
DROP INDEX "entity_entity_type_id_idx";
-- reverse: drop index "entity_owner_id" from table: "entities"
CREATE INDEX "entity_owner_id" ON "entities" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "email_verification_tokens_owner_id_fk" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_owner_id_fk";
-- reverse: create index "email_template_workflow_instance_id_idx" to table: "email_templates"
DROP INDEX "email_template_workflow_instance_id_idx";
-- reverse: create index "email_template_workflow_definition_id_idx" to table: "email_templates"
DROP INDEX "email_template_workflow_definition_id_idx";
-- reverse: create index "email_template_trust_center_id_idx" to table: "email_templates"
DROP INDEX "email_template_trust_center_id_idx";
-- reverse: create index "email_template_owner_id_idx" to table: "email_templates"
DROP INDEX "email_template_owner_id_idx";
-- reverse: create index "email_template_integration_id_idx" to table: "email_templates"
DROP INDEX "email_template_integration_id_idx";
-- reverse: drop index "emailtemplate_owner_id" from table: "email_templates"
CREATE INDEX "emailtemplate_owner_id" ON "email_templates" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "document_data_files_file_id_idx" to table: "document_data_files"
DROP INDEX "document_data_files_file_id_idx";
-- reverse: create index "document_template_id_idx" to table: "document_data"
DROP INDEX "document_template_id_idx";
-- reverse: create index "document_owner_id_idx" to table: "document_data"
DROP INDEX "document_owner_id_idx";
-- reverse: drop index "documentdata_owner_id" from table: "document_data"
CREATE INDEX "documentdata_owner_id" ON "document_data" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "dns_verification_owner_id_idx" to table: "dns_verifications"
DROP INDEX "dns_verification_owner_id_idx";
-- reverse: drop index "dnsverification_owner_id" from table: "dns_verifications"
CREATE INDEX "dnsverification_owner_id" ON "dns_verifications" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "discussion_owner_id_idx" to table: "discussions"
DROP INDEX "discussion_owner_id_idx";
-- reverse: drop index "discussion_owner_id" from table: "discussions"
CREATE INDEX "discussion_owner_id" ON "discussions" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "directory_sync_run_owner_id_idx" to table: "directory_sync_runs"
DROP INDEX "directory_sync_run_owner_id_idx";
-- reverse: create index "directory_membership_owner_id_idx" to table: "directory_memberships"
DROP INDEX "directory_membership_owner_id_idx";
-- reverse: create index "directory_membership_directory_group_id_idx" to table: "directory_memberships"
DROP INDEX "directory_membership_directory_group_id_idx";
-- reverse: create index "directory_group_owner_id_idx" to table: "directory_groups"
DROP INDEX "directory_group_owner_id_idx";
-- reverse: create index "directory_account_owner_id_idx" to table: "directory_accounts"
DROP INDEX "directory_account_owner_id_idx";
-- reverse: create index "directory_account_avatar_local_file_id_idx" to table: "directory_accounts"
DROP INDEX "directory_account_avatar_local_file_id_idx";
-- reverse: create index "custom_type_enum_owner_id_idx" to table: "custom_type_enums"
DROP INDEX "custom_type_enum_owner_id_idx";
-- reverse: drop index "customtypeenum_owner_id" from table: "custom_type_enums"
CREATE INDEX "customtypeenum_owner_id" ON "custom_type_enums" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "custom_domain_owner_id_idx" to table: "custom_domains"
DROP INDEX "custom_domain_owner_id_idx";
-- reverse: create index "custom_domain_mappable_domain_id_idx" to table: "custom_domains"
DROP INDEX "custom_domain_mappable_domain_id_idx";
-- reverse: create index "custom_domain_dns_verification_id_idx" to table: "custom_domains"
DROP INDEX "custom_domain_dns_verification_id_idx";
-- reverse: drop index "customdomain_owner_id" from table: "custom_domains"
CREATE INDEX "customdomain_owner_id" ON "custom_domains" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "control_owner_id_idx" to table: "controls"
DROP INDEX "control_owner_id_idx";
-- reverse: drop index "control_owner_id" from table: "controls"
CREATE INDEX "control_owner_id" ON "controls" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "control_tasks_task_id_idx" to table: "control_tasks"
DROP INDEX "control_tasks_task_id_idx";
-- reverse: create index "control_scans_scan_id_idx" to table: "control_scans"
DROP INDEX "control_scans_scan_id_idx";
-- reverse: create index "control_risks_risk_id_idx" to table: "control_risks"
DROP INDEX "control_risks_risk_id_idx";
-- reverse: create index "control_procedures_procedure_id_idx" to table: "control_procedures"
DROP INDEX "control_procedures_procedure_id_idx";
-- reverse: create index "control_objective_owner_id_idx" to table: "control_objectives"
DROP INDEX "control_objective_owner_id_idx";
-- reverse: drop index "controlobjective_owner_id" from table: "control_objectives"
CREATE INDEX "controlobjective_owner_id" ON "control_objectives" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "control_objective_viewers_group_id_idx" to table: "control_objective_viewers"
DROP INDEX "control_objective_viewers_group_id_idx";
-- reverse: create index "control_objective_tasks_task_id_idx" to table: "control_objective_tasks"
DROP INDEX "control_objective_tasks_task_id_idx";
-- reverse: create index "control_objective_editors_group_id_idx" to table: "control_objective_editors"
DROP INDEX "control_objective_editors_group_id_idx";
-- reverse: create index "control_objective_blocked_groups_group_id_idx" to table: "control_objective_blocked_groups"
DROP INDEX "control_objective_blocked_groups_group_id_idx";
-- reverse: create index "control_narratives_narrative_id_idx" to table: "control_narratives"
DROP INDEX "control_narratives_narrative_id_idx";
-- reverse: create index "control_implementation_owner_id_idx" to table: "control_implementations"
DROP INDEX "control_implementation_owner_id_idx";
-- reverse: drop index "controlimplementation_owner_id" from table: "control_implementations"
CREATE INDEX "controlimplementation_owner_id" ON "control_implementations" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "control_implementation_viewers_group_id_idx" to table: "control_implementation_viewers"
DROP INDEX "control_implementation_viewers_group_id_idx";
-- reverse: create index "control_implementation_tasks_task_id_idx" to table: "control_implementation_tasks"
DROP INDEX "control_implementation_tasks_task_id_idx";
-- reverse: create index "control_implementation_editors_group_id_idx" to table: "control_implementation_editors"
DROP INDEX "control_implementation_editors_group_id_idx";
-- reverse: create index "control_implementation_blocked_groups_group_id_idx" to table: "control_implementation_blocked_groups"
DROP INDEX "control_implementation_blocked_groups_group_id_idx";
-- reverse: create index "control_identity_holders_identity_holder_id_idx" to table: "control_identity_holders"
DROP INDEX "control_identity_holders_identity_holder_id_idx";
-- reverse: create index "control_entities_entity_id_idx" to table: "control_entities"
DROP INDEX "control_entities_entity_id_idx";
-- reverse: create index "control_editors_group_id_idx" to table: "control_editors"
DROP INDEX "control_editors_group_id_idx";
-- reverse: create index "control_control_objectives_control_objective_id_idx" to table: "control_control_objectives"
DROP INDEX "control_control_objectives_control_objective_id_idx";
-- reverse: create index "control_control_implementations_control_implementation_id_idx" to table: "control_control_implementations"
DROP INDEX "control_control_implementations_control_implementation_id_idx";
-- reverse: create index "control_campaigns_campaign_id_idx" to table: "control_campaigns"
DROP INDEX "control_campaigns_campaign_id_idx";
-- reverse: create index "control_blocked_groups_group_id_idx" to table: "control_blocked_groups"
DROP INDEX "control_blocked_groups_group_id_idx";
-- reverse: create index "control_assets_asset_id_idx" to table: "control_assets"
DROP INDEX "control_assets_asset_id_idx";
-- reverse: create index "control_action_plans_action_plan_id_idx" to table: "control_action_plans"
DROP INDEX "control_action_plans_action_plan_id_idx";
-- reverse: create index "contact_owner_id_idx" to table: "contacts"
DROP INDEX "contact_owner_id_idx";
-- reverse: drop index "contact_owner_id" from table: "contacts"
CREATE INDEX "contact_owner_id" ON "contacts" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "contact_files_file_id_idx" to table: "contact_files"
DROP INDEX "contact_files_file_id_idx";
-- reverse: create index "check_result_integration_id_idx" to table: "check_results"
DROP INDEX "check_result_integration_id_idx";
-- reverse: create index "check_result_controls_control_id_idx" to table: "check_result_controls"
DROP INDEX "check_result_controls_control_id_idx";
-- reverse: create index "campaign_trust_center_id_idx" to table: "campaigns"
DROP INDEX "campaign_trust_center_id_idx";
-- reverse: create index "campaign_template_id_idx" to table: "campaigns"
DROP INDEX "campaign_template_id_idx";
-- reverse: create index "campaign_owner_id_idx" to table: "campaigns"
DROP INDEX "campaign_owner_id_idx";
-- reverse: create index "campaign_integration_id_idx" to table: "campaigns"
DROP INDEX "campaign_integration_id_idx";
-- reverse: create index "campaign_email_template_id_idx" to table: "campaigns"
DROP INDEX "campaign_email_template_id_idx";
-- reverse: create index "campaign_assessment_id_idx" to table: "campaigns"
DROP INDEX "campaign_assessment_id_idx";
-- reverse: drop index "campaign_owner_id" from table: "campaigns"
CREATE INDEX "campaign_owner_id" ON "campaigns" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "campaign_viewers_group_id_idx" to table: "campaign_viewers"
DROP INDEX "campaign_viewers_group_id_idx";
-- reverse: create index "campaign_users_user_id_idx" to table: "campaign_users"
DROP INDEX "campaign_users_user_id_idx";
-- reverse: create index "campaign_target_owner_id_idx" to table: "campaign_targets"
DROP INDEX "campaign_target_owner_id_idx";
-- reverse: drop index "campaigntarget_owner_id" from table: "campaign_targets"
CREATE INDEX "campaigntarget_owner_id" ON "campaign_targets" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "campaign_identity_holders_identity_holder_id_idx" to table: "campaign_identity_holders"
DROP INDEX "campaign_identity_holders_identity_holder_id_idx";
-- reverse: create index "campaign_groups_group_id_idx" to table: "campaign_groups"
DROP INDEX "campaign_groups_group_id_idx";
-- reverse: create index "campaign_editors_group_id_idx" to table: "campaign_editors"
DROP INDEX "campaign_editors_group_id_idx";
-- reverse: create index "campaign_contacts_contact_id_idx" to table: "campaign_contacts"
DROP INDEX "campaign_contacts_contact_id_idx";
-- reverse: create index "campaign_blocked_groups_group_id_idx" to table: "campaign_blocked_groups"
DROP INDEX "campaign_blocked_groups_group_id_idx";
-- reverse: create index "asset_source_platform_id_idx" to table: "assets"
DROP INDEX "asset_source_platform_id_idx";
-- reverse: create index "asset_owner_id_idx" to table: "assets"
DROP INDEX "asset_owner_id_idx";
-- reverse: create index "asset_integration_id_idx" to table: "assets"
DROP INDEX "asset_integration_id_idx";
-- reverse: drop index "asset_owner_id" from table: "assets"
CREATE INDEX "asset_owner_id" ON "assets" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "asset_connected_assets_connected_from_id_idx" to table: "asset_connected_assets"
DROP INDEX "asset_connected_assets_connected_from_id_idx";
-- reverse: create index "assessment_template_id_idx" to table: "assessments"
DROP INDEX "assessment_template_id_idx";
-- reverse: create index "assessment_owner_id_idx" to table: "assessments"
DROP INDEX "assessment_owner_id_idx";
-- reverse: drop index "assessment_owner_id" from table: "assessments"
CREATE INDEX "assessment_owner_id" ON "assessments" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "assessment_response_owner_id_idx" to table: "assessment_responses"
DROP INDEX "assessment_response_owner_id_idx";
-- reverse: create index "assessment_response_document_data_id_idx" to table: "assessment_responses"
DROP INDEX "assessment_response_document_data_id_idx";
-- reverse: drop index "assessmentresponse_owner_id" from table: "assessment_responses"
CREATE INDEX "assessmentresponse_owner_id" ON "assessment_responses" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "api_token_owner_id_idx" to table: "api_tokens"
DROP INDEX "api_token_owner_id_idx";
-- reverse: drop index "apitoken_owner_id" from table: "api_tokens"
CREATE INDEX "apitoken_owner_id" ON "api_tokens" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "action_plan_owner_id_idx" to table: "action_plans"
DROP INDEX "action_plan_owner_id_idx";
-- reverse: create index "action_plan_file_id_idx" to table: "action_plans"
DROP INDEX "action_plan_file_id_idx";
-- reverse: drop index "actionplan_owner_id" from table: "action_plans"
CREATE INDEX "actionplan_owner_id" ON "action_plans" ("owner_id") WHERE (deleted_at IS NULL);
-- reverse: create index "action_plan_viewers_group_id_idx" to table: "action_plan_viewers"
DROP INDEX "action_plan_viewers_group_id_idx";
-- reverse: create index "action_plan_tasks_task_id_idx" to table: "action_plan_tasks"
DROP INDEX "action_plan_tasks_task_id_idx";
-- reverse: create index "action_plan_editors_group_id_idx" to table: "action_plan_editors"
DROP INDEX "action_plan_editors_group_id_idx";
-- reverse: create index "action_plan_blocked_groups_group_id_idx" to table: "action_plan_blocked_groups"
DROP INDEX "action_plan_blocked_groups_group_id_idx";
