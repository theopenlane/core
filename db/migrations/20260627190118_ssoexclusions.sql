-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "check_results" table
ALTER TABLE "check_results" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "discussions" table
ALTER TABLE "discussions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "dns_verifications" table
ALTER TABLE "dns_verifications" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "entities" table
ALTER TABLE "entities" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "entity_types" table
ALTER TABLE "entity_types" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "exports" table
ALTER TABLE "exports" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "finding_controls" table
ALTER TABLE "finding_controls" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "findings" table
ALTER TABLE "findings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "group_memberships" table
ALTER TABLE "group_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "group_settings" table
ALTER TABLE "group_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "impersonation_events" table
ALTER TABLE "impersonation_events" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "invites" table
ALTER TABLE "invites" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt" boolean NULL DEFAULT false;
-- Modify "job_results" table
ALTER TABLE "job_results" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "job_runner_registration_tokens" table
ALTER TABLE "job_runner_registration_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "job_runner_tokens" table
ALTER TABLE "job_runner_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "job_runners" table
ALTER TABLE "job_runners" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "job_templates" table
ALTER TABLE "job_templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "mappable_domains" table
ALTER TABLE "mappable_domains" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "notes" table
ALTER TABLE "notes" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "notification_preferences" table
ALTER TABLE "notification_preferences" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "notifications" table
ALTER TABLE "notifications" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "org_memberships" table
ALTER TABLE "org_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "sso_exempt" boolean NULL DEFAULT false, ADD COLUMN "sso_exempt_reason" character varying NULL, ADD COLUMN "sso_exempt_granted_by" character varying NULL, ADD COLUMN "sso_exempt_granted_at" timestamptz NULL;
-- Modify "org_modules" table
ALTER TABLE "org_modules" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "org_prices" table
ALTER TABLE "org_prices" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "org_products" table
ALTER TABLE "org_products" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "identity_provider_jit_provisioning" boolean NOT NULL DEFAULT true, ADD COLUMN "jit_allowed_email_domains" jsonb NULL, ADD COLUMN "sso_exempt_domains" jsonb NULL, ADD COLUMN "allow_support_access" boolean NULL DEFAULT false;
-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "updated_by_impersonator" character varying NULL, ADD COLUMN "slug_name" character varying NULL;
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "scans" table
ALTER TABLE "scans" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "sla_definitions" table
ALTER TABLE "sla_definitions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "standards" table
ALTER TABLE "standards" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "subprocessors" table
ALTER TABLE "subprocessors" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_entities" table
ALTER TABLE "trust_center_entities" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "user_settings" table
ALTER TABLE "user_settings" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "vendor_scoring_configs" table
ALTER TABLE "vendor_scoring_configs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_assignment_targets" table
ALTER TABLE "workflow_assignment_targets" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_events" table
ALTER TABLE "workflow_events" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "updated_by_impersonator" character varying NULL;
-- Modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" ADD COLUMN "updated_by_impersonator" character varying NULL;
