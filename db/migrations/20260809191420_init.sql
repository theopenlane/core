-- Create "api_tokens" table
CREATE TABLE "api_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "sso_authorizations" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "api_token_owner_id_idx" to table: "api_tokens"
CREATE INDEX "api_token_owner_id_idx" ON "api_tokens" ("owner_id");
-- Create index "api_tokens_token_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_token_key" ON "api_tokens" ("token");
-- Create index "apitoken_token" to table: "api_tokens"
CREATE INDEX "apitoken_token" ON "api_tokens" ("token");
-- Create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED', "details" text NULL, "details_json" jsonb NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "summary" character varying NULL, "tag_suggestions" jsonb NULL, "dismissed_tag_suggestions" jsonb NULL, "control_suggestions" jsonb NULL, "dismissed_control_suggestions" jsonb NULL, "improvement_suggestions" jsonb NULL, "dismissed_improvement_suggestions" jsonb NULL, "url" character varying NULL, "external_file_id" character varying NULL, "external_contents" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "action_plan_kind_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "title" character varying NOT NULL, "description" text NULL, "due_date" timestamptz NULL, "completed_at" timestamptz NULL, "priority" character varying NULL, "requires_approval" boolean NOT NULL DEFAULT false, "blocked" boolean NOT NULL DEFAULT false, "blocker_reason" text NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "source" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "action_plan_kind_id" character varying NULL, "file_id" character varying NULL, "custom_type_enum_action_plans" character varying NULL, "owner_id" character varying NULL, "subcontrol_action_plans" character varying NULL, "user_action_plans" character varying NULL, PRIMARY KEY ("id"));
-- Create index "action_plan_file_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_file_id_idx" ON "action_plans" ("file_id");
-- Create index "action_plan_owner_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_owner_id_idx" ON "action_plans" ("owner_id");
-- Create "assessments" table
CREATE TABLE "assessments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "response_due_duration" bigint NULL, "owner_id" character varying NULL, "template_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "assessment_name_owner_id" to table: "assessments"
CREATE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "assessment_owner_id_idx" to table: "assessments"
CREATE INDEX "assessment_owner_id_idx" ON "assessments" ("owner_id");
-- Create index "assessment_template_id_idx" to table: "assessments"
CREATE INDEX "assessment_template_id_idx" ON "assessments" ("template_id");
-- Create "assessment_responses" table
CREATE TABLE "assessment_responses" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "is_test" boolean NOT NULL DEFAULT false, "display_name" character varying NULL, "email" character varying NULL, "send_attempts" bigint NOT NULL DEFAULT 1, "email_delivered_at" timestamptz NULL, "email_opened_at" timestamptz NULL, "email_clicked_at" timestamptz NULL, "email_open_count" bigint NULL DEFAULT 0, "email_click_count" bigint NULL DEFAULT 0, "last_email_event_at" timestamptz NULL, "email_metadata" jsonb NULL, "status" character varying NOT NULL DEFAULT 'SENT', "assigned_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "is_draft" boolean NOT NULL DEFAULT false, "assessment_id" character varying NOT NULL, "document_data_id" character varying NULL, "campaign_id" character varying NULL, "entity_id" character varying NULL, "identity_holder_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "assessment_response_document_data_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_document_data_id_idx" ON "assessment_responses" ("document_data_id");
-- Create index "assessment_response_owner_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_owner_id_idx" ON "assessment_responses" ("owner_id");
-- Create index "assessmentresponse_assessment_id_email_is_test" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_email_is_test" ON "assessment_responses" ("assessment_id", "email", "is_test") WHERE ((deleted_at IS NULL) AND (campaign_id IS NULL));
-- Create index "assessmentresponse_assigned_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_assigned_at" ON "assessment_responses" ("assigned_at");
-- Create index "assessmentresponse_campaign_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_campaign_id" ON "assessment_responses" ("campaign_id");
-- Create index "assessmentresponse_campaign_id_assessment_id_email_is_test" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_campaign_id_assessment_id_email_is_test" ON "assessment_responses" ("campaign_id", "assessment_id", "email", "is_test") WHERE ((deleted_at IS NULL) AND (campaign_id IS NOT NULL));
-- Create index "assessmentresponse_completed_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_completed_at" ON "assessment_responses" ("completed_at");
-- Create index "assessmentresponse_due_date" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_due_date" ON "assessment_responses" ("due_date");
-- Create index "assessmentresponse_entity_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_entity_id" ON "assessment_responses" ("entity_id");
-- Create index "assessmentresponse_identity_holder_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_identity_holder_id" ON "assessment_responses" ("identity_holder_id");
-- Create index "assessmentresponse_status" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_status" ON "assessment_responses" ("status");
-- Create "assets" table
CREATE TABLE "assets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "asset_subtype_name" character varying NULL, "asset_data_classification_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "access_model_name" character varying NULL, "encryption_status_name" character varying NULL, "security_tier_name" character varying NULL, "criticality_name" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "asset_type" character varying NOT NULL DEFAULT 'TECHNOLOGY', "name" character varying NOT NULL, "display_name" character varying NULL, "description" character varying NULL, "identifier" character varying NULL, "website" character varying NULL, "physical_location" character varying NULL, "region" character varying NULL, "contains_pii" boolean NULL DEFAULT false, "source_type" character varying NOT NULL DEFAULT 'MANUAL', "source_identifier" character varying NULL, "cost_center" character varying NULL, "estimated_monthly_cost" double precision NULL, "purchase_date" timestamptz NULL, "cpe" character varying NULL, "categories" jsonb NULL, "observed_at" timestamptz NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "asset_subtype_id" character varying NULL, "asset_data_classification_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "access_model_id" character varying NULL, "encryption_status_id" character varying NULL, "security_tier_id" character varying NULL, "criticality_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "source_platform_id" character varying NULL, "risk_assets" character varying NULL, PRIMARY KEY ("id"));
-- Create index "asset_integration_id_idx" to table: "assets"
CREATE INDEX "asset_integration_id_idx" ON "assets" ("integration_id");
-- Create index "asset_owner_id_idx" to table: "assets"
CREATE INDEX "asset_owner_id_idx" ON "assets" ("owner_id");
-- Create index "asset_source_platform_id_idx" to table: "assets"
CREATE INDEX "asset_source_platform_id_idx" ON "assets" ("source_platform_id");
-- Create "campaigns" table
CREATE TABLE "campaigns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "name" character varying NOT NULL, "description" character varying NULL, "campaign_type" character varying NOT NULL DEFAULT 'QUESTIONNAIRE', "status" character varying NOT NULL DEFAULT 'DRAFT', "is_active" boolean NOT NULL DEFAULT false, "scheduled_at" timestamptz NULL, "launched_at" timestamptz NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "is_recurring" boolean NOT NULL DEFAULT false, "recurrence_frequency" character varying NULL DEFAULT 'NONE', "recurrence_interval" bigint NULL DEFAULT 1, "recurrence_timezone" character varying NULL, "recurrence_cron" character varying NULL, "last_run_at" timestamptz NULL, "next_run_at" timestamptz NULL, "recurrence_end_at" timestamptz NULL, "recipient_count" bigint NULL DEFAULT 0, "resend_count" bigint NULL DEFAULT 0, "last_resent_at" timestamptz NULL, "metadata" jsonb NULL, "email_branding_id" character varying NULL, "assessment_id" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "email_template_id" character varying NULL, "entity_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "campaign_assessment_id_idx" to table: "campaigns"
CREATE INDEX "campaign_assessment_id_idx" ON "campaigns" ("assessment_id");
-- Create index "campaign_display_id_owner_id" to table: "campaigns"
CREATE UNIQUE INDEX "campaign_display_id_owner_id" ON "campaigns" ("display_id", "owner_id");
-- Create index "campaign_email_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_email_template_id_idx" ON "campaigns" ("email_template_id");
-- Create index "campaign_entity_id" to table: "campaigns"
CREATE INDEX "campaign_entity_id" ON "campaigns" ("entity_id");
-- Create index "campaign_integration_id_idx" to table: "campaigns"
CREATE INDEX "campaign_integration_id_idx" ON "campaigns" ("integration_id");
-- Create index "campaign_name_owner_id" to table: "campaigns"
CREATE INDEX "campaign_name_owner_id" ON "campaigns" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "campaign_owner_id_idx" to table: "campaigns"
CREATE INDEX "campaign_owner_id_idx" ON "campaigns" ("owner_id");
-- Create index "campaign_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_template_id_idx" ON "campaigns" ("template_id");
-- Create index "campaign_trust_center_id_idx" to table: "campaigns"
CREATE INDEX "campaign_trust_center_id_idx" ON "campaigns" ("trust_center_id");
-- Create "campaign_targets" table
CREATE TABLE "campaign_targets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "email" character varying NOT NULL, "full_name" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "sent_at" timestamptz NULL, "completed_at" timestamptz NULL, "metadata" jsonb NULL, "campaign_id" character varying NULL, "contact_id" character varying NULL, "group_id" character varying NULL, "owner_id" character varying NULL, "subscriber_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "campaign_target_owner_id_idx" to table: "campaign_targets"
CREATE INDEX "campaign_target_owner_id_idx" ON "campaign_targets" ("owner_id");
-- Create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
CREATE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
-- Create index "campaigntarget_contact_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_contact_id" ON "campaign_targets" ("contact_id");
-- Create index "campaigntarget_group_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_group_id" ON "campaign_targets" ("group_id");
-- Create index "campaigntarget_status" to table: "campaign_targets"
CREATE INDEX "campaigntarget_status" ON "campaign_targets" ("status");
-- Create index "campaigntarget_subscriber_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_subscriber_id" ON "campaign_targets" ("subscriber_id");
-- Create index "campaigntarget_user_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_user_id" ON "campaign_targets" ("user_id");
-- Create "check_results" table
CREATE TABLE "check_results" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'UNKNOWN', "source" character varying NOT NULL, "last_observed_at" timestamptz NULL, "external_uri" character varying NULL, "details" text NULL, "parent_external_id" character varying NULL, "integration_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "check_result_integration_id_idx" to table: "check_results"
CREATE INDEX "check_result_integration_id_idx" ON "check_results" ("integration_id");
-- Create "contacts" table
CREATE TABLE "contacts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "full_name" character varying NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "external_id" character varying NULL, "integration_id" character varying NULL, "observed_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "contact_owner_id_idx" to table: "contacts"
CREATE INDEX "contact_owner_id_idx" ON "contacts" ("owner_id");
-- Create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "external_uuid" character varying NULL, "title" character varying NULL, "description" text NULL, "description_json" jsonb NULL, "aliases" jsonb NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NOT_IMPLEMENTED', "implementation_status" character varying NULL DEFAULT 'PLANNED', "implementation_description" text NULL, "public_representation" text NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "source_name" character varying NULL, "reference_framework" character varying NULL, "reference_framework_revision" character varying NULL, "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "testing_procedures" jsonb NULL, "evidence_requests" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "control_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "ref_code" character varying NOT NULL, "trust_center_visibility" character varying NULL DEFAULT 'NOT_VISIBLE', "is_trust_center_control" boolean NULL DEFAULT false, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "responsible_party_id" character varying NULL, "control_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "custom_type_enum_controls" character varying NULL, "owner_id" character varying NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "control_auditor_reference_id" to table: "controls"
CREATE INDEX "control_auditor_reference_id" ON "controls" ("auditor_reference_id") WHERE (deleted_at IS NULL);
-- Create index "control_display_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_display_id_owner_id" ON "controls" ("display_id", "owner_id");
-- Create index "control_external_uuid_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_external_uuid_owner_id" ON "controls" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "control_owner_id_idx" to table: "controls"
CREATE INDEX "control_owner_id_idx" ON "controls" ("owner_id");
-- Create index "control_ref_code" to table: "controls"
CREATE INDEX "control_ref_code" ON "controls" ("ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND ((status)::text <> 'ARCHIVED'::text));
-- Create index "control_ref_code_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_ref_code_owner_id" ON "controls" ("ref_code", "owner_id") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND (standard_id IS NULL));
-- Create index "control_reference_id" to table: "controls"
CREATE INDEX "control_reference_id" ON "controls" ("reference_id") WHERE (deleted_at IS NULL);
-- Create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NULL));
-- Create index "control_standard_id_ref_code_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code_owner_id" ON "controls" ("standard_id", "ref_code", "owner_id") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND (standard_id IS NOT NULL));
-- Create "control_implementations" table
CREATE TABLE "control_implementations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "status" character varying NULL DEFAULT 'DRAFT', "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, "details_json" jsonb NULL, "evidence_control_implementations" character varying NULL, "internal_policy_control_implementations" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "control_implementation_owner_id_idx" to table: "control_implementations"
CREATE INDEX "control_implementation_owner_id_idx" ON "control_implementations" ("owner_id");
-- Create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "desired_outcome" text NULL, "desired_outcome_json" jsonb NULL, "status" character varying NULL DEFAULT 'DRAFT', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_objective_type" character varying NULL, "category" character varying NULL, "subcategory" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "control_objective_owner_id_idx" to table: "control_objectives"
CREATE INDEX "control_objective_owner_id_idx" ON "control_objectives" ("owner_id");
-- Create index "controlobjective_display_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_display_id_owner_id" ON "control_objectives" ("display_id", "owner_id");
-- Create "custom_domains" table
CREATE TABLE "custom_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "cname_record" character varying NOT NULL, "trust_center_id" character varying NULL, "domain_type" character varying NOT NULL DEFAULT 'UNKNOWN', "mappable_domain_id" character varying NOT NULL, "dns_verification_id" character varying NULL, "dns_verification_custom_domains" character varying NULL, "mappable_domain_custom_domains" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "custom_domain_dns_verification_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_dns_verification_id_idx" ON "custom_domains" ("dns_verification_id");
-- Create index "custom_domain_mappable_domain_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_mappable_domain_id_idx" ON "custom_domains" ("mappable_domain_id");
-- Create index "custom_domain_owner_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_owner_id_idx" ON "custom_domains" ("owner_id");
-- Create index "customdomain_cname_record" to table: "custom_domains"
CREATE UNIQUE INDEX "customdomain_cname_record" ON "custom_domains" ("cname_record") WHERE (deleted_at IS NULL);
-- Create "custom_type_enums" table
CREATE TABLE "custom_type_enums" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "object_type" character varying NOT NULL, "field" character varying NOT NULL DEFAULT 'kind', "name" citext NOT NULL, "description" character varying NULL, "color" character varying NULL, "icon" character varying NULL, "entity_auth_methods" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "custom_type_enum_owner_id_idx" to table: "custom_type_enums"
CREATE INDEX "custom_type_enum_owner_id_idx" ON "custom_type_enums" ("owner_id");
-- Create index "customtypeenum_name_field" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_name_field" ON "custom_type_enums" ("name", "field") WHERE (deleted_at IS NULL);
-- Create index "customtypeenum_name_object_type_field_owner_id" to table: "custom_type_enums"
CREATE UNIQUE INDEX "customtypeenum_name_object_type_field_owner_id" ON "custom_type_enums" ("name", "object_type", "field", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "customtypeenum_object_type" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_object_type" ON "custom_type_enums" ("object_type") WHERE (deleted_at IS NULL);
-- Create "dns_verifications" table
CREATE TABLE "dns_verifications" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "cloudflare_hostname_id" character varying NOT NULL, "dns_txt_record" character varying NOT NULL, "dns_txt_value" character varying NOT NULL, "dns_verification_status" character varying NOT NULL DEFAULT 'PENDING', "dns_verification_status_reason" character varying NULL, "acme_challenge_path" character varying NULL, "expected_acme_challenge_value" character varying NULL, "acme_challenge_status" character varying NOT NULL DEFAULT 'INITIALIZING', "acme_challenge_status_reason" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "dns_verification_owner_id_idx" to table: "dns_verifications"
CREATE INDEX "dns_verification_owner_id_idx" ON "dns_verifications" ("owner_id");
-- Create index "dnsverification_cloudflare_hostname_id" to table: "dns_verifications"
CREATE UNIQUE INDEX "dnsverification_cloudflare_hostname_id" ON "dns_verifications" ("cloudflare_hostname_id") WHERE (deleted_at IS NULL);
-- Create "directory_accounts" table
CREATE TABLE "directory_accounts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "directory_name" character varying NULL, "external_id" character varying NOT NULL, "secondary_key" character varying NULL, "canonical_email" character varying NULL, "email_aliases" jsonb NULL, "phone_number" character varying NULL, "display_name" character varying NULL, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "given_name" character varying NULL, "family_name" character varying NULL, "job_title" character varying NULL, "department" character varying NULL, "organization_unit" character varying NULL, "account_type" character varying NULL DEFAULT 'USER', "status" character varying NOT NULL DEFAULT 'ACTIVE', "mfa_state" character varying NOT NULL DEFAULT 'UNKNOWN', "last_seen_ip" character varying NULL, "last_login_at" timestamptz NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "added_at" timestamptz NULL, "removed_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "metadata" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, "primary_source" boolean NOT NULL DEFAULT false, "environment_id" character varying NULL, "scope_id" character varying NULL, "avatar_local_file_id" character varying NULL, "directory_sync_run_id" character varying NULL, "identity_holder_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "directory_account_avatar_local_file_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_avatar_local_file_id_idx" ON "directory_accounts" ("avatar_local_file_id");
-- Create index "directory_account_owner_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_owner_id_idx" ON "directory_accounts" ("owner_id");
-- Create index "directoryaccount_directory_instance_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_instance_id_canonical_email" ON "directory_accounts" ("directory_instance_id", "canonical_email");
-- Create index "directoryaccount_directory_instance_id_external_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_instance_id_external_id" ON "directory_accounts" ("directory_instance_id", "external_id");
-- Create index "directoryaccount_directory_sync_run_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_sync_run_id_canonical_email" ON "directory_accounts" ("directory_sync_run_id", "canonical_email");
-- Create index "directoryaccount_display_id_owner_id" to table: "directory_accounts"
CREATE UNIQUE INDEX "directoryaccount_display_id_owner_id" ON "directory_accounts" ("display_id", "owner_id");
-- Create index "directoryaccount_identity_holder_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_identity_holder_id" ON "directory_accounts" ("identity_holder_id");
-- Create index "directoryaccount_identity_holder_id_directory_name" to table: "directory_accounts"
CREATE INDEX "directoryaccount_identity_holder_id_directory_name" ON "directory_accounts" ("identity_holder_id", "directory_name");
-- Create index "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" to table: "directory_accounts"
CREATE UNIQUE INDEX "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" ON "directory_accounts" ("integration_id", "external_id", "directory_sync_run_id");
-- Create index "directoryaccount_integration_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_integration_id_canonical_email" ON "directory_accounts" ("integration_id", "canonical_email");
-- Create index "directoryaccount_owner_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_owner_id_canonical_email" ON "directory_accounts" ("owner_id", "canonical_email");
-- Create index "directoryaccount_platform_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_platform_id_canonical_email" ON "directory_accounts" ("platform_id", "canonical_email");
-- Create index "directoryaccount_platform_id_external_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_platform_id_external_id" ON "directory_accounts" ("platform_id", "external_id");
-- Create "directory_groups" table
CREATE TABLE "directory_groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "external_id" character varying NOT NULL, "email" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "classification" character varying NOT NULL DEFAULT 'TEAM', "status" character varying NOT NULL DEFAULT 'ACTIVE', "external_sharing_allowed" boolean NULL DEFAULT false, "member_count" bigint NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "added_at" timestamptz NULL, "removed_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "metadata" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, "directory_name" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "directory_sync_run_id" character varying NOT NULL, "integration_id" character varying NOT NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "directory_group_owner_id_idx" to table: "directory_groups"
CREATE INDEX "directory_group_owner_id_idx" ON "directory_groups" ("owner_id");
-- Create index "directorygroup_directory_instance_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_instance_id_email" ON "directory_groups" ("directory_instance_id", "email");
-- Create index "directorygroup_directory_instance_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_instance_id_external_id" ON "directory_groups" ("directory_instance_id", "external_id");
-- Create index "directorygroup_directory_sync_run_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_sync_run_id_email" ON "directory_groups" ("directory_sync_run_id", "email");
-- Create index "directorygroup_display_id_owner_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "owner_id");
-- Create index "directorygroup_integration_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("integration_id", "email");
-- Create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("integration_id", "external_id", "directory_sync_run_id");
-- Create index "directorygroup_owner_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("owner_id", "email");
-- Create index "directorygroup_platform_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_email" ON "directory_groups" ("platform_id", "email");
-- Create index "directorygroup_platform_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_external_id" ON "directory_groups" ("platform_id", "external_id");
-- Create "directory_memberships" table
CREATE TABLE "directory_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "role" character varying NULL DEFAULT 'MEMBER', "source" character varying NULL, "directory_name" character varying NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "added_at" timestamptz NULL, "removed_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "last_confirmed_run_id" character varying NULL, "metadata" jsonb NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "directory_account_id" character varying NOT NULL, "directory_group_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "integration_id" character varying NOT NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "directory_membership_directory_group_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_directory_group_id_idx" ON "directory_memberships" ("directory_group_id");
-- Create index "directory_membership_owner_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_owner_id_idx" ON "directory_memberships" ("owner_id");
-- Create index "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" to table: "directory_memberships"
CREATE INDEX "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" ON "directory_memberships" ("directory_instance_id", "directory_account_id", "directory_group_id");
-- Create index "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" ON "directory_memberships" ("directory_account_id", "directory_group_id", "directory_sync_run_id");
-- Create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory_account_id_directory_group_id" ON "directory_memberships" ("directory_account_id", "directory_group_id") WHERE (removed_at IS NULL);
-- Create index "directorymembership_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_directory_sync_run_id" ON "directory_memberships" ("directory_sync_run_id");
-- Create index "directorymembership_display_id_owner_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_display_id_owner_id" ON "directory_memberships" ("display_id", "owner_id");
-- Create index "directorymembership_integration_id_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_integration_id_directory_sync_run_id" ON "directory_memberships" ("integration_id", "directory_sync_run_id");
-- Create index "directorymembership_platform_id_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_platform_id_directory_sync_run_id" ON "directory_memberships" ("platform_id", "directory_sync_run_id");
-- Create "directory_sync_runs" table
CREATE TABLE "directory_sync_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "source_cursor" character varying NULL, "full_count" bigint NOT NULL DEFAULT 0, "delta_count" bigint NOT NULL DEFAULT 0, "error" text NULL, "raw_manifest_file_id" character varying NULL, "stats" jsonb NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "integration_id" character varying NOT NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "directory_sync_run_owner_id_idx" to table: "directory_sync_runs"
CREATE INDEX "directory_sync_run_owner_id_idx" ON "directory_sync_runs" ("owner_id");
-- Create index "directorysyncrun_directory_instance_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_directory_instance_id_started_at" ON "directory_sync_runs" ("directory_instance_id", "started_at");
-- Create index "directorysyncrun_display_id_owner_id" to table: "directory_sync_runs"
CREATE UNIQUE INDEX "directorysyncrun_display_id_owner_id" ON "directory_sync_runs" ("display_id", "owner_id");
-- Create index "directorysyncrun_integration_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_integration_id_started_at" ON "directory_sync_runs" ("integration_id", "started_at");
-- Create index "directorysyncrun_platform_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_platform_id_started_at" ON "directory_sync_runs" ("platform_id", "started_at");
-- Create "discussions" table
CREATE TABLE "discussions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "external_id" character varying NULL, "is_resolved" boolean NOT NULL DEFAULT false, "control_discussions" character varying NULL, "internal_policy_discussions" character varying NULL, "owner_id" character varying NULL, "procedure_discussions" character varying NULL, "risk_discussions" character varying NULL, "subcontrol_discussions" character varying NULL, PRIMARY KEY ("id"));
-- Create index "discussion_owner_id_idx" to table: "discussions"
CREATE INDEX "discussion_owner_id_idx" ON "discussions" ("owner_id");
-- Create index "discussions_external_id_key" to table: "discussions"
CREATE UNIQUE INDEX "discussions_external_id_key" ON "discussions" ("external_id");
-- Create "document_data" table
CREATE TABLE "document_data" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "data" jsonb NOT NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "document_owner_id_idx" to table: "document_data"
CREATE INDEX "document_owner_id_idx" ON "document_data" ("owner_id");
-- Create index "document_template_id_idx" to table: "document_data"
CREATE INDEX "document_template_id_idx" ON "document_data" ("template_id");
-- Create "email_templates" table
CREATE TABLE "email_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "format" character varying NULL DEFAULT 'HTML', "locale" character varying NOT NULL DEFAULT 'en-US', "subject_template" character varying NULL, "preheader_template" character varying NULL, "body_template" text NULL, "text_template" text NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "template_context" character varying NULL, "defaults" jsonb NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "trust_center_id" character varying NULL, "workflow_definition_id" character varying NULL, "workflow_instance_id" character varying NULL, PRIMARY KEY ("id"));
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
-- Create index "emailtemplate_owner_id_key" to table: "email_templates"
CREATE INDEX "emailtemplate_owner_id_key" ON "email_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- Create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "email_verification_tokens_owner_id_fk" to table: "email_verification_tokens"
CREATE INDEX "email_verification_tokens_owner_id_fk" ON "email_verification_tokens" ("owner_id");
-- Create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_token_key" ON "email_verification_tokens" ("token");
-- Create index "emailverificationtoken_token" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_token" ON "email_verification_tokens" ("token") WHERE (deleted_at IS NULL);
-- Create "entities" table
CREATE TABLE "entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "reviewed_by" character varying NULL, "last_reviewed_at" timestamptz NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "entity_relationship_state_name" character varying NULL, "entity_security_questionnaire_status_name" character varying NULL, "entity_source_type_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "name" citext NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "aliases" jsonb NULL, "status" character varying NULL DEFAULT 'ACTIVE', "approved_for_use" boolean NULL DEFAULT false, "linked_asset_ids" jsonb NULL, "has_soc2" boolean NULL DEFAULT false, "soc2_period_end" timestamptz NULL, "contract_start_date" timestamptz NULL, "contract_end_date" timestamptz NULL, "auto_renews" boolean NULL DEFAULT false, "termination_notice_days" bigint NULL, "annual_spend" double precision NULL, "spend_currency" character varying NULL DEFAULT 'USD', "billing_model" character varying NULL, "renewal_risk" character varying NULL, "sso_enforced" boolean NULL DEFAULT false, "mfa_supported" boolean NULL DEFAULT false, "mfa_enforced" boolean NULL DEFAULT false, "status_page_url" character varying NULL, "provided_services" jsonb NULL, "links" jsonb NULL, "risk_rating" character varying NULL, "risk_score" bigint NULL, "risk_score_coverage" bigint NULL, "tier" character varying NULL DEFAULT 'LOW', "review_frequency" character varying NULL DEFAULT 'YEARLY', "next_review_at" timestamptz NULL, "contract_renewal_at" timestamptz NULL, "vendor_metadata" jsonb NULL, "logo_remote_url" character varying NULL, "external_id" character varying NULL, "observed_at" timestamptz NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "entity_relationship_state_id" character varying NULL, "entity_security_questionnaire_status_id" character varying NULL, "entity_source_type_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "entity_type_id" character varying NULL, "logo_file_id" character varying NULL, "entity_type_entities" character varying NULL, "owner_id" character varying NULL, "risk_entities" character varying NULL, PRIMARY KEY ("id"));
-- Create index "entity_entity_type_id_idx" to table: "entities"
CREATE INDEX "entity_entity_type_id_idx" ON "entities" ("entity_type_id");
-- Create index "entity_logo_file_id_idx" to table: "entities"
CREATE INDEX "entity_logo_file_id_idx" ON "entities" ("logo_file_id");
-- Create index "entity_name_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "entity_owner_id_idx" to table: "entities"
CREATE INDEX "entity_owner_id_idx" ON "entities" ("owner_id");
-- Create index "entity_reviewed_by_user_id" to table: "entities"
CREATE INDEX "entity_reviewed_by_user_id" ON "entities" ("reviewed_by_user_id");
-- Create "entity_types" table
CREATE TABLE "entity_types" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" citext NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "entity_type_owner_id_idx" to table: "entity_types"
CREATE INDEX "entity_type_owner_id_idx" ON "entity_types" ("owner_id");
-- Create index "entitytype_name_owner_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_name_owner_id" ON "entity_types" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create "events" table
CREATE TABLE "events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, "directory_membership_events" character varying NULL, "export_events" character varying NULL, PRIMARY KEY ("id"));
-- Create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "status" character varying NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "environment_id" character varying NULL, "scope_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- Create index "evidence_external_uuid_owner_id" to table: "evidences"
CREATE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "evidence_owner_id_idx" to table: "evidences"
CREATE INDEX "evidence_owner_id_idx" ON "evidences" ("owner_id");
-- Create "exports" table
CREATE TABLE "exports" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "requestor_id" character varying NULL, "export_type" character varying NOT NULL, "format" character varying NOT NULL DEFAULT 'CSV', "status" character varying NOT NULL DEFAULT 'PENDING', "fields" jsonb NULL, "filters" character varying NULL, "error_message" character varying NULL, "mode" character varying NOT NULL DEFAULT 'FLAT', "export_metadata" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "export_owner_id_idx" to table: "exports"
CREATE INDEX "export_owner_id_idx" ON "exports" ("owner_id");
-- Create "files" table
CREATE TABLE "files" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "category_name" character varying NULL, "name" character varying NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, "metadata" jsonb NULL, "storage_region" character varying NULL, "storage_provider" character varying NULL, "last_accessed_at" timestamptz NULL, "email_template_files" character varying NULL, "export_files" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "category_id" character varying NULL, "finding_files" character varying NULL, "integration_files" character varying NULL, "note_files" character varying NULL, "platform_architecture_diagrams" character varying NULL, "platform_data_flow_diagrams" character varying NULL, "platform_trust_boundary_diagrams" character varying NULL, "remediation_files" character varying NULL, "review_files" character varying NULL, "vulnerability_files" character varying NULL, PRIMARY KEY ("id"));
-- Create "file_download_tokens" table
CREATE TABLE "file_download_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NULL, "ttl" timestamptz NULL, "user_id" character varying NULL, "organization_id" character varying NULL, "file_id" character varying NULL, "secret" bytea NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "file_download_tokens_owner_id_fk" to table: "file_download_tokens"
CREATE INDEX "file_download_tokens_owner_id_fk" ON "file_download_tokens" ("owner_id");
-- Create index "file_download_tokens_token_key" to table: "file_download_tokens"
CREATE UNIQUE INDEX "file_download_tokens_token_key" ON "file_download_tokens" ("token");
-- Create index "filedownloadtoken_token" to table: "file_download_tokens"
CREATE UNIQUE INDEX "filedownloadtoken_token" ON "file_download_tokens" ("token") WHERE (deleted_at IS NULL);
-- Create "findings" table
CREATE TABLE "findings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "reviewed_by" character varying NULL, "assigned_to" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "finding_status_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_id" character varying NULL, "security_level" character varying NULL DEFAULT 'NONE', "external_owner_id" character varying NULL, "source" character varying NULL, "resource_name" character varying NULL, "display_name" character varying NULL, "state" character varying NULL, "category" character varying NULL, "categories" jsonb NULL, "finding_class" character varying NULL, "severity" character varying NULL, "numeric_severity" double precision NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "open" boolean NULL DEFAULT true, "blocks_production" boolean NULL, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "assessment_id" character varying NULL, "description" text NULL, "recommendation" text NULL, "recommended_actions" text NULL, "references" jsonb NULL, "steps_to_reproduce" jsonb NULL, "targets" jsonb NULL, "target_details" jsonb NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "event_time" timestamptz NULL, "reported_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "assigned_to_user_id" character varying NULL, "assigned_to_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "finding_status_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "finding_display_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_display_id_owner_id" ON "findings" ("display_id", "owner_id");
-- Create index "finding_external_id_external_owner_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_external_id_external_owner_id_owner_id" ON "findings" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "finding_owner_id_idx" to table: "findings"
CREATE INDEX "finding_owner_id_idx" ON "findings" ("owner_id");
-- Create "finding_controls" table
CREATE TABLE "finding_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "external_standard" character varying NULL, "external_standard_version" character varying NULL, "external_control_id" character varying NULL, "source" character varying NULL, "metadata" jsonb NULL, "discovered_at" timestamptz NULL, "finding_id" character varying NOT NULL, "control_id" character varying NOT NULL, "standard_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "finding_control_control_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_control_id_idx" ON "finding_controls" ("control_id");
-- Create index "finding_control_owner_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_owner_id_idx" ON "finding_controls" ("owner_id");
-- Create index "finding_control_standard_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_standard_id_idx" ON "finding_controls" ("standard_id");
-- Create index "findingcontrol_finding_id_control_id" to table: "finding_controls"
CREATE UNIQUE INDEX "findingcontrol_finding_id_control_id" ON "finding_controls" ("finding_id", "control_id");
-- Create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" citext NOT NULL, "description" character varying NULL, "is_managed" boolean NULL DEFAULT false, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', "oscal_role" character varying NULL, "oscal_party_uuid" character varying NULL, "oscal_contact_uuids" jsonb NULL, "scim_external_id" character varying NULL, "scim_display_name" character varying NULL, "scim_active" boolean NULL DEFAULT true, "scim_group_mailing" character varying NULL, "assessment_blocked_groups" character varying NULL, "assessment_editors" character varying NULL, "assessment_viewers" character varying NULL, "asset_blocked_groups" character varying NULL, "asset_editors" character varying NULL, "asset_viewers" character varying NULL, "check_result_blocked_groups" character varying NULL, "check_result_editors" character varying NULL, "check_result_viewers" character varying NULL, "email_template_blocked_groups" character varying NULL, "email_template_editors" character varying NULL, "email_template_viewers" character varying NULL, "avatar_local_file_id" character varying NULL, "identity_holder_blocked_groups" character varying NULL, "identity_holder_editors" character varying NULL, "identity_holder_viewers" character varying NULL, "organization_action_plan_creators" character varying NULL, "organization_api_token_creators" character varying NULL, "organization_assessment_creators" character varying NULL, "organization_asset_creators" character varying NULL, "organization_campaign_creators" character varying NULL, "organization_campaign_target_creators" character varying NULL, "organization_check_result_creators" character varying NULL, "organization_contact_creators" character varying NULL, "organization_control_creators" character varying NULL, "organization_control_implementation_creators" character varying NULL, "organization_control_objective_creators" character varying NULL, "organization_custom_domain_creators" character varying NULL, "organization_custom_type_enum_creators" character varying NULL, "organization_directory_account_creators" character varying NULL, "organization_directory_group_creators" character varying NULL, "organization_directory_membership_creators" character varying NULL, "organization_directory_sync_run_creators" character varying NULL, "organization_discussion_creators" character varying NULL, "organization_document_data_creators" character varying NULL, "organization_email_template_creators" character varying NULL, "organization_entity_creators" character varying NULL, "organization_entity_type_creators" character varying NULL, "organization_evidence_creators" character varying NULL, "organization_file_creators" character varying NULL, "organization_finding_creators" character varying NULL, "organization_finding_control_creators" character varying NULL, "organization_group_creators" character varying NULL, "organization_group_membership_creators" character varying NULL, "organization_group_setting_creators" character varying NULL, "organization_hush_creators" character varying NULL, "organization_identity_holder_creators" character varying NULL, "organization_internal_policy_creators" character varying NULL, "organization_invite_creators" character varying NULL, "organization_job_runner_creators" character varying NULL, "organization_job_runner_registration_token_creators" character varying NULL, "organization_job_runner_token_creators" character varying NULL, "organization_job_template_creators" character varying NULL, "organization_mapped_control_creators" character varying NULL, "organization_narrative_creators" character varying NULL, "organization_note_creators" character varying NULL, "organization_notification_template_creators" character varying NULL, "organization_org_membership_creators" character varying NULL, "organization_platform_creators" character varying NULL, "organization_procedure_creators" character varying NULL, "organization_program_creators" character varying NULL, "organization_program_membership_creators" character varying NULL, "organization_remediation_creators" character varying NULL, "organization_review_creators" character varying NULL, "organization_risk_creators" character varying NULL, "organization_scan_creators" character varying NULL, "organization_scheduled_job_creators" character varying NULL, "organization_scheduled_job_run_creators" character varying NULL, "organization_sla_definition_creators" character varying NULL, "organization_standard_creators" character varying NULL, "organization_subcontrol_creators" character varying NULL, "organization_subprocessor_creators" character varying NULL, "organization_subscriber_creators" character varying NULL, "organization_system_detail_creators" character varying NULL, "organization_tag_definition_creators" character varying NULL, "organization_task_creators" character varying NULL, "organization_template_creators" character varying NULL, "organization_trust_center_creators" character varying NULL, "organization_trust_center_compliance_creators" character varying NULL, "organization_trust_center_doc_creators" character varying NULL, "organization_trust_center_entity_creators" character varying NULL, "organization_trust_center_faq_creators" character varying NULL, "organization_trust_center_nda_request_creators" character varying NULL, "organization_trust_center_subprocessor_creators" character varying NULL, "organization_trust_center_watermark_config_creators" character varying NULL, "organization_vendor_risk_score_creators" character varying NULL, "organization_vendor_scoring_config_creators" character varying NULL, "organization_vulnerability_creators" character varying NULL, "organization_workflow_definition_creators" character varying NULL, "organization_campaigns_manager" character varying NULL, "organization_compliance_manager" character varying NULL, "organization_group_manager" character varying NULL, "organization_policies_manager" character varying NULL, "organization_registry_manager" character varying NULL, "organization_risk_manager" character varying NULL, "organization_trust_center_manager" character varying NULL, "organization_workflows_manager" character varying NULL, "owner_id" character varying NULL, "sla_definition_blocked_groups" character varying NULL, "sla_definition_editors" character varying NULL, "trust_center_blocked_groups" character varying NULL, "trust_center_editors" character varying NULL, "trust_center_compliance_blocked_groups" character varying NULL, "trust_center_compliance_editors" character varying NULL, "trust_center_doc_blocked_groups" character varying NULL, "trust_center_doc_editors" character varying NULL, "trust_center_entity_blocked_groups" character varying NULL, "trust_center_entity_editors" character varying NULL, "trust_center_faq_blocked_groups" character varying NULL, "trust_center_faq_editors" character varying NULL, "trust_center_nda_request_blocked_groups" character varying NULL, "trust_center_nda_request_editors" character varying NULL, "trust_center_setting_blocked_groups" character varying NULL, "trust_center_setting_editors" character varying NULL, "trust_center_subprocessor_blocked_groups" character varying NULL, "trust_center_subprocessor_editors" character varying NULL, "trust_center_watermark_config_blocked_groups" character varying NULL, "trust_center_watermark_config_editors" character varying NULL, "vulnerability_blocked_groups" character varying NULL, "vulnerability_editors" character varying NULL, "vulnerability_viewers" character varying NULL, "workflow_definition_blocked_groups" character varying NULL, "workflow_definition_editors" character varying NULL, "workflow_definition_viewers" character varying NULL, "workflow_definition_groups" character varying NULL, PRIMARY KEY ("id"));
-- Create index "group_avatar_local_file_id_idx" to table: "groups"
CREATE INDEX "group_avatar_local_file_id_idx" ON "groups" ("avatar_local_file_id");
-- Create index "group_display_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_display_id_owner_id" ON "groups" ("display_id", "owner_id");
-- Create index "group_name_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_name_owner_id" ON "groups" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "group_owner_id_idx" to table: "groups"
CREATE INDEX "group_owner_id_idx" ON "groups" ("owner_id");
-- Create "group_memberships" table
CREATE TABLE "group_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, "group_membership_org_membership" character varying NULL, PRIMARY KEY ("id"));
-- Create index "group_membership_group_id_idx" to table: "group_memberships"
CREATE INDEX "group_membership_group_id_idx" ON "group_memberships" ("group_id");
-- Create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id");
-- Create "group_settings" table
CREATE TABLE "group_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "group_setting_group_id_idx" to table: "group_settings"
CREATE INDEX "group_setting_group_id_idx" ON "group_settings" ("group_id");
-- Create index "group_settings_group_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_group_id_key" ON "group_settings" ("group_id");
-- Create "hushes" table
CREATE TABLE "hushes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, "credential_set" jsonb NULL, "metadata" jsonb NULL, "last_used_at" timestamptz NULL, "expires_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "secret_owner_id_idx" to table: "hushes"
CREATE INDEX "secret_owner_id_idx" ON "hushes" ("owner_id");
-- Create "identity_holders" table
CREATE TABLE "identity_holders" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "full_name" character varying NOT NULL, "email" character varying NOT NULL, "alternate_email" character varying NULL, "email_aliases" jsonb NULL, "phone_number" character varying NULL, "is_openlane_user" boolean NULL DEFAULT false, "identity_holder_type" character varying NOT NULL DEFAULT 'UNSPECIFIED', "status" character varying NOT NULL DEFAULT 'ACTIVE', "is_active" boolean NOT NULL DEFAULT true, "title" character varying NULL, "department" character varying NULL, "team" character varying NULL, "location" character varying NULL, "start_date" timestamptz NULL, "end_date" timestamptz NULL, "external_user_id" character varying NULL, "external_reference_id" character varying NULL, "metadata" jsonb NULL, "avatar_remote_url" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "employer_entity_id" character varying NULL, "owner_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "identity_holder_employer_entity_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_employer_entity_id_idx" ON "identity_holders" ("employer_entity_id");
-- Create index "identity_holder_owner_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_owner_id_idx" ON "identity_holders" ("owner_id");
-- Create index "identityholder_display_id_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_display_id_owner_id" ON "identity_holders" ("display_id", "owner_id");
-- Create index "identityholder_email_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_email_owner_id" ON "identity_holders" ("email", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "identityholder_external_user_id" to table: "identity_holders"
CREATE INDEX "identityholder_external_user_id" ON "identity_holders" ("external_user_id");
-- Create index "identityholder_user_id" to table: "identity_holders"
CREATE INDEX "identityholder_user_id" ON "identity_holders" ("user_id");
-- Create "impersonation_events" table
CREATE TABLE "impersonation_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "impersonation_type" character varying NOT NULL, "action" character varying NOT NULL, "reason" character varying NULL, "ip_address" character varying NULL, "user_agent" character varying NULL, "scopes" jsonb NULL, "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, "target_user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "impersonation_event_organization_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_organization_id_idx" ON "impersonation_events" ("organization_id");
-- Create index "impersonation_event_target_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_target_user_id_idx" ON "impersonation_events" ("target_user_id");
-- Create index "impersonation_event_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_user_id_idx" ON "impersonation_events" ("user_id");
-- Create "integrations" table
CREATE TABLE "integrations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "integration_type" character varying NULL, "provider_metadata" jsonb NULL, "config" jsonb NULL, "installation_metadata" jsonb NULL, "provider_state" jsonb NULL, "metadata" jsonb NULL, "definition_id" character varying NULL, "definition_version" character varying NULL, "definition_slug" character varying NULL, "family" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "provider_metadata_snapshot" jsonb NULL, "primary_directory" boolean NOT NULL DEFAULT false, "campaign_email" boolean NOT NULL DEFAULT false, "file_integrations" character varying NULL, "group_integrations" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "integration_owner_id_idx" to table: "integrations"
CREATE INDEX "integration_owner_id_idx" ON "integrations" ("owner_id");
-- Create index "integration_platform_id_idx" to table: "integrations"
CREATE INDEX "integration_platform_id_idx" ON "integrations" ("platform_id");
-- Create "integration_runs" table
CREATE TABLE "integration_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "operation_name" character varying NULL, "operation_kind" character varying NULL, "run_type" character varying NULL, "operation_config" jsonb NULL, "mapping_version" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "started_at" timestamptz NOT NULL, "finished_at" timestamptz NULL, "duration_ms" bigint NULL, "summary" character varying NULL, "error" text NULL, "metrics" jsonb NULL, "integration_id" character varying NULL, "request_file_id" character varying NULL, "response_file_id" character varying NULL, "event_id" character varying NULL, "assessment_response_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "integration_run_event_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_event_id_idx" ON "integration_runs" ("event_id");
-- Create index "integration_run_owner_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_owner_id_idx" ON "integration_runs" ("owner_id");
-- Create index "integration_run_request_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_request_file_id_idx" ON "integration_runs" ("request_file_id");
-- Create index "integration_run_response_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_response_file_id_idx" ON "integration_runs" ("response_file_id");
-- Create index "integrationrun_assessment_response_id_operation_name" to table: "integration_runs"
CREATE UNIQUE INDEX "integrationrun_assessment_response_id_operation_name" ON "integration_runs" ("assessment_response_id", "operation_name") WHERE ((deleted_at IS NULL) AND (assessment_response_id IS NOT NULL));
-- Create index "integrationrun_assessment_response_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_assessment_response_id_started_at" ON "integration_runs" ("assessment_response_id", "started_at") WHERE (deleted_at IS NULL);
-- Create index "integrationrun_integration_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_integration_id_started_at" ON "integration_runs" ("integration_id", "started_at") WHERE (deleted_at IS NULL);
-- Create "integration_webhooks" table
CREATE TABLE "integration_webhooks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "provider" character varying NOT NULL, "name" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "endpoint_id" character varying NULL, "endpoint_url" character varying NULL, "secret_token" character varying NULL, "allowed_events" jsonb NULL, "last_delivery_id" character varying NULL, "last_delivery_at" timestamptz NULL, "last_delivery_status" character varying NULL, "last_delivery_error" text NULL, "external_event_id" character varying NULL, "metadata" jsonb NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "integration_webhook_owner_id_idx" to table: "integration_webhooks"
CREATE INDEX "integration_webhook_owner_id_idx" ON "integration_webhooks" ("owner_id");
-- Create index "integrationwebhook_endpoint_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_endpoint_id" ON "integration_webhooks" ("endpoint_id") WHERE ((deleted_at IS NULL) AND (endpoint_id IS NOT NULL));
-- Create index "integrationwebhook_integration_id_name_external_event_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_integration_id_name_external_event_id" ON "integration_webhooks" ("integration_id", "name", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- Create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED', "details" text NULL, "details_json" jsonb NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "summary" character varying NULL, "tag_suggestions" jsonb NULL, "dismissed_tag_suggestions" jsonb NULL, "control_suggestions" jsonb NULL, "dismissed_control_suggestions" jsonb NULL, "improvement_suggestions" jsonb NULL, "dismissed_improvement_suggestions" jsonb NULL, "url" character varying NULL, "external_file_id" character varying NULL, "external_contents" character varying NULL, "internal_policy_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "custom_type_enum_internal_policies" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "internal_policy_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "file_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "internal_policy_file_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_file_id_idx" ON "internal_policies" ("file_id");
-- Create index "internal_policy_owner_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_owner_id_idx" ON "internal_policies" ("owner_id");
-- Create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_display_id_owner_id" ON "internal_policies" ("display_id", "owner_id");
-- Create index "internalpolicy_external_uuid_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_external_uuid_owner_id" ON "internal_policies" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create "invites" table
CREATE TABLE "invites" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "requestor_id" character varying NULL, "token" character varying NOT NULL, "expires" timestamptz NULL, "recipient" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'INVITATION_SENT', "role" character varying NOT NULL DEFAULT 'MEMBER', "send_attempts" bigint NOT NULL DEFAULT 1, "secret" bytea NOT NULL, "ownership_transfer" boolean NULL DEFAULT false, "sso_exempt" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "invite_owner_id_idx" to table: "invites"
CREATE INDEX "invite_owner_id_idx" ON "invites" ("owner_id");
-- Create index "invite_recipient_owner_id" to table: "invites"
CREATE UNIQUE INDEX "invite_recipient_owner_id" ON "invites" ("recipient", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "invites_token_key" to table: "invites"
CREATE UNIQUE INDEX "invites_token_key" ON "invites" ("token");
-- Create "job_results" table
CREATE TABLE "job_results" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "status" character varying NOT NULL, "exit_code" bigint NOT NULL, "finished_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "log" text NULL, "scheduled_job_id" character varying NOT NULL, "file_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "job_result_file_id_idx" to table: "job_results"
CREATE INDEX "job_result_file_id_idx" ON "job_results" ("file_id");
-- Create index "job_result_owner_id_idx" to table: "job_results"
CREATE INDEX "job_result_owner_id_idx" ON "job_results" ("owner_id");
-- Create index "job_result_scheduled_job_id_idx" to table: "job_results"
CREATE INDEX "job_result_scheduled_job_id_idx" ON "job_results" ("scheduled_job_id");
-- Create "job_runners" table
CREATE TABLE "job_runners" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'OFFLINE', "ip_address" character varying NULL, "last_seen" timestamptz NULL, "version" character varying NULL, "os" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "job_runner_owner_id_idx" to table: "job_runners"
CREATE INDEX "job_runner_owner_id_idx" ON "job_runners" ("owner_id");
-- Create index "jobrunner_display_id_owner_id" to table: "job_runners"
CREATE UNIQUE INDEX "jobrunner_display_id_owner_id" ON "job_runners" ("display_id", "owner_id");
-- Create "job_runner_registration_tokens" table
CREATE TABLE "job_runner_registration_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "token" character varying NOT NULL, "expires_at" timestamptz NOT NULL, "last_used_at" timestamptz NULL, "job_runner_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "job_runner_registration_token_job_runner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_job_runner_id_idx" ON "job_runner_registration_tokens" ("job_runner_id");
-- Create index "job_runner_registration_token_owner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_owner_id_idx" ON "job_runner_registration_tokens" ("owner_id");
-- Create index "job_runner_registration_tokens_token_key" to table: "job_runner_registration_tokens"
CREATE UNIQUE INDEX "job_runner_registration_tokens_token_key" ON "job_runner_registration_tokens" ("token");
-- Create "job_runner_tokens" table
CREATE TABLE "job_runner_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "job_runner_token_owner_id_idx" to table: "job_runner_tokens"
CREATE INDEX "job_runner_token_owner_id_idx" ON "job_runner_tokens" ("owner_id");
-- Create index "job_runner_tokens_token_key" to table: "job_runner_tokens"
CREATE UNIQUE INDEX "job_runner_tokens_token_key" ON "job_runner_tokens" ("token");
-- Create index "jobrunnertoken_token_expires_at_is_active" to table: "job_runner_tokens"
CREATE INDEX "jobrunnertoken_token_expires_at_is_active" ON "job_runner_tokens" ("token", "expires_at", "is_active");
-- Create "job_templates" table
CREATE TABLE "job_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "title" character varying NOT NULL, "description" character varying NULL, "platform" character varying NOT NULL, "windmill_path" character varying NULL, "download_url" character varying NOT NULL, "configuration" jsonb NULL, "cron" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "job_template_owner_id_idx" to table: "job_templates"
CREATE INDEX "job_template_owner_id_idx" ON "job_templates" ("owner_id");
-- Create index "jobtemplate_display_id_owner_id" to table: "job_templates"
CREATE UNIQUE INDEX "jobtemplate_display_id_owner_id" ON "job_templates" ("display_id", "owner_id");
-- Create "mappable_domains" table
CREATE TABLE "mappable_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "zone_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "mappabledomain_name" to table: "mappable_domains"
CREATE UNIQUE INDEX "mappabledomain_name" ON "mappable_domains" ("name") WHERE (deleted_at IS NULL);
-- Create "mapped_controls" table
CREATE TABLE "mapped_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "mapping_type" character varying NOT NULL DEFAULT 'EQUAL', "relation" character varying NULL, "confidence" bigint NULL, "source" character varying NULL DEFAULT 'MANUAL', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "mapped_control_owner_id_idx" to table: "mapped_controls"
CREATE INDEX "mapped_control_owner_id_idx" ON "mapped_controls" ("owner_id");
-- Create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "details" text NULL, "control_objective_narratives" character varying NULL, "owner_id" character varying NULL, "subcontrol_narratives" character varying NULL, PRIMARY KEY ("id"));
-- Create index "narrative_display_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_display_id_owner_id" ON "narratives" ("display_id", "owner_id");
-- Create index "narrative_owner_id_idx" to table: "narratives"
CREATE INDEX "narrative_owner_id_idx" ON "narratives" ("owner_id");
-- Create "notes" table
CREATE TABLE "notes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "title" character varying NULL, "text" text NOT NULL, "text_json" jsonb NULL, "note_ref" character varying NULL, "is_edited" boolean NOT NULL DEFAULT false, "notify_subscribers" boolean NULL DEFAULT false, "notified_at" timestamptz NULL, "control_comments" character varying NULL, "discussion_id" character varying NULL, "entity_notes" character varying NULL, "evidence_comments" character varying NULL, "finding_comments" character varying NULL, "internal_policy_comments" character varying NULL, "owner_id" character varying NULL, "procedure_comments" character varying NULL, "program_notes" character varying NULL, "remediation_comments" character varying NULL, "review_comments" character varying NULL, "risk_comments" character varying NULL, "subcontrol_comments" character varying NULL, "task_comments" character varying NULL, "trust_center_id" character varying NULL, "vulnerability_comments" character varying NULL, PRIMARY KEY ("id"));
-- Create index "note_discussion_id_idx" to table: "notes"
CREATE INDEX "note_discussion_id_idx" ON "notes" ("discussion_id");
-- Create index "note_display_id_owner_id" to table: "notes"
CREATE UNIQUE INDEX "note_display_id_owner_id" ON "notes" ("display_id", "owner_id");
-- Create index "note_owner_id_idx" to table: "notes"
CREATE INDEX "note_owner_id_idx" ON "notes" ("owner_id");
-- Create index "note_trust_center_id_idx" to table: "notes"
CREATE INDEX "note_trust_center_id_idx" ON "notes" ("trust_center_id");
-- Create "notifications" table
CREATE TABLE "notifications" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "tags" jsonb NULL, "user_id" character varying NULL, "notification_type" character varying NOT NULL, "object_type" character varying NOT NULL, "title" character varying NOT NULL, "body" text NOT NULL, "data" jsonb NULL, "read_at" timestamptz NULL, "channels" jsonb NULL, "topic" character varying NULL, "template_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "notification_owner_id_idx" to table: "notifications"
CREATE INDEX "notification_owner_id_idx" ON "notifications" ("owner_id");
-- Create index "notification_template_id_idx" to table: "notifications"
CREATE INDEX "notification_template_id_idx" ON "notifications" ("template_id");
-- Create index "notification_user_id_read_at_owner_id" to table: "notifications"
CREATE INDEX "notification_user_id_read_at_owner_id" ON "notifications" ("user_id", "read_at", "owner_id");
-- Create "notification_preferences" table
CREATE TABLE "notification_preferences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "channel" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'ENABLED', "provider" character varying NULL, "destination" character varying NULL, "config" jsonb NULL, "enabled" boolean NOT NULL DEFAULT true, "cadence" character varying NOT NULL DEFAULT 'IMMEDIATE', "priority" character varying NULL, "topic_patterns" jsonb NULL, "topic_overrides" jsonb NULL, "mute_until" timestamptz NULL, "quiet_hours_start" character varying NULL, "quiet_hours_end" character varying NULL, "timezone" character varying NULL, "is_default" boolean NOT NULL DEFAULT false, "verified_at" timestamptz NULL, "last_used_at" timestamptz NULL, "last_error" text NULL, "metadata" jsonb NULL, "user_id" character varying NOT NULL, "template_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "notification_preference_owner_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_owner_id_idx" ON "notification_preferences" ("owner_id");
-- Create index "notification_preference_template_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_template_id_idx" ON "notification_preferences" ("template_id");
-- Create index "notification_preference_user_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_user_id_idx" ON "notification_preferences" ("user_id");
-- Create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
CREATE INDEX "notificationpreference_owner_id_user_id_channel" ON "notification_preferences" ("owner_id", "user_id", "channel") WHERE (deleted_at IS NULL);
-- Create "notification_templates" table
CREATE TABLE "notification_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "channel" character varying NULL, "format" character varying NOT NULL DEFAULT 'MARKDOWN', "locale" character varying NOT NULL DEFAULT 'en-US', "topic_pattern" character varying NOT NULL, "destinations" jsonb NULL, "title_template" character varying NULL, "subject_template" character varying NULL, "body_template" text NULL, "blocks" jsonb NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "template_context" character varying NULL, "defaults" jsonb NULL, "email_template_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "workflow_definition_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "notification_template_email_template_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_email_template_id_idx" ON "notification_templates" ("email_template_id");
-- Create index "notification_template_integration_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_integration_id_idx" ON "notification_templates" ("integration_id");
-- Create index "notification_template_owner_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_owner_id_idx" ON "notification_templates" ("owner_id");
-- Create index "notification_template_workflow_definition_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_workflow_definition_id_idx" ON "notification_templates" ("workflow_definition_id");
-- Create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- Create index "notificationtemplate_owner_id_key" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_key" ON "notification_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- Create "onboardings" table
CREATE TABLE "onboardings" ("id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "company_name" character varying NOT NULL, "domains" jsonb NULL, "company_details" jsonb NULL, "user_details" jsonb NULL, "compliance" jsonb NULL, "demo_requested" boolean NULL DEFAULT false, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "org_memberships" table
CREATE TABLE "org_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "sso_exempt" boolean NULL DEFAULT false, "sso_exempt_reason" character varying NULL, "sso_exempt_granted_by" character varying NULL, "sso_exempt_granted_at" timestamptz NULL, "tfa_enforced" boolean NULL DEFAULT false, "tfa_enforced_reason" character varying NULL, "tfa_enforced_by" character varying NULL, "tfa_enforced_at" timestamptz NULL, "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "org_membership_organization_id_idx" to table: "org_memberships"
CREATE INDEX "org_membership_organization_id_idx" ON "org_memberships" ("organization_id");
-- Create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id");
-- Create "org_modules" table
CREATE TABLE "org_modules" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "module" character varying NOT NULL, "price" jsonb NULL, "stripe_price_id" character varying NULL, "status" character varying NULL, "visibility" character varying NULL, "active" boolean NOT NULL DEFAULT true, "module_lookup_key" character varying NULL, "price_id" character varying NULL, "org_product_org_modules" character varying NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "org_module_owner_id_idx" to table: "org_modules"
CREATE INDEX "org_module_owner_id_idx" ON "org_modules" ("owner_id");
-- Create index "org_module_subscription_id_idx" to table: "org_modules"
CREATE INDEX "org_module_subscription_id_idx" ON "org_modules" ("subscription_id");
-- Create "org_prices" table
CREATE TABLE "org_prices" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "price" jsonb NULL, "stripe_price_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "product_id" character varying NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "org_price_owner_id_idx" to table: "org_prices"
CREATE INDEX "org_price_owner_id_idx" ON "org_prices" ("owner_id");
-- Create index "org_price_subscription_id_idx" to table: "org_prices"
CREATE INDEX "org_price_subscription_id_idx" ON "org_prices" ("subscription_id");
-- Create "org_products" table
CREATE TABLE "org_products" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "module" character varying NOT NULL, "stripe_product_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "price_id" character varying NULL, "org_module_org_products" character varying NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "org_product_owner_id_idx" to table: "org_products"
CREATE INDEX "org_product_owner_id_idx" ON "org_products" ("owner_id");
-- Create index "org_product_subscription_id_idx" to table: "org_products"
CREATE INDEX "org_product_subscription_id_idx" ON "org_products" ("subscription_id");
-- Create "org_subscriptions" table
CREATE TABLE "org_subscriptions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "stripe_subscription_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "expires_at" timestamptz NULL, "trial_expires_at" timestamptz NULL, "days_until_due" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "org_subscription_owner_id_idx" to table: "org_subscriptions"
CREATE INDEX "org_subscription_owner_id_idx" ON "org_subscriptions" ("owner_id");
-- Create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "stripe_customer_id" character varying NULL, "slug_name" character varying NULL, "parent_organization_id" character varying NULL, "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "organization_avatar_local_file_id_idx" to table: "organizations"
CREATE INDEX "organization_avatar_local_file_id_idx" ON "organizations" ("avatar_local_file_id");
-- Create index "organization_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_name" ON "organizations" ("name") WHERE (deleted_at IS NULL);
-- Create index "organization_parent_organization_id_idx" to table: "organizations"
CREATE INDEX "organization_parent_organization_id_idx" ON "organizations" ("parent_organization_id");
-- Create index "organizations_stripe_customer_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_stripe_customer_id_key" ON "organizations" ("stripe_customer_id");
-- Create "organization_settings" table
CREATE TABLE "organization_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" jsonb NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "billing_notifications_enabled" boolean NOT NULL DEFAULT true, "allowed_email_domains" jsonb NULL, "allow_matching_domains_autojoin" boolean NULL DEFAULT false, "identity_provider" character varying NULL DEFAULT 'NONE', "identity_provider_client_id" character varying NULL, "identity_provider_client_secret" character varying NULL, "identity_provider_metadata_endpoint" character varying NULL, "identity_provider_auth_tested" boolean NOT NULL DEFAULT false, "identity_provider_entity_id" character varying NULL, "oidc_discovery_endpoint" character varying NULL, "saml_signin_url" character varying NULL, "saml_issuer" character varying NULL, "saml_cert" text NULL, "identity_provider_login_enforced" boolean NOT NULL DEFAULT false, "identity_provider_jit_provisioning" boolean NOT NULL DEFAULT true, "jit_allowed_email_domains" jsonb NULL, "multifactor_auth_enforced" boolean NULL DEFAULT false, "sso_exempt_domains" jsonb NULL, "allow_support_access" boolean NULL DEFAULT false, "compliance_webhook_token" character varying NULL, "payment_method_added" boolean NOT NULL DEFAULT false, "pending_deletion_at" timestamptz NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "organization_setting_organization_id_idx" to table: "organization_settings"
CREATE INDEX "organization_setting_organization_id_idx" ON "organization_settings" ("organization_id");
-- Create index "organization_settings_compliance_webhook_token_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_compliance_webhook_token_key" ON "organization_settings" ("compliance_webhook_token");
-- Create index "organization_settings_organization_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_organization_id_key" ON "organization_settings" ("organization_id");
-- Create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "password_reset_tokens_owner_id_fk" to table: "password_reset_tokens"
CREATE INDEX "password_reset_tokens_owner_id_fk" ON "password_reset_tokens" ("owner_id");
-- Create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
-- Create index "passwordresettoken_token" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_token" ON "password_reset_tokens" ("token") WHERE (deleted_at IS NULL);
-- Create "personal_access_tokens" table
CREATE TABLE "personal_access_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "sso_authorizations" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "personal_access_tokens_owner_id_fk" to table: "personal_access_tokens"
CREATE INDEX "personal_access_tokens_owner_id_fk" ON "personal_access_tokens" ("owner_id");
-- Create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_token_key" ON "personal_access_tokens" ("token");
-- Create index "personalaccesstoken_token" to table: "personal_access_tokens"
CREATE INDEX "personalaccesstoken_token" ON "personal_access_tokens" ("token");
-- Create "platforms" table
CREATE TABLE "platforms" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "business_owner" character varying NULL, "technical_owner" character varying NULL, "security_owner" character varying NULL, "platform_kind_name" character varying NULL, "platform_data_classification_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "access_model_name" character varying NULL, "encryption_status_name" character varying NULL, "security_tier_name" character varying NULL, "criticality_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "business_purpose" character varying NULL, "scope_statement" text NULL, "trust_boundary_description" text NULL, "data_flow_summary" text NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "physical_location" character varying NULL, "region" character varying NULL, "contains_pii" boolean NULL DEFAULT false, "source_type" character varying NOT NULL DEFAULT 'MANUAL', "source_identifier" character varying NULL, "cost_center" character varying NULL, "estimated_monthly_cost" double precision NULL, "purchase_date" timestamptz NULL, "external_reference_id" character varying NULL, "metadata" jsonb NULL, "custom_type_enum_platforms" character varying NULL, "identity_holder_access_platforms" character varying NULL, "owner_id" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "business_owner_user_id" character varying NULL, "business_owner_group_id" character varying NULL, "technical_owner_user_id" character varying NULL, "technical_owner_group_id" character varying NULL, "security_owner_user_id" character varying NULL, "security_owner_group_id" character varying NULL, "platform_kind_id" character varying NULL, "platform_data_classification_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "access_model_id" character varying NULL, "encryption_status_id" character varying NULL, "security_tier_id" character varying NULL, "criticality_id" character varying NULL, "platform_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "platform_display_id_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_display_id_owner_id" ON "platforms" ("display_id", "owner_id");
-- Create index "platform_external_uuid_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_external_uuid_owner_id" ON "platforms" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "platform_name_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_name_owner_id" ON "platforms" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_owner_id_idx" ON "platforms" ("owner_id");
-- Create index "platform_platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_platform_owner_id_idx" ON "platforms" ("platform_owner_id");
-- Create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED', "details" text NULL, "details_json" jsonb NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "summary" character varying NULL, "tag_suggestions" jsonb NULL, "dismissed_tag_suggestions" jsonb NULL, "control_suggestions" jsonb NULL, "dismissed_control_suggestions" jsonb NULL, "improvement_suggestions" jsonb NULL, "dismissed_improvement_suggestions" jsonb NULL, "url" character varying NULL, "external_file_id" character varying NULL, "external_contents" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "procedure_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "control_objective_procedures" character varying NULL, "custom_type_enum_procedures" character varying NULL, "owner_id" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "procedure_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "file_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "procedure_display_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_display_id_owner_id" ON "procedures" ("display_id", "owner_id");
-- Create index "procedure_file_id_idx" to table: "procedures"
CREATE INDEX "procedure_file_id_idx" ON "procedures" ("file_id");
-- Create index "procedure_owner_id_idx" to table: "procedures"
CREATE INDEX "procedure_owner_id_idx" ON "procedures" ("owner_id");
-- Create "programs" table
CREATE TABLE "programs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "program_kind_name" character varying NULL, "external_uuid" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "framework_name" character varying NULL, "start_date" timestamptz NULL, "end_date" timestamptz NULL, "observation_period_start_date" timestamptz NULL, "observation_period_end_date" timestamptz NULL, "fieldwork_start_date" timestamptz NULL, "fieldwork_end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, "audit_firm" character varying NULL, "auditor" character varying NULL, "auditor_email" character varying NULL, "custom_type_enum_programs" character varying NULL, "owner_id" character varying NULL, "program_kind_id" character varying NULL, "program_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "program_display_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_display_id_owner_id" ON "programs" ("display_id", "owner_id");
-- Create index "program_external_uuid_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_external_uuid_owner_id" ON "programs" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "program_owner_id_idx" to table: "programs"
CREATE INDEX "program_owner_id_idx" ON "programs" ("owner_id");
-- Create index "program_program_owner_id_idx" to table: "programs"
CREATE INDEX "program_program_owner_id_idx" ON "programs" ("program_owner_id");
-- Create "program_memberships" table
CREATE TABLE "program_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, "program_membership_org_membership" character varying NULL, PRIMARY KEY ("id"));
-- Create index "program_membership_program_id_idx" to table: "program_memberships"
CREATE INDEX "program_membership_program_id_idx" ON "program_memberships" ("program_id");
-- Create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id");
-- Create "remediations" table
CREATE TABLE "remediations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NULL, "status" character varying NULL DEFAULT 'IN_PROGRESS', "state" character varying NULL, "intent" character varying NULL, "summary" text NULL, "explanation" text NULL, "instructions" text NULL, "owner_reference" character varying NULL, "repository_uri" character varying NULL, "pull_request_uri" character varying NULL, "ticket_reference" character varying NULL, "due_at" timestamptz NULL, "completed_at" timestamptz NULL, "pr_generated_at" timestamptz NULL, "error" text NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "owner_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "remediation_display_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_display_id_owner_id" ON "remediations" ("display_id", "owner_id");
-- Create index "remediation_external_id_external_owner_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_external_id_external_owner_id_owner_id" ON "remediations" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "remediation_owner_id_idx" to table: "remediations"
CREATE INDEX "remediation_owner_id_idx" ON "remediations" ("owner_id");
-- Create "reviews" table
CREATE TABLE "reviews" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NOT NULL, "state" character varying NULL, "status" character varying NULL DEFAULT 'OPEN', "category" character varying NULL, "classification" character varying NULL, "summary" text NULL, "details" text NULL, "reporter" character varying NULL, "approved" boolean NULL DEFAULT false, "reviewed_at" timestamptz NULL, "reported_at" timestamptz NULL, "approved_at" timestamptz NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "owner_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "reviewer_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "review_external_id_external_owner_id_owner_id" to table: "reviews"
CREATE UNIQUE INDEX "review_external_id_external_owner_id_owner_id" ON "reviews" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "review_owner_id_idx" to table: "reviews"
CREATE INDEX "review_owner_id_idx" ON "reviews" ("owner_id");
-- Create index "review_reviewer_id_idx" to table: "reviews"
CREATE INDEX "review_reviewer_id_idx" ON "reviews" ("reviewer_id");
-- Create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "risk_kind_name" character varying NULL, "risk_category_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_id" character varying NULL, "integration_id" character varying NULL, "observed_at" timestamptz NULL, "external_uuid" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL, "impact" character varying NULL, "likelihood" character varying NULL DEFAULT 'LIKELY', "score" bigint NULL, "mitigation" text NULL, "mitigation_json" jsonb NULL, "details" text NULL, "details_json" jsonb NULL, "business_costs" text NULL, "business_costs_json" jsonb NULL, "mitigated_at" timestamptz NULL, "review_required" boolean NULL DEFAULT true, "last_reviewed_at" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "due_date" timestamptz NULL, "next_review_due_at" timestamptz NULL, "residual_score" bigint NULL, "risk_decision" character varying NULL DEFAULT 'NONE', "control_objective_risks" character varying NULL, "custom_type_enum_risks" character varying NULL, "custom_type_enum_risk_categories" character varying NULL, "owner_id" character varying NULL, "risk_kind_id" character varying NULL, "risk_category_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "stakeholder_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "risk_delegate_id_idx" to table: "risks"
CREATE INDEX "risk_delegate_id_idx" ON "risks" ("delegate_id");
-- Create index "risk_display_id_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_display_id_owner_id" ON "risks" ("display_id", "owner_id");
-- Create index "risk_external_uuid_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_external_uuid_owner_id" ON "risks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "risk_owner_id_idx" to table: "risks"
CREATE INDEX "risk_owner_id_idx" ON "risks" ("owner_id");
-- Create index "risk_stakeholder_id_idx" to table: "risks"
CREATE INDEX "risk_stakeholder_id_idx" ON "risks" ("stakeholder_id");
-- Create "sla_definitions" table
CREATE TABLE "sla_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "sla_days" bigint NOT NULL, "security_level" character varying NOT NULL DEFAULT 'NONE', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "sla_definition_owner_id_idx" to table: "sla_definitions"
CREATE INDEX "sla_definition_owner_id_idx" ON "sla_definitions" ("owner_id");
-- Create index "sladefinition_display_id_owner_id" to table: "sla_definitions"
CREATE UNIQUE INDEX "sladefinition_display_id_owner_id" ON "sla_definitions" ("display_id", "owner_id");
-- Create index "sladefinition_security_level_owner_id" to table: "sla_definitions"
CREATE UNIQUE INDEX "sladefinition_security_level_owner_id" ON "sla_definitions" ("security_level", "owner_id") WHERE (deleted_at IS NULL);
-- Create "scans" table
CREATE TABLE "scans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "reviewed_by" character varying NULL, "assigned_to" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "target" character varying NOT NULL, "scan_type" character varying NOT NULL DEFAULT 'DOMAIN', "metadata" jsonb NULL, "scan_date" timestamptz NULL, "scan_schedule" character varying NULL, "next_scan_run_at" timestamptz NULL, "performed_by" character varying NULL, "discovered_vulnerability_ids" jsonb NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "owner_id" character varying NULL, "generated_by_platform_id" character varying NULL, "risk_scans" character varying NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "assigned_to_user_id" character varying NULL, "assigned_to_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "performed_by_user_id" character varying NULL, "performed_by_group_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "scan_generated_by_platform_id_idx" to table: "scans"
CREATE INDEX "scan_generated_by_platform_id_idx" ON "scans" ("generated_by_platform_id");
-- Create index "scan_owner_id_idx" to table: "scans"
CREATE INDEX "scan_owner_id_idx" ON "scans" ("owner_id");
-- Create index "scan_performed_by_group_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_group_id_idx" ON "scans" ("performed_by_group_id");
-- Create index "scan_performed_by_user_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_user_id_idx" ON "scans" ("performed_by_user_id");
-- Create "scheduled_jobs" table
CREATE TABLE "scheduled_jobs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "active" boolean NOT NULL DEFAULT true, "configuration" jsonb NULL, "cron" character varying NULL, "job_id" character varying NOT NULL, "owner_id" character varying NULL, "job_runner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "scheduled_job_job_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_id_idx" ON "scheduled_jobs" ("job_id");
-- Create index "scheduled_job_job_runner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_runner_id_idx" ON "scheduled_jobs" ("job_runner_id");
-- Create index "scheduled_job_owner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_owner_id_idx" ON "scheduled_jobs" ("owner_id");
-- Create index "scheduledjob_display_id_owner_id" to table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_display_id_owner_id" ON "scheduled_jobs" ("display_id", "owner_id");
-- Create "scheduled_job_runs" table
CREATE TABLE "scheduled_job_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "expected_execution_time" timestamptz NOT NULL, "script" character varying NOT NULL, "owner_id" character varying NULL, "scheduled_job_id" character varying NOT NULL, "job_runner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "scheduled_job_run_job_runner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_job_runner_id_idx" ON "scheduled_job_runs" ("job_runner_id");
-- Create index "scheduled_job_run_owner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_owner_id_idx" ON "scheduled_job_runs" ("owner_id");
-- Create index "scheduled_job_run_scheduled_job_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_scheduled_job_id_idx" ON "scheduled_job_runs" ("scheduled_job_id");
-- Create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "short_name" character varying NULL, "framework" text NULL, "description" text NULL, "governing_body_logo_url" character varying NULL, "governing_body" character varying NULL, "domains" jsonb NULL, "link" character varying NULL, "status" character varying NULL DEFAULT 'ACTIVE', "is_public" boolean NULL DEFAULT false, "free_to_use" boolean NULL DEFAULT false, "standard_type" character varying NULL, "version" character varying NULL, "owner_id" character varying NULL, "logo_file_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "standard_logo_file_id_idx" to table: "standards"
CREATE INDEX "standard_logo_file_id_idx" ON "standards" ("logo_file_id");
-- Create index "standard_owner_id_idx" to table: "standards"
CREATE INDEX "standard_owner_id_idx" ON "standards" ("owner_id");
-- Create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "external_uuid" character varying NULL, "title" character varying NULL, "description" text NULL, "description_json" jsonb NULL, "aliases" jsonb NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NOT_IMPLEMENTED', "implementation_status" character varying NULL DEFAULT 'PLANNED', "implementation_description" text NULL, "public_representation" text NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "source_name" character varying NULL, "reference_framework" character varying NULL, "reference_framework_revision" character varying NULL, "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "testing_procedures" jsonb NULL, "evidence_requests" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "subcontrol_kind_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "ref_code" character varying NOT NULL, "control_id" character varying NOT NULL, "custom_type_enum_subcontrols" character varying NULL, "owner_id" character varying NULL, "program_subcontrols" character varying NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "responsible_party_id" character varying NULL, "subcontrol_kind_id" character varying NULL, "user_subcontrols" character varying NULL, PRIMARY KEY ("id"));
-- Create index "subcontrol_auditor_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_auditor_reference_id_deleted_at_owner_id" ON "subcontrols" ("auditor_reference_id", "deleted_at", "owner_id");
-- Create index "subcontrol_control_id_ref_code" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_control_id_ref_code" ON "subcontrols" ("control_id", "ref_code") WHERE (deleted_at IS NULL);
-- Create index "subcontrol_control_id_ref_code_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_control_id_ref_code_owner_id" ON "subcontrols" ("control_id", "ref_code", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "subcontrol_display_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_display_id_owner_id" ON "subcontrols" ("display_id", "owner_id");
-- Create index "subcontrol_external_uuid_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_external_uuid_owner_id" ON "subcontrols" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "subcontrol_owner_id_idx" to table: "subcontrols"
CREATE INDEX "subcontrol_owner_id_idx" ON "subcontrols" ("owner_id");
-- Create index "subcontrol_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_reference_id_deleted_at_owner_id" ON "subcontrols" ("reference_id", "deleted_at", "owner_id");
-- Create "subprocessors" table
CREATE TABLE "subprocessors" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "logo_remote_url" character varying NULL, "owner_id" character varying NULL, "logo_file_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "subprocessor_logo_file_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_logo_file_id_idx" ON "subprocessors" ("logo_file_id");
-- Create index "subprocessor_name_owner_id" to table: "subprocessors"
CREATE UNIQUE INDEX "subprocessor_name_owner_id" ON "subprocessors" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "subprocessor_owner_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_owner_id_idx" ON "subprocessors" ("owner_id");
-- Create "subscribers" table
CREATE TABLE "subscribers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "phone_number" character varying NULL, "verified_email" boolean NOT NULL DEFAULT false, "verified_phone" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT false, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "secret" bytea NOT NULL, "unsubscribed" boolean NOT NULL DEFAULT false, "send_attempts" bigint NOT NULL DEFAULT 1, "contact_id" character varying NULL, "owner_id" character varying NULL, "trust_center_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "subscriber_contact_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_contact_id_idx" ON "subscribers" ("contact_id");
-- Create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NULL));
-- Create index "subscriber_email_trust_center_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_trust_center_id" ON "subscribers" ("email", "trust_center_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NOT NULL));
-- Create index "subscriber_owner_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_owner_id_idx" ON "subscribers" ("owner_id");
-- Create index "subscriber_trust_center_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_trust_center_id_idx" ON "subscribers" ("trust_center_id");
-- Create index "subscriber_user_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_user_id_idx" ON "subscribers" ("user_id");
-- Create index "subscribers_token_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_token_key" ON "subscribers" ("token");
-- Create "system_details" table
CREATE TABLE "system_details" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_name" character varying NOT NULL, "version" character varying NULL, "description" text NULL, "authorization_boundary" text NULL, "sensitivity_level" character varying NULL DEFAULT 'UNKNOWN', "last_reviewed" timestamptz NULL, "revision_history" jsonb NULL, "oscal_metadata_json" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "system_detail_owner_id_idx" to table: "system_details"
CREATE INDEX "system_detail_owner_id_idx" ON "system_details" ("owner_id");
-- Create index "systemdetail_display_id_owner_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_display_id_owner_id" ON "system_details" ("display_id", "owner_id");
-- Create "tfa_settings" table
CREATE TABLE "tfa_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tfa_secret" character varying NULL, "verified" boolean NOT NULL DEFAULT false, "recovery_codes" jsonb NULL, "phone_otp_allowed" boolean NULL DEFAULT false, "email_otp_allowed" boolean NULL DEFAULT false, "totp_allowed" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "tfa_settings_owner_id_fk" to table: "tfa_settings"
CREATE INDEX "tfa_settings_owner_id_fk" ON "tfa_settings" ("owner_id");
-- Create index "tfasetting_owner_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_owner_id" ON "tfa_settings" ("owner_id") WHERE (deleted_at IS NULL);
-- Create "tag_definitions" table
CREATE TABLE "tag_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" citext NOT NULL, "aliases" jsonb NULL, "slug" citext NULL, "description" character varying NULL, "color" character varying NULL, "owner_id" character varying NULL, "workflow_definition_tag_definitions" character varying NULL, PRIMARY KEY ("id"));
-- Create index "tag_definition_owner_id_idx" to table: "tag_definitions"
CREATE INDEX "tag_definition_owner_id_idx" ON "tag_definitions" ("owner_id");
-- Create index "tagdefinition_name_owner_id" to table: "tag_definitions"
CREATE UNIQUE INDEX "tagdefinition_name_owner_id" ON "tag_definitions" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "tagdefinition_slug_owner_id" to table: "tag_definitions"
CREATE UNIQUE INDEX "tagdefinition_slug_owner_id" ON "tag_definitions" ("slug", "owner_id") WHERE (deleted_at IS NULL);
-- Create "tasks" table
CREATE TABLE "tasks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "task_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "title" character varying NOT NULL, "details" text NULL, "details_json" jsonb NULL, "metadata" jsonb NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "due" timestamptz NULL, "completed" timestamptz NULL, "system_generated" boolean NOT NULL DEFAULT false, "is_template" boolean NOT NULL DEFAULT false, "is_suggested" boolean NOT NULL DEFAULT false, "priority" bigint NOT NULL DEFAULT 0, "source" character varying NULL, "source_key" character varying NULL, "idempotency_key" character varying NULL, "external_reference_url" jsonb NULL, "custom_type_enum_tasks" character varying NULL, "integration_tasks" character varying NULL, "owner_id" character varying NULL, "remediation_tasks" character varying NULL, "review_tasks" character varying NULL, "task_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "parent_task_id" character varying NULL, "assigner_id" character varying NULL, "assignee_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "tasks_tasks_tasks" FOREIGN KEY ("parent_task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "task_assignee_id_idx" to table: "tasks"
CREATE INDEX "task_assignee_id_idx" ON "tasks" ("assignee_id");
-- Create index "task_assigner_id_idx" to table: "tasks"
CREATE INDEX "task_assigner_id_idx" ON "tasks" ("assigner_id");
-- Create index "task_display_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_display_id_owner_id" ON "tasks" ("display_id", "owner_id");
-- Create index "task_external_uuid_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_external_uuid_owner_id" ON "tasks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "task_owner_id_idempotency_key" to table: "tasks"
CREATE UNIQUE INDEX "task_owner_id_idempotency_key" ON "tasks" ("owner_id", "idempotency_key") WHERE ((deleted_at IS NULL) AND (idempotency_key IS NOT NULL));
-- Create index "task_owner_id_idx" to table: "tasks"
CREATE INDEX "task_owner_id_idx" ON "tasks" ("owner_id");
-- Create index "task_owner_id_is_suggested_priority" to table: "tasks"
CREATE INDEX "task_owner_id_is_suggested_priority" ON "tasks" ("owner_id", "is_suggested", "priority") WHERE (deleted_at IS NULL);
-- Create index "task_parent_task_id_idx" to table: "tasks"
CREATE INDEX "task_parent_task_id_idx" ON "tasks" ("parent_task_id");
-- Create "templates" table
CREATE TABLE "templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "kind" character varying NULL DEFAULT 'QUESTIONNAIRE', "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, "transform_configuration" jsonb NULL, "owner_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "templates_check" CHECK ((trust_center_id IS NOT NULL) OR ((kind)::text <> 'TRUSTCENTER_NDA'::text)));
-- Create index "template_name_owner_id_template_type" to table: "templates"
CREATE UNIQUE INDEX "template_name_owner_id_template_type" ON "templates" ("name", "owner_id", "template_type") WHERE (deleted_at IS NULL);
-- Create index "template_owner_id_idx" to table: "templates"
CREATE INDEX "template_owner_id_idx" ON "templates" ("owner_id");
-- Create index "template_trust_center_id" to table: "templates"
CREATE UNIQUE INDEX "template_trust_center_id" ON "templates" ("trust_center_id") WHERE ((deleted_at IS NULL) AND ((kind)::text = 'TRUSTCENTER_NDA'::text));
-- Create "trust_centers" table
CREATE TABLE "trust_centers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "slug" character varying NULL, "pirsch_domain_id" character varying NULL, "pirsch_identification_code" character varying NULL, "pirsch_access_link" character varying NULL, "preview_status" character varying NULL DEFAULT 'NONE', "subprocessor_url" character varying NULL, "owner_id" character varying NULL, "custom_domain_id" character varying NULL, "preview_domain_id" character varying NULL, "trust_center_setting" character varying NULL, "trust_center_preview_setting" character varying NULL, "trust_center_watermark_config" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_custom_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_custom_domain_id_idx" ON "trust_centers" ("custom_domain_id");
-- Create index "trust_center_owner_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_owner_id_idx" ON "trust_centers" ("owner_id");
-- Create index "trust_center_preview_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_preview_domain_id_idx" ON "trust_centers" ("preview_domain_id");
-- Create index "trustcenter_slug" to table: "trust_centers"
CREATE UNIQUE INDEX "trustcenter_slug" ON "trust_centers" ("slug") WHERE (deleted_at IS NULL);
-- Create "trust_center_compliances" table
CREATE TABLE "trust_center_compliances" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "standard_id" character varying NOT NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_compliance_trust_center_id_idx" to table: "trust_center_compliances"
CREATE INDEX "trust_center_compliance_trust_center_id_idx" ON "trust_center_compliances" ("trust_center_id");
-- Create index "trustcentercompliance_standard_id_trust_center_id" to table: "trust_center_compliances"
CREATE UNIQUE INDEX "trustcentercompliance_standard_id_trust_center_id" ON "trust_center_compliances" ("standard_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- Create "trust_center_docs" table
CREATE TABLE "trust_center_docs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "trust_center_doc_kind_name" character varying NULL, "title" character varying NOT NULL, "watermarking_enabled" boolean NULL, "watermark_status" character varying NULL DEFAULT 'PENDING', "visibility" character varying NULL DEFAULT 'NOT_VISIBLE', "standard_id" character varying NULL, "trust_center_id" character varying NULL, "trust_center_doc_kind_id" character varying NULL, "file_id" character varying NULL, "original_file_id" character varying NULL, "trust_center_nda_request_trust_center_docs" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_doc_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_file_id_idx" ON "trust_center_docs" ("file_id");
-- Create index "trust_center_doc_original_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_original_file_id_idx" ON "trust_center_docs" ("original_file_id");
-- Create index "trust_center_doc_standard_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_standard_id_idx" ON "trust_center_docs" ("standard_id");
-- Create index "trust_center_doc_trust_center_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_trust_center_id_idx" ON "trust_center_docs" ("trust_center_id");
-- Create "trust_center_entities" table
CREATE TABLE "trust_center_entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "url" character varying NULL, "name" character varying NOT NULL, "file_trust_center_entities" character varying NULL, "trust_center_id" character varying NULL, "logo_file_id" character varying NULL, "entity_type_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_entity_entity_type_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_entity_type_id_idx" ON "trust_center_entities" ("entity_type_id");
-- Create index "trust_center_entity_logo_file_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_logo_file_id_idx" ON "trust_center_entities" ("logo_file_id");
-- Create index "trust_center_entity_trust_center_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_trust_center_id_idx" ON "trust_center_entities" ("trust_center_id");
-- Create "trust_center_faqs" table
CREATE TABLE "trust_center_faqs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_faq_kind_name" character varying NULL, "reference_link" character varying NULL, "display_order" bigint NULL DEFAULT 0, "note_id" character varying NOT NULL, "trust_center_id" character varying NULL, "trust_center_faq_kind_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_faq_trust_center_id_idx" to table: "trust_center_faqs"
CREATE INDEX "trust_center_faq_trust_center_id_idx" ON "trust_center_faqs" ("trust_center_id");
-- Create index "trustcenterfaq_note_id_trust_center_id" to table: "trust_center_faqs"
CREATE UNIQUE INDEX "trustcenterfaq_note_id_trust_center_id" ON "trust_center_faqs" ("note_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- Create "trust_center_nda_requests" table
CREATE TABLE "trust_center_nda_requests" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "first_name" character varying NOT NULL, "last_name" character varying NOT NULL, "email" character varying NOT NULL, "company_name" character varying NULL, "reason" character varying NULL, "access_level" character varying NULL DEFAULT 'FULL', "status" character varying NULL DEFAULT 'REQUESTED', "approved_at" timestamptz NULL, "signed_at" timestamptz NULL, "trust_center_id" character varying NULL, "document_data_id" character varying NULL, "file_id" character varying NULL, "approved_by_user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_nda_request_approved_by_user_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_approved_by_user_id_idx" ON "trust_center_nda_requests" ("approved_by_user_id");
-- Create index "trust_center_nda_request_document_data_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_document_data_id_idx" ON "trust_center_nda_requests" ("document_data_id");
-- Create index "trust_center_nda_request_file_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_file_id_idx" ON "trust_center_nda_requests" ("file_id");
-- Create index "trust_center_nda_request_trust_center_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_trust_center_id_idx" ON "trust_center_nda_requests" ("trust_center_id");
-- Create "trust_center_settings" table
CREATE TABLE "trust_center_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "title" character varying NULL, "company_name" character varying NULL, "company_description" character varying NULL, "overview" text NULL, "logo_remote_url" character varying NULL, "favicon_remote_url" character varying NULL, "theme_mode" character varying NULL DEFAULT 'EASY', "primary_color" character varying NULL, "font" character varying NULL, "foreground_color" character varying NULL, "background_color" character varying NULL, "accent_color" character varying NULL, "secondary_background_color" character varying NULL, "secondary_foreground_color" character varying NULL, "environment" character varying NULL DEFAULT 'LIVE', "remove_branding" boolean NULL DEFAULT false, "company_domain" character varying NULL, "security_contact" character varying NULL, "nda_approval_required" boolean NULL DEFAULT false, "allow_subscribers" boolean NULL DEFAULT true, "notify_subscribers_on_subprocessor_change" boolean NULL DEFAULT false, "subprocessors_notified_at" timestamptz NULL, "status_page_url" character varying NULL, "logo_local_file_id" character varying NULL, "favicon_local_file_id" character varying NULL, "hero_image_local_file_id" character varying NULL, "nda_approver_group_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_setting_favicon_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_favicon_local_file_id_idx" ON "trust_center_settings" ("favicon_local_file_id");
-- Create index "trust_center_setting_hero_image_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_hero_image_local_file_id_idx" ON "trust_center_settings" ("hero_image_local_file_id");
-- Create index "trust_center_setting_logo_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_logo_local_file_id_idx" ON "trust_center_settings" ("logo_local_file_id");
-- Create index "trust_center_setting_nda_approver_group_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_nda_approver_group_id_idx" ON "trust_center_settings" ("nda_approver_group_id");
-- Create index "trustcentersetting_trust_center_id_environment" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_trust_center_id_environment" ON "trust_center_settings" ("trust_center_id", "environment") WHERE (deleted_at IS NULL);
-- Create "trust_center_subprocessors" table
CREATE TABLE "trust_center_subprocessors" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_subprocessor_kind_name" character varying NULL, "countries" jsonb NULL, "subprocessor_id" character varying NOT NULL, "trust_center_id" character varying NULL, "trust_center_subprocessor_kind_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "trust_center_subprocessor_trust_center_id_idx" to table: "trust_center_subprocessors"
CREATE INDEX "trust_center_subprocessor_trust_center_id_idx" ON "trust_center_subprocessors" ("trust_center_id");
-- Create index "trustcentersubprocessor_subprocessor_id_trust_center_id" to table: "trust_center_subprocessors"
CREATE UNIQUE INDEX "trustcentersubprocessor_subprocessor_id_trust_center_id" ON "trust_center_subprocessors" ("subprocessor_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- Create "trust_center_watermark_configs" table
CREATE TABLE "trust_center_watermark_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "is_enabled" boolean NULL DEFAULT true, "text" character varying NULL, "font_size" double precision NULL DEFAULT 48, "opacity" double precision NULL DEFAULT 0.3, "rotation" double precision NULL DEFAULT 45, "color" character varying NULL DEFAULT '#808080', "font" character varying NULL DEFAULT 'HELVETICA', "owner_id" character varying NULL, "logo_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "text_or_logo_id_not_null" CHECK ((text IS NOT NULL) OR (logo_id IS NOT NULL)));
-- Create index "trust_center_watermark_config_logo_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_logo_id_idx" ON "trust_center_watermark_configs" ("logo_id");
-- Create index "trust_center_watermark_config_owner_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_owner_id_idx" ON "trust_center_watermark_configs" ("owner_id");
-- Create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
CREATE UNIQUE INDEX "trustcenterwatermarkconfig_trust_center_id" ON "trust_center_watermark_configs" ("trust_center_id") WHERE (deleted_at IS NULL);
-- Create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "last_login_provider" character varying NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', "scim_external_id" character varying NULL, "scim_username" character varying NULL, "scim_active" boolean NULL DEFAULT true, "scim_preferred_language" character varying NULL, "scim_locale" character varying NULL, "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "user_email" to table: "users"
CREATE UNIQUE INDEX "user_email" ON "users" ("email") WHERE (deleted_at IS NULL);
-- Create index "users_display_id_key" to table: "users"
CREATE UNIQUE INDEX "users_display_id_key" ON "users" ("display_id");
-- Create index "users_sub_key" to table: "users"
CREATE UNIQUE INDEX "users_sub_key" ON "users" ("sub");
-- Create "user_settings" table
CREATE TABLE "user_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "delegate_user_id" character varying NULL, "delegate_start_at" timestamptz NULL, "delegate_end_at" timestamptz NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, "user_id" character varying NULL, "user_setting_default_org" character varying NULL, PRIMARY KEY ("id"));
-- Create index "user_setting_user_id_idx" to table: "user_settings"
CREATE INDEX "user_setting_user_id_idx" ON "user_settings" ("user_id");
-- Create index "user_settings_user_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_user_id_key" ON "user_settings" ("user_id");
-- Create "vendor_risk_scores" table
CREATE TABLE "vendor_risk_scores" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "question_key" character varying NOT NULL, "question_name" character varying NOT NULL, "question_description" character varying NULL, "question_category" character varying NOT NULL, "answer_type" character varying NOT NULL, "impact" character varying NOT NULL, "likelihood" character varying NOT NULL, "score" double precision NOT NULL DEFAULT 0, "answer" character varying NULL, "notes" character varying NULL, "assessment_response_vendor_risk_scores" character varying NULL, "entity_vendor_risk_scores" character varying NULL, "owner_id" character varying NULL, "vendor_scoring_config_id" character varying NULL, "entity_id" character varying NOT NULL, "assessment_response_id" character varying NULL, "vendor_scoring_config_vendor_risk_scores" character varying NULL, PRIMARY KEY ("id"));
-- Create index "vendor_risk_score_assessment_response_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_assessment_response_id_idx" ON "vendor_risk_scores" ("assessment_response_id");
-- Create index "vendor_risk_score_entity_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_entity_id_idx" ON "vendor_risk_scores" ("entity_id");
-- Create index "vendor_risk_score_owner_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_owner_id_idx" ON "vendor_risk_scores" ("owner_id");
-- Create index "vendor_risk_score_vendor_scoring_config_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_vendor_scoring_config_id_idx" ON "vendor_risk_scores" ("vendor_scoring_config_id");
-- Create "vendor_scoring_configs" table
CREATE TABLE "vendor_scoring_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "questions" jsonb NOT NULL, "scoring_mode" character varying NOT NULL DEFAULT 'ANSWERED_ONLY', "risk_thresholds" jsonb NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "vendor_scoring_config_owner_id_idx" to table: "vendor_scoring_configs"
CREATE INDEX "vendor_scoring_config_owner_id_idx" ON "vendor_scoring_configs" ("owner_id");
-- Create "vulnerabilities" table
CREATE TABLE "vulnerabilities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "reviewed_by" character varying NULL, "assigned_to" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "vulnerability_status_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_owner_id" character varying NULL, "security_level" character varying NULL DEFAULT 'NONE', "external_id" character varying NOT NULL, "cve_id" character varying NULL, "source" character varying NULL, "display_name" character varying NULL, "category" character varying NULL, "severity" character varying NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "summary" text NULL, "description" text NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "open" boolean NULL DEFAULT true, "blocking" boolean NULL DEFAULT false, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "references" jsonb NULL, "impacts" jsonb NULL, "cwe_ids" jsonb NULL, "vulnerable_version_range" character varying NULL, "first_patched_version" character varying NULL, "fix_available" boolean NULL, "package_name" character varying NULL, "package_ecosystem" character varying NULL, "manifest_path" character varying NULL, "dependency_scope" character varying NULL, "published_at" timestamptz NULL, "discovered_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "dismissed_at" timestamptz NULL, "dismissed_reason" character varying NULL, "dismissed_comment" text NULL, "fixed_at" timestamptz NULL, "auto_dismissed_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "owner_id" character varying NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "assigned_to_user_id" character varying NULL, "assigned_to_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "vulnerability_status_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_cve_id_owner_id" ON "vulnerabilities" ("cve_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "vulnerability_display_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_display_id_owner_id" ON "vulnerabilities" ("display_id", "owner_id");
-- Create index "vulnerability_external_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_external_id_owner_id" ON "vulnerabilities" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "vulnerability_owner_id_idx" to table: "vulnerabilities"
CREATE INDEX "vulnerability_owner_id_idx" ON "vulnerabilities" ("owner_id");
-- Create "webauthns" table
CREATE TABLE "webauthns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "credential_id" bytea NULL, "public_key" bytea NULL, "attestation_type" character varying NULL, "aaguid" bytea NOT NULL, "sign_count" integer NOT NULL, "transports" jsonb NOT NULL, "backup_eligible" boolean NOT NULL DEFAULT false, "backup_state" boolean NOT NULL DEFAULT false, "user_present" boolean NOT NULL DEFAULT false, "user_verified" boolean NOT NULL DEFAULT false, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "webauthns_credential_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_credential_id_key" ON "webauthns" ("credential_id");
-- Create index "webauthns_owner_id_fk" to table: "webauthns"
CREATE INDEX "webauthns_owner_id_fk" ON "webauthns" ("owner_id");
-- Create "workflow_assignments" table
CREATE TABLE "workflow_assignments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "assignment_key" character varying NOT NULL, "role" character varying NOT NULL DEFAULT 'APPROVER', "label" character varying NULL, "required" boolean NOT NULL DEFAULT true, "status" character varying NOT NULL DEFAULT 'PENDING', "metadata" jsonb NULL, "approval_metadata" jsonb NULL, "rejection_metadata" jsonb NULL, "invalidation_metadata" jsonb NULL, "outcome_metadata" jsonb NULL, "decided_at" timestamptz NULL, "notes" text NULL, "due_at" timestamptz NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "actor_user_id" character varying NULL, "actor_group_id" character varying NULL, "workflow_instance_workflow_assignments" character varying NULL, PRIMARY KEY ("id"));
-- Create index "workflow_assignment_actor_group_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_group_id_idx" ON "workflow_assignments" ("actor_group_id");
-- Create index "workflow_assignment_actor_user_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_user_id_idx" ON "workflow_assignments" ("actor_user_id");
-- Create index "workflow_assignment_owner_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_owner_id_idx" ON "workflow_assignments" ("owner_id");
-- Create index "workflowassignment_display_id_owner_id" to table: "workflow_assignments"
CREATE UNIQUE INDEX "workflowassignment_display_id_owner_id" ON "workflow_assignments" ("display_id", "owner_id");
-- Create index "workflowassignment_workflow_instance_id_assignment_key" to table: "workflow_assignments"
CREATE UNIQUE INDEX "workflowassignment_workflow_instance_id_assignment_key" ON "workflow_assignments" ("workflow_instance_id", "assignment_key");
-- Create "workflow_assignment_targets" table
CREATE TABLE "workflow_assignment_targets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "target_type" character varying NOT NULL, "resolver_key" character varying NULL, "owner_id" character varying NULL, "workflow_assignment_workflow_assignment_targets" character varying NULL, "workflow_assignment_id" character varying NOT NULL, "target_user_id" character varying NULL, "target_group_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "workflow_assignment_target_owner_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_owner_id_idx" ON "workflow_assignment_targets" ("owner_id");
-- Create index "workflow_assignment_target_target_group_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_group_id_idx" ON "workflow_assignment_targets" ("target_group_id");
-- Create index "workflow_assignment_target_target_user_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_user_id_idx" ON "workflow_assignment_targets" ("target_user_id");
-- Create index "workflowassignmenttarget_display_id_owner_id" to table: "workflow_assignment_targets"
CREATE UNIQUE INDEX "workflowassignmenttarget_display_id_owner_id" ON "workflow_assignment_targets" ("display_id", "owner_id");
-- Create index "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" to table: "workflow_assignment_targets"
CREATE UNIQUE INDEX "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" ON "workflow_assignment_targets" ("workflow_assignment_id", "target_type", "target_user_id", "target_group_id", "resolver_key") WHERE (deleted_at IS NULL);
-- Create index "workflowassignmenttarget_workflow_assignment_id" to table: "workflow_assignment_targets"
CREATE INDEX "workflowassignmenttarget_workflow_assignment_id" ON "workflow_assignment_targets" ("workflow_assignment_id") WHERE (deleted_at IS NULL);
-- Create "workflow_definitions" table
CREATE TABLE "workflow_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "workflow_kind" character varying NOT NULL, "schema_type" character varying NOT NULL, "revision" bigint NOT NULL DEFAULT 1, "draft" boolean NOT NULL DEFAULT true, "published_at" timestamptz NULL, "cooldown_seconds" bigint NOT NULL DEFAULT 0, "is_default" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT true, "trigger_operations" jsonb NULL, "trigger_fields" jsonb NULL, "approval_fields" jsonb NULL, "approval_edges" jsonb NULL, "approval_submission_mode" character varying NULL DEFAULT 'AUTO_SUBMIT', "definition_json" jsonb NULL, "tracked_fields" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "workflow_definition_owner_id_idx" to table: "workflow_definitions"
CREATE INDEX "workflow_definition_owner_id_idx" ON "workflow_definitions" ("owner_id");
-- Create index "workflowdefinition_display_id_owner_id" to table: "workflow_definitions"
CREATE UNIQUE INDEX "workflowdefinition_display_id_owner_id" ON "workflow_definitions" ("display_id", "owner_id");
-- Create "workflow_events" table
CREATE TABLE "workflow_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "event_type" character varying NOT NULL, "payload" jsonb NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "workflow_instance_workflow_events" character varying NULL, PRIMARY KEY ("id"));
-- Create index "workflow_event_owner_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_owner_id_idx" ON "workflow_events" ("owner_id");
-- Create index "workflow_event_workflow_instance_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_workflow_instance_id_idx" ON "workflow_events" ("workflow_instance_id");
-- Create index "workflowevent_display_id_owner_id" to table: "workflow_events"
CREATE UNIQUE INDEX "workflowevent_display_id_owner_id" ON "workflow_events" ("display_id", "owner_id");
-- Create "workflow_instances" table
CREATE TABLE "workflow_instances" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "state" character varying NOT NULL DEFAULT 'RUNNING', "context" jsonb NULL, "last_evaluated_at" timestamptz NULL, "definition_snapshot" jsonb NULL, "current_action_index" bigint NOT NULL DEFAULT 0, "owner_id" character varying NULL, "workflow_definition_id" character varying NOT NULL, "control_id" character varying NULL, "internal_policy_id" character varying NULL, "evidence_id" character varying NULL, "subcontrol_id" character varying NULL, "action_plan_id" character varying NULL, "procedure_id" character varying NULL, "campaign_id" character varying NULL, "campaign_target_id" character varying NULL, "identity_holder_id" character varying NULL, "platform_id" character varying NULL, "assessment_id" character varying NULL, "assessment_response_id" character varying NULL, "finding_id" character varying NULL, "integration_id" character varying NULL, "remediation_id" character varying NULL, "risk_id" character varying NULL, "task_id" character varying NULL, "vulnerability_id" character varying NULL, "workflow_proposal_id" character varying NULL, PRIMARY KEY ("id"));
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
-- Create index "workflowinstance_display_id_owner_id" to table: "workflow_instances"
CREATE UNIQUE INDEX "workflowinstance_display_id_owner_id" ON "workflow_instances" ("display_id", "owner_id");
-- Create index "workflowinstance_workflow_definition_id" to table: "workflow_instances"
CREATE INDEX "workflowinstance_workflow_definition_id" ON "workflow_instances" ("workflow_definition_id") WHERE (deleted_at IS NULL);
-- Create "workflow_object_refs" table
CREATE TABLE "workflow_object_refs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "workflow_instance_workflow_object_refs" character varying NULL, "workflow_instance_id" character varying NOT NULL, "control_id" character varying NULL, "task_id" character varying NULL, "internal_policy_id" character varying NULL, "finding_id" character varying NULL, "directory_account_id" character varying NULL, "directory_group_id" character varying NULL, "directory_membership_id" character varying NULL, "evidence_id" character varying NULL, "subcontrol_id" character varying NULL, "action_plan_id" character varying NULL, "procedure_id" character varying NULL, "campaign_id" character varying NULL, "campaign_target_id" character varying NULL, "identity_holder_id" character varying NULL, "platform_id" character varying NULL, "vulnerability_id" character varying NULL, "risk_id" character varying NULL, "assessment_id" character varying NULL, "assessment_response_id" character varying NULL, "remediation_id" character varying NULL, PRIMARY KEY ("id"));
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
-- Create index "workflowobjectref_display_id_owner_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_display_id_owner_id" ON "workflow_object_refs" ("display_id", "owner_id");
-- Create index "workflowobjectref_workflow_instance_id_action_plan_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_action_plan_id" ON "workflow_object_refs" ("workflow_instance_id", "action_plan_id");
-- Create index "workflowobjectref_workflow_instance_id_assessment_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_id");
-- Create index "workflowobjectref_workflow_instance_id_assessment_response_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_response_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_response_id");
-- Create index "workflowobjectref_workflow_instance_id_campaign_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_campaign_id" ON "workflow_object_refs" ("workflow_instance_id", "campaign_id");
-- Create index "workflowobjectref_workflow_instance_id_campaign_target_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_campaign_target_id" ON "workflow_object_refs" ("workflow_instance_id", "campaign_target_id");
-- Create index "workflowobjectref_workflow_instance_id_control_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_control_id" ON "workflow_object_refs" ("workflow_instance_id", "control_id");
-- Create index "workflowobjectref_workflow_instance_id_directory_account_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_account_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_account_id");
-- Create index "workflowobjectref_workflow_instance_id_directory_group_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_group_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_group_id");
-- Create index "workflowobjectref_workflow_instance_id_directory_membership_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_membership_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_membership_id");
-- Create index "workflowobjectref_workflow_instance_id_evidence_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_evidence_id" ON "workflow_object_refs" ("workflow_instance_id", "evidence_id");
-- Create index "workflowobjectref_workflow_instance_id_finding_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_finding_id" ON "workflow_object_refs" ("workflow_instance_id", "finding_id");
-- Create index "workflowobjectref_workflow_instance_id_identity_holder_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_identity_holder_id" ON "workflow_object_refs" ("workflow_instance_id", "identity_holder_id");
-- Create index "workflowobjectref_workflow_instance_id_internal_policy_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_internal_policy_id" ON "workflow_object_refs" ("workflow_instance_id", "internal_policy_id");
-- Create index "workflowobjectref_workflow_instance_id_platform_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_platform_id" ON "workflow_object_refs" ("workflow_instance_id", "platform_id");
-- Create index "workflowobjectref_workflow_instance_id_procedure_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_procedure_id" ON "workflow_object_refs" ("workflow_instance_id", "procedure_id");
-- Create index "workflowobjectref_workflow_instance_id_remediation_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_remediation_id" ON "workflow_object_refs" ("workflow_instance_id", "remediation_id");
-- Create index "workflowobjectref_workflow_instance_id_risk_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_risk_id" ON "workflow_object_refs" ("workflow_instance_id", "risk_id");
-- Create index "workflowobjectref_workflow_instance_id_subcontrol_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_subcontrol_id" ON "workflow_object_refs" ("workflow_instance_id", "subcontrol_id");
-- Create index "workflowobjectref_workflow_instance_id_task_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_task_id" ON "workflow_object_refs" ("workflow_instance_id", "task_id");
-- Create index "workflowobjectref_workflow_instance_id_vulnerability_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_vulnerability_id" ON "workflow_object_refs" ("workflow_instance_id", "vulnerability_id");
-- Create "workflow_proposals" table
CREATE TABLE "workflow_proposals" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "tags" jsonb NULL, "domain_key" character varying NOT NULL, "state" character varying NOT NULL DEFAULT 'DRAFT', "revision" bigint NOT NULL DEFAULT 1, "changes" jsonb NULL, "proposed_changes" jsonb NULL, "proposed_hash" character varying NULL, "approved_hash" character varying NULL, "submitted_at" timestamptz NULL, "owner_id" character varying NULL, "workflow_object_ref_id" character varying NOT NULL, "submitted_by_user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "workflow_proposal_owner_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_owner_id_idx" ON "workflow_proposals" ("owner_id");
-- Create index "workflow_proposal_submitted_by_user_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_submitted_by_user_id_idx" ON "workflow_proposals" ("submitted_by_user_id");
-- Create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY (ARRAY[('DRAFT'::character varying)::text, ('SUBMITTED'::character varying)::text]));
-- Create "action_plan_blocked_groups" table
CREATE TABLE "action_plan_blocked_groups" ("action_plan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "group_id"));
-- Create index "action_plan_blocked_groups_group_id_idx" to table: "action_plan_blocked_groups"
CREATE INDEX "action_plan_blocked_groups_group_id_idx" ON "action_plan_blocked_groups" ("group_id");
-- Create "action_plan_editors" table
CREATE TABLE "action_plan_editors" ("action_plan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "group_id"));
-- Create index "action_plan_editors_group_id_idx" to table: "action_plan_editors"
CREATE INDEX "action_plan_editors_group_id_idx" ON "action_plan_editors" ("group_id");
-- Create "action_plan_viewers" table
CREATE TABLE "action_plan_viewers" ("action_plan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "group_id"));
-- Create index "action_plan_viewers_group_id_idx" to table: "action_plan_viewers"
CREATE INDEX "action_plan_viewers_group_id_idx" ON "action_plan_viewers" ("group_id");
-- Create "action_plan_tasks" table
CREATE TABLE "action_plan_tasks" ("action_plan_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "task_id"));
-- Create index "action_plan_tasks_task_id_idx" to table: "action_plan_tasks"
CREATE INDEX "action_plan_tasks_task_id_idx" ON "action_plan_tasks" ("task_id");
-- Create "asset_connected_assets" table
CREATE TABLE "asset_connected_assets" ("asset_id" character varying NOT NULL, "connected_from_id" character varying NOT NULL, PRIMARY KEY ("asset_id", "connected_from_id"));
-- Create index "asset_connected_assets_connected_from_id_idx" to table: "asset_connected_assets"
CREATE INDEX "asset_connected_assets_connected_from_id_idx" ON "asset_connected_assets" ("connected_from_id");
-- Create "campaign_blocked_groups" table
CREATE TABLE "campaign_blocked_groups" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create index "campaign_blocked_groups_group_id_idx" to table: "campaign_blocked_groups"
CREATE INDEX "campaign_blocked_groups_group_id_idx" ON "campaign_blocked_groups" ("group_id");
-- Create "campaign_editors" table
CREATE TABLE "campaign_editors" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create index "campaign_editors_group_id_idx" to table: "campaign_editors"
CREATE INDEX "campaign_editors_group_id_idx" ON "campaign_editors" ("group_id");
-- Create "campaign_viewers" table
CREATE TABLE "campaign_viewers" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create index "campaign_viewers_group_id_idx" to table: "campaign_viewers"
CREATE INDEX "campaign_viewers_group_id_idx" ON "campaign_viewers" ("group_id");
-- Create "campaign_contacts" table
CREATE TABLE "campaign_contacts" ("campaign_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "contact_id"));
-- Create index "campaign_contacts_contact_id_idx" to table: "campaign_contacts"
CREATE INDEX "campaign_contacts_contact_id_idx" ON "campaign_contacts" ("contact_id");
-- Create "campaign_users" table
CREATE TABLE "campaign_users" ("campaign_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "user_id"));
-- Create index "campaign_users_user_id_idx" to table: "campaign_users"
CREATE INDEX "campaign_users_user_id_idx" ON "campaign_users" ("user_id");
-- Create "campaign_groups" table
CREATE TABLE "campaign_groups" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create index "campaign_groups_group_id_idx" to table: "campaign_groups"
CREATE INDEX "campaign_groups_group_id_idx" ON "campaign_groups" ("group_id");
-- Create "campaign_identity_holders" table
CREATE TABLE "campaign_identity_holders" ("campaign_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "identity_holder_id"));
-- Create index "campaign_identity_holders_identity_holder_id_idx" to table: "campaign_identity_holders"
CREATE INDEX "campaign_identity_holders_identity_holder_id_idx" ON "campaign_identity_holders" ("identity_holder_id");
-- Create "check_result_controls" table
CREATE TABLE "check_result_controls" ("check_result_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("check_result_id", "control_id"));
-- Create index "check_result_controls_control_id_idx" to table: "check_result_controls"
CREATE INDEX "check_result_controls_control_id_idx" ON "check_result_controls" ("control_id");
-- Create "contact_files" table
CREATE TABLE "contact_files" ("contact_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "file_id"));
-- Create index "contact_files_file_id_idx" to table: "contact_files"
CREATE INDEX "contact_files_file_id_idx" ON "contact_files" ("file_id");
-- Create "control_control_objectives" table
CREATE TABLE "control_control_objectives" ("control_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_objective_id"));
-- Create index "control_control_objectives_control_objective_id_idx" to table: "control_control_objectives"
CREATE INDEX "control_control_objectives_control_objective_id_idx" ON "control_control_objectives" ("control_objective_id");
-- Create "control_tasks" table
CREATE TABLE "control_tasks" ("control_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_id", "task_id"));
-- Create index "control_tasks_task_id_idx" to table: "control_tasks"
CREATE INDEX "control_tasks_task_id_idx" ON "control_tasks" ("task_id");
-- Create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- Create index "control_narratives_narrative_id_idx" to table: "control_narratives"
CREATE INDEX "control_narratives_narrative_id_idx" ON "control_narratives" ("narrative_id");
-- Create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- Create index "control_risks_risk_id_idx" to table: "control_risks"
CREATE INDEX "control_risks_risk_id_idx" ON "control_risks" ("risk_id");
-- Create "control_action_plans" table
CREATE TABLE "control_action_plans" ("control_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "action_plan_id"));
-- Create index "control_action_plans_action_plan_id_idx" to table: "control_action_plans"
CREATE INDEX "control_action_plans_action_plan_id_idx" ON "control_action_plans" ("action_plan_id");
-- Create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- Create index "control_procedures_procedure_id_idx" to table: "control_procedures"
CREATE INDEX "control_procedures_procedure_id_idx" ON "control_procedures" ("procedure_id");
-- Create "control_scans" table
CREATE TABLE "control_scans" ("control_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "scan_id"));
-- Create index "control_scans_scan_id_idx" to table: "control_scans"
CREATE INDEX "control_scans_scan_id_idx" ON "control_scans" ("scan_id");
-- Create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- Create index "control_blocked_groups_group_id_idx" to table: "control_blocked_groups"
CREATE INDEX "control_blocked_groups_group_id_idx" ON "control_blocked_groups" ("group_id");
-- Create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- Create index "control_editors_group_id_idx" to table: "control_editors"
CREATE INDEX "control_editors_group_id_idx" ON "control_editors" ("group_id");
-- Create "control_assets" table
CREATE TABLE "control_assets" ("control_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("control_id", "asset_id"));
-- Create index "control_assets_asset_id_idx" to table: "control_assets"
CREATE INDEX "control_assets_asset_id_idx" ON "control_assets" ("asset_id");
-- Create "control_entities" table
CREATE TABLE "control_entities" ("control_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("control_id", "entity_id"));
-- Create index "control_entities_entity_id_idx" to table: "control_entities"
CREATE INDEX "control_entities_entity_id_idx" ON "control_entities" ("entity_id");
-- Create "control_identity_holders" table
CREATE TABLE "control_identity_holders" ("control_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("control_id", "identity_holder_id"));
-- Create index "control_identity_holders_identity_holder_id_idx" to table: "control_identity_holders"
CREATE INDEX "control_identity_holders_identity_holder_id_idx" ON "control_identity_holders" ("identity_holder_id");
-- Create "control_campaigns" table
CREATE TABLE "control_campaigns" ("control_id" character varying NOT NULL, "campaign_id" character varying NOT NULL, PRIMARY KEY ("control_id", "campaign_id"));
-- Create index "control_campaigns_campaign_id_idx" to table: "control_campaigns"
CREATE INDEX "control_campaigns_campaign_id_idx" ON "control_campaigns" ("campaign_id");
-- Create "control_control_implementations" table
CREATE TABLE "control_control_implementations" ("control_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_implementation_id"));
-- Create index "control_control_implementations_control_implementation_id_idx" to table: "control_control_implementations"
CREATE INDEX "control_control_implementations_control_implementation_id_idx" ON "control_control_implementations" ("control_implementation_id");
-- Create "control_implementation_blocked_groups" table
CREATE TABLE "control_implementation_blocked_groups" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"));
-- Create index "control_implementation_blocked_groups_group_id_idx" to table: "control_implementation_blocked_groups"
CREATE INDEX "control_implementation_blocked_groups_group_id_idx" ON "control_implementation_blocked_groups" ("group_id");
-- Create "control_implementation_editors" table
CREATE TABLE "control_implementation_editors" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"));
-- Create index "control_implementation_editors_group_id_idx" to table: "control_implementation_editors"
CREATE INDEX "control_implementation_editors_group_id_idx" ON "control_implementation_editors" ("group_id");
-- Create "control_implementation_viewers" table
CREATE TABLE "control_implementation_viewers" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"));
-- Create index "control_implementation_viewers_group_id_idx" to table: "control_implementation_viewers"
CREATE INDEX "control_implementation_viewers_group_id_idx" ON "control_implementation_viewers" ("group_id");
-- Create "control_implementation_tasks" table
CREATE TABLE "control_implementation_tasks" ("control_implementation_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "task_id"));
-- Create index "control_implementation_tasks_task_id_idx" to table: "control_implementation_tasks"
CREATE INDEX "control_implementation_tasks_task_id_idx" ON "control_implementation_tasks" ("task_id");
-- Create "control_objective_blocked_groups" table
CREATE TABLE "control_objective_blocked_groups" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- Create index "control_objective_blocked_groups_group_id_idx" to table: "control_objective_blocked_groups"
CREATE INDEX "control_objective_blocked_groups_group_id_idx" ON "control_objective_blocked_groups" ("group_id");
-- Create "control_objective_editors" table
CREATE TABLE "control_objective_editors" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- Create index "control_objective_editors_group_id_idx" to table: "control_objective_editors"
CREATE INDEX "control_objective_editors_group_id_idx" ON "control_objective_editors" ("group_id");
-- Create "control_objective_viewers" table
CREATE TABLE "control_objective_viewers" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- Create index "control_objective_viewers_group_id_idx" to table: "control_objective_viewers"
CREATE INDEX "control_objective_viewers_group_id_idx" ON "control_objective_viewers" ("group_id");
-- Create "control_objective_tasks" table
CREATE TABLE "control_objective_tasks" ("control_objective_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "task_id"));
-- Create index "control_objective_tasks_task_id_idx" to table: "control_objective_tasks"
CREATE INDEX "control_objective_tasks_task_id_idx" ON "control_objective_tasks" ("task_id");
-- Create "document_data_files" table
CREATE TABLE "document_data_files" ("document_data_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("document_data_id", "file_id"));
-- Create index "document_data_files_file_id_idx" to table: "document_data_files"
CREATE INDEX "document_data_files_file_id_idx" ON "document_data_files" ("file_id");
-- Create "entity_blocked_groups" table
CREATE TABLE "entity_blocked_groups" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"));
-- Create index "entity_blocked_groups_group_id_idx" to table: "entity_blocked_groups"
CREATE INDEX "entity_blocked_groups_group_id_idx" ON "entity_blocked_groups" ("group_id");
-- Create "entity_editors" table
CREATE TABLE "entity_editors" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"));
-- Create index "entity_editors_group_id_idx" to table: "entity_editors"
CREATE INDEX "entity_editors_group_id_idx" ON "entity_editors" ("group_id");
-- Create "entity_contacts" table
CREATE TABLE "entity_contacts" ("entity_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "contact_id"));
-- Create index "entity_contacts_contact_id_idx" to table: "entity_contacts"
CREATE INDEX "entity_contacts_contact_id_idx" ON "entity_contacts" ("contact_id");
-- Create "entity_documents" table
CREATE TABLE "entity_documents" ("entity_id" character varying NOT NULL, "document_data_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "document_data_id"));
-- Create index "entity_documents_document_data_id_idx" to table: "entity_documents"
CREATE INDEX "entity_documents_document_data_id_idx" ON "entity_documents" ("document_data_id");
-- Create "entity_files" table
CREATE TABLE "entity_files" ("entity_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "file_id"));
-- Create index "entity_files_file_id_idx" to table: "entity_files"
CREATE INDEX "entity_files_file_id_idx" ON "entity_files" ("file_id");
-- Create "entity_assets" table
CREATE TABLE "entity_assets" ("entity_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "asset_id"));
-- Create index "entity_assets_asset_id_idx" to table: "entity_assets"
CREATE INDEX "entity_assets_asset_id_idx" ON "entity_assets" ("asset_id");
-- Create "entity_system_details" table
CREATE TABLE "entity_system_details" ("entity_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "system_detail_id"));
-- Create index "entity_system_details_system_detail_id_idx" to table: "entity_system_details"
CREATE INDEX "entity_system_details_system_detail_id_idx" ON "entity_system_details" ("system_detail_id");
-- Create "entity_integrations" table
CREATE TABLE "entity_integrations" ("entity_id" character varying NOT NULL, "integration_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "integration_id"));
-- Create index "entity_integrations_integration_id_idx" to table: "entity_integrations"
CREATE INDEX "entity_integrations_integration_id_idx" ON "entity_integrations" ("integration_id");
-- Create "entity_subprocessors" table
CREATE TABLE "entity_subprocessors" ("entity_id" character varying NOT NULL, "subprocessor_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "subprocessor_id"));
-- Create index "entity_subprocessors_subprocessor_id_idx" to table: "entity_subprocessors"
CREATE INDEX "entity_subprocessors_subprocessor_id_idx" ON "entity_subprocessors" ("subprocessor_id");
-- Create "evidence_controls" table
CREATE TABLE "evidence_controls" ("evidence_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_id"));
-- Create index "evidence_controls_control_id_idx" to table: "evidence_controls"
CREATE INDEX "evidence_controls_control_id_idx" ON "evidence_controls" ("control_id");
-- Create "evidence_subcontrols" table
CREATE TABLE "evidence_subcontrols" ("evidence_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "subcontrol_id"));
-- Create index "evidence_subcontrols_subcontrol_id_idx" to table: "evidence_subcontrols"
CREATE INDEX "evidence_subcontrols_subcontrol_id_idx" ON "evidence_subcontrols" ("subcontrol_id");
-- Create "evidence_control_objectives" table
CREATE TABLE "evidence_control_objectives" ("evidence_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_objective_id"));
-- Create index "evidence_control_objectives_control_objective_id_idx" to table: "evidence_control_objectives"
CREATE INDEX "evidence_control_objectives_control_objective_id_idx" ON "evidence_control_objectives" ("control_objective_id");
-- Create "evidence_files" table
CREATE TABLE "evidence_files" ("evidence_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "file_id"));
-- Create index "evidence_files_file_id_idx" to table: "evidence_files"
CREATE INDEX "evidence_files_file_id_idx" ON "evidence_files" ("file_id");
-- Create "file_events" table
CREATE TABLE "file_events" ("file_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("file_id", "event_id"));
-- Create index "file_events_event_id_idx" to table: "file_events"
CREATE INDEX "file_events_event_id_idx" ON "file_events" ("event_id");
-- Create "file_secrets" table
CREATE TABLE "file_secrets" ("file_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("file_id", "hush_id"));
-- Create index "file_secrets_hush_id_idx" to table: "file_secrets"
CREATE INDEX "file_secrets_hush_id_idx" ON "file_secrets" ("hush_id");
-- Create "finding_blocked_groups" table
CREATE TABLE "finding_blocked_groups" ("finding_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "group_id"));
-- Create index "finding_blocked_groups_group_id_idx" to table: "finding_blocked_groups"
CREATE INDEX "finding_blocked_groups_group_id_idx" ON "finding_blocked_groups" ("group_id");
-- Create "finding_editors" table
CREATE TABLE "finding_editors" ("finding_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "group_id"));
-- Create index "finding_editors_group_id_idx" to table: "finding_editors"
CREATE INDEX "finding_editors_group_id_idx" ON "finding_editors" ("group_id");
-- Create "finding_vulnerabilities" table
CREATE TABLE "finding_vulnerabilities" ("finding_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "vulnerability_id"));
-- Create index "finding_vulnerabilities_vulnerability_id_idx" to table: "finding_vulnerabilities"
CREATE INDEX "finding_vulnerabilities_vulnerability_id_idx" ON "finding_vulnerabilities" ("vulnerability_id");
-- Create "finding_action_plans" table
CREATE TABLE "finding_action_plans" ("finding_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "action_plan_id"));
-- Create index "finding_action_plans_action_plan_id_idx" to table: "finding_action_plans"
CREATE INDEX "finding_action_plans_action_plan_id_idx" ON "finding_action_plans" ("action_plan_id");
-- Create "finding_subcontrols" table
CREATE TABLE "finding_subcontrols" ("finding_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "subcontrol_id"));
-- Create index "finding_subcontrols_subcontrol_id_idx" to table: "finding_subcontrols"
CREATE INDEX "finding_subcontrols_subcontrol_id_idx" ON "finding_subcontrols" ("subcontrol_id");
-- Create "finding_risks" table
CREATE TABLE "finding_risks" ("finding_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "risk_id"));
-- Create index "finding_risks_risk_id_idx" to table: "finding_risks"
CREATE INDEX "finding_risks_risk_id_idx" ON "finding_risks" ("risk_id");
-- Create "finding_programs" table
CREATE TABLE "finding_programs" ("finding_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "program_id"));
-- Create index "finding_programs_program_id_idx" to table: "finding_programs"
CREATE INDEX "finding_programs_program_id_idx" ON "finding_programs" ("program_id");
-- Create "finding_assets" table
CREATE TABLE "finding_assets" ("finding_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "asset_id"));
-- Create index "finding_assets_asset_id_idx" to table: "finding_assets"
CREATE INDEX "finding_assets_asset_id_idx" ON "finding_assets" ("asset_id");
-- Create "finding_entities" table
CREATE TABLE "finding_entities" ("finding_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "entity_id"));
-- Create index "finding_entities_entity_id_idx" to table: "finding_entities"
CREATE INDEX "finding_entities_entity_id_idx" ON "finding_entities" ("entity_id");
-- Create "finding_scans" table
CREATE TABLE "finding_scans" ("finding_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "scan_id"));
-- Create index "finding_scans_scan_id_idx" to table: "finding_scans"
CREATE INDEX "finding_scans_scan_id_idx" ON "finding_scans" ("scan_id");
-- Create "finding_tasks" table
CREATE TABLE "finding_tasks" ("finding_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "task_id"));
-- Create index "finding_tasks_task_id_idx" to table: "finding_tasks"
CREATE INDEX "finding_tasks_task_id_idx" ON "finding_tasks" ("task_id");
-- Create "finding_directory_accounts" table
CREATE TABLE "finding_directory_accounts" ("finding_id" character varying NOT NULL, "directory_account_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "directory_account_id"));
-- Create index "finding_directory_accounts_directory_account_id_idx" to table: "finding_directory_accounts"
CREATE INDEX "finding_directory_accounts_directory_account_id_idx" ON "finding_directory_accounts" ("directory_account_id");
-- Create "finding_identity_holders" table
CREATE TABLE "finding_identity_holders" ("finding_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "identity_holder_id"));
-- Create index "finding_identity_holders_identity_holder_id_idx" to table: "finding_identity_holders"
CREATE INDEX "finding_identity_holders_identity_holder_id_idx" ON "finding_identity_holders" ("identity_holder_id");
-- Create "finding_check_results" table
CREATE TABLE "finding_check_results" ("finding_id" character varying NOT NULL, "check_result_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "check_result_id"));
-- Create index "finding_check_results_check_result_id_idx" to table: "finding_check_results"
CREATE INDEX "finding_check_results_check_result_id_idx" ON "finding_check_results" ("check_result_id");
-- Create "group_events" table
CREATE TABLE "group_events" ("group_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_id", "event_id"));
-- Create index "group_events_event_id_idx" to table: "group_events"
CREATE INDEX "group_events_event_id_idx" ON "group_events" ("event_id");
-- Create "group_files" table
CREATE TABLE "group_files" ("group_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("group_id", "file_id"));
-- Create index "group_files_file_id_idx" to table: "group_files"
CREATE INDEX "group_files_file_id_idx" ON "group_files" ("file_id");
-- Create "group_tasks" table
CREATE TABLE "group_tasks" ("group_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("group_id", "task_id"));
-- Create index "group_tasks_task_id_idx" to table: "group_tasks"
CREATE INDEX "group_tasks_task_id_idx" ON "group_tasks" ("task_id");
-- Create "group_membership_events" table
CREATE TABLE "group_membership_events" ("group_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_membership_id", "event_id"));
-- Create index "group_membership_events_event_id_idx" to table: "group_membership_events"
CREATE INDEX "group_membership_events_event_id_idx" ON "group_membership_events" ("event_id");
-- Create "hush_events" table
CREATE TABLE "hush_events" ("hush_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("hush_id", "event_id"));
-- Create index "hush_events_event_id_idx" to table: "hush_events"
CREATE INDEX "hush_events_event_id_idx" ON "hush_events" ("event_id");
-- Create "identity_holder_assessments" table
CREATE TABLE "identity_holder_assessments" ("identity_holder_id" character varying NOT NULL, "assessment_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "assessment_id"));
-- Create index "identity_holder_assessments_assessment_id_idx" to table: "identity_holder_assessments"
CREATE INDEX "identity_holder_assessments_assessment_id_idx" ON "identity_holder_assessments" ("assessment_id");
-- Create "identity_holder_templates" table
CREATE TABLE "identity_holder_templates" ("identity_holder_id" character varying NOT NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "template_id"));
-- Create index "identity_holder_templates_template_id_idx" to table: "identity_holder_templates"
CREATE INDEX "identity_holder_templates_template_id_idx" ON "identity_holder_templates" ("template_id");
-- Create "identity_holder_assets" table
CREATE TABLE "identity_holder_assets" ("identity_holder_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "asset_id"));
-- Create index "identity_holder_assets_asset_id_idx" to table: "identity_holder_assets"
CREATE INDEX "identity_holder_assets_asset_id_idx" ON "identity_holder_assets" ("asset_id");
-- Create "identity_holder_entities" table
CREATE TABLE "identity_holder_entities" ("identity_holder_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "entity_id"));
-- Create index "identity_holder_entities_entity_id_idx" to table: "identity_holder_entities"
CREATE INDEX "identity_holder_entities_entity_id_idx" ON "identity_holder_entities" ("entity_id");
-- Create "identity_holder_tasks" table
CREATE TABLE "identity_holder_tasks" ("identity_holder_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "task_id"));
-- Create index "identity_holder_tasks_task_id_idx" to table: "identity_holder_tasks"
CREATE INDEX "identity_holder_tasks_task_id_idx" ON "identity_holder_tasks" ("task_id");
-- Create "identity_holder_files" table
CREATE TABLE "identity_holder_files" ("identity_holder_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "file_id"));
-- Create index "identity_holder_files_file_id_idx" to table: "identity_holder_files"
CREATE INDEX "identity_holder_files_file_id_idx" ON "identity_holder_files" ("file_id");
-- Create "integration_secrets" table
CREATE TABLE "integration_secrets" ("integration_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "hush_id"));
-- Create index "integration_secrets_hush_id_idx" to table: "integration_secrets"
CREATE INDEX "integration_secrets_hush_id_idx" ON "integration_secrets" ("hush_id");
-- Create "integration_events" table
CREATE TABLE "integration_events" ("integration_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "event_id"));
-- Create index "integration_events_event_id_idx" to table: "integration_events"
CREATE INDEX "integration_events_event_id_idx" ON "integration_events" ("event_id");
-- Create "integration_findings" table
CREATE TABLE "integration_findings" ("integration_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "finding_id"));
-- Create index "integration_findings_finding_id_idx" to table: "integration_findings"
CREATE INDEX "integration_findings_finding_id_idx" ON "integration_findings" ("finding_id");
-- Create "integration_vulnerabilities" table
CREATE TABLE "integration_vulnerabilities" ("integration_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "vulnerability_id"));
-- Create index "integration_vulnerabilities_vulnerability_id_idx" to table: "integration_vulnerabilities"
CREATE INDEX "integration_vulnerabilities_vulnerability_id_idx" ON "integration_vulnerabilities" ("vulnerability_id");
-- Create "integration_internal_policies" table
CREATE TABLE "integration_internal_policies" ("integration_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "internal_policy_id"));
-- Create index "integration_internal_policies_internal_policy_id_idx" to table: "integration_internal_policies"
CREATE INDEX "integration_internal_policies_internal_policy_id_idx" ON "integration_internal_policies" ("internal_policy_id");
-- Create "integration_reviews" table
CREATE TABLE "integration_reviews" ("integration_id" character varying NOT NULL, "review_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "review_id"));
-- Create index "integration_reviews_review_id_idx" to table: "integration_reviews"
CREATE INDEX "integration_reviews_review_id_idx" ON "integration_reviews" ("review_id");
-- Create "integration_remediations" table
CREATE TABLE "integration_remediations" ("integration_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "remediation_id"));
-- Create index "integration_remediations_remediation_id_idx" to table: "integration_remediations"
CREATE INDEX "integration_remediations_remediation_id_idx" ON "integration_remediations" ("remediation_id");
-- Create "integration_action_plans" table
CREATE TABLE "integration_action_plans" ("integration_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "action_plan_id"));
-- Create index "integration_action_plans_action_plan_id_idx" to table: "integration_action_plans"
CREATE INDEX "integration_action_plans_action_plan_id_idx" ON "integration_action_plans" ("action_plan_id");
-- Create "internal_policy_blocked_groups" table
CREATE TABLE "internal_policy_blocked_groups" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- Create index "internal_policy_blocked_groups_group_id_idx" to table: "internal_policy_blocked_groups"
CREATE INDEX "internal_policy_blocked_groups_group_id_idx" ON "internal_policy_blocked_groups" ("group_id");
-- Create "internal_policy_editors" table
CREATE TABLE "internal_policy_editors" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- Create index "internal_policy_editors_group_id_idx" to table: "internal_policy_editors"
CREATE INDEX "internal_policy_editors_group_id_idx" ON "internal_policy_editors" ("group_id");
-- Create "internal_policy_control_objectives" table
CREATE TABLE "internal_policy_control_objectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- Create index "internal_policy_control_objectives_control_objective_id_idx" to table: "internal_policy_control_objectives"
CREATE INDEX "internal_policy_control_objectives_control_objective_id_idx" ON "internal_policy_control_objectives" ("control_objective_id");
-- Create "internal_policy_controls" table
CREATE TABLE "internal_policy_controls" ("internal_policy_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_id"));
-- Create index "internal_policy_controls_control_id_idx" to table: "internal_policy_controls"
CREATE INDEX "internal_policy_controls_control_id_idx" ON "internal_policy_controls" ("control_id");
-- Create "internal_policy_subcontrols" table
CREATE TABLE "internal_policy_subcontrols" ("internal_policy_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "subcontrol_id"));
-- Create index "internal_policy_subcontrols_subcontrol_id_idx" to table: "internal_policy_subcontrols"
CREATE INDEX "internal_policy_subcontrols_subcontrol_id_idx" ON "internal_policy_subcontrols" ("subcontrol_id");
-- Create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- Create index "internal_policy_procedures_procedure_id_idx" to table: "internal_policy_procedures"
CREATE INDEX "internal_policy_procedures_procedure_id_idx" ON "internal_policy_procedures" ("procedure_id");
-- Create "internal_policy_narratives" table
CREATE TABLE "internal_policy_narratives" ("internal_policy_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "narrative_id"));
-- Create index "internal_policy_narratives_narrative_id_idx" to table: "internal_policy_narratives"
CREATE INDEX "internal_policy_narratives_narrative_id_idx" ON "internal_policy_narratives" ("narrative_id");
-- Create "internal_policy_tasks" table
CREATE TABLE "internal_policy_tasks" ("internal_policy_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "task_id"));
-- Create index "internal_policy_tasks_task_id_idx" to table: "internal_policy_tasks"
CREATE INDEX "internal_policy_tasks_task_id_idx" ON "internal_policy_tasks" ("task_id");
-- Create "internal_policy_risks" table
CREATE TABLE "internal_policy_risks" ("internal_policy_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "risk_id"));
-- Create index "internal_policy_risks_risk_id_idx" to table: "internal_policy_risks"
CREATE INDEX "internal_policy_risks_risk_id_idx" ON "internal_policy_risks" ("risk_id");
-- Create "internal_policy_assets" table
CREATE TABLE "internal_policy_assets" ("internal_policy_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "asset_id"));
-- Create index "internal_policy_assets_asset_id_idx" to table: "internal_policy_assets"
CREATE INDEX "internal_policy_assets_asset_id_idx" ON "internal_policy_assets" ("asset_id");
-- Create "internal_policy_entities" table
CREATE TABLE "internal_policy_entities" ("internal_policy_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "entity_id"));
-- Create index "internal_policy_entities_entity_id_idx" to table: "internal_policy_entities"
CREATE INDEX "internal_policy_entities_entity_id_idx" ON "internal_policy_entities" ("entity_id");
-- Create "internal_policy_identity_holders" table
CREATE TABLE "internal_policy_identity_holders" ("internal_policy_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "identity_holder_id"));
-- Create index "internal_policy_identity_holders_identity_holder_id_idx" to table: "internal_policy_identity_holders"
CREATE INDEX "internal_policy_identity_holders_identity_holder_id_idx" ON "internal_policy_identity_holders" ("identity_holder_id");
-- Create "invite_events" table
CREATE TABLE "invite_events" ("invite_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "event_id"));
-- Create index "invite_events_event_id_idx" to table: "invite_events"
CREATE INDEX "invite_events_event_id_idx" ON "invite_events" ("event_id");
-- Create "invite_groups" table
CREATE TABLE "invite_groups" ("invite_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "group_id"));
-- Create index "invite_groups_group_id_idx" to table: "invite_groups"
CREATE INDEX "invite_groups_group_id_idx" ON "invite_groups" ("group_id");
-- Create "job_runner_job_runner_tokens" table
CREATE TABLE "job_runner_job_runner_tokens" ("job_runner_id" character varying NOT NULL, "job_runner_token_id" character varying NOT NULL, PRIMARY KEY ("job_runner_id", "job_runner_token_id"));
-- Create index "job_runner_job_runner_tokens_job_runner_token_id_idx" to table: "job_runner_job_runner_tokens"
CREATE INDEX "job_runner_job_runner_tokens_job_runner_token_id_idx" ON "job_runner_job_runner_tokens" ("job_runner_token_id");
-- Create "mapped_control_blocked_groups" table
CREATE TABLE "mapped_control_blocked_groups" ("mapped_control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "group_id"));
-- Create index "mapped_control_blocked_groups_group_id_idx" to table: "mapped_control_blocked_groups"
CREATE INDEX "mapped_control_blocked_groups_group_id_idx" ON "mapped_control_blocked_groups" ("group_id");
-- Create "mapped_control_editors" table
CREATE TABLE "mapped_control_editors" ("mapped_control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "group_id"));
-- Create index "mapped_control_editors_group_id_idx" to table: "mapped_control_editors"
CREATE INDEX "mapped_control_editors_group_id_idx" ON "mapped_control_editors" ("group_id");
-- Create "mapped_control_from_controls" table
CREATE TABLE "mapped_control_from_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"));
-- Create index "mapped_control_from_controls_control_id_idx" to table: "mapped_control_from_controls"
CREATE INDEX "mapped_control_from_controls_control_id_idx" ON "mapped_control_from_controls" ("control_id");
-- Create "mapped_control_to_controls" table
CREATE TABLE "mapped_control_to_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"));
-- Create index "mapped_control_to_controls_control_id_idx" to table: "mapped_control_to_controls"
CREATE INDEX "mapped_control_to_controls_control_id_idx" ON "mapped_control_to_controls" ("control_id");
-- Create "mapped_control_from_subcontrols" table
CREATE TABLE "mapped_control_from_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"));
-- Create index "mapped_control_from_subcontrols_subcontrol_id_idx" to table: "mapped_control_from_subcontrols"
CREATE INDEX "mapped_control_from_subcontrols_subcontrol_id_idx" ON "mapped_control_from_subcontrols" ("subcontrol_id");
-- Create "mapped_control_to_subcontrols" table
CREATE TABLE "mapped_control_to_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"));
-- Create index "mapped_control_to_subcontrols_subcontrol_id_idx" to table: "mapped_control_to_subcontrols"
CREATE INDEX "mapped_control_to_subcontrols_subcontrol_id_idx" ON "mapped_control_to_subcontrols" ("subcontrol_id");
-- Create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- Create index "narrative_blocked_groups_group_id_idx" to table: "narrative_blocked_groups"
CREATE INDEX "narrative_blocked_groups_group_id_idx" ON "narrative_blocked_groups" ("group_id");
-- Create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- Create index "narrative_editors_group_id_idx" to table: "narrative_editors"
CREATE INDEX "narrative_editors_group_id_idx" ON "narrative_editors" ("group_id");
-- Create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- Create index "narrative_viewers_group_id_idx" to table: "narrative_viewers"
CREATE INDEX "narrative_viewers_group_id_idx" ON "narrative_viewers" ("group_id");
-- Create "org_membership_events" table
CREATE TABLE "org_membership_events" ("org_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_membership_id", "event_id"));
-- Create index "org_membership_events_event_id_idx" to table: "org_membership_events"
CREATE INDEX "org_membership_events_event_id_idx" ON "org_membership_events" ("event_id");
-- Create "org_module_org_prices" table
CREATE TABLE "org_module_org_prices" ("org_module_id" character varying NOT NULL, "org_price_id" character varying NOT NULL, PRIMARY KEY ("org_module_id", "org_price_id"));
-- Create index "org_module_org_prices_org_price_id_idx" to table: "org_module_org_prices"
CREATE INDEX "org_module_org_prices_org_price_id_idx" ON "org_module_org_prices" ("org_price_id");
-- Create "org_product_org_prices" table
CREATE TABLE "org_product_org_prices" ("org_product_id" character varying NOT NULL, "org_price_id" character varying NOT NULL, PRIMARY KEY ("org_product_id", "org_price_id"));
-- Create index "org_product_org_prices_org_price_id_idx" to table: "org_product_org_prices"
CREATE INDEX "org_product_org_prices_org_price_id_idx" ON "org_product_org_prices" ("org_price_id");
-- Create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"));
-- Create index "org_subscription_events_event_id_idx" to table: "org_subscription_events"
CREATE INDEX "org_subscription_events_event_id_idx" ON "org_subscription_events" ("event_id");
-- Create "organization_personal_access_tokens" table
CREATE TABLE "organization_personal_access_tokens" ("organization_id" character varying NOT NULL, "personal_access_token_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "personal_access_token_id"));
-- Create index "organization_personal_access_tokens_personal_access_token_id_id" to table: "organization_personal_access_tokens"
CREATE INDEX "organization_personal_access_tokens_personal_access_token_id_id" ON "organization_personal_access_tokens" ("personal_access_token_id");
-- Create "organization_files" table
CREATE TABLE "organization_files" ("organization_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "file_id"));
-- Create index "organization_files_file_id_idx" to table: "organization_files"
CREATE INDEX "organization_files_file_id_idx" ON "organization_files" ("file_id");
-- Create "organization_events" table
CREATE TABLE "organization_events" ("organization_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "event_id"));
-- Create index "organization_events_event_id_idx" to table: "organization_events"
CREATE INDEX "organization_events_event_id_idx" ON "organization_events" ("event_id");
-- Create "organization_setting_files" table
CREATE TABLE "organization_setting_files" ("organization_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_setting_id", "file_id"));
-- Create index "organization_setting_files_file_id_idx" to table: "organization_setting_files"
CREATE INDEX "organization_setting_files_file_id_idx" ON "organization_setting_files" ("file_id");
-- Create "personal_access_token_events" table
CREATE TABLE "personal_access_token_events" ("personal_access_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("personal_access_token_id", "event_id"));
-- Create index "personal_access_token_events_event_id_idx" to table: "personal_access_token_events"
CREATE INDEX "personal_access_token_events_event_id_idx" ON "personal_access_token_events" ("event_id");
-- Create "platform_blocked_groups" table
CREATE TABLE "platform_blocked_groups" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- Create index "platform_blocked_groups_group_id_idx" to table: "platform_blocked_groups"
CREATE INDEX "platform_blocked_groups_group_id_idx" ON "platform_blocked_groups" ("group_id");
-- Create "platform_editors" table
CREATE TABLE "platform_editors" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- Create index "platform_editors_group_id_idx" to table: "platform_editors"
CREATE INDEX "platform_editors_group_id_idx" ON "platform_editors" ("group_id");
-- Create "platform_viewers" table
CREATE TABLE "platform_viewers" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- Create index "platform_viewers_group_id_idx" to table: "platform_viewers"
CREATE INDEX "platform_viewers_group_id_idx" ON "platform_viewers" ("group_id");
-- Create "platform_assets" table
CREATE TABLE "platform_assets" ("platform_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "asset_id"));
-- Create index "platform_assets_asset_id_idx" to table: "platform_assets"
CREATE INDEX "platform_assets_asset_id_idx" ON "platform_assets" ("asset_id");
-- Create "platform_entities" table
CREATE TABLE "platform_entities" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- Create index "platform_entities_entity_id_idx" to table: "platform_entities"
CREATE INDEX "platform_entities_entity_id_idx" ON "platform_entities" ("entity_id");
-- Create "platform_evidence" table
CREATE TABLE "platform_evidence" ("platform_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "evidence_id"));
-- Create index "platform_evidence_evidence_id_idx" to table: "platform_evidence"
CREATE INDEX "platform_evidence_evidence_id_idx" ON "platform_evidence" ("evidence_id");
-- Create "platform_files" table
CREATE TABLE "platform_files" ("platform_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "file_id"));
-- Create index "platform_files_file_id_idx" to table: "platform_files"
CREATE INDEX "platform_files_file_id_idx" ON "platform_files" ("file_id");
-- Create "platform_risks" table
CREATE TABLE "platform_risks" ("platform_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "risk_id"));
-- Create index "platform_risks_risk_id_idx" to table: "platform_risks"
CREATE INDEX "platform_risks_risk_id_idx" ON "platform_risks" ("risk_id");
-- Create "platform_controls" table
CREATE TABLE "platform_controls" ("platform_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "control_id"));
-- Create index "platform_controls_control_id_idx" to table: "platform_controls"
CREATE INDEX "platform_controls_control_id_idx" ON "platform_controls" ("control_id");
-- Create "platform_assessments" table
CREATE TABLE "platform_assessments" ("platform_id" character varying NOT NULL, "assessment_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "assessment_id"));
-- Create index "platform_assessments_assessment_id_idx" to table: "platform_assessments"
CREATE INDEX "platform_assessments_assessment_id_idx" ON "platform_assessments" ("assessment_id");
-- Create "platform_scans" table
CREATE TABLE "platform_scans" ("platform_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "scan_id"));
-- Create index "platform_scans_scan_id_idx" to table: "platform_scans"
CREATE INDEX "platform_scans_scan_id_idx" ON "platform_scans" ("scan_id");
-- Create "platform_tasks" table
CREATE TABLE "platform_tasks" ("platform_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "task_id"));
-- Create index "platform_tasks_task_id_idx" to table: "platform_tasks"
CREATE INDEX "platform_tasks_task_id_idx" ON "platform_tasks" ("task_id");
-- Create "platform_identity_holders" table
CREATE TABLE "platform_identity_holders" ("platform_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "identity_holder_id"));
-- Create index "platform_identity_holders_identity_holder_id_idx" to table: "platform_identity_holders"
CREATE INDEX "platform_identity_holders_identity_holder_id_idx" ON "platform_identity_holders" ("identity_holder_id");
-- Create "platform_source_entities" table
CREATE TABLE "platform_source_entities" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- Create index "platform_source_entities_entity_id_idx" to table: "platform_source_entities"
CREATE INDEX "platform_source_entities_entity_id_idx" ON "platform_source_entities" ("entity_id");
-- Create "platform_out_of_scope_assets" table
CREATE TABLE "platform_out_of_scope_assets" ("platform_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "asset_id"));
-- Create index "platform_out_of_scope_assets_asset_id_idx" to table: "platform_out_of_scope_assets"
CREATE INDEX "platform_out_of_scope_assets_asset_id_idx" ON "platform_out_of_scope_assets" ("asset_id");
-- Create "platform_out_of_scope_vendors" table
CREATE TABLE "platform_out_of_scope_vendors" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- Create index "platform_out_of_scope_vendors_entity_id_idx" to table: "platform_out_of_scope_vendors"
CREATE INDEX "platform_out_of_scope_vendors_entity_id_idx" ON "platform_out_of_scope_vendors" ("entity_id");
-- Create "platform_applicable_frameworks" table
CREATE TABLE "platform_applicable_frameworks" ("platform_id" character varying NOT NULL, "standard_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "standard_id"));
-- Create index "platform_applicable_frameworks_standard_id_idx" to table: "platform_applicable_frameworks"
CREATE INDEX "platform_applicable_frameworks_standard_id_idx" ON "platform_applicable_frameworks" ("standard_id");
-- Create "platform_system_details" table
CREATE TABLE "platform_system_details" ("platform_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "system_detail_id"));
-- Create index "platform_system_details_system_detail_id_idx" to table: "platform_system_details"
CREATE INDEX "platform_system_details_system_detail_id_idx" ON "platform_system_details" ("system_detail_id");
-- Create "procedure_blocked_groups" table
CREATE TABLE "procedure_blocked_groups" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- Create index "procedure_blocked_groups_group_id_idx" to table: "procedure_blocked_groups"
CREATE INDEX "procedure_blocked_groups_group_id_idx" ON "procedure_blocked_groups" ("group_id");
-- Create "procedure_editors" table
CREATE TABLE "procedure_editors" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- Create index "procedure_editors_group_id_idx" to table: "procedure_editors"
CREATE INDEX "procedure_editors_group_id_idx" ON "procedure_editors" ("group_id");
-- Create "procedure_narratives" table
CREATE TABLE "procedure_narratives" ("procedure_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "narrative_id"));
-- Create index "procedure_narratives_narrative_id_idx" to table: "procedure_narratives"
CREATE INDEX "procedure_narratives_narrative_id_idx" ON "procedure_narratives" ("narrative_id");
-- Create "procedure_risks" table
CREATE TABLE "procedure_risks" ("procedure_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "risk_id"));
-- Create index "procedure_risks_risk_id_idx" to table: "procedure_risks"
CREATE INDEX "procedure_risks_risk_id_idx" ON "procedure_risks" ("risk_id");
-- Create "procedure_tasks" table
CREATE TABLE "procedure_tasks" ("procedure_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "task_id"));
-- Create index "procedure_tasks_task_id_idx" to table: "procedure_tasks"
CREATE INDEX "procedure_tasks_task_id_idx" ON "procedure_tasks" ("task_id");
-- Create "program_blocked_groups" table
CREATE TABLE "program_blocked_groups" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- Create index "program_blocked_groups_group_id_idx" to table: "program_blocked_groups"
CREATE INDEX "program_blocked_groups_group_id_idx" ON "program_blocked_groups" ("group_id");
-- Create "program_editors" table
CREATE TABLE "program_editors" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- Create index "program_editors_group_id_idx" to table: "program_editors"
CREATE INDEX "program_editors_group_id_idx" ON "program_editors" ("group_id");
-- Create "program_viewers" table
CREATE TABLE "program_viewers" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- Create index "program_viewers_group_id_idx" to table: "program_viewers"
CREATE INDEX "program_viewers_group_id_idx" ON "program_viewers" ("group_id");
-- Create "program_controls" table
CREATE TABLE "program_controls" ("program_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_id"));
-- Create index "program_controls_control_id_idx" to table: "program_controls"
CREATE INDEX "program_controls_control_id_idx" ON "program_controls" ("control_id");
-- Create "program_control_objectives" table
CREATE TABLE "program_control_objectives" ("program_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_objective_id"));
-- Create index "program_control_objectives_control_objective_id_idx" to table: "program_control_objectives"
CREATE INDEX "program_control_objectives_control_objective_id_idx" ON "program_control_objectives" ("control_objective_id");
-- Create "program_internal_policies" table
CREATE TABLE "program_internal_policies" ("program_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("program_id", "internal_policy_id"));
-- Create index "program_internal_policies_internal_policy_id_idx" to table: "program_internal_policies"
CREATE INDEX "program_internal_policies_internal_policy_id_idx" ON "program_internal_policies" ("internal_policy_id");
-- Create "program_procedures" table
CREATE TABLE "program_procedures" ("program_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("program_id", "procedure_id"));
-- Create index "program_procedures_procedure_id_idx" to table: "program_procedures"
CREATE INDEX "program_procedures_procedure_id_idx" ON "program_procedures" ("procedure_id");
-- Create "program_risks" table
CREATE TABLE "program_risks" ("program_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("program_id", "risk_id"));
-- Create index "program_risks_risk_id_idx" to table: "program_risks"
CREATE INDEX "program_risks_risk_id_idx" ON "program_risks" ("risk_id");
-- Create "program_tasks" table
CREATE TABLE "program_tasks" ("program_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("program_id", "task_id"));
-- Create index "program_tasks_task_id_idx" to table: "program_tasks"
CREATE INDEX "program_tasks_task_id_idx" ON "program_tasks" ("task_id");
-- Create "program_files" table
CREATE TABLE "program_files" ("program_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("program_id", "file_id"));
-- Create index "program_files_file_id_idx" to table: "program_files"
CREATE INDEX "program_files_file_id_idx" ON "program_files" ("file_id");
-- Create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"));
-- Create index "program_evidence_evidence_id_idx" to table: "program_evidence"
CREATE INDEX "program_evidence_evidence_id_idx" ON "program_evidence" ("evidence_id");
-- Create "program_narratives" table
CREATE TABLE "program_narratives" ("program_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("program_id", "narrative_id"));
-- Create index "program_narratives_narrative_id_idx" to table: "program_narratives"
CREATE INDEX "program_narratives_narrative_id_idx" ON "program_narratives" ("narrative_id");
-- Create "program_action_plans" table
CREATE TABLE "program_action_plans" ("program_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("program_id", "action_plan_id"));
-- Create index "program_action_plans_action_plan_id_idx" to table: "program_action_plans"
CREATE INDEX "program_action_plans_action_plan_id_idx" ON "program_action_plans" ("action_plan_id");
-- Create "program_system_details" table
CREATE TABLE "program_system_details" ("program_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("program_id", "system_detail_id"));
-- Create index "program_system_details_system_detail_id_idx" to table: "program_system_details"
CREATE INDEX "program_system_details_system_detail_id_idx" ON "program_system_details" ("system_detail_id");
-- Create "remediation_blocked_groups" table
CREATE TABLE "remediation_blocked_groups" ("remediation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "group_id"));
-- Create index "remediation_blocked_groups_group_id_idx" to table: "remediation_blocked_groups"
CREATE INDEX "remediation_blocked_groups_group_id_idx" ON "remediation_blocked_groups" ("group_id");
-- Create "remediation_editors" table
CREATE TABLE "remediation_editors" ("remediation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "group_id"));
-- Create index "remediation_editors_group_id_idx" to table: "remediation_editors"
CREATE INDEX "remediation_editors_group_id_idx" ON "remediation_editors" ("group_id");
-- Create "remediation_findings" table
CREATE TABLE "remediation_findings" ("remediation_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "finding_id"));
-- Create index "remediation_findings_finding_id_idx" to table: "remediation_findings"
CREATE INDEX "remediation_findings_finding_id_idx" ON "remediation_findings" ("finding_id");
-- Create "remediation_vulnerabilities" table
CREATE TABLE "remediation_vulnerabilities" ("remediation_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "vulnerability_id"));
-- Create index "remediation_vulnerabilities_vulnerability_id_idx" to table: "remediation_vulnerabilities"
CREATE INDEX "remediation_vulnerabilities_vulnerability_id_idx" ON "remediation_vulnerabilities" ("vulnerability_id");
-- Create "remediation_action_plans" table
CREATE TABLE "remediation_action_plans" ("remediation_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "action_plan_id"));
-- Create index "remediation_action_plans_action_plan_id_idx" to table: "remediation_action_plans"
CREATE INDEX "remediation_action_plans_action_plan_id_idx" ON "remediation_action_plans" ("action_plan_id");
-- Create "remediation_controls" table
CREATE TABLE "remediation_controls" ("remediation_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "control_id"));
-- Create index "remediation_controls_control_id_idx" to table: "remediation_controls"
CREATE INDEX "remediation_controls_control_id_idx" ON "remediation_controls" ("control_id");
-- Create "remediation_subcontrols" table
CREATE TABLE "remediation_subcontrols" ("remediation_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "subcontrol_id"));
-- Create index "remediation_subcontrols_subcontrol_id_idx" to table: "remediation_subcontrols"
CREATE INDEX "remediation_subcontrols_subcontrol_id_idx" ON "remediation_subcontrols" ("subcontrol_id");
-- Create "remediation_risks" table
CREATE TABLE "remediation_risks" ("remediation_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "risk_id"));
-- Create index "remediation_risks_risk_id_idx" to table: "remediation_risks"
CREATE INDEX "remediation_risks_risk_id_idx" ON "remediation_risks" ("risk_id");
-- Create "remediation_programs" table
CREATE TABLE "remediation_programs" ("remediation_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "program_id"));
-- Create index "remediation_programs_program_id_idx" to table: "remediation_programs"
CREATE INDEX "remediation_programs_program_id_idx" ON "remediation_programs" ("program_id");
-- Create "remediation_assets" table
CREATE TABLE "remediation_assets" ("remediation_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "asset_id"));
-- Create index "remediation_assets_asset_id_idx" to table: "remediation_assets"
CREATE INDEX "remediation_assets_asset_id_idx" ON "remediation_assets" ("asset_id");
-- Create "remediation_entities" table
CREATE TABLE "remediation_entities" ("remediation_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "entity_id"));
-- Create index "remediation_entities_entity_id_idx" to table: "remediation_entities"
CREATE INDEX "remediation_entities_entity_id_idx" ON "remediation_entities" ("entity_id");
-- Create "review_blocked_groups" table
CREATE TABLE "review_blocked_groups" ("review_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("review_id", "group_id"));
-- Create index "review_blocked_groups_group_id_idx" to table: "review_blocked_groups"
CREATE INDEX "review_blocked_groups_group_id_idx" ON "review_blocked_groups" ("group_id");
-- Create "review_editors" table
CREATE TABLE "review_editors" ("review_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("review_id", "group_id"));
-- Create index "review_editors_group_id_idx" to table: "review_editors"
CREATE INDEX "review_editors_group_id_idx" ON "review_editors" ("group_id");
-- Create "review_findings" table
CREATE TABLE "review_findings" ("review_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("review_id", "finding_id"));
-- Create index "review_findings_finding_id_idx" to table: "review_findings"
CREATE INDEX "review_findings_finding_id_idx" ON "review_findings" ("finding_id");
-- Create "review_vulnerabilities" table
CREATE TABLE "review_vulnerabilities" ("review_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("review_id", "vulnerability_id"));
-- Create index "review_vulnerabilities_vulnerability_id_idx" to table: "review_vulnerabilities"
CREATE INDEX "review_vulnerabilities_vulnerability_id_idx" ON "review_vulnerabilities" ("vulnerability_id");
-- Create "review_action_plans" table
CREATE TABLE "review_action_plans" ("review_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("review_id", "action_plan_id"));
-- Create index "review_action_plans_action_plan_id_idx" to table: "review_action_plans"
CREATE INDEX "review_action_plans_action_plan_id_idx" ON "review_action_plans" ("action_plan_id");
-- Create "review_remediations" table
CREATE TABLE "review_remediations" ("review_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("review_id", "remediation_id"));
-- Create index "review_remediations_remediation_id_idx" to table: "review_remediations"
CREATE INDEX "review_remediations_remediation_id_idx" ON "review_remediations" ("remediation_id");
-- Create "review_controls" table
CREATE TABLE "review_controls" ("review_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("review_id", "control_id"));
-- Create index "review_controls_control_id_idx" to table: "review_controls"
CREATE INDEX "review_controls_control_id_idx" ON "review_controls" ("control_id");
-- Create "review_subcontrols" table
CREATE TABLE "review_subcontrols" ("review_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("review_id", "subcontrol_id"));
-- Create index "review_subcontrols_subcontrol_id_idx" to table: "review_subcontrols"
CREATE INDEX "review_subcontrols_subcontrol_id_idx" ON "review_subcontrols" ("subcontrol_id");
-- Create "review_risks" table
CREATE TABLE "review_risks" ("review_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("review_id", "risk_id"));
-- Create index "review_risks_risk_id_idx" to table: "review_risks"
CREATE INDEX "review_risks_risk_id_idx" ON "review_risks" ("risk_id");
-- Create "review_programs" table
CREATE TABLE "review_programs" ("review_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("review_id", "program_id"));
-- Create index "review_programs_program_id_idx" to table: "review_programs"
CREATE INDEX "review_programs_program_id_idx" ON "review_programs" ("program_id");
-- Create "review_assets" table
CREATE TABLE "review_assets" ("review_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("review_id", "asset_id"));
-- Create index "review_assets_asset_id_idx" to table: "review_assets"
CREATE INDEX "review_assets_asset_id_idx" ON "review_assets" ("asset_id");
-- Create "review_entities" table
CREATE TABLE "review_entities" ("review_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("review_id", "entity_id"));
-- Create index "review_entities_entity_id_idx" to table: "review_entities"
CREATE INDEX "review_entities_entity_id_idx" ON "review_entities" ("entity_id");
-- Create "review_internal_policies" table
CREATE TABLE "review_internal_policies" ("review_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("review_id", "internal_policy_id"));
-- Create index "review_internal_policies_internal_policy_id_idx" to table: "review_internal_policies"
CREATE INDEX "review_internal_policies_internal_policy_id_idx" ON "review_internal_policies" ("internal_policy_id");
-- Create "risk_blocked_groups" table
CREATE TABLE "risk_blocked_groups" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- Create index "risk_blocked_groups_group_id_idx" to table: "risk_blocked_groups"
CREATE INDEX "risk_blocked_groups_group_id_idx" ON "risk_blocked_groups" ("group_id");
-- Create "risk_editors" table
CREATE TABLE "risk_editors" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- Create index "risk_editors_group_id_idx" to table: "risk_editors"
CREATE INDEX "risk_editors_group_id_idx" ON "risk_editors" ("group_id");
-- Create "risk_viewers" table
CREATE TABLE "risk_viewers" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- Create index "risk_viewers_group_id_idx" to table: "risk_viewers"
CREATE INDEX "risk_viewers_group_id_idx" ON "risk_viewers" ("group_id");
-- Create "risk_action_plans" table
CREATE TABLE "risk_action_plans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- Create index "risk_action_plans_action_plan_id_idx" to table: "risk_action_plans"
CREATE INDEX "risk_action_plans_action_plan_id_idx" ON "risk_action_plans" ("action_plan_id");
-- Create "risk_tasks" table
CREATE TABLE "risk_tasks" ("risk_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "task_id"));
-- Create index "risk_tasks_task_id_idx" to table: "risk_tasks"
CREATE INDEX "risk_tasks_task_id_idx" ON "risk_tasks" ("task_id");
-- Create "scan_blocked_groups" table
CREATE TABLE "scan_blocked_groups" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- Create index "scan_blocked_groups_group_id_idx" to table: "scan_blocked_groups"
CREATE INDEX "scan_blocked_groups_group_id_idx" ON "scan_blocked_groups" ("group_id");
-- Create "scan_editors" table
CREATE TABLE "scan_editors" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- Create index "scan_editors_group_id_idx" to table: "scan_editors"
CREATE INDEX "scan_editors_group_id_idx" ON "scan_editors" ("group_id");
-- Create "scan_assets" table
CREATE TABLE "scan_assets" ("scan_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "asset_id"));
-- Create index "scan_assets_asset_id_idx" to table: "scan_assets"
CREATE INDEX "scan_assets_asset_id_idx" ON "scan_assets" ("asset_id");
-- Create "scan_entities" table
CREATE TABLE "scan_entities" ("scan_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "entity_id"));
-- Create index "scan_entities_entity_id_idx" to table: "scan_entities"
CREATE INDEX "scan_entities_entity_id_idx" ON "scan_entities" ("entity_id");
-- Create "scan_evidence" table
CREATE TABLE "scan_evidence" ("scan_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "evidence_id"));
-- Create index "scan_evidence_evidence_id_idx" to table: "scan_evidence"
CREATE INDEX "scan_evidence_evidence_id_idx" ON "scan_evidence" ("evidence_id");
-- Create "scan_files" table
CREATE TABLE "scan_files" ("scan_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "file_id"));
-- Create index "scan_files_file_id_idx" to table: "scan_files"
CREATE INDEX "scan_files_file_id_idx" ON "scan_files" ("file_id");
-- Create "scan_remediations" table
CREATE TABLE "scan_remediations" ("scan_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "remediation_id"));
-- Create index "scan_remediations_remediation_id_idx" to table: "scan_remediations"
CREATE INDEX "scan_remediations_remediation_id_idx" ON "scan_remediations" ("remediation_id");
-- Create "scan_action_plans" table
CREATE TABLE "scan_action_plans" ("scan_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "action_plan_id"));
-- Create index "scan_action_plans_action_plan_id_idx" to table: "scan_action_plans"
CREATE INDEX "scan_action_plans_action_plan_id_idx" ON "scan_action_plans" ("action_plan_id");
-- Create "scan_tasks" table
CREATE TABLE "scan_tasks" ("scan_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "task_id"));
-- Create index "scan_tasks_task_id_idx" to table: "scan_tasks"
CREATE INDEX "scan_tasks_task_id_idx" ON "scan_tasks" ("task_id");
-- Create "scheduled_job_controls" table
CREATE TABLE "scheduled_job_controls" ("scheduled_job_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("scheduled_job_id", "control_id"));
-- Create index "scheduled_job_controls_control_id_idx" to table: "scheduled_job_controls"
CREATE INDEX "scheduled_job_controls_control_id_idx" ON "scheduled_job_controls" ("control_id");
-- Create "scheduled_job_subcontrols" table
CREATE TABLE "scheduled_job_subcontrols" ("scheduled_job_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("scheduled_job_id", "subcontrol_id"));
-- Create index "scheduled_job_subcontrols_subcontrol_id_idx" to table: "scheduled_job_subcontrols"
CREATE INDEX "scheduled_job_subcontrols_subcontrol_id_idx" ON "scheduled_job_subcontrols" ("subcontrol_id");
-- Create "subcontrol_control_objectives" table
CREATE TABLE "subcontrol_control_objectives" ("subcontrol_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_objective_id"));
-- Create index "subcontrol_control_objectives_control_objective_id_idx" to table: "subcontrol_control_objectives"
CREATE INDEX "subcontrol_control_objectives_control_objective_id_idx" ON "subcontrol_control_objectives" ("control_objective_id");
-- Create "subcontrol_tasks" table
CREATE TABLE "subcontrol_tasks" ("subcontrol_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "task_id"));
-- Create index "subcontrol_tasks_task_id_idx" to table: "subcontrol_tasks"
CREATE INDEX "subcontrol_tasks_task_id_idx" ON "subcontrol_tasks" ("task_id");
-- Create "subcontrol_risks" table
CREATE TABLE "subcontrol_risks" ("subcontrol_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "risk_id"));
-- Create index "subcontrol_risks_risk_id_idx" to table: "subcontrol_risks"
CREATE INDEX "subcontrol_risks_risk_id_idx" ON "subcontrol_risks" ("risk_id");
-- Create "subcontrol_procedures" table
CREATE TABLE "subcontrol_procedures" ("subcontrol_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "procedure_id"));
-- Create index "subcontrol_procedures_procedure_id_idx" to table: "subcontrol_procedures"
CREATE INDEX "subcontrol_procedures_procedure_id_idx" ON "subcontrol_procedures" ("procedure_id");
-- Create "subcontrol_scans" table
CREATE TABLE "subcontrol_scans" ("subcontrol_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "scan_id"));
-- Create index "subcontrol_scans_scan_id_idx" to table: "subcontrol_scans"
CREATE INDEX "subcontrol_scans_scan_id_idx" ON "subcontrol_scans" ("scan_id");
-- Create "subcontrol_control_implementations" table
CREATE TABLE "subcontrol_control_implementations" ("subcontrol_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_implementation_id"));
-- Create index "subcontrol_control_implementations_control_implementation_id_id" to table: "subcontrol_control_implementations"
CREATE INDEX "subcontrol_control_implementations_control_implementation_id_id" ON "subcontrol_control_implementations" ("control_implementation_id");
-- Create "subcontrol_assets" table
CREATE TABLE "subcontrol_assets" ("subcontrol_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "asset_id"));
-- Create index "subcontrol_assets_asset_id_idx" to table: "subcontrol_assets"
CREATE INDEX "subcontrol_assets_asset_id_idx" ON "subcontrol_assets" ("asset_id");
-- Create "subcontrol_entities" table
CREATE TABLE "subcontrol_entities" ("subcontrol_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "entity_id"));
-- Create index "subcontrol_entities_entity_id_idx" to table: "subcontrol_entities"
CREATE INDEX "subcontrol_entities_entity_id_idx" ON "subcontrol_entities" ("entity_id");
-- Create "subcontrol_identity_holders" table
CREATE TABLE "subcontrol_identity_holders" ("subcontrol_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "identity_holder_id"));
-- Create index "subcontrol_identity_holders_identity_holder_id_idx" to table: "subcontrol_identity_holders"
CREATE INDEX "subcontrol_identity_holders_identity_holder_id_idx" ON "subcontrol_identity_holders" ("identity_holder_id");
-- Create "subscriber_events" table
CREATE TABLE "subscriber_events" ("subscriber_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("subscriber_id", "event_id"));
-- Create index "subscriber_events_event_id_idx" to table: "subscriber_events"
CREATE INDEX "subscriber_events_event_id_idx" ON "subscriber_events" ("event_id");
-- Create "system_detail_assets" table
CREATE TABLE "system_detail_assets" ("system_detail_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("system_detail_id", "asset_id"));
-- Create index "system_detail_assets_asset_id_idx" to table: "system_detail_assets"
CREATE INDEX "system_detail_assets_asset_id_idx" ON "system_detail_assets" ("asset_id");
-- Create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"));
-- Create index "task_evidence_evidence_id_idx" to table: "task_evidence"
CREATE INDEX "task_evidence_evidence_id_idx" ON "task_evidence" ("evidence_id");
-- Create "template_files" table
CREATE TABLE "template_files" ("template_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("template_id", "file_id"));
-- Create index "template_files_file_id_idx" to table: "template_files"
CREATE INDEX "template_files_file_id_idx" ON "template_files" ("file_id");
-- Create "user_events" table
CREATE TABLE "user_events" ("user_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("user_id", "event_id"));
-- Create index "user_events_event_id_idx" to table: "user_events"
CREATE INDEX "user_events_event_id_idx" ON "user_events" ("event_id");
-- Create "vulnerability_action_plans" table
CREATE TABLE "vulnerability_action_plans" ("vulnerability_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "action_plan_id"));
-- Create index "vulnerability_action_plans_action_plan_id_idx" to table: "vulnerability_action_plans"
CREATE INDEX "vulnerability_action_plans_action_plan_id_idx" ON "vulnerability_action_plans" ("action_plan_id");
-- Create "vulnerability_controls" table
CREATE TABLE "vulnerability_controls" ("vulnerability_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "control_id"));
-- Create index "vulnerability_controls_control_id_idx" to table: "vulnerability_controls"
CREATE INDEX "vulnerability_controls_control_id_idx" ON "vulnerability_controls" ("control_id");
-- Create "vulnerability_subcontrols" table
CREATE TABLE "vulnerability_subcontrols" ("vulnerability_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "subcontrol_id"));
-- Create index "vulnerability_subcontrols_subcontrol_id_idx" to table: "vulnerability_subcontrols"
CREATE INDEX "vulnerability_subcontrols_subcontrol_id_idx" ON "vulnerability_subcontrols" ("subcontrol_id");
-- Create "vulnerability_risks" table
CREATE TABLE "vulnerability_risks" ("vulnerability_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "risk_id"));
-- Create index "vulnerability_risks_risk_id_idx" to table: "vulnerability_risks"
CREATE INDEX "vulnerability_risks_risk_id_idx" ON "vulnerability_risks" ("risk_id");
-- Create "vulnerability_programs" table
CREATE TABLE "vulnerability_programs" ("vulnerability_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "program_id"));
-- Create index "vulnerability_programs_program_id_idx" to table: "vulnerability_programs"
CREATE INDEX "vulnerability_programs_program_id_idx" ON "vulnerability_programs" ("program_id");
-- Create "vulnerability_assets" table
CREATE TABLE "vulnerability_assets" ("vulnerability_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "asset_id"));
-- Create index "vulnerability_assets_asset_id_idx" to table: "vulnerability_assets"
CREATE INDEX "vulnerability_assets_asset_id_idx" ON "vulnerability_assets" ("asset_id");
-- Create "vulnerability_entities" table
CREATE TABLE "vulnerability_entities" ("vulnerability_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "entity_id"));
-- Create index "vulnerability_entities_entity_id_idx" to table: "vulnerability_entities"
CREATE INDEX "vulnerability_entities_entity_id_idx" ON "vulnerability_entities" ("entity_id");
-- Create "vulnerability_scans" table
CREATE TABLE "vulnerability_scans" ("vulnerability_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "scan_id"));
-- Create index "vulnerability_scans_scan_id_idx" to table: "vulnerability_scans"
CREATE INDEX "vulnerability_scans_scan_id_idx" ON "vulnerability_scans" ("scan_id");
-- Create "vulnerability_tasks" table
CREATE TABLE "vulnerability_tasks" ("vulnerability_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "task_id"));
-- Create index "vulnerability_tasks_task_id_idx" to table: "vulnerability_tasks"
CREATE INDEX "vulnerability_tasks_task_id_idx" ON "vulnerability_tasks" ("task_id");
-- Modify "api_tokens" table
ALTER TABLE "api_tokens" ADD CONSTRAINT "api_tokens_organizations_api_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD CONSTRAINT "action_plans_custom_type_enums_action_plan_kind" FOREIGN KEY ("action_plan_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_custom_type_enums_action_plans" FOREIGN KEY ("custom_type_enum_action_plans") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_organizations_action_plans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_subcontrols_action_plans" FOREIGN KEY ("subcontrol_action_plans") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_users_action_plans" FOREIGN KEY ("user_action_plans") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" ADD CONSTRAINT "assessments_organizations_assessments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD CONSTRAINT "assessment_responses_assessments_assessment_responses" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "assessment_responses_campaigns_assessment_responses" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_entities_assessment_responses" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_identity_holders_assessment_responses" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_organizations_assessment_responses" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "assets" table
ALTER TABLE "assets" ADD CONSTRAINT "assets_custom_type_enums_access_model" FOREIGN KEY ("access_model_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_asset_data_classification" FOREIGN KEY ("asset_data_classification_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_asset_subtype" FOREIGN KEY ("asset_subtype_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_criticality" FOREIGN KEY ("criticality_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_encryption_status" FOREIGN KEY ("encryption_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_security_tier" FOREIGN KEY ("security_tier_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_integrations_assets" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_organizations_assets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_platforms_source_assets" FOREIGN KEY ("source_platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_risks_assets" FOREIGN KEY ("risk_assets") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "campaigns" table
ALTER TABLE "campaigns" ADD CONSTRAINT "campaigns_assessments_campaigns" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_email_templates_campaigns" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_entities_campaigns" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_integrations_campaigns" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_organizations_campaigns" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_templates_campaigns" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_trust_centers_campaigns" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD CONSTRAINT "campaign_targets_campaigns_campaign_targets" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_contacts_campaign_targets" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_groups_campaign_targets" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_organizations_campaign_targets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_subscribers_campaign_targets" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_users_campaign_targets" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "check_results" table
ALTER TABLE "check_results" ADD CONSTRAINT "check_results_integrations_check_results" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" ADD CONSTRAINT "contacts_organizations_contacts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_custom_type_enums_control_kind" FOREIGN KEY ("control_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_controls" FOREIGN KEY ("custom_type_enum_controls") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_entities_responsible_party" FOREIGN KEY ("responsible_party_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_standards_controls" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD CONSTRAINT "control_implementations_evidences_control_implementations" FOREIGN KEY ("evidence_control_implementations") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_implementations_intern_78a7d74302db6f99776c0594111f170b" FOREIGN KEY ("internal_policy_control_implementations") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_implementations_organizations_control_implementations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_organizations_control_objectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "custom_domains" table
ALTER TABLE "custom_domains" ADD CONSTRAINT "custom_domains_dns_verifications_custom_domains" FOREIGN KEY ("dns_verification_custom_domains") REFERENCES "dns_verifications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_dns_verifications_dns_verification" FOREIGN KEY ("dns_verification_id") REFERENCES "dns_verifications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_mappable_domains_custom_domains" FOREIGN KEY ("mappable_domain_custom_domains") REFERENCES "mappable_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_mappable_domains_mappable_domain" FOREIGN KEY ("mappable_domain_id") REFERENCES "mappable_domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "custom_domains_organizations_custom_domains" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD CONSTRAINT "custom_type_enums_entities_auth_methods" FOREIGN KEY ("entity_auth_methods") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_type_enums_organizations_custom_type_enums" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "dns_verifications" table
ALTER TABLE "dns_verifications" ADD CONSTRAINT "dns_verifications_organizations_dns_verifications" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD CONSTRAINT "directory_accounts_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_directory_sync_runs_directory_accounts" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_identity_holders_directory_accounts" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_integrations_directory_accounts" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_organizations_directory_accounts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_platforms_directory_accounts" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD CONSTRAINT "directory_groups_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_directory_sync_runs_directory_groups" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_platforms_directory_groups" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD CONSTRAINT "directory_memberships_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_directory_accounts_directory_account" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_directory_groups_directory_group" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_directory_sync_runs_directory_memberships" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_integrations_directory_memberships" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_organizations_directory_memberships" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_platforms_directory_memberships" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD CONSTRAINT "directory_sync_runs_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_integrations_directory_sync_runs" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_sync_runs_organizations_directory_sync_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_platforms_directory_sync_runs" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "discussions" table
ALTER TABLE "discussions" ADD CONSTRAINT "discussions_controls_discussions" FOREIGN KEY ("control_discussions") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_internal_policies_discussions" FOREIGN KEY ("internal_policy_discussions") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_organizations_discussions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_procedures_discussions" FOREIGN KEY ("procedure_discussions") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_risks_discussions" FOREIGN KEY ("risk_discussions") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_subcontrols_discussions" FOREIGN KEY ("subcontrol_discussions") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" ADD CONSTRAINT "document_data_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_organizations_documents" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD CONSTRAINT "email_templates_integrations_email_templates" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_organizations_email_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_trust_centers_email_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_workflow_definitions_email_templates" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_workflow_instances_email_templates" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD CONSTRAINT "email_verification_tokens_users_email_verification_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_custom_type_enums_entity_relationship_state" FOREIGN KEY ("entity_relationship_state_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_entity_security_questionnaire_status" FOREIGN KEY ("entity_security_questionnaire_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_entity_source_type" FOREIGN KEY ("entity_source_type_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_organizations_entities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_risks_entities" FOREIGN KEY ("risk_entities") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "entity_types" table
ALTER TABLE "entity_types" ADD CONSTRAINT "entity_types_organizations_entity_types" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "events" table
ALTER TABLE "events" ADD CONSTRAINT "events_directory_memberships_events" FOREIGN KEY ("directory_membership_events") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "events_exports_events" FOREIGN KEY ("export_events") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "evidences" table
ALTER TABLE "evidences" ADD CONSTRAINT "evidences_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "exports" table
ALTER TABLE "exports" ADD CONSTRAINT "exports_organizations_exports" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_custom_type_enums_category" FOREIGN KEY ("category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_email_templates_files" FOREIGN KEY ("email_template_files") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_exports_files" FOREIGN KEY ("export_files") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_findings_files" FOREIGN KEY ("finding_files") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_integrations_files" FOREIGN KEY ("integration_files") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_notes_files" FOREIGN KEY ("note_files") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_architecture_diagrams" FOREIGN KEY ("platform_architecture_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_data_flow_diagrams" FOREIGN KEY ("platform_data_flow_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_trust_boundary_diagrams" FOREIGN KEY ("platform_trust_boundary_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_remediations_files" FOREIGN KEY ("remediation_files") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_reviews_files" FOREIGN KEY ("review_files") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_vulnerabilities_files" FOREIGN KEY ("vulnerability_files") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "file_download_tokens" table
ALTER TABLE "file_download_tokens" ADD CONSTRAINT "file_download_tokens_users_file_download_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "findings" table
ALTER TABLE "findings" ADD CONSTRAINT "findings_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_custom_type_enums_finding_status" FOREIGN KEY ("finding_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_organizations_findings" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "finding_controls" table
ALTER TABLE "finding_controls" ADD CONSTRAINT "finding_controls_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "finding_controls_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "finding_controls_organizations_finding_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "finding_controls_standards_standard" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_assessments_blocked_groups" FOREIGN KEY ("assessment_blocked_groups") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_editors" FOREIGN KEY ("assessment_editors") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_viewers" FOREIGN KEY ("assessment_viewers") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_blocked_groups" FOREIGN KEY ("asset_blocked_groups") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_editors" FOREIGN KEY ("asset_editors") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_viewers" FOREIGN KEY ("asset_viewers") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_blocked_groups" FOREIGN KEY ("check_result_blocked_groups") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_editors" FOREIGN KEY ("check_result_editors") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_viewers" FOREIGN KEY ("check_result_viewers") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_blocked_groups" FOREIGN KEY ("email_template_blocked_groups") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_editors" FOREIGN KEY ("email_template_editors") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_viewers" FOREIGN KEY ("email_template_viewers") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_blocked_groups" FOREIGN KEY ("identity_holder_blocked_groups") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_editors" FOREIGN KEY ("identity_holder_editors") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_viewers" FOREIGN KEY ("identity_holder_viewers") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_action_plan_creators" FOREIGN KEY ("organization_action_plan_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_api_token_creators" FOREIGN KEY ("organization_api_token_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_assessment_creators" FOREIGN KEY ("organization_assessment_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_asset_creators" FOREIGN KEY ("organization_asset_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_campaign_creators" FOREIGN KEY ("organization_campaign_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_campaign_target_creators" FOREIGN KEY ("organization_campaign_target_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_campaigns_manager" FOREIGN KEY ("organization_campaigns_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_check_result_creators" FOREIGN KEY ("organization_check_result_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_compliance_manager" FOREIGN KEY ("organization_compliance_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_contact_creators" FOREIGN KEY ("organization_contact_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_creators" FOREIGN KEY ("organization_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_implementation_creators" FOREIGN KEY ("organization_control_implementation_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_objective_creators" FOREIGN KEY ("organization_control_objective_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_custom_domain_creators" FOREIGN KEY ("organization_custom_domain_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_custom_type_enum_creators" FOREIGN KEY ("organization_custom_type_enum_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_account_creators" FOREIGN KEY ("organization_directory_account_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_group_creators" FOREIGN KEY ("organization_directory_group_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_membership_creators" FOREIGN KEY ("organization_directory_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_sync_run_creators" FOREIGN KEY ("organization_directory_sync_run_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_discussion_creators" FOREIGN KEY ("organization_discussion_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_document_data_creators" FOREIGN KEY ("organization_document_data_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_email_template_creators" FOREIGN KEY ("organization_email_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_entity_creators" FOREIGN KEY ("organization_entity_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_entity_type_creators" FOREIGN KEY ("organization_entity_type_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_evidence_creators" FOREIGN KEY ("organization_evidence_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_file_creators" FOREIGN KEY ("organization_file_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_finding_control_creators" FOREIGN KEY ("organization_finding_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_finding_creators" FOREIGN KEY ("organization_finding_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_creators" FOREIGN KEY ("organization_group_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_manager" FOREIGN KEY ("organization_group_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_membership_creators" FOREIGN KEY ("organization_group_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_setting_creators" FOREIGN KEY ("organization_group_setting_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_hush_creators" FOREIGN KEY ("organization_hush_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_identity_holder_creators" FOREIGN KEY ("organization_identity_holder_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_internal_policy_creators" FOREIGN KEY ("organization_internal_policy_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_invite_creators" FOREIGN KEY ("organization_invite_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_runner_creators" FOREIGN KEY ("organization_job_runner_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_runner_registration_token_creators" FOREIGN KEY ("organization_job_runner_registration_token_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_runner_token_creators" FOREIGN KEY ("organization_job_runner_token_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_template_creators" FOREIGN KEY ("organization_job_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_mapped_control_creators" FOREIGN KEY ("organization_mapped_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_narrative_creators" FOREIGN KEY ("organization_narrative_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_note_creators" FOREIGN KEY ("organization_note_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_notification_template_creators" FOREIGN KEY ("organization_notification_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_org_membership_creators" FOREIGN KEY ("organization_org_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_platform_creators" FOREIGN KEY ("organization_platform_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_policies_manager" FOREIGN KEY ("organization_policies_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_procedure_creators" FOREIGN KEY ("organization_procedure_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_program_creators" FOREIGN KEY ("organization_program_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_program_membership_creators" FOREIGN KEY ("organization_program_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_registry_manager" FOREIGN KEY ("organization_registry_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_remediation_creators" FOREIGN KEY ("organization_remediation_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_review_creators" FOREIGN KEY ("organization_review_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_risk_creators" FOREIGN KEY ("organization_risk_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_risk_manager" FOREIGN KEY ("organization_risk_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scan_creators" FOREIGN KEY ("organization_scan_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scheduled_job_creators" FOREIGN KEY ("organization_scheduled_job_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scheduled_job_run_creators" FOREIGN KEY ("organization_scheduled_job_run_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_sla_definition_creators" FOREIGN KEY ("organization_sla_definition_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_standard_creators" FOREIGN KEY ("organization_standard_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_subcontrol_creators" FOREIGN KEY ("organization_subcontrol_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_subprocessor_creators" FOREIGN KEY ("organization_subprocessor_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_subscriber_creators" FOREIGN KEY ("organization_subscriber_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_system_detail_creators" FOREIGN KEY ("organization_system_detail_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_tag_definition_creators" FOREIGN KEY ("organization_tag_definition_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_task_creators" FOREIGN KEY ("organization_task_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_template_creators" FOREIGN KEY ("organization_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_compliance_creators" FOREIGN KEY ("organization_trust_center_compliance_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_creators" FOREIGN KEY ("organization_trust_center_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_doc_creators" FOREIGN KEY ("organization_trust_center_doc_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_entity_creators" FOREIGN KEY ("organization_trust_center_entity_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_faq_creators" FOREIGN KEY ("organization_trust_center_faq_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_manager" FOREIGN KEY ("organization_trust_center_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_nda_request_creators" FOREIGN KEY ("organization_trust_center_nda_request_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_subprocessor_creators" FOREIGN KEY ("organization_trust_center_subprocessor_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_watermark_config_creators" FOREIGN KEY ("organization_trust_center_watermark_config_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_vendor_risk_score_creators" FOREIGN KEY ("organization_vendor_risk_score_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_vendor_scoring_config_creators" FOREIGN KEY ("organization_vendor_scoring_config_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_vulnerability_creators" FOREIGN KEY ("organization_vulnerability_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_workflow_definition_creators" FOREIGN KEY ("organization_workflow_definition_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_workflows_manager" FOREIGN KEY ("organization_workflows_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_sla_definitions_blocked_groups" FOREIGN KEY ("sla_definition_blocked_groups") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_sla_definitions_editors" FOREIGN KEY ("sla_definition_editors") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_compliances_blocked_groups" FOREIGN KEY ("trust_center_compliance_blocked_groups") REFERENCES "trust_center_compliances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_compliances_editors" FOREIGN KEY ("trust_center_compliance_editors") REFERENCES "trust_center_compliances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_docs_blocked_groups" FOREIGN KEY ("trust_center_doc_blocked_groups") REFERENCES "trust_center_docs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_docs_editors" FOREIGN KEY ("trust_center_doc_editors") REFERENCES "trust_center_docs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_entities_blocked_groups" FOREIGN KEY ("trust_center_entity_blocked_groups") REFERENCES "trust_center_entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_entities_editors" FOREIGN KEY ("trust_center_entity_editors") REFERENCES "trust_center_entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_faqs_blocked_groups" FOREIGN KEY ("trust_center_faq_blocked_groups") REFERENCES "trust_center_faqs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_faqs_editors" FOREIGN KEY ("trust_center_faq_editors") REFERENCES "trust_center_faqs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_nda_requests_blocked_groups" FOREIGN KEY ("trust_center_nda_request_blocked_groups") REFERENCES "trust_center_nda_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_nda_requests_editors" FOREIGN KEY ("trust_center_nda_request_editors") REFERENCES "trust_center_nda_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_settings_blocked_groups" FOREIGN KEY ("trust_center_setting_blocked_groups") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_settings_editors" FOREIGN KEY ("trust_center_setting_editors") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_subprocessors_blocked_groups" FOREIGN KEY ("trust_center_subprocessor_blocked_groups") REFERENCES "trust_center_subprocessors" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_subprocessors_editors" FOREIGN KEY ("trust_center_subprocessor_editors") REFERENCES "trust_center_subprocessors" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_watermark_configs_blocked_groups" FOREIGN KEY ("trust_center_watermark_config_blocked_groups") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_watermark_configs_editors" FOREIGN KEY ("trust_center_watermark_config_editors") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_centers_blocked_groups" FOREIGN KEY ("trust_center_blocked_groups") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_centers_editors" FOREIGN KEY ("trust_center_editors") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_blocked_groups" FOREIGN KEY ("vulnerability_blocked_groups") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_editors" FOREIGN KEY ("vulnerability_editors") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_viewers" FOREIGN KEY ("vulnerability_viewers") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_blocked_groups" FOREIGN KEY ("workflow_definition_blocked_groups") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_editors" FOREIGN KEY ("workflow_definition_editors") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_groups" FOREIGN KEY ("workflow_definition_groups") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_viewers" FOREIGN KEY ("workflow_definition_viewers") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "group_memberships" table
ALTER TABLE "group_memberships" ADD CONSTRAINT "group_memberships_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "group_memberships_org_memberships_org_membership" FOREIGN KEY ("group_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "group_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "group_settings" table
ALTER TABLE "group_settings" ADD CONSTRAINT "group_settings_groups_setting" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "hushes" table
ALTER TABLE "hushes" ADD CONSTRAINT "hushes_organizations_secrets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD CONSTRAINT "identity_holders_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_entities_employer" FOREIGN KEY ("employer_entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_organizations_identity_holders" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_users_identity_holder_profiles" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "impersonation_events" table
ALTER TABLE "impersonation_events" ADD CONSTRAINT "impersonation_events_organizations_impersonation_events" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "impersonation_events_users_impersonation_events" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "impersonation_events_users_targeted_impersonations" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_files_integrations" FOREIGN KEY ("file_integrations") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_groups_integrations" FOREIGN KEY ("group_integrations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_organizations_integrations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_platforms_integrations" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD CONSTRAINT "integration_runs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_events_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_files_request_file" FOREIGN KEY ("request_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_files_response_file" FOREIGN KEY ("response_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_integrations_integration_runs" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_organizations_integration_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD CONSTRAINT "integration_webhooks_integrations_integration_webhooks" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_webhooks_organizations_integration_webhooks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_internal_policies" FOREIGN KEY ("custom_type_enum_internal_policies") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_internal_policy_kind" FOREIGN KEY ("internal_policy_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_organizations_internal_policies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "invites" table
ALTER TABLE "invites" ADD CONSTRAINT "invites_organizations_invites" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "job_results" table
ALTER TABLE "job_results" ADD CONSTRAINT "job_results_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "job_results_organizations_job_results" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "job_results_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "job_runners" table
ALTER TABLE "job_runners" ADD CONSTRAINT "job_runners_organizations_job_runners" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "job_runner_registration_tokens" table
ALTER TABLE "job_runner_registration_tokens" ADD CONSTRAINT "job_runner_registration_tokens_daddf3e078805108b2d174df258ddb4b" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "job_runner_registration_tokens_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "job_runner_tokens" table
ALTER TABLE "job_runner_tokens" ADD CONSTRAINT "job_runner_tokens_organizations_job_runner_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "job_templates" table
ALTER TABLE "job_templates" ADD CONSTRAINT "job_templates_organizations_job_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD CONSTRAINT "mapped_controls_organizations_mapped_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "narratives" table
ALTER TABLE "narratives" ADD CONSTRAINT "narratives_control_objectives_narratives" FOREIGN KEY ("control_objective_narratives") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_subcontrols_narratives" FOREIGN KEY ("subcontrol_narratives") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "notes" table
ALTER TABLE "notes" ADD CONSTRAINT "notes_controls_comments" FOREIGN KEY ("control_comments") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_discussions_comments" FOREIGN KEY ("discussion_id") REFERENCES "discussions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_entities_notes" FOREIGN KEY ("entity_notes") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_evidences_comments" FOREIGN KEY ("evidence_comments") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_findings_comments" FOREIGN KEY ("finding_comments") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_internal_policies_comments" FOREIGN KEY ("internal_policy_comments") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_organizations_notes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_procedures_comments" FOREIGN KEY ("procedure_comments") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_programs_notes" FOREIGN KEY ("program_notes") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_remediations_comments" FOREIGN KEY ("remediation_comments") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_reviews_comments" FOREIGN KEY ("review_comments") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_risks_comments" FOREIGN KEY ("risk_comments") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_subcontrols_comments" FOREIGN KEY ("subcontrol_comments") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_tasks_comments" FOREIGN KEY ("task_comments") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_trust_centers_posts" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_vulnerabilities_comments" FOREIGN KEY ("vulnerability_comments") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "notifications" table
ALTER TABLE "notifications" ADD CONSTRAINT "notifications_notification_templates_notifications" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notifications_organizations_notifications" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "notification_preferences" table
ALTER TABLE "notification_preferences" ADD CONSTRAINT "notification_preferences_notif_aabd0a3ca9e335110ce7c2348e4f4cf0" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_preferences_organizations_notification_preferences" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_preferences_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" ADD CONSTRAINT "notification_templates_email_templates_notification_templates" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_templates_integrations_notification_templates" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_templates_organizations_notification_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_templates_workflo_439a17f2830fbf868eeb61d3d3fdac37" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "onboardings" table
ALTER TABLE "onboardings" ADD CONSTRAINT "onboardings_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "org_memberships" table
ALTER TABLE "org_memberships" ADD CONSTRAINT "org_memberships_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "org_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "org_modules" table
ALTER TABLE "org_modules" ADD CONSTRAINT "org_modules_org_products_org_modules" FOREIGN KEY ("org_product_org_modules") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_modules_org_subscriptions_modules" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_modules_organizations_org_modules" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "org_prices" table
ALTER TABLE "org_prices" ADD CONSTRAINT "org_prices_org_subscriptions_prices" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_prices_organizations_org_prices" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "org_products" table
ALTER TABLE "org_products" ADD CONSTRAINT "org_products_org_modules_org_products" FOREIGN KEY ("org_module_org_products") REFERENCES "org_modules" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_products_org_subscriptions_products" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_products_organizations_org_products" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD CONSTRAINT "org_subscriptions_organizations_org_subscriptions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "organizations" table
ALTER TABLE "organizations" ADD CONSTRAINT "organizations_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD CONSTRAINT "organization_settings_organizations_setting" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" ADD CONSTRAINT "password_reset_tokens_users_password_reset_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD CONSTRAINT "personal_access_tokens_users_personal_access_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "platforms" table
ALTER TABLE "platforms" ADD CONSTRAINT "platforms_custom_type_enums_access_model" FOREIGN KEY ("access_model_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_criticality" FOREIGN KEY ("criticality_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_encryption_status" FOREIGN KEY ("encryption_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platform_data_classification" FOREIGN KEY ("platform_data_classification_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platform_kind" FOREIGN KEY ("platform_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platforms" FOREIGN KEY ("custom_type_enum_platforms") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_security_tier" FOREIGN KEY ("security_tier_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_business_owner_group" FOREIGN KEY ("business_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_security_owner_group" FOREIGN KEY ("security_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_technical_owner_group" FOREIGN KEY ("technical_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_identity_holders_access_platforms" FOREIGN KEY ("identity_holder_access_platforms") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_organizations_platforms" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_business_owner_user" FOREIGN KEY ("business_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_platforms_owned" FOREIGN KEY ("platform_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_security_owner_user" FOREIGN KEY ("security_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_technical_owner_user" FOREIGN KEY ("technical_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_procedure_kind" FOREIGN KEY ("procedure_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_procedures" FOREIGN KEY ("custom_type_enum_procedures") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_organizations_procedures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "programs" table
ALTER TABLE "programs" ADD CONSTRAINT "programs_custom_type_enums_program_kind" FOREIGN KEY ("program_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_custom_type_enums_programs" FOREIGN KEY ("custom_type_enum_programs") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_organizations_programs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_users_programs_owned" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" ADD CONSTRAINT "program_memberships_org_memberships_org_membership" FOREIGN KEY ("program_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "program_memberships_programs_program" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "program_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "remediations" table
ALTER TABLE "remediations" ADD CONSTRAINT "remediations_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_organizations_remediations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "reviews" table
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_organizations_reviews" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_users_reviewer" FOREIGN KEY ("reviewer_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risk_categories" FOREIGN KEY ("custom_type_enum_risk_categories") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risk_category" FOREIGN KEY ("risk_category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risk_kind" FOREIGN KEY ("risk_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risks" FOREIGN KEY ("custom_type_enum_risks") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("stakeholder_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "sla_definitions" table
ALTER TABLE "sla_definitions" ADD CONSTRAINT "sla_definitions_organizations_sla_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "scans" table
ALTER TABLE "scans" ADD CONSTRAINT "scans_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_performed_by_group" FOREIGN KEY ("performed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_organizations_scans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_platforms_generated_scans" FOREIGN KEY ("generated_by_platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_risks_scans" FOREIGN KEY ("risk_scans") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_performed_by_user" FOREIGN KEY ("performed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD CONSTRAINT "scheduled_jobs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs" FOREIGN KEY ("job_id") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "scheduled_jobs_organizations_scheduled_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ADD CONSTRAINT "scheduled_job_runs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "scheduled_job_runs_organizations_scheduled_job_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scheduled_job_runs_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "standards" table
ALTER TABLE "standards" ADD CONSTRAINT "standards_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "standards_organizations_standards" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_controls_subcontrols" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "subcontrols_custom_type_enums_subcontrol_kind" FOREIGN KEY ("subcontrol_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_custom_type_enums_subcontrols" FOREIGN KEY ("custom_type_enum_subcontrols") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_entities_responsible_party" FOREIGN KEY ("responsible_party_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_programs_subcontrols" FOREIGN KEY ("program_subcontrols") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_subcontrols" FOREIGN KEY ("user_subcontrols") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subprocessors" table
ALTER TABLE "subprocessors" ADD CONSTRAINT "subprocessors_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subprocessors_organizations_subprocessors" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subscribers" table
ALTER TABLE "subscribers" ADD CONSTRAINT "subscribers_contacts_subscribers" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_organizations_subscribers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_trust_centers_subscribers" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_users_subscribers" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "system_details" table
ALTER TABLE "system_details" ADD CONSTRAINT "system_details_organizations_system_details" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD CONSTRAINT "tfa_settings_users_tfa_settings" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD CONSTRAINT "tag_definitions_organizations_tag_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tag_definitions_workflow_definitions_tag_definitions" FOREIGN KEY ("workflow_definition_tag_definitions") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_task_kind" FOREIGN KEY ("task_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_tasks" FOREIGN KEY ("custom_type_enum_tasks") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_integrations_tasks" FOREIGN KEY ("integration_tasks") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_remediations_tasks" FOREIGN KEY ("remediation_tasks") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_reviews_tasks" FOREIGN KEY ("review_tasks") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_organizations_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_trust_centers_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD CONSTRAINT "trust_centers_custom_domains_custom_domain" FOREIGN KEY ("custom_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_custom_domains_preview_domain" FOREIGN KEY ("preview_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_organizations_trust_centers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_preview_setting" FOREIGN KEY ("trust_center_preview_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_setting" FOREIGN KEY ("trust_center_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_watermark_configs_watermark_config" FOREIGN KEY ("trust_center_watermark_config") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" ADD CONSTRAINT "trust_center_compliances_standards_trust_center_compliances" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_compliances_trust_centers_trust_center_compliances" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD CONSTRAINT "trust_center_docs_custom_type_enums_trust_center_doc_kind" FOREIGN KEY ("trust_center_doc_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_files_original_file" FOREIGN KEY ("original_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_standards_trust_center_docs" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_trust_center_nda_requests_trust_center_docs" FOREIGN KEY ("trust_center_nda_request_trust_center_docs") REFERENCES "trust_center_nda_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_trust_centers_trust_center_docs" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_entities" table
ALTER TABLE "trust_center_entities" ADD CONSTRAINT "trust_center_entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_entities_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_entities_files_trust_center_entities" FOREIGN KEY ("file_trust_center_entities") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_entities_trust_centers_trust_center_entities" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" ADD CONSTRAINT "trust_center_faqs_custom_type_enums_trust_center_faq_kind" FOREIGN KEY ("trust_center_faq_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_faqs_notes_trust_center_faqs" FOREIGN KEY ("note_id") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_faqs_trust_centers_trust_center_faqs" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD CONSTRAINT "trust_center_nda_requests_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_nda_requests_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_nda_requests_trus_166c4573710ee5957bac7d4b99111f81" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_nda_requests_users_approved_by_user" FOREIGN KEY ("approved_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD CONSTRAINT "trust_center_settings_files_favicon_file" FOREIGN KEY ("favicon_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_settings_files_hero_image_file" FOREIGN KEY ("hero_image_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_settings_files_logo_file" FOREIGN KEY ("logo_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_settings_groups_nda_approver_group" FOREIGN KEY ("nda_approver_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" ADD CONSTRAINT "trust_center_subprocessors_cus_d5ebb915269b07a0bf77b5b0ec180583" FOREIGN KEY ("trust_center_subprocessor_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_subprocessors_sub_24055b695e9bd0e49b3edea05d355a0b" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_subprocessors_tru_bb0fd7936579c86ecda7d42ebfe60199" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ADD CONSTRAINT "trust_center_watermark_configs_e2f038ca8412a7e2b03e1fad46be2f7f" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_watermark_configs_files_file" FOREIGN KEY ("logo_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "user_settings" table
ALTER TABLE "user_settings" ADD CONSTRAINT "user_settings_organizations_default_org" FOREIGN KEY ("user_setting_default_org") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_settings_users_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" ADD CONSTRAINT "vendor_risk_scores_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_assessment_responses_vendor_risk_scores" FOREIGN KEY ("assessment_response_vendor_risk_scores") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_entities_entity" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "vendor_risk_scores_entities_vendor_risk_scores" FOREIGN KEY ("entity_vendor_risk_scores") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_organizations_vendor_risk_scores" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_risk_scores" FOREIGN KEY ("vendor_scoring_config_vendor_risk_scores") REFERENCES "vendor_scoring_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_scoring_config" FOREIGN KEY ("vendor_scoring_config_id") REFERENCES "vendor_scoring_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "vendor_scoring_configs" table
ALTER TABLE "vendor_scoring_configs" ADD CONSTRAINT "vendor_scoring_configs_organizations_vendor_scoring_configs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD CONSTRAINT "vulnerabilities_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_custom_type_enums_vulnerability_status" FOREIGN KEY ("vulnerability_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_organizations_vulnerabilities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "webauthns" table
ALTER TABLE "webauthns" ADD CONSTRAINT "webauthns_users_webauthns" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD CONSTRAINT "workflow_assignments_groups_group" FOREIGN KEY ("actor_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_organizations_workflow_assignments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_users_user" FOREIGN KEY ("actor_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_workflow_instances_workflow_assignments" FOREIGN KEY ("workflow_instance_workflow_assignments") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "workflow_assignment_targets" table
ALTER TABLE "workflow_assignment_targets" ADD CONSTRAINT "workflow_assignment_targets_groups_group" FOREIGN KEY ("target_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_or_8bb74468c70e1b9fcce1d5b038516f9a" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_users_user" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_wo_35919ebc89c62ef82cb5889ff40ce351" FOREIGN KEY ("workflow_assignment_workflow_assignment_targets") REFERENCES "workflow_assignments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_wo_6077e6f4bf744947c345bb2733c1c240" FOREIGN KEY ("workflow_assignment_id") REFERENCES "workflow_assignments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ADD CONSTRAINT "workflow_definitions_organizations_workflow_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_events" table
ALTER TABLE "workflow_events" ADD CONSTRAINT "workflow_events_organizations_workflow_events" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_events_workflow_instances_workflow_events" FOREIGN KEY ("workflow_instance_workflow_events") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_events_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD CONSTRAINT "workflow_instances_action_plans_action_plan" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_campaign_targets_campaign_target" FOREIGN KEY ("campaign_target_id") REFERENCES "campaign_targets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_campaigns_campaign" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_identity_holders_identity_holder" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_organizations_workflow_instances" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_platforms_platform" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_procedures_procedure" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_subcontrols_subcontrol" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_workflow_definitions_workflow_definition" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "workflow_instances_workflow_proposals_workflow_proposal" FOREIGN KEY ("workflow_proposal_id") REFERENCES "workflow_proposals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD CONSTRAINT "workflow_object_refs_action_plans_action_plan" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_campaign_targets_campaign_target" FOREIGN KEY ("campaign_target_id") REFERENCES "campaign_targets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_campaigns_campaign" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_directory_accounts_directory_account" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_directory_groups_directory_group" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_directory_memberships_directory_membership" FOREIGN KEY ("directory_membership_id") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_identity_holders_identity_holder" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_organizations_workflow_object_refs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_platforms_platform" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_procedures_procedure" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_subcontrols_subcontrol" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "workflow_object_refs_workflow_instances_workflow_object_refs" FOREIGN KEY ("workflow_instance_workflow_object_refs") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" ADD CONSTRAINT "workflow_proposals_organizations_workflow_proposals" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_proposals_users_user" FOREIGN KEY ("submitted_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_proposals_workflow_object_refs_workflow_object_ref" FOREIGN KEY ("workflow_object_ref_id") REFERENCES "workflow_object_refs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "action_plan_blocked_groups" table
ALTER TABLE "action_plan_blocked_groups" ADD CONSTRAINT "action_plan_blocked_groups_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "action_plan_editors" table
ALTER TABLE "action_plan_editors" ADD CONSTRAINT "action_plan_editors_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "action_plan_viewers" table
ALTER TABLE "action_plan_viewers" ADD CONSTRAINT "action_plan_viewers_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "action_plan_tasks" table
ALTER TABLE "action_plan_tasks" ADD CONSTRAINT "action_plan_tasks_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "asset_connected_assets" table
ALTER TABLE "asset_connected_assets" ADD CONSTRAINT "asset_connected_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "asset_connected_assets_connected_from_id" FOREIGN KEY ("connected_from_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_blocked_groups" table
ALTER TABLE "campaign_blocked_groups" ADD CONSTRAINT "campaign_blocked_groups_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_editors" table
ALTER TABLE "campaign_editors" ADD CONSTRAINT "campaign_editors_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_viewers" table
ALTER TABLE "campaign_viewers" ADD CONSTRAINT "campaign_viewers_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_contacts" table
ALTER TABLE "campaign_contacts" ADD CONSTRAINT "campaign_contacts_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_users" table
ALTER TABLE "campaign_users" ADD CONSTRAINT "campaign_users_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_groups" table
ALTER TABLE "campaign_groups" ADD CONSTRAINT "campaign_groups_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "campaign_identity_holders" table
ALTER TABLE "campaign_identity_holders" ADD CONSTRAINT "campaign_identity_holders_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "check_result_controls" table
ALTER TABLE "check_result_controls" ADD CONSTRAINT "check_result_controls_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "check_result_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "contact_files" table
ALTER TABLE "contact_files" ADD CONSTRAINT "contact_files_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "contact_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_control_objectives" table
ALTER TABLE "control_control_objectives" ADD CONSTRAINT "control_control_objectives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_tasks" table
ALTER TABLE "control_tasks" ADD CONSTRAINT "control_tasks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_narratives" table
ALTER TABLE "control_narratives" ADD CONSTRAINT "control_narratives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_risks" table
ALTER TABLE "control_risks" ADD CONSTRAINT "control_risks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_action_plans" table
ALTER TABLE "control_action_plans" ADD CONSTRAINT "control_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_action_plans_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_procedures" table
ALTER TABLE "control_procedures" ADD CONSTRAINT "control_procedures_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_scans" table
ALTER TABLE "control_scans" ADD CONSTRAINT "control_scans_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" ADD CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_editors" table
ALTER TABLE "control_editors" ADD CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_assets" table
ALTER TABLE "control_assets" ADD CONSTRAINT "control_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_assets_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_entities" table
ALTER TABLE "control_entities" ADD CONSTRAINT "control_entities_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_identity_holders" table
ALTER TABLE "control_identity_holders" ADD CONSTRAINT "control_identity_holders_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_campaigns" table
ALTER TABLE "control_campaigns" ADD CONSTRAINT "control_campaigns_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_campaigns_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_control_implementations" table
ALTER TABLE "control_control_implementations" ADD CONSTRAINT "control_control_implementations_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_implementation_blocked_groups" table
ALTER TABLE "control_implementation_blocked_groups" ADD CONSTRAINT "control_implementation_blocked_groups_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_implementation_editors" table
ALTER TABLE "control_implementation_editors" ADD CONSTRAINT "control_implementation_editors_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_implementation_viewers" table
ALTER TABLE "control_implementation_viewers" ADD CONSTRAINT "control_implementation_viewers_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_implementation_tasks" table
ALTER TABLE "control_implementation_tasks" ADD CONSTRAINT "control_implementation_tasks_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_objective_blocked_groups" table
ALTER TABLE "control_objective_blocked_groups" ADD CONSTRAINT "control_objective_blocked_groups_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_objective_editors" table
ALTER TABLE "control_objective_editors" ADD CONSTRAINT "control_objective_editors_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_objective_viewers" table
ALTER TABLE "control_objective_viewers" ADD CONSTRAINT "control_objective_viewers_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_objective_tasks" table
ALTER TABLE "control_objective_tasks" ADD CONSTRAINT "control_objective_tasks_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "document_data_files" table
ALTER TABLE "document_data_files" ADD CONSTRAINT "document_data_files_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "document_data_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_blocked_groups" table
ALTER TABLE "entity_blocked_groups" ADD CONSTRAINT "entity_blocked_groups_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_editors" table
ALTER TABLE "entity_editors" ADD CONSTRAINT "entity_editors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_contacts" table
ALTER TABLE "entity_contacts" ADD CONSTRAINT "entity_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_contacts_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_documents" table
ALTER TABLE "entity_documents" ADD CONSTRAINT "entity_documents_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_documents_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_files" table
ALTER TABLE "entity_files" ADD CONSTRAINT "entity_files_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_assets" table
ALTER TABLE "entity_assets" ADD CONSTRAINT "entity_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_assets_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_system_details" table
ALTER TABLE "entity_system_details" ADD CONSTRAINT "entity_system_details_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_integrations" table
ALTER TABLE "entity_integrations" ADD CONSTRAINT "entity_integrations_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_integrations_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_subprocessors" table
ALTER TABLE "entity_subprocessors" ADD CONSTRAINT "entity_subprocessors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_subprocessors_subprocessor_id" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_controls" table
ALTER TABLE "evidence_controls" ADD CONSTRAINT "evidence_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_controls_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_subcontrols" table
ALTER TABLE "evidence_subcontrols" ADD CONSTRAINT "evidence_subcontrols_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_control_objectives" table
ALTER TABLE "evidence_control_objectives" ADD CONSTRAINT "evidence_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_control_objectives_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_files" table
ALTER TABLE "evidence_files" ADD CONSTRAINT "evidence_files_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "file_events" table
ALTER TABLE "file_events" ADD CONSTRAINT "file_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_events_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "file_secrets" table
ALTER TABLE "file_secrets" ADD CONSTRAINT "file_secrets_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_blocked_groups" table
ALTER TABLE "finding_blocked_groups" ADD CONSTRAINT "finding_blocked_groups_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_editors" table
ALTER TABLE "finding_editors" ADD CONSTRAINT "finding_editors_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_vulnerabilities" table
ALTER TABLE "finding_vulnerabilities" ADD CONSTRAINT "finding_vulnerabilities_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_action_plans" table
ALTER TABLE "finding_action_plans" ADD CONSTRAINT "finding_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_action_plans_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_subcontrols" table
ALTER TABLE "finding_subcontrols" ADD CONSTRAINT "finding_subcontrols_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_risks" table
ALTER TABLE "finding_risks" ADD CONSTRAINT "finding_risks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_programs" table
ALTER TABLE "finding_programs" ADD CONSTRAINT "finding_programs_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_assets" table
ALTER TABLE "finding_assets" ADD CONSTRAINT "finding_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_assets_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_entities" table
ALTER TABLE "finding_entities" ADD CONSTRAINT "finding_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_entities_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_scans" table
ALTER TABLE "finding_scans" ADD CONSTRAINT "finding_scans_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_tasks" table
ALTER TABLE "finding_tasks" ADD CONSTRAINT "finding_tasks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_directory_accounts" table
ALTER TABLE "finding_directory_accounts" ADD CONSTRAINT "finding_directory_accounts_directory_account_id" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_directory_accounts_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_identity_holders" table
ALTER TABLE "finding_identity_holders" ADD CONSTRAINT "finding_identity_holders_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "finding_check_results" table
ALTER TABLE "finding_check_results" ADD CONSTRAINT "finding_check_results_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_check_results_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "group_events" table
ALTER TABLE "group_events" ADD CONSTRAINT "group_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_events_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "group_files" table
ALTER TABLE "group_files" ADD CONSTRAINT "group_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_files_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "group_tasks" table
ALTER TABLE "group_tasks" ADD CONSTRAINT "group_tasks_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "group_membership_events" table
ALTER TABLE "group_membership_events" ADD CONSTRAINT "group_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_membership_events_group_membership_id" FOREIGN KEY ("group_membership_id") REFERENCES "group_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "hush_events" table
ALTER TABLE "hush_events" ADD CONSTRAINT "hush_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "hush_events_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "identity_holder_assessments" table
ALTER TABLE "identity_holder_assessments" ADD CONSTRAINT "identity_holder_assessments_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_assessments_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "identity_holder_templates" table
ALTER TABLE "identity_holder_templates" ADD CONSTRAINT "identity_holder_templates_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_templates_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "identity_holder_assets" table
ALTER TABLE "identity_holder_assets" ADD CONSTRAINT "identity_holder_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_assets_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "identity_holder_entities" table
ALTER TABLE "identity_holder_entities" ADD CONSTRAINT "identity_holder_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_entities_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "identity_holder_tasks" table
ALTER TABLE "identity_holder_tasks" ADD CONSTRAINT "identity_holder_tasks_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "identity_holder_files" table
ALTER TABLE "identity_holder_files" ADD CONSTRAINT "identity_holder_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_files_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_secrets" table
ALTER TABLE "integration_secrets" ADD CONSTRAINT "integration_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_secrets_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_events" table
ALTER TABLE "integration_events" ADD CONSTRAINT "integration_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_events_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_findings" table
ALTER TABLE "integration_findings" ADD CONSTRAINT "integration_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_findings_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_vulnerabilities" table
ALTER TABLE "integration_vulnerabilities" ADD CONSTRAINT "integration_vulnerabilities_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_internal_policies" table
ALTER TABLE "integration_internal_policies" ADD CONSTRAINT "integration_internal_policies_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_reviews" table
ALTER TABLE "integration_reviews" ADD CONSTRAINT "integration_reviews_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_reviews_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_remediations" table
ALTER TABLE "integration_remediations" ADD CONSTRAINT "integration_remediations_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_action_plans" table
ALTER TABLE "integration_action_plans" ADD CONSTRAINT "integration_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_action_plans_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_blocked_groups" table
ALTER TABLE "internal_policy_blocked_groups" ADD CONSTRAINT "internal_policy_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_blocked_groups_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_editors" table
ALTER TABLE "internal_policy_editors" ADD CONSTRAINT "internal_policy_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_editors_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_control_objectives" table
ALTER TABLE "internal_policy_control_objectives" ADD CONSTRAINT "internal_policy_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_control_objectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_controls" table
ALTER TABLE "internal_policy_controls" ADD CONSTRAINT "internal_policy_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_controls_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_subcontrols" table
ALTER TABLE "internal_policy_subcontrols" ADD CONSTRAINT "internal_policy_subcontrols_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" ADD CONSTRAINT "internal_policy_narratives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" ADD CONSTRAINT "internal_policy_tasks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_risks" table
ALTER TABLE "internal_policy_risks" ADD CONSTRAINT "internal_policy_risks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_assets" table
ALTER TABLE "internal_policy_assets" ADD CONSTRAINT "internal_policy_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_assets_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_entities" table
ALTER TABLE "internal_policy_entities" ADD CONSTRAINT "internal_policy_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_entities_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_identity_holders" table
ALTER TABLE "internal_policy_identity_holders" ADD CONSTRAINT "internal_policy_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_identity_holders_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "invite_events" table
ALTER TABLE "invite_events" ADD CONSTRAINT "invite_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_events_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "invite_groups" table
ALTER TABLE "invite_groups" ADD CONSTRAINT "invite_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_groups_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "job_runner_job_runner_tokens" table
ALTER TABLE "job_runner_job_runner_tokens" ADD CONSTRAINT "job_runner_job_runner_tokens_job_runner_id" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "job_runner_job_runner_tokens_job_runner_token_id" FOREIGN KEY ("job_runner_token_id") REFERENCES "job_runner_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_blocked_groups" table
ALTER TABLE "mapped_control_blocked_groups" ADD CONSTRAINT "mapped_control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_blocked_groups_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_editors" table
ALTER TABLE "mapped_control_editors" ADD CONSTRAINT "mapped_control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_editors_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_from_controls" table
ALTER TABLE "mapped_control_from_controls" ADD CONSTRAINT "mapped_control_from_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_from_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_to_controls" table
ALTER TABLE "mapped_control_to_controls" ADD CONSTRAINT "mapped_control_to_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_to_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_from_subcontrols" table
ALTER TABLE "mapped_control_from_subcontrols" ADD CONSTRAINT "mapped_control_from_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_from_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_to_subcontrols" table
ALTER TABLE "mapped_control_to_subcontrols" ADD CONSTRAINT "mapped_control_to_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_to_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" ADD CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "narrative_editors" table
ALTER TABLE "narrative_editors" ADD CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" ADD CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "org_membership_events" table
ALTER TABLE "org_membership_events" ADD CONSTRAINT "org_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_membership_events_org_membership_id" FOREIGN KEY ("org_membership_id") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "org_module_org_prices" table
ALTER TABLE "org_module_org_prices" ADD CONSTRAINT "org_module_org_prices_org_module_id" FOREIGN KEY ("org_module_id") REFERENCES "org_modules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_module_org_prices_org_price_id" FOREIGN KEY ("org_price_id") REFERENCES "org_prices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "org_product_org_prices" table
ALTER TABLE "org_product_org_prices" ADD CONSTRAINT "org_product_org_prices_org_price_id" FOREIGN KEY ("org_price_id") REFERENCES "org_prices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_product_org_prices_org_product_id" FOREIGN KEY ("org_product_id") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "org_subscription_events" table
ALTER TABLE "org_subscription_events" ADD CONSTRAINT "org_subscription_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_subscription_events_org_subscription_id" FOREIGN KEY ("org_subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "organization_personal_access_tokens" table
ALTER TABLE "organization_personal_access_tokens" ADD CONSTRAINT "organization_personal_access_tokens_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_personal_access_tokens_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "organization_files" table
ALTER TABLE "organization_files" ADD CONSTRAINT "organization_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_files_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "organization_events" table
ALTER TABLE "organization_events" ADD CONSTRAINT "organization_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_events_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "organization_setting_files" table
ALTER TABLE "organization_setting_files" ADD CONSTRAINT "organization_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_setting_files_organization_setting_id" FOREIGN KEY ("organization_setting_id") REFERENCES "organization_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "personal_access_token_events" table
ALTER TABLE "personal_access_token_events" ADD CONSTRAINT "personal_access_token_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "personal_access_token_events_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_blocked_groups" table
ALTER TABLE "platform_blocked_groups" ADD CONSTRAINT "platform_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_blocked_groups_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_editors" table
ALTER TABLE "platform_editors" ADD CONSTRAINT "platform_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_editors_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_viewers" table
ALTER TABLE "platform_viewers" ADD CONSTRAINT "platform_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_viewers_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_assets" table
ALTER TABLE "platform_assets" ADD CONSTRAINT "platform_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_assets_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_entities" table
ALTER TABLE "platform_entities" ADD CONSTRAINT "platform_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_entities_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_evidence" table
ALTER TABLE "platform_evidence" ADD CONSTRAINT "platform_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_evidence_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_files" table
ALTER TABLE "platform_files" ADD CONSTRAINT "platform_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_files_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_risks" table
ALTER TABLE "platform_risks" ADD CONSTRAINT "platform_risks_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_controls" table
ALTER TABLE "platform_controls" ADD CONSTRAINT "platform_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_controls_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_assessments" table
ALTER TABLE "platform_assessments" ADD CONSTRAINT "platform_assessments_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_assessments_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_scans" table
ALTER TABLE "platform_scans" ADD CONSTRAINT "platform_scans_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_tasks" table
ALTER TABLE "platform_tasks" ADD CONSTRAINT "platform_tasks_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_identity_holders" table
ALTER TABLE "platform_identity_holders" ADD CONSTRAINT "platform_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_identity_holders_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_source_entities" table
ALTER TABLE "platform_source_entities" ADD CONSTRAINT "platform_source_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_source_entities_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_out_of_scope_assets" table
ALTER TABLE "platform_out_of_scope_assets" ADD CONSTRAINT "platform_out_of_scope_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_out_of_scope_assets_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_out_of_scope_vendors" table
ALTER TABLE "platform_out_of_scope_vendors" ADD CONSTRAINT "platform_out_of_scope_vendors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_out_of_scope_vendors_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_applicable_frameworks" table
ALTER TABLE "platform_applicable_frameworks" ADD CONSTRAINT "platform_applicable_frameworks_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_applicable_frameworks_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "platform_system_details" table
ALTER TABLE "platform_system_details" ADD CONSTRAINT "platform_system_details_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_blocked_groups" table
ALTER TABLE "procedure_blocked_groups" ADD CONSTRAINT "procedure_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_blocked_groups_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_editors" table
ALTER TABLE "procedure_editors" ADD CONSTRAINT "procedure_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_editors_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_narratives" table
ALTER TABLE "procedure_narratives" ADD CONSTRAINT "procedure_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_narratives_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_risks" table
ALTER TABLE "procedure_risks" ADD CONSTRAINT "procedure_risks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_tasks" table
ALTER TABLE "procedure_tasks" ADD CONSTRAINT "procedure_tasks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_blocked_groups" table
ALTER TABLE "program_blocked_groups" ADD CONSTRAINT "program_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_blocked_groups_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_editors" table
ALTER TABLE "program_editors" ADD CONSTRAINT "program_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_editors_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_viewers" table
ALTER TABLE "program_viewers" ADD CONSTRAINT "program_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_viewers_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_controls" table
ALTER TABLE "program_controls" ADD CONSTRAINT "program_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_controls_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_control_objectives" table
ALTER TABLE "program_control_objectives" ADD CONSTRAINT "program_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_control_objectives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_internal_policies" table
ALTER TABLE "program_internal_policies" ADD CONSTRAINT "program_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_internal_policies_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_procedures" table
ALTER TABLE "program_procedures" ADD CONSTRAINT "program_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_procedures_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_risks" table
ALTER TABLE "program_risks" ADD CONSTRAINT "program_risks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_tasks" table
ALTER TABLE "program_tasks" ADD CONSTRAINT "program_tasks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_files" table
ALTER TABLE "program_files" ADD CONSTRAINT "program_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_files_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_evidence" table
ALTER TABLE "program_evidence" ADD CONSTRAINT "program_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_evidence_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_narratives" table
ALTER TABLE "program_narratives" ADD CONSTRAINT "program_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_narratives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_action_plans" table
ALTER TABLE "program_action_plans" ADD CONSTRAINT "program_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_action_plans_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "program_system_details" table
ALTER TABLE "program_system_details" ADD CONSTRAINT "program_system_details_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_blocked_groups" table
ALTER TABLE "remediation_blocked_groups" ADD CONSTRAINT "remediation_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_blocked_groups_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_editors" table
ALTER TABLE "remediation_editors" ADD CONSTRAINT "remediation_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_editors_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_findings" table
ALTER TABLE "remediation_findings" ADD CONSTRAINT "remediation_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_findings_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_vulnerabilities" table
ALTER TABLE "remediation_vulnerabilities" ADD CONSTRAINT "remediation_vulnerabilities_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_action_plans" table
ALTER TABLE "remediation_action_plans" ADD CONSTRAINT "remediation_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_action_plans_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_controls" table
ALTER TABLE "remediation_controls" ADD CONSTRAINT "remediation_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_controls_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_subcontrols" table
ALTER TABLE "remediation_subcontrols" ADD CONSTRAINT "remediation_subcontrols_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_risks" table
ALTER TABLE "remediation_risks" ADD CONSTRAINT "remediation_risks_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_programs" table
ALTER TABLE "remediation_programs" ADD CONSTRAINT "remediation_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_programs_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_assets" table
ALTER TABLE "remediation_assets" ADD CONSTRAINT "remediation_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_assets_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "remediation_entities" table
ALTER TABLE "remediation_entities" ADD CONSTRAINT "remediation_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_entities_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_blocked_groups" table
ALTER TABLE "review_blocked_groups" ADD CONSTRAINT "review_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_blocked_groups_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_editors" table
ALTER TABLE "review_editors" ADD CONSTRAINT "review_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_editors_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_findings" table
ALTER TABLE "review_findings" ADD CONSTRAINT "review_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_findings_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_vulnerabilities" table
ALTER TABLE "review_vulnerabilities" ADD CONSTRAINT "review_vulnerabilities_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_action_plans" table
ALTER TABLE "review_action_plans" ADD CONSTRAINT "review_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_action_plans_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_remediations" table
ALTER TABLE "review_remediations" ADD CONSTRAINT "review_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_remediations_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_controls" table
ALTER TABLE "review_controls" ADD CONSTRAINT "review_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_controls_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_subcontrols" table
ALTER TABLE "review_subcontrols" ADD CONSTRAINT "review_subcontrols_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_risks" table
ALTER TABLE "review_risks" ADD CONSTRAINT "review_risks_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_programs" table
ALTER TABLE "review_programs" ADD CONSTRAINT "review_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_programs_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_assets" table
ALTER TABLE "review_assets" ADD CONSTRAINT "review_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_assets_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_entities" table
ALTER TABLE "review_entities" ADD CONSTRAINT "review_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_entities_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "review_internal_policies" table
ALTER TABLE "review_internal_policies" ADD CONSTRAINT "review_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_internal_policies_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" ADD CONSTRAINT "risk_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_blocked_groups_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_editors" table
ALTER TABLE "risk_editors" ADD CONSTRAINT "risk_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_editors_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_viewers" table
ALTER TABLE "risk_viewers" ADD CONSTRAINT "risk_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_viewers_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_action_plans" table
ALTER TABLE "risk_action_plans" ADD CONSTRAINT "risk_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_action_plans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_tasks" table
ALTER TABLE "risk_tasks" ADD CONSTRAINT "risk_tasks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_blocked_groups" table
ALTER TABLE "scan_blocked_groups" ADD CONSTRAINT "scan_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_blocked_groups_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_editors" table
ALTER TABLE "scan_editors" ADD CONSTRAINT "scan_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_editors_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_assets" table
ALTER TABLE "scan_assets" ADD CONSTRAINT "scan_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_assets_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_entities" table
ALTER TABLE "scan_entities" ADD CONSTRAINT "scan_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_entities_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_evidence" table
ALTER TABLE "scan_evidence" ADD CONSTRAINT "scan_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_evidence_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_files" table
ALTER TABLE "scan_files" ADD CONSTRAINT "scan_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_files_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_remediations" table
ALTER TABLE "scan_remediations" ADD CONSTRAINT "scan_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_remediations_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_action_plans" table
ALTER TABLE "scan_action_plans" ADD CONSTRAINT "scan_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_action_plans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scan_tasks" table
ALTER TABLE "scan_tasks" ADD CONSTRAINT "scan_tasks_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scheduled_job_controls" table
ALTER TABLE "scheduled_job_controls" ADD CONSTRAINT "scheduled_job_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scheduled_job_controls_scheduled_job_id" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "scheduled_job_subcontrols" table
ALTER TABLE "scheduled_job_subcontrols" ADD CONSTRAINT "scheduled_job_subcontrols_scheduled_job_id" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scheduled_job_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" ADD CONSTRAINT "subcontrol_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_objectives_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" ADD CONSTRAINT "subcontrol_tasks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_risks" table
ALTER TABLE "subcontrol_risks" ADD CONSTRAINT "subcontrol_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_risks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_procedures" table
ALTER TABLE "subcontrol_procedures" ADD CONSTRAINT "subcontrol_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_procedures_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_scans" table
ALTER TABLE "subcontrol_scans" ADD CONSTRAINT "subcontrol_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_scans_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_control_implementations" table
ALTER TABLE "subcontrol_control_implementations" ADD CONSTRAINT "subcontrol_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_implementations_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_assets" table
ALTER TABLE "subcontrol_assets" ADD CONSTRAINT "subcontrol_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_assets_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_entities" table
ALTER TABLE "subcontrol_entities" ADD CONSTRAINT "subcontrol_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_entities_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_identity_holders" table
ALTER TABLE "subcontrol_identity_holders" ADD CONSTRAINT "subcontrol_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_identity_holders_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subscriber_events" table
ALTER TABLE "subscriber_events" ADD CONSTRAINT "subscriber_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subscriber_events_subscriber_id" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "system_detail_assets" table
ALTER TABLE "system_detail_assets" ADD CONSTRAINT "system_detail_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "system_detail_assets_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "task_evidence" table
ALTER TABLE "task_evidence" ADD CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "template_files" table
ALTER TABLE "template_files" ADD CONSTRAINT "template_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "template_files_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "user_events" table
ALTER TABLE "user_events" ADD CONSTRAINT "user_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_action_plans" table
ALTER TABLE "vulnerability_action_plans" ADD CONSTRAINT "vulnerability_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_action_plans_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_controls" table
ALTER TABLE "vulnerability_controls" ADD CONSTRAINT "vulnerability_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_controls_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_subcontrols" table
ALTER TABLE "vulnerability_subcontrols" ADD CONSTRAINT "vulnerability_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_subcontrols_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_risks" table
ALTER TABLE "vulnerability_risks" ADD CONSTRAINT "vulnerability_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_risks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_programs" table
ALTER TABLE "vulnerability_programs" ADD CONSTRAINT "vulnerability_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_programs_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_assets" table
ALTER TABLE "vulnerability_assets" ADD CONSTRAINT "vulnerability_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_assets_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_entities" table
ALTER TABLE "vulnerability_entities" ADD CONSTRAINT "vulnerability_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_entities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_scans" table
ALTER TABLE "vulnerability_scans" ADD CONSTRAINT "vulnerability_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_scans_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "vulnerability_tasks" table
ALTER TABLE "vulnerability_tasks" ADD CONSTRAINT "vulnerability_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_tasks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
