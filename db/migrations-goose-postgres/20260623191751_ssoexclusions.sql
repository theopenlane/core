-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "check_results" table
ALTER TABLE "check_results" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "discussions" table
ALTER TABLE "discussions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "dns_verifications" table
ALTER TABLE "dns_verifications" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "document_data" table
ALTER TABLE "document_data" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "exports" table
ALTER TABLE "exports" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "finding_controls" table
ALTER TABLE "finding_controls" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "findings" table
ALTER TABLE "findings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "group_memberships" table
ALTER TABLE "group_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "group_settings" table
ALTER TABLE "group_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "impersonation_events" table
ALTER TABLE "impersonation_events" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "invites" table
ALTER TABLE "invites" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt" boolean NULL DEFAULT false;
-- modify "job_results" table
ALTER TABLE "job_results" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "job_runner_registration_tokens" table
ALTER TABLE "job_runner_registration_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "job_runner_tokens" table
ALTER TABLE "job_runner_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "job_runners" table
ALTER TABLE "job_runners" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "job_templates" table
ALTER TABLE "job_templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "mappable_domains" table
ALTER TABLE "mappable_domains" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "notification_preferences" table
ALTER TABLE "notification_preferences" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "notifications" table
ALTER TABLE "notifications" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "org_memberships" table
ALTER TABLE "org_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt" boolean NULL DEFAULT false, ADD COLUMN "sso_exempt_reason" character varying NULL, ADD COLUMN "sso_exempt_granted_by" character varying NULL, ADD COLUMN "sso_exempt_granted_at" timestamptz NULL;
-- modify "org_modules" table
ALTER TABLE "org_modules" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "org_prices" table
ALTER TABLE "org_prices" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "org_products" table
ALTER TABLE "org_products" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt_domains" jsonb NULL, ADD COLUMN "allow_support_access" boolean NULL DEFAULT false;
-- modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "sla_definitions" table
ALTER TABLE "sla_definitions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "subprocessors" table
ALTER TABLE "subprocessors" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_entities" table
ALTER TABLE "trust_center_entities" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "user_settings" table
ALTER TABLE "user_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "vendor_scoring_configs" table
ALTER TABLE "vendor_scoring_configs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_assignment_targets" table
ALTER TABLE "workflow_assignment_targets" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_events" table
ALTER TABLE "workflow_events" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" ADD COLUMN "updated_by_impersonator" character varying NULL;

-- +goose Down
-- reverse: modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_instances" table
ALTER TABLE "workflow_instances" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_events" table
ALTER TABLE "workflow_events" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "workflow_assignment_targets" table
ALTER TABLE "workflow_assignment_targets" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "vendor_scoring_configs" table
ALTER TABLE "vendor_scoring_configs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_entities" table
ALTER TABLE "trust_center_entities" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "tag_definitions" table
ALTER TABLE "tag_definitions" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "system_details" table
ALTER TABLE "system_details" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "subprocessors" table
ALTER TABLE "subprocessors" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "sla_definitions" table
ALTER TABLE "sla_definitions" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "platforms" table
ALTER TABLE "platforms" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "allow_support_access", DROP COLUMN "sso_exempt_domains", DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_products" table
ALTER TABLE "org_products" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_prices" table
ALTER TABLE "org_prices" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_modules" table
ALTER TABLE "org_modules" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "sso_exempt_granted_at", DROP COLUMN "sso_exempt_granted_by", DROP COLUMN "sso_exempt_reason", DROP COLUMN "sso_exempt", DROP COLUMN "updated_by_impersonator";
-- reverse: modify "notifications" table
ALTER TABLE "notifications" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "notification_preferences" table
ALTER TABLE "notification_preferences" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "mappable_domains" table
ALTER TABLE "mappable_domains" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "job_templates" table
ALTER TABLE "job_templates" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "job_runners" table
ALTER TABLE "job_runners" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "job_runner_tokens" table
ALTER TABLE "job_runner_tokens" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "job_runner_registration_tokens" table
ALTER TABLE "job_runner_registration_tokens" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "job_results" table
ALTER TABLE "job_results" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP COLUMN "sso_exempt", DROP COLUMN "updated_by_impersonator";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "impersonation_events" table
ALTER TABLE "impersonation_events" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "identity_holders" table
ALTER TABLE "identity_holders" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "finding_controls" table
ALTER TABLE "finding_controls" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "exports" table
ALTER TABLE "exports" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "dns_verifications" table
ALTER TABLE "dns_verifications" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "discussions" table
ALTER TABLE "discussions" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_memberships" table
ALTER TABLE "directory_memberships" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_groups" table
ALTER TABLE "directory_groups" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "directory_accounts" table
ALTER TABLE "directory_accounts" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "check_results" table
ALTER TABLE "check_results" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "campaign_targets" table
ALTER TABLE "campaign_targets" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "updated_by_impersonator";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "updated_by_impersonator";
