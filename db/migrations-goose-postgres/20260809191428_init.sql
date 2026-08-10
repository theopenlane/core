-- +goose Up
-- create "api_tokens" table
CREATE TABLE "api_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "sso_authorizations" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "api_token_owner_id_idx" to table: "api_tokens"
CREATE INDEX "api_token_owner_id_idx" ON "api_tokens" ("owner_id");
-- create index "api_tokens_token_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_token_key" ON "api_tokens" ("token");
-- create index "apitoken_token" to table: "api_tokens"
CREATE INDEX "apitoken_token" ON "api_tokens" ("token");
-- create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED', "details" text NULL, "details_json" jsonb NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "summary" character varying NULL, "tag_suggestions" jsonb NULL, "dismissed_tag_suggestions" jsonb NULL, "control_suggestions" jsonb NULL, "dismissed_control_suggestions" jsonb NULL, "improvement_suggestions" jsonb NULL, "dismissed_improvement_suggestions" jsonb NULL, "url" character varying NULL, "external_file_id" character varying NULL, "external_contents" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "action_plan_kind_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "title" character varying NOT NULL, "description" text NULL, "due_date" timestamptz NULL, "completed_at" timestamptz NULL, "priority" character varying NULL, "requires_approval" boolean NOT NULL DEFAULT false, "blocked" boolean NOT NULL DEFAULT false, "blocker_reason" text NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "source" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "action_plan_kind_id" character varying NULL, "file_id" character varying NULL, "custom_type_enum_action_plans" character varying NULL, "owner_id" character varying NULL, "subcontrol_action_plans" character varying NULL, "user_action_plans" character varying NULL, PRIMARY KEY ("id"));
-- create index "action_plan_file_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_file_id_idx" ON "action_plans" ("file_id");
-- create index "action_plan_owner_id_idx" to table: "action_plans"
CREATE INDEX "action_plan_owner_id_idx" ON "action_plans" ("owner_id");
-- create "assessments" table
CREATE TABLE "assessments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "response_due_duration" bigint NULL, "owner_id" character varying NULL, "template_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessment_name_owner_id" to table: "assessments"
CREATE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "assessment_owner_id_idx" to table: "assessments"
CREATE INDEX "assessment_owner_id_idx" ON "assessments" ("owner_id");
-- create index "assessment_template_id_idx" to table: "assessments"
CREATE INDEX "assessment_template_id_idx" ON "assessments" ("template_id");
-- create "assessment_responses" table
CREATE TABLE "assessment_responses" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "is_test" boolean NOT NULL DEFAULT false, "display_name" character varying NULL, "email" character varying NULL, "send_attempts" bigint NOT NULL DEFAULT 1, "email_delivered_at" timestamptz NULL, "email_opened_at" timestamptz NULL, "email_clicked_at" timestamptz NULL, "email_open_count" bigint NULL DEFAULT 0, "email_click_count" bigint NULL DEFAULT 0, "last_email_event_at" timestamptz NULL, "email_metadata" jsonb NULL, "status" character varying NOT NULL DEFAULT 'SENT', "assigned_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "is_draft" boolean NOT NULL DEFAULT false, "assessment_id" character varying NOT NULL, "document_data_id" character varying NULL, "campaign_id" character varying NULL, "entity_id" character varying NULL, "identity_holder_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessment_response_document_data_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_document_data_id_idx" ON "assessment_responses" ("document_data_id");
-- create index "assessment_response_owner_id_idx" to table: "assessment_responses"
CREATE INDEX "assessment_response_owner_id_idx" ON "assessment_responses" ("owner_id");
-- create index "assessmentresponse_assessment_id_email_is_test" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_email_is_test" ON "assessment_responses" ("assessment_id", "email", "is_test") WHERE ((deleted_at IS NULL) AND (campaign_id IS NULL));
-- create index "assessmentresponse_assigned_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_assigned_at" ON "assessment_responses" ("assigned_at");
-- create index "assessmentresponse_campaign_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_campaign_id" ON "assessment_responses" ("campaign_id");
-- create index "assessmentresponse_campaign_id_assessment_id_email_is_test" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_campaign_id_assessment_id_email_is_test" ON "assessment_responses" ("campaign_id", "assessment_id", "email", "is_test") WHERE ((deleted_at IS NULL) AND (campaign_id IS NOT NULL));
-- create index "assessmentresponse_completed_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_completed_at" ON "assessment_responses" ("completed_at");
-- create index "assessmentresponse_due_date" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_due_date" ON "assessment_responses" ("due_date");
-- create index "assessmentresponse_entity_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_entity_id" ON "assessment_responses" ("entity_id");
-- create index "assessmentresponse_identity_holder_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_identity_holder_id" ON "assessment_responses" ("identity_holder_id");
-- create index "assessmentresponse_status" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_status" ON "assessment_responses" ("status");
-- create "assets" table
CREATE TABLE "assets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "asset_subtype_name" character varying NULL, "asset_data_classification_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "access_model_name" character varying NULL, "encryption_status_name" character varying NULL, "security_tier_name" character varying NULL, "criticality_name" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "asset_type" character varying NOT NULL DEFAULT 'TECHNOLOGY', "name" character varying NOT NULL, "display_name" character varying NULL, "description" character varying NULL, "identifier" character varying NULL, "website" character varying NULL, "physical_location" character varying NULL, "region" character varying NULL, "contains_pii" boolean NULL DEFAULT false, "source_type" character varying NOT NULL DEFAULT 'MANUAL', "source_identifier" character varying NULL, "cost_center" character varying NULL, "estimated_monthly_cost" double precision NULL, "purchase_date" timestamptz NULL, "cpe" character varying NULL, "categories" jsonb NULL, "observed_at" timestamptz NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "asset_subtype_id" character varying NULL, "asset_data_classification_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "access_model_id" character varying NULL, "encryption_status_id" character varying NULL, "security_tier_id" character varying NULL, "criticality_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "source_platform_id" character varying NULL, "risk_assets" character varying NULL, PRIMARY KEY ("id"));
-- create index "asset_integration_id_idx" to table: "assets"
CREATE INDEX "asset_integration_id_idx" ON "assets" ("integration_id");
-- create index "asset_owner_id_idx" to table: "assets"
CREATE INDEX "asset_owner_id_idx" ON "assets" ("owner_id");
-- create index "asset_source_platform_id_idx" to table: "assets"
CREATE INDEX "asset_source_platform_id_idx" ON "assets" ("source_platform_id");
-- create "campaigns" table
CREATE TABLE "campaigns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "name" character varying NOT NULL, "description" character varying NULL, "campaign_type" character varying NOT NULL DEFAULT 'QUESTIONNAIRE', "status" character varying NOT NULL DEFAULT 'DRAFT', "is_active" boolean NOT NULL DEFAULT false, "scheduled_at" timestamptz NULL, "launched_at" timestamptz NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "is_recurring" boolean NOT NULL DEFAULT false, "recurrence_frequency" character varying NULL DEFAULT 'NONE', "recurrence_interval" bigint NULL DEFAULT 1, "recurrence_timezone" character varying NULL, "recurrence_cron" character varying NULL, "last_run_at" timestamptz NULL, "next_run_at" timestamptz NULL, "recurrence_end_at" timestamptz NULL, "recipient_count" bigint NULL DEFAULT 0, "resend_count" bigint NULL DEFAULT 0, "last_resent_at" timestamptz NULL, "metadata" jsonb NULL, "email_branding_id" character varying NULL, "assessment_id" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "email_template_id" character varying NULL, "entity_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "campaign_assessment_id_idx" to table: "campaigns"
CREATE INDEX "campaign_assessment_id_idx" ON "campaigns" ("assessment_id");
-- create index "campaign_display_id_owner_id" to table: "campaigns"
CREATE UNIQUE INDEX "campaign_display_id_owner_id" ON "campaigns" ("display_id", "owner_id");
-- create index "campaign_email_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_email_template_id_idx" ON "campaigns" ("email_template_id");
-- create index "campaign_entity_id" to table: "campaigns"
CREATE INDEX "campaign_entity_id" ON "campaigns" ("entity_id");
-- create index "campaign_integration_id_idx" to table: "campaigns"
CREATE INDEX "campaign_integration_id_idx" ON "campaigns" ("integration_id");
-- create index "campaign_name_owner_id" to table: "campaigns"
CREATE INDEX "campaign_name_owner_id" ON "campaigns" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "campaign_owner_id_idx" to table: "campaigns"
CREATE INDEX "campaign_owner_id_idx" ON "campaigns" ("owner_id");
-- create index "campaign_template_id_idx" to table: "campaigns"
CREATE INDEX "campaign_template_id_idx" ON "campaigns" ("template_id");
-- create index "campaign_trust_center_id_idx" to table: "campaigns"
CREATE INDEX "campaign_trust_center_id_idx" ON "campaigns" ("trust_center_id");
-- create "campaign_targets" table
CREATE TABLE "campaign_targets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "email" character varying NOT NULL, "full_name" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "sent_at" timestamptz NULL, "completed_at" timestamptz NULL, "metadata" jsonb NULL, "campaign_id" character varying NULL, "contact_id" character varying NULL, "group_id" character varying NULL, "owner_id" character varying NULL, "subscriber_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "campaign_target_owner_id_idx" to table: "campaign_targets"
CREATE INDEX "campaign_target_owner_id_idx" ON "campaign_targets" ("owner_id");
-- create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
CREATE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
-- create index "campaigntarget_contact_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_contact_id" ON "campaign_targets" ("contact_id");
-- create index "campaigntarget_group_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_group_id" ON "campaign_targets" ("group_id");
-- create index "campaigntarget_status" to table: "campaign_targets"
CREATE INDEX "campaigntarget_status" ON "campaign_targets" ("status");
-- create index "campaigntarget_subscriber_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_subscriber_id" ON "campaign_targets" ("subscriber_id");
-- create index "campaigntarget_user_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_user_id" ON "campaign_targets" ("user_id");
-- create "check_results" table
CREATE TABLE "check_results" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'UNKNOWN', "source" character varying NOT NULL, "last_observed_at" timestamptz NULL, "external_uri" character varying NULL, "details" text NULL, "parent_external_id" character varying NULL, "integration_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "check_result_integration_id_idx" to table: "check_results"
CREATE INDEX "check_result_integration_id_idx" ON "check_results" ("integration_id");
-- create "contacts" table
CREATE TABLE "contacts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "full_name" character varying NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "external_id" character varying NULL, "integration_id" character varying NULL, "observed_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "contact_owner_id_idx" to table: "contacts"
CREATE INDEX "contact_owner_id_idx" ON "contacts" ("owner_id");
-- create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "external_uuid" character varying NULL, "title" character varying NULL, "description" text NULL, "description_json" jsonb NULL, "aliases" jsonb NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NOT_IMPLEMENTED', "implementation_status" character varying NULL DEFAULT 'PLANNED', "implementation_description" text NULL, "public_representation" text NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "source_name" character varying NULL, "reference_framework" character varying NULL, "reference_framework_revision" character varying NULL, "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "testing_procedures" jsonb NULL, "evidence_requests" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "control_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "ref_code" character varying NOT NULL, "trust_center_visibility" character varying NULL DEFAULT 'NOT_VISIBLE', "is_trust_center_control" boolean NULL DEFAULT false, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "responsible_party_id" character varying NULL, "control_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "custom_type_enum_controls" character varying NULL, "owner_id" character varying NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "control_auditor_reference_id" to table: "controls"
CREATE INDEX "control_auditor_reference_id" ON "controls" ("auditor_reference_id") WHERE (deleted_at IS NULL);
-- create index "control_display_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_display_id_owner_id" ON "controls" ("display_id", "owner_id");
-- create index "control_external_uuid_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_external_uuid_owner_id" ON "controls" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "control_owner_id_idx" to table: "controls"
CREATE INDEX "control_owner_id_idx" ON "controls" ("owner_id");
-- create index "control_ref_code" to table: "controls"
CREATE INDEX "control_ref_code" ON "controls" ("ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND ((status)::text <> 'ARCHIVED'::text));
-- create index "control_ref_code_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_ref_code_owner_id" ON "controls" ("ref_code", "owner_id") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND (standard_id IS NULL));
-- create index "control_reference_id" to table: "controls"
CREATE INDEX "control_reference_id" ON "controls" ("reference_id") WHERE (deleted_at IS NULL);
-- create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NULL));
-- create index "control_standard_id_ref_code_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code_owner_id" ON "controls" ("standard_id", "ref_code", "owner_id") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND (standard_id IS NOT NULL));
-- create "control_implementations" table
CREATE TABLE "control_implementations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "status" character varying NULL DEFAULT 'DRAFT', "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, "details_json" jsonb NULL, "evidence_control_implementations" character varying NULL, "internal_policy_control_implementations" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "control_implementation_owner_id_idx" to table: "control_implementations"
CREATE INDEX "control_implementation_owner_id_idx" ON "control_implementations" ("owner_id");
-- create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "desired_outcome" text NULL, "desired_outcome_json" jsonb NULL, "status" character varying NULL DEFAULT 'DRAFT', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_objective_type" character varying NULL, "category" character varying NULL, "subcategory" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "control_objective_owner_id_idx" to table: "control_objectives"
CREATE INDEX "control_objective_owner_id_idx" ON "control_objectives" ("owner_id");
-- create index "controlobjective_display_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_display_id_owner_id" ON "control_objectives" ("display_id", "owner_id");
-- create "custom_domains" table
CREATE TABLE "custom_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "cname_record" character varying NOT NULL, "trust_center_id" character varying NULL, "domain_type" character varying NOT NULL DEFAULT 'UNKNOWN', "mappable_domain_id" character varying NOT NULL, "dns_verification_id" character varying NULL, "dns_verification_custom_domains" character varying NULL, "mappable_domain_custom_domains" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "custom_domain_dns_verification_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_dns_verification_id_idx" ON "custom_domains" ("dns_verification_id");
-- create index "custom_domain_mappable_domain_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_mappable_domain_id_idx" ON "custom_domains" ("mappable_domain_id");
-- create index "custom_domain_owner_id_idx" to table: "custom_domains"
CREATE INDEX "custom_domain_owner_id_idx" ON "custom_domains" ("owner_id");
-- create index "customdomain_cname_record" to table: "custom_domains"
CREATE UNIQUE INDEX "customdomain_cname_record" ON "custom_domains" ("cname_record") WHERE (deleted_at IS NULL);
-- create "custom_type_enums" table
CREATE TABLE "custom_type_enums" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "object_type" character varying NOT NULL, "field" character varying NOT NULL DEFAULT 'kind', "name" citext NOT NULL, "description" character varying NULL, "color" character varying NULL, "icon" character varying NULL, "entity_auth_methods" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "custom_type_enum_owner_id_idx" to table: "custom_type_enums"
CREATE INDEX "custom_type_enum_owner_id_idx" ON "custom_type_enums" ("owner_id");
-- create index "customtypeenum_name_field" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_name_field" ON "custom_type_enums" ("name", "field") WHERE (deleted_at IS NULL);
-- create index "customtypeenum_name_object_type_field_owner_id" to table: "custom_type_enums"
CREATE UNIQUE INDEX "customtypeenum_name_object_type_field_owner_id" ON "custom_type_enums" ("name", "object_type", "field", "owner_id") WHERE (deleted_at IS NULL);
-- create index "customtypeenum_object_type" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_object_type" ON "custom_type_enums" ("object_type") WHERE (deleted_at IS NULL);
-- create "dns_verifications" table
CREATE TABLE "dns_verifications" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "cloudflare_hostname_id" character varying NOT NULL, "dns_txt_record" character varying NOT NULL, "dns_txt_value" character varying NOT NULL, "dns_verification_status" character varying NOT NULL DEFAULT 'PENDING', "dns_verification_status_reason" character varying NULL, "acme_challenge_path" character varying NULL, "expected_acme_challenge_value" character varying NULL, "acme_challenge_status" character varying NOT NULL DEFAULT 'INITIALIZING', "acme_challenge_status_reason" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "dns_verification_owner_id_idx" to table: "dns_verifications"
CREATE INDEX "dns_verification_owner_id_idx" ON "dns_verifications" ("owner_id");
-- create index "dnsverification_cloudflare_hostname_id" to table: "dns_verifications"
CREATE UNIQUE INDEX "dnsverification_cloudflare_hostname_id" ON "dns_verifications" ("cloudflare_hostname_id") WHERE (deleted_at IS NULL);
-- create "directory_accounts" table
CREATE TABLE "directory_accounts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "directory_name" character varying NULL, "external_id" character varying NOT NULL, "secondary_key" character varying NULL, "canonical_email" character varying NULL, "email_aliases" jsonb NULL, "phone_number" character varying NULL, "display_name" character varying NULL, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "given_name" character varying NULL, "family_name" character varying NULL, "job_title" character varying NULL, "department" character varying NULL, "organization_unit" character varying NULL, "account_type" character varying NULL DEFAULT 'USER', "status" character varying NOT NULL DEFAULT 'ACTIVE', "mfa_state" character varying NOT NULL DEFAULT 'UNKNOWN', "last_seen_ip" character varying NULL, "last_login_at" timestamptz NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "added_at" timestamptz NULL, "removed_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "metadata" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, "primary_source" boolean NOT NULL DEFAULT false, "environment_id" character varying NULL, "scope_id" character varying NULL, "avatar_local_file_id" character varying NULL, "directory_sync_run_id" character varying NULL, "identity_holder_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "directory_account_avatar_local_file_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_avatar_local_file_id_idx" ON "directory_accounts" ("avatar_local_file_id");
-- create index "directory_account_owner_id_idx" to table: "directory_accounts"
CREATE INDEX "directory_account_owner_id_idx" ON "directory_accounts" ("owner_id");
-- create index "directoryaccount_directory_instance_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_instance_id_canonical_email" ON "directory_accounts" ("directory_instance_id", "canonical_email");
-- create index "directoryaccount_directory_instance_id_external_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_instance_id_external_id" ON "directory_accounts" ("directory_instance_id", "external_id");
-- create index "directoryaccount_directory_sync_run_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_sync_run_id_canonical_email" ON "directory_accounts" ("directory_sync_run_id", "canonical_email");
-- create index "directoryaccount_display_id_owner_id" to table: "directory_accounts"
CREATE UNIQUE INDEX "directoryaccount_display_id_owner_id" ON "directory_accounts" ("display_id", "owner_id");
-- create index "directoryaccount_identity_holder_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_identity_holder_id" ON "directory_accounts" ("identity_holder_id");
-- create index "directoryaccount_identity_holder_id_directory_name" to table: "directory_accounts"
CREATE INDEX "directoryaccount_identity_holder_id_directory_name" ON "directory_accounts" ("identity_holder_id", "directory_name");
-- create index "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" to table: "directory_accounts"
CREATE UNIQUE INDEX "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" ON "directory_accounts" ("integration_id", "external_id", "directory_sync_run_id");
-- create index "directoryaccount_integration_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_integration_id_canonical_email" ON "directory_accounts" ("integration_id", "canonical_email");
-- create index "directoryaccount_owner_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_owner_id_canonical_email" ON "directory_accounts" ("owner_id", "canonical_email");
-- create index "directoryaccount_platform_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_platform_id_canonical_email" ON "directory_accounts" ("platform_id", "canonical_email");
-- create index "directoryaccount_platform_id_external_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_platform_id_external_id" ON "directory_accounts" ("platform_id", "external_id");
-- create "directory_groups" table
CREATE TABLE "directory_groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "external_id" character varying NOT NULL, "email" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "classification" character varying NOT NULL DEFAULT 'TEAM', "status" character varying NOT NULL DEFAULT 'ACTIVE', "external_sharing_allowed" boolean NULL DEFAULT false, "member_count" bigint NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "added_at" timestamptz NULL, "removed_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "metadata" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, "directory_name" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "directory_sync_run_id" character varying NOT NULL, "integration_id" character varying NOT NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "directory_group_owner_id_idx" to table: "directory_groups"
CREATE INDEX "directory_group_owner_id_idx" ON "directory_groups" ("owner_id");
-- create index "directorygroup_directory_instance_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_instance_id_email" ON "directory_groups" ("directory_instance_id", "email");
-- create index "directorygroup_directory_instance_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_instance_id_external_id" ON "directory_groups" ("directory_instance_id", "external_id");
-- create index "directorygroup_directory_sync_run_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_sync_run_id_email" ON "directory_groups" ("directory_sync_run_id", "email");
-- create index "directorygroup_display_id_owner_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "owner_id");
-- create index "directorygroup_integration_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("integration_id", "email");
-- create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("integration_id", "external_id", "directory_sync_run_id");
-- create index "directorygroup_owner_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("owner_id", "email");
-- create index "directorygroup_platform_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_email" ON "directory_groups" ("platform_id", "email");
-- create index "directorygroup_platform_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_external_id" ON "directory_groups" ("platform_id", "external_id");
-- create "directory_memberships" table
CREATE TABLE "directory_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "role" character varying NULL DEFAULT 'MEMBER', "source" character varying NULL, "directory_name" character varying NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "added_at" timestamptz NULL, "removed_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "last_confirmed_run_id" character varying NULL, "metadata" jsonb NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "directory_account_id" character varying NOT NULL, "directory_group_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "integration_id" character varying NOT NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "directory_membership_directory_group_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_directory_group_id_idx" ON "directory_memberships" ("directory_group_id");
-- create index "directory_membership_owner_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_owner_id_idx" ON "directory_memberships" ("owner_id");
-- create index "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" to table: "directory_memberships"
CREATE INDEX "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" ON "directory_memberships" ("directory_instance_id", "directory_account_id", "directory_group_id");
-- create index "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" ON "directory_memberships" ("directory_account_id", "directory_group_id", "directory_sync_run_id");
-- create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory_account_id_directory_group_id" ON "directory_memberships" ("directory_account_id", "directory_group_id") WHERE (removed_at IS NULL);
-- create index "directorymembership_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_directory_sync_run_id" ON "directory_memberships" ("directory_sync_run_id");
-- create index "directorymembership_display_id_owner_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_display_id_owner_id" ON "directory_memberships" ("display_id", "owner_id");
-- create index "directorymembership_integration_id_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_integration_id_directory_sync_run_id" ON "directory_memberships" ("integration_id", "directory_sync_run_id");
-- create index "directorymembership_platform_id_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_platform_id_directory_sync_run_id" ON "directory_memberships" ("platform_id", "directory_sync_run_id");
-- create "directory_sync_runs" table
CREATE TABLE "directory_sync_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "directory_instance_id" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "source_cursor" character varying NULL, "full_count" bigint NOT NULL DEFAULT 0, "delta_count" bigint NOT NULL DEFAULT 0, "error" text NULL, "raw_manifest_file_id" character varying NULL, "stats" jsonb NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "integration_id" character varying NOT NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "directory_sync_run_owner_id_idx" to table: "directory_sync_runs"
CREATE INDEX "directory_sync_run_owner_id_idx" ON "directory_sync_runs" ("owner_id");
-- create index "directorysyncrun_directory_instance_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_directory_instance_id_started_at" ON "directory_sync_runs" ("directory_instance_id", "started_at");
-- create index "directorysyncrun_display_id_owner_id" to table: "directory_sync_runs"
CREATE UNIQUE INDEX "directorysyncrun_display_id_owner_id" ON "directory_sync_runs" ("display_id", "owner_id");
-- create index "directorysyncrun_integration_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_integration_id_started_at" ON "directory_sync_runs" ("integration_id", "started_at");
-- create index "directorysyncrun_platform_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_platform_id_started_at" ON "directory_sync_runs" ("platform_id", "started_at");
-- create "discussions" table
CREATE TABLE "discussions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "external_id" character varying NULL, "is_resolved" boolean NOT NULL DEFAULT false, "control_discussions" character varying NULL, "internal_policy_discussions" character varying NULL, "owner_id" character varying NULL, "procedure_discussions" character varying NULL, "risk_discussions" character varying NULL, "subcontrol_discussions" character varying NULL, PRIMARY KEY ("id"));
-- create index "discussion_owner_id_idx" to table: "discussions"
CREATE INDEX "discussion_owner_id_idx" ON "discussions" ("owner_id");
-- create index "discussions_external_id_key" to table: "discussions"
CREATE UNIQUE INDEX "discussions_external_id_key" ON "discussions" ("external_id");
-- create "document_data" table
CREATE TABLE "document_data" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "data" jsonb NOT NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "document_owner_id_idx" to table: "document_data"
CREATE INDEX "document_owner_id_idx" ON "document_data" ("owner_id");
-- create index "document_template_id_idx" to table: "document_data"
CREATE INDEX "document_template_id_idx" ON "document_data" ("template_id");
-- create "email_templates" table
CREATE TABLE "email_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "format" character varying NULL DEFAULT 'HTML', "locale" character varying NOT NULL DEFAULT 'en-US', "subject_template" character varying NULL, "preheader_template" character varying NULL, "body_template" text NULL, "text_template" text NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "template_context" character varying NULL, "defaults" jsonb NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "trust_center_id" character varying NULL, "workflow_definition_id" character varying NULL, "workflow_instance_id" character varying NULL, PRIMARY KEY ("id"));
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
-- create index "emailtemplate_owner_id_key" to table: "email_templates"
CREATE INDEX "emailtemplate_owner_id_key" ON "email_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "email_verification_tokens_owner_id_fk" to table: "email_verification_tokens"
CREATE INDEX "email_verification_tokens_owner_id_fk" ON "email_verification_tokens" ("owner_id");
-- create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_token_key" ON "email_verification_tokens" ("token");
-- create index "emailverificationtoken_token" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_token" ON "email_verification_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "entities" table
CREATE TABLE "entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "reviewed_by" character varying NULL, "last_reviewed_at" timestamptz NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "entity_relationship_state_name" character varying NULL, "entity_security_questionnaire_status_name" character varying NULL, "entity_source_type_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "name" citext NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "aliases" jsonb NULL, "status" character varying NULL DEFAULT 'ACTIVE', "approved_for_use" boolean NULL DEFAULT false, "linked_asset_ids" jsonb NULL, "has_soc2" boolean NULL DEFAULT false, "soc2_period_end" timestamptz NULL, "contract_start_date" timestamptz NULL, "contract_end_date" timestamptz NULL, "auto_renews" boolean NULL DEFAULT false, "termination_notice_days" bigint NULL, "annual_spend" double precision NULL, "spend_currency" character varying NULL DEFAULT 'USD', "billing_model" character varying NULL, "renewal_risk" character varying NULL, "sso_enforced" boolean NULL DEFAULT false, "mfa_supported" boolean NULL DEFAULT false, "mfa_enforced" boolean NULL DEFAULT false, "status_page_url" character varying NULL, "provided_services" jsonb NULL, "links" jsonb NULL, "risk_rating" character varying NULL, "risk_score" bigint NULL, "risk_score_coverage" bigint NULL, "tier" character varying NULL DEFAULT 'LOW', "review_frequency" character varying NULL DEFAULT 'YEARLY', "next_review_at" timestamptz NULL, "contract_renewal_at" timestamptz NULL, "vendor_metadata" jsonb NULL, "logo_remote_url" character varying NULL, "external_id" character varying NULL, "observed_at" timestamptz NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "entity_relationship_state_id" character varying NULL, "entity_security_questionnaire_status_id" character varying NULL, "entity_source_type_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "entity_type_id" character varying NULL, "logo_file_id" character varying NULL, "entity_type_entities" character varying NULL, "owner_id" character varying NULL, "risk_entities" character varying NULL, PRIMARY KEY ("id"));
-- create index "entity_entity_type_id_idx" to table: "entities"
CREATE INDEX "entity_entity_type_id_idx" ON "entities" ("entity_type_id");
-- create index "entity_logo_file_id_idx" to table: "entities"
CREATE INDEX "entity_logo_file_id_idx" ON "entities" ("logo_file_id");
-- create index "entity_name_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "entity_owner_id_idx" to table: "entities"
CREATE INDEX "entity_owner_id_idx" ON "entities" ("owner_id");
-- create index "entity_reviewed_by_user_id" to table: "entities"
CREATE INDEX "entity_reviewed_by_user_id" ON "entities" ("reviewed_by_user_id");
-- create "entity_types" table
CREATE TABLE "entity_types" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" citext NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "entity_type_owner_id_idx" to table: "entity_types"
CREATE INDEX "entity_type_owner_id_idx" ON "entity_types" ("owner_id");
-- create index "entitytype_name_owner_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_name_owner_id" ON "entity_types" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "events" table
CREATE TABLE "events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, "directory_membership_events" character varying NULL, "export_events" character varying NULL, PRIMARY KEY ("id"));
-- create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "status" character varying NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "environment_id" character varying NULL, "scope_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- create index "evidence_external_uuid_owner_id" to table: "evidences"
CREATE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "evidence_owner_id_idx" to table: "evidences"
CREATE INDEX "evidence_owner_id_idx" ON "evidences" ("owner_id");
-- create "exports" table
CREATE TABLE "exports" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "requestor_id" character varying NULL, "export_type" character varying NOT NULL, "format" character varying NOT NULL DEFAULT 'CSV', "status" character varying NOT NULL DEFAULT 'PENDING', "fields" jsonb NULL, "filters" character varying NULL, "error_message" character varying NULL, "mode" character varying NOT NULL DEFAULT 'FLAT', "export_metadata" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "export_owner_id_idx" to table: "exports"
CREATE INDEX "export_owner_id_idx" ON "exports" ("owner_id");
-- create "files" table
CREATE TABLE "files" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "category_name" character varying NULL, "name" character varying NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, "metadata" jsonb NULL, "storage_region" character varying NULL, "storage_provider" character varying NULL, "last_accessed_at" timestamptz NULL, "email_template_files" character varying NULL, "export_files" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "category_id" character varying NULL, "finding_files" character varying NULL, "integration_files" character varying NULL, "note_files" character varying NULL, "platform_architecture_diagrams" character varying NULL, "platform_data_flow_diagrams" character varying NULL, "platform_trust_boundary_diagrams" character varying NULL, "remediation_files" character varying NULL, "review_files" character varying NULL, "vulnerability_files" character varying NULL, PRIMARY KEY ("id"));
-- create "file_download_tokens" table
CREATE TABLE "file_download_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NULL, "ttl" timestamptz NULL, "user_id" character varying NULL, "organization_id" character varying NULL, "file_id" character varying NULL, "secret" bytea NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "file_download_tokens_owner_id_fk" to table: "file_download_tokens"
CREATE INDEX "file_download_tokens_owner_id_fk" ON "file_download_tokens" ("owner_id");
-- create index "file_download_tokens_token_key" to table: "file_download_tokens"
CREATE UNIQUE INDEX "file_download_tokens_token_key" ON "file_download_tokens" ("token");
-- create index "filedownloadtoken_token" to table: "file_download_tokens"
CREATE UNIQUE INDEX "filedownloadtoken_token" ON "file_download_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "findings" table
CREATE TABLE "findings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "reviewed_by" character varying NULL, "assigned_to" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "finding_status_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_id" character varying NULL, "security_level" character varying NULL DEFAULT 'NONE', "external_owner_id" character varying NULL, "source" character varying NULL, "resource_name" character varying NULL, "display_name" character varying NULL, "state" character varying NULL, "category" character varying NULL, "categories" jsonb NULL, "finding_class" character varying NULL, "severity" character varying NULL, "numeric_severity" double precision NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "open" boolean NULL DEFAULT true, "blocks_production" boolean NULL, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "assessment_id" character varying NULL, "description" text NULL, "recommendation" text NULL, "recommended_actions" text NULL, "references" jsonb NULL, "steps_to_reproduce" jsonb NULL, "targets" jsonb NULL, "target_details" jsonb NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "event_time" timestamptz NULL, "reported_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "assigned_to_user_id" character varying NULL, "assigned_to_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "finding_status_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "finding_display_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_display_id_owner_id" ON "findings" ("display_id", "owner_id");
-- create index "finding_external_id_external_owner_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_external_id_external_owner_id_owner_id" ON "findings" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "finding_owner_id_idx" to table: "findings"
CREATE INDEX "finding_owner_id_idx" ON "findings" ("owner_id");
-- create "finding_controls" table
CREATE TABLE "finding_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "external_standard" character varying NULL, "external_standard_version" character varying NULL, "external_control_id" character varying NULL, "source" character varying NULL, "metadata" jsonb NULL, "discovered_at" timestamptz NULL, "finding_id" character varying NOT NULL, "control_id" character varying NOT NULL, "standard_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "finding_control_control_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_control_id_idx" ON "finding_controls" ("control_id");
-- create index "finding_control_owner_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_owner_id_idx" ON "finding_controls" ("owner_id");
-- create index "finding_control_standard_id_idx" to table: "finding_controls"
CREATE INDEX "finding_control_standard_id_idx" ON "finding_controls" ("standard_id");
-- create index "findingcontrol_finding_id_control_id" to table: "finding_controls"
CREATE UNIQUE INDEX "findingcontrol_finding_id_control_id" ON "finding_controls" ("finding_id", "control_id");
-- create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" citext NOT NULL, "description" character varying NULL, "is_managed" boolean NULL DEFAULT false, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', "oscal_role" character varying NULL, "oscal_party_uuid" character varying NULL, "oscal_contact_uuids" jsonb NULL, "scim_external_id" character varying NULL, "scim_display_name" character varying NULL, "scim_active" boolean NULL DEFAULT true, "scim_group_mailing" character varying NULL, "assessment_blocked_groups" character varying NULL, "assessment_editors" character varying NULL, "assessment_viewers" character varying NULL, "asset_blocked_groups" character varying NULL, "asset_editors" character varying NULL, "asset_viewers" character varying NULL, "check_result_blocked_groups" character varying NULL, "check_result_editors" character varying NULL, "check_result_viewers" character varying NULL, "email_template_blocked_groups" character varying NULL, "email_template_editors" character varying NULL, "email_template_viewers" character varying NULL, "avatar_local_file_id" character varying NULL, "identity_holder_blocked_groups" character varying NULL, "identity_holder_editors" character varying NULL, "identity_holder_viewers" character varying NULL, "organization_action_plan_creators" character varying NULL, "organization_api_token_creators" character varying NULL, "organization_assessment_creators" character varying NULL, "organization_asset_creators" character varying NULL, "organization_campaign_creators" character varying NULL, "organization_campaign_target_creators" character varying NULL, "organization_check_result_creators" character varying NULL, "organization_contact_creators" character varying NULL, "organization_control_creators" character varying NULL, "organization_control_implementation_creators" character varying NULL, "organization_control_objective_creators" character varying NULL, "organization_custom_domain_creators" character varying NULL, "organization_custom_type_enum_creators" character varying NULL, "organization_directory_account_creators" character varying NULL, "organization_directory_group_creators" character varying NULL, "organization_directory_membership_creators" character varying NULL, "organization_directory_sync_run_creators" character varying NULL, "organization_discussion_creators" character varying NULL, "organization_document_data_creators" character varying NULL, "organization_email_template_creators" character varying NULL, "organization_entity_creators" character varying NULL, "organization_entity_type_creators" character varying NULL, "organization_evidence_creators" character varying NULL, "organization_file_creators" character varying NULL, "organization_finding_creators" character varying NULL, "organization_finding_control_creators" character varying NULL, "organization_group_creators" character varying NULL, "organization_group_membership_creators" character varying NULL, "organization_group_setting_creators" character varying NULL, "organization_hush_creators" character varying NULL, "organization_identity_holder_creators" character varying NULL, "organization_internal_policy_creators" character varying NULL, "organization_invite_creators" character varying NULL, "organization_job_runner_creators" character varying NULL, "organization_job_runner_registration_token_creators" character varying NULL, "organization_job_runner_token_creators" character varying NULL, "organization_job_template_creators" character varying NULL, "organization_mapped_control_creators" character varying NULL, "organization_narrative_creators" character varying NULL, "organization_note_creators" character varying NULL, "organization_notification_template_creators" character varying NULL, "organization_org_membership_creators" character varying NULL, "organization_platform_creators" character varying NULL, "organization_procedure_creators" character varying NULL, "organization_program_creators" character varying NULL, "organization_program_membership_creators" character varying NULL, "organization_remediation_creators" character varying NULL, "organization_review_creators" character varying NULL, "organization_risk_creators" character varying NULL, "organization_scan_creators" character varying NULL, "organization_scheduled_job_creators" character varying NULL, "organization_scheduled_job_run_creators" character varying NULL, "organization_sla_definition_creators" character varying NULL, "organization_standard_creators" character varying NULL, "organization_subcontrol_creators" character varying NULL, "organization_subprocessor_creators" character varying NULL, "organization_subscriber_creators" character varying NULL, "organization_system_detail_creators" character varying NULL, "organization_tag_definition_creators" character varying NULL, "organization_task_creators" character varying NULL, "organization_template_creators" character varying NULL, "organization_trust_center_creators" character varying NULL, "organization_trust_center_compliance_creators" character varying NULL, "organization_trust_center_doc_creators" character varying NULL, "organization_trust_center_entity_creators" character varying NULL, "organization_trust_center_faq_creators" character varying NULL, "organization_trust_center_nda_request_creators" character varying NULL, "organization_trust_center_subprocessor_creators" character varying NULL, "organization_trust_center_watermark_config_creators" character varying NULL, "organization_vendor_risk_score_creators" character varying NULL, "organization_vendor_scoring_config_creators" character varying NULL, "organization_vulnerability_creators" character varying NULL, "organization_workflow_definition_creators" character varying NULL, "organization_campaigns_manager" character varying NULL, "organization_compliance_manager" character varying NULL, "organization_group_manager" character varying NULL, "organization_policies_manager" character varying NULL, "organization_registry_manager" character varying NULL, "organization_risk_manager" character varying NULL, "organization_trust_center_manager" character varying NULL, "organization_workflows_manager" character varying NULL, "owner_id" character varying NULL, "sla_definition_blocked_groups" character varying NULL, "sla_definition_editors" character varying NULL, "trust_center_blocked_groups" character varying NULL, "trust_center_editors" character varying NULL, "trust_center_compliance_blocked_groups" character varying NULL, "trust_center_compliance_editors" character varying NULL, "trust_center_doc_blocked_groups" character varying NULL, "trust_center_doc_editors" character varying NULL, "trust_center_entity_blocked_groups" character varying NULL, "trust_center_entity_editors" character varying NULL, "trust_center_faq_blocked_groups" character varying NULL, "trust_center_faq_editors" character varying NULL, "trust_center_nda_request_blocked_groups" character varying NULL, "trust_center_nda_request_editors" character varying NULL, "trust_center_setting_blocked_groups" character varying NULL, "trust_center_setting_editors" character varying NULL, "trust_center_subprocessor_blocked_groups" character varying NULL, "trust_center_subprocessor_editors" character varying NULL, "trust_center_watermark_config_blocked_groups" character varying NULL, "trust_center_watermark_config_editors" character varying NULL, "vulnerability_blocked_groups" character varying NULL, "vulnerability_editors" character varying NULL, "vulnerability_viewers" character varying NULL, "workflow_definition_blocked_groups" character varying NULL, "workflow_definition_editors" character varying NULL, "workflow_definition_viewers" character varying NULL, "workflow_definition_groups" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_avatar_local_file_id_idx" to table: "groups"
CREATE INDEX "group_avatar_local_file_id_idx" ON "groups" ("avatar_local_file_id");
-- create index "group_display_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_display_id_owner_id" ON "groups" ("display_id", "owner_id");
-- create index "group_name_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_name_owner_id" ON "groups" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "group_owner_id_idx" to table: "groups"
CREATE INDEX "group_owner_id_idx" ON "groups" ("owner_id");
-- create "group_memberships" table
CREATE TABLE "group_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, "group_membership_org_membership" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_membership_group_id_idx" to table: "group_memberships"
CREATE INDEX "group_membership_group_id_idx" ON "group_memberships" ("group_id");
-- create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id");
-- create "group_settings" table
CREATE TABLE "group_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_setting_group_id_idx" to table: "group_settings"
CREATE INDEX "group_setting_group_id_idx" ON "group_settings" ("group_id");
-- create index "group_settings_group_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_group_id_key" ON "group_settings" ("group_id");
-- create "hushes" table
CREATE TABLE "hushes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, "credential_set" jsonb NULL, "metadata" jsonb NULL, "last_used_at" timestamptz NULL, "expires_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "secret_owner_id_idx" to table: "hushes"
CREATE INDEX "secret_owner_id_idx" ON "hushes" ("owner_id");
-- create "identity_holders" table
CREATE TABLE "identity_holders" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "full_name" character varying NOT NULL, "email" character varying NOT NULL, "alternate_email" character varying NULL, "email_aliases" jsonb NULL, "phone_number" character varying NULL, "is_openlane_user" boolean NULL DEFAULT false, "identity_holder_type" character varying NOT NULL DEFAULT 'UNSPECIFIED', "status" character varying NOT NULL DEFAULT 'ACTIVE', "is_active" boolean NOT NULL DEFAULT true, "title" character varying NULL, "department" character varying NULL, "team" character varying NULL, "location" character varying NULL, "start_date" timestamptz NULL, "end_date" timestamptz NULL, "external_user_id" character varying NULL, "external_reference_id" character varying NULL, "metadata" jsonb NULL, "avatar_remote_url" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "employer_entity_id" character varying NULL, "owner_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "identity_holder_employer_entity_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_employer_entity_id_idx" ON "identity_holders" ("employer_entity_id");
-- create index "identity_holder_owner_id_idx" to table: "identity_holders"
CREATE INDEX "identity_holder_owner_id_idx" ON "identity_holders" ("owner_id");
-- create index "identityholder_display_id_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_display_id_owner_id" ON "identity_holders" ("display_id", "owner_id");
-- create index "identityholder_email_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_email_owner_id" ON "identity_holders" ("email", "owner_id") WHERE (deleted_at IS NULL);
-- create index "identityholder_external_user_id" to table: "identity_holders"
CREATE INDEX "identityholder_external_user_id" ON "identity_holders" ("external_user_id");
-- create index "identityholder_user_id" to table: "identity_holders"
CREATE INDEX "identityholder_user_id" ON "identity_holders" ("user_id");
-- create "impersonation_events" table
CREATE TABLE "impersonation_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "impersonation_type" character varying NOT NULL, "action" character varying NOT NULL, "reason" character varying NULL, "ip_address" character varying NULL, "user_agent" character varying NULL, "scopes" jsonb NULL, "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, "target_user_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "impersonation_event_organization_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_organization_id_idx" ON "impersonation_events" ("organization_id");
-- create index "impersonation_event_target_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_target_user_id_idx" ON "impersonation_events" ("target_user_id");
-- create index "impersonation_event_user_id_idx" to table: "impersonation_events"
CREATE INDEX "impersonation_event_user_id_idx" ON "impersonation_events" ("user_id");
-- create "integrations" table
CREATE TABLE "integrations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "integration_type" character varying NULL, "provider_metadata" jsonb NULL, "config" jsonb NULL, "installation_metadata" jsonb NULL, "provider_state" jsonb NULL, "metadata" jsonb NULL, "definition_id" character varying NULL, "definition_version" character varying NULL, "definition_slug" character varying NULL, "family" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "provider_metadata_snapshot" jsonb NULL, "primary_directory" boolean NOT NULL DEFAULT false, "campaign_email" boolean NOT NULL DEFAULT false, "file_integrations" character varying NULL, "group_integrations" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "integration_owner_id_idx" to table: "integrations"
CREATE INDEX "integration_owner_id_idx" ON "integrations" ("owner_id");
-- create index "integration_platform_id_idx" to table: "integrations"
CREATE INDEX "integration_platform_id_idx" ON "integrations" ("platform_id");
-- create "integration_runs" table
CREATE TABLE "integration_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "operation_name" character varying NULL, "operation_kind" character varying NULL, "run_type" character varying NULL, "operation_config" jsonb NULL, "mapping_version" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "started_at" timestamptz NOT NULL, "finished_at" timestamptz NULL, "duration_ms" bigint NULL, "summary" character varying NULL, "error" text NULL, "metrics" jsonb NULL, "integration_id" character varying NULL, "request_file_id" character varying NULL, "response_file_id" character varying NULL, "event_id" character varying NULL, "assessment_response_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "integration_run_event_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_event_id_idx" ON "integration_runs" ("event_id");
-- create index "integration_run_owner_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_owner_id_idx" ON "integration_runs" ("owner_id");
-- create index "integration_run_request_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_request_file_id_idx" ON "integration_runs" ("request_file_id");
-- create index "integration_run_response_file_id_idx" to table: "integration_runs"
CREATE INDEX "integration_run_response_file_id_idx" ON "integration_runs" ("response_file_id");
-- create index "integrationrun_assessment_response_id_operation_name" to table: "integration_runs"
CREATE UNIQUE INDEX "integrationrun_assessment_response_id_operation_name" ON "integration_runs" ("assessment_response_id", "operation_name") WHERE ((deleted_at IS NULL) AND (assessment_response_id IS NOT NULL));
-- create index "integrationrun_assessment_response_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_assessment_response_id_started_at" ON "integration_runs" ("assessment_response_id", "started_at") WHERE (deleted_at IS NULL);
-- create index "integrationrun_integration_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_integration_id_started_at" ON "integration_runs" ("integration_id", "started_at") WHERE (deleted_at IS NULL);
-- create "integration_webhooks" table
CREATE TABLE "integration_webhooks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "provider" character varying NOT NULL, "name" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "endpoint_id" character varying NULL, "endpoint_url" character varying NULL, "secret_token" character varying NULL, "allowed_events" jsonb NULL, "last_delivery_id" character varying NULL, "last_delivery_at" timestamptz NULL, "last_delivery_status" character varying NULL, "last_delivery_error" text NULL, "external_event_id" character varying NULL, "metadata" jsonb NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "integration_webhook_owner_id_idx" to table: "integration_webhooks"
CREATE INDEX "integration_webhook_owner_id_idx" ON "integration_webhooks" ("owner_id");
-- create index "integrationwebhook_endpoint_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_endpoint_id" ON "integration_webhooks" ("endpoint_id") WHERE ((deleted_at IS NULL) AND (endpoint_id IS NOT NULL));
-- create index "integrationwebhook_integration_id_name_external_event_id" to table: "integration_webhooks"
CREATE UNIQUE INDEX "integrationwebhook_integration_id_name_external_event_id" ON "integration_webhooks" ("integration_id", "name", "external_event_id") WHERE ((deleted_at IS NULL) AND (external_event_id IS NOT NULL));
-- create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED', "details" text NULL, "details_json" jsonb NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "summary" character varying NULL, "tag_suggestions" jsonb NULL, "dismissed_tag_suggestions" jsonb NULL, "control_suggestions" jsonb NULL, "dismissed_control_suggestions" jsonb NULL, "improvement_suggestions" jsonb NULL, "dismissed_improvement_suggestions" jsonb NULL, "url" character varying NULL, "external_file_id" character varying NULL, "external_contents" character varying NULL, "internal_policy_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "custom_type_enum_internal_policies" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "internal_policy_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "file_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "internal_policy_file_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_file_id_idx" ON "internal_policies" ("file_id");
-- create index "internal_policy_owner_id_idx" to table: "internal_policies"
CREATE INDEX "internal_policy_owner_id_idx" ON "internal_policies" ("owner_id");
-- create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_display_id_owner_id" ON "internal_policies" ("display_id", "owner_id");
-- create index "internalpolicy_external_uuid_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_external_uuid_owner_id" ON "internal_policies" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create "invites" table
CREATE TABLE "invites" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "requestor_id" character varying NULL, "token" character varying NOT NULL, "expires" timestamptz NULL, "recipient" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'INVITATION_SENT', "role" character varying NOT NULL DEFAULT 'MEMBER', "send_attempts" bigint NOT NULL DEFAULT 1, "secret" bytea NOT NULL, "ownership_transfer" boolean NULL DEFAULT false, "sso_exempt" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "invite_owner_id_idx" to table: "invites"
CREATE INDEX "invite_owner_id_idx" ON "invites" ("owner_id");
-- create index "invite_recipient_owner_id" to table: "invites"
CREATE UNIQUE INDEX "invite_recipient_owner_id" ON "invites" ("recipient", "owner_id") WHERE (deleted_at IS NULL);
-- create index "invites_token_key" to table: "invites"
CREATE UNIQUE INDEX "invites_token_key" ON "invites" ("token");
-- create "job_results" table
CREATE TABLE "job_results" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "status" character varying NOT NULL, "exit_code" bigint NOT NULL, "finished_at" timestamptz NOT NULL, "started_at" timestamptz NOT NULL, "log" text NULL, "scheduled_job_id" character varying NOT NULL, "file_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "job_result_file_id_idx" to table: "job_results"
CREATE INDEX "job_result_file_id_idx" ON "job_results" ("file_id");
-- create index "job_result_owner_id_idx" to table: "job_results"
CREATE INDEX "job_result_owner_id_idx" ON "job_results" ("owner_id");
-- create index "job_result_scheduled_job_id_idx" to table: "job_results"
CREATE INDEX "job_result_scheduled_job_id_idx" ON "job_results" ("scheduled_job_id");
-- create "job_runners" table
CREATE TABLE "job_runners" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'OFFLINE', "ip_address" character varying NULL, "last_seen" timestamptz NULL, "version" character varying NULL, "os" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "job_runner_owner_id_idx" to table: "job_runners"
CREATE INDEX "job_runner_owner_id_idx" ON "job_runners" ("owner_id");
-- create index "jobrunner_display_id_owner_id" to table: "job_runners"
CREATE UNIQUE INDEX "jobrunner_display_id_owner_id" ON "job_runners" ("display_id", "owner_id");
-- create "job_runner_registration_tokens" table
CREATE TABLE "job_runner_registration_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "token" character varying NOT NULL, "expires_at" timestamptz NOT NULL, "last_used_at" timestamptz NULL, "job_runner_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "job_runner_registration_token_job_runner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_job_runner_id_idx" ON "job_runner_registration_tokens" ("job_runner_id");
-- create index "job_runner_registration_token_owner_id_idx" to table: "job_runner_registration_tokens"
CREATE INDEX "job_runner_registration_token_owner_id_idx" ON "job_runner_registration_tokens" ("owner_id");
-- create index "job_runner_registration_tokens_token_key" to table: "job_runner_registration_tokens"
CREATE UNIQUE INDEX "job_runner_registration_tokens_token_key" ON "job_runner_registration_tokens" ("token");
-- create "job_runner_tokens" table
CREATE TABLE "job_runner_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "job_runner_token_owner_id_idx" to table: "job_runner_tokens"
CREATE INDEX "job_runner_token_owner_id_idx" ON "job_runner_tokens" ("owner_id");
-- create index "job_runner_tokens_token_key" to table: "job_runner_tokens"
CREATE UNIQUE INDEX "job_runner_tokens_token_key" ON "job_runner_tokens" ("token");
-- create index "jobrunnertoken_token_expires_at_is_active" to table: "job_runner_tokens"
CREATE INDEX "jobrunnertoken_token_expires_at_is_active" ON "job_runner_tokens" ("token", "expires_at", "is_active");
-- create "job_templates" table
CREATE TABLE "job_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "title" character varying NOT NULL, "description" character varying NULL, "platform" character varying NOT NULL, "windmill_path" character varying NULL, "download_url" character varying NOT NULL, "configuration" jsonb NULL, "cron" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "job_template_owner_id_idx" to table: "job_templates"
CREATE INDEX "job_template_owner_id_idx" ON "job_templates" ("owner_id");
-- create index "jobtemplate_display_id_owner_id" to table: "job_templates"
CREATE UNIQUE INDEX "jobtemplate_display_id_owner_id" ON "job_templates" ("display_id", "owner_id");
-- create "mappable_domains" table
CREATE TABLE "mappable_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "zone_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "mappabledomain_name" to table: "mappable_domains"
CREATE UNIQUE INDEX "mappabledomain_name" ON "mappable_domains" ("name") WHERE (deleted_at IS NULL);
-- create "mapped_controls" table
CREATE TABLE "mapped_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "mapping_type" character varying NOT NULL DEFAULT 'EQUAL', "relation" character varying NULL, "confidence" bigint NULL, "source" character varying NULL DEFAULT 'MANUAL', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "mapped_control_owner_id_idx" to table: "mapped_controls"
CREATE INDEX "mapped_control_owner_id_idx" ON "mapped_controls" ("owner_id");
-- create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "details" text NULL, "control_objective_narratives" character varying NULL, "owner_id" character varying NULL, "subcontrol_narratives" character varying NULL, PRIMARY KEY ("id"));
-- create index "narrative_display_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_display_id_owner_id" ON "narratives" ("display_id", "owner_id");
-- create index "narrative_owner_id_idx" to table: "narratives"
CREATE INDEX "narrative_owner_id_idx" ON "narratives" ("owner_id");
-- create "notes" table
CREATE TABLE "notes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "title" character varying NULL, "text" text NOT NULL, "text_json" jsonb NULL, "note_ref" character varying NULL, "is_edited" boolean NOT NULL DEFAULT false, "notify_subscribers" boolean NULL DEFAULT false, "notified_at" timestamptz NULL, "control_comments" character varying NULL, "discussion_id" character varying NULL, "entity_notes" character varying NULL, "evidence_comments" character varying NULL, "finding_comments" character varying NULL, "internal_policy_comments" character varying NULL, "owner_id" character varying NULL, "procedure_comments" character varying NULL, "program_notes" character varying NULL, "remediation_comments" character varying NULL, "review_comments" character varying NULL, "risk_comments" character varying NULL, "subcontrol_comments" character varying NULL, "task_comments" character varying NULL, "trust_center_id" character varying NULL, "vulnerability_comments" character varying NULL, PRIMARY KEY ("id"));
-- create index "note_discussion_id_idx" to table: "notes"
CREATE INDEX "note_discussion_id_idx" ON "notes" ("discussion_id");
-- create index "note_display_id_owner_id" to table: "notes"
CREATE UNIQUE INDEX "note_display_id_owner_id" ON "notes" ("display_id", "owner_id");
-- create index "note_owner_id_idx" to table: "notes"
CREATE INDEX "note_owner_id_idx" ON "notes" ("owner_id");
-- create index "note_trust_center_id_idx" to table: "notes"
CREATE INDEX "note_trust_center_id_idx" ON "notes" ("trust_center_id");
-- create "notifications" table
CREATE TABLE "notifications" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "tags" jsonb NULL, "user_id" character varying NULL, "notification_type" character varying NOT NULL, "object_type" character varying NOT NULL, "title" character varying NOT NULL, "body" text NOT NULL, "data" jsonb NULL, "read_at" timestamptz NULL, "channels" jsonb NULL, "topic" character varying NULL, "template_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "notification_owner_id_idx" to table: "notifications"
CREATE INDEX "notification_owner_id_idx" ON "notifications" ("owner_id");
-- create index "notification_template_id_idx" to table: "notifications"
CREATE INDEX "notification_template_id_idx" ON "notifications" ("template_id");
-- create index "notification_user_id_read_at_owner_id" to table: "notifications"
CREATE INDEX "notification_user_id_read_at_owner_id" ON "notifications" ("user_id", "read_at", "owner_id");
-- create "notification_preferences" table
CREATE TABLE "notification_preferences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "channel" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'ENABLED', "provider" character varying NULL, "destination" character varying NULL, "config" jsonb NULL, "enabled" boolean NOT NULL DEFAULT true, "cadence" character varying NOT NULL DEFAULT 'IMMEDIATE', "priority" character varying NULL, "topic_patterns" jsonb NULL, "topic_overrides" jsonb NULL, "mute_until" timestamptz NULL, "quiet_hours_start" character varying NULL, "quiet_hours_end" character varying NULL, "timezone" character varying NULL, "is_default" boolean NOT NULL DEFAULT false, "verified_at" timestamptz NULL, "last_used_at" timestamptz NULL, "last_error" text NULL, "metadata" jsonb NULL, "user_id" character varying NOT NULL, "template_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "notification_preference_owner_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_owner_id_idx" ON "notification_preferences" ("owner_id");
-- create index "notification_preference_template_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_template_id_idx" ON "notification_preferences" ("template_id");
-- create index "notification_preference_user_id_idx" to table: "notification_preferences"
CREATE INDEX "notification_preference_user_id_idx" ON "notification_preferences" ("user_id");
-- create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
CREATE INDEX "notificationpreference_owner_id_user_id_channel" ON "notification_preferences" ("owner_id", "user_id", "channel") WHERE (deleted_at IS NULL);
-- create "notification_templates" table
CREATE TABLE "notification_templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "key" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "channel" character varying NULL, "format" character varying NOT NULL DEFAULT 'MARKDOWN', "locale" character varying NOT NULL DEFAULT 'en-US', "topic_pattern" character varying NOT NULL, "destinations" jsonb NULL, "title_template" character varying NULL, "subject_template" character varying NULL, "body_template" text NULL, "blocks" jsonb NULL, "jsonconfig" jsonb NULL, "uischema" jsonb NULL, "metadata" jsonb NULL, "active" boolean NOT NULL DEFAULT true, "version" bigint NOT NULL DEFAULT 1, "template_context" character varying NULL, "defaults" jsonb NULL, "email_template_id" character varying NULL, "integration_id" character varying NULL, "owner_id" character varying NULL, "workflow_definition_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "notification_template_email_template_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_email_template_id_idx" ON "notification_templates" ("email_template_id");
-- create index "notification_template_integration_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_integration_id_idx" ON "notification_templates" ("integration_id");
-- create index "notification_template_owner_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_owner_id_idx" ON "notification_templates" ("owner_id");
-- create index "notification_template_workflow_definition_id_idx" to table: "notification_templates"
CREATE INDEX "notification_template_workflow_definition_id_idx" ON "notification_templates" ("workflow_definition_id");
-- create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern" ON "notification_templates" ("owner_id", "channel", "locale", "topic_pattern") WHERE (deleted_at IS NULL);
-- create index "notificationtemplate_owner_id_key" to table: "notification_templates"
CREATE INDEX "notificationtemplate_owner_id_key" ON "notification_templates" ("owner_id", "key") WHERE (deleted_at IS NULL);
-- create "onboardings" table
CREATE TABLE "onboardings" ("id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "company_name" character varying NOT NULL, "domains" jsonb NULL, "company_details" jsonb NULL, "user_details" jsonb NULL, "compliance" jsonb NULL, "demo_requested" boolean NULL DEFAULT false, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- create "org_memberships" table
CREATE TABLE "org_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "sso_exempt" boolean NULL DEFAULT false, "sso_exempt_reason" character varying NULL, "sso_exempt_granted_by" character varying NULL, "sso_exempt_granted_at" timestamptz NULL, "tfa_enforced" boolean NULL DEFAULT false, "tfa_enforced_reason" character varying NULL, "tfa_enforced_by" character varying NULL, "tfa_enforced_at" timestamptz NULL, "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "org_membership_organization_id_idx" to table: "org_memberships"
CREATE INDEX "org_membership_organization_id_idx" ON "org_memberships" ("organization_id");
-- create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id");
-- create "org_modules" table
CREATE TABLE "org_modules" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "module" character varying NOT NULL, "price" jsonb NULL, "stripe_price_id" character varying NULL, "status" character varying NULL, "visibility" character varying NULL, "active" boolean NOT NULL DEFAULT true, "module_lookup_key" character varying NULL, "price_id" character varying NULL, "org_product_org_modules" character varying NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "org_module_owner_id_idx" to table: "org_modules"
CREATE INDEX "org_module_owner_id_idx" ON "org_modules" ("owner_id");
-- create index "org_module_subscription_id_idx" to table: "org_modules"
CREATE INDEX "org_module_subscription_id_idx" ON "org_modules" ("subscription_id");
-- create "org_prices" table
CREATE TABLE "org_prices" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "price" jsonb NULL, "stripe_price_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "product_id" character varying NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "org_price_owner_id_idx" to table: "org_prices"
CREATE INDEX "org_price_owner_id_idx" ON "org_prices" ("owner_id");
-- create index "org_price_subscription_id_idx" to table: "org_prices"
CREATE INDEX "org_price_subscription_id_idx" ON "org_prices" ("subscription_id");
-- create "org_products" table
CREATE TABLE "org_products" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "module" character varying NOT NULL, "stripe_product_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "price_id" character varying NULL, "org_module_org_products" character varying NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "org_product_owner_id_idx" to table: "org_products"
CREATE INDEX "org_product_owner_id_idx" ON "org_products" ("owner_id");
-- create index "org_product_subscription_id_idx" to table: "org_products"
CREATE INDEX "org_product_subscription_id_idx" ON "org_products" ("subscription_id");
-- create "org_subscriptions" table
CREATE TABLE "org_subscriptions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "stripe_subscription_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "expires_at" timestamptz NULL, "trial_expires_at" timestamptz NULL, "days_until_due" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "org_subscription_owner_id_idx" to table: "org_subscriptions"
CREATE INDEX "org_subscription_owner_id_idx" ON "org_subscriptions" ("owner_id");
-- create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "stripe_customer_id" character varying NULL, "slug_name" character varying NULL, "parent_organization_id" character varying NULL, "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "organization_avatar_local_file_id_idx" to table: "organizations"
CREATE INDEX "organization_avatar_local_file_id_idx" ON "organizations" ("avatar_local_file_id");
-- create index "organization_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_name" ON "organizations" ("name") WHERE (deleted_at IS NULL);
-- create index "organization_parent_organization_id_idx" to table: "organizations"
CREATE INDEX "organization_parent_organization_id_idx" ON "organizations" ("parent_organization_id");
-- create index "organizations_stripe_customer_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_stripe_customer_id_key" ON "organizations" ("stripe_customer_id");
-- create "organization_settings" table
CREATE TABLE "organization_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" jsonb NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "billing_notifications_enabled" boolean NOT NULL DEFAULT true, "allowed_email_domains" jsonb NULL, "allow_matching_domains_autojoin" boolean NULL DEFAULT false, "identity_provider" character varying NULL DEFAULT 'NONE', "identity_provider_client_id" character varying NULL, "identity_provider_client_secret" character varying NULL, "identity_provider_metadata_endpoint" character varying NULL, "identity_provider_auth_tested" boolean NOT NULL DEFAULT false, "identity_provider_entity_id" character varying NULL, "oidc_discovery_endpoint" character varying NULL, "saml_signin_url" character varying NULL, "saml_issuer" character varying NULL, "saml_cert" text NULL, "identity_provider_login_enforced" boolean NOT NULL DEFAULT false, "identity_provider_jit_provisioning" boolean NOT NULL DEFAULT true, "jit_allowed_email_domains" jsonb NULL, "multifactor_auth_enforced" boolean NULL DEFAULT false, "sso_exempt_domains" jsonb NULL, "allow_support_access" boolean NULL DEFAULT false, "compliance_webhook_token" character varying NULL, "payment_method_added" boolean NOT NULL DEFAULT false, "pending_deletion_at" timestamptz NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "organization_setting_organization_id_idx" to table: "organization_settings"
CREATE INDEX "organization_setting_organization_id_idx" ON "organization_settings" ("organization_id");
-- create index "organization_settings_compliance_webhook_token_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_compliance_webhook_token_key" ON "organization_settings" ("compliance_webhook_token");
-- create index "organization_settings_organization_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_organization_id_key" ON "organization_settings" ("organization_id");
-- create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "password_reset_tokens_owner_id_fk" to table: "password_reset_tokens"
CREATE INDEX "password_reset_tokens_owner_id_fk" ON "password_reset_tokens" ("owner_id");
-- create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
-- create index "passwordresettoken_token" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_token" ON "password_reset_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "personal_access_tokens" table
CREATE TABLE "personal_access_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "sso_authorizations" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "personal_access_tokens_owner_id_fk" to table: "personal_access_tokens"
CREATE INDEX "personal_access_tokens_owner_id_fk" ON "personal_access_tokens" ("owner_id");
-- create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_token_key" ON "personal_access_tokens" ("token");
-- create index "personalaccesstoken_token" to table: "personal_access_tokens"
CREATE INDEX "personalaccesstoken_token" ON "personal_access_tokens" ("token");
-- create "platforms" table
CREATE TABLE "platforms" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "business_owner" character varying NULL, "technical_owner" character varying NULL, "security_owner" character varying NULL, "platform_kind_name" character varying NULL, "platform_data_classification_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "access_model_name" character varying NULL, "encryption_status_name" character varying NULL, "security_tier_name" character varying NULL, "criticality_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "business_purpose" character varying NULL, "scope_statement" text NULL, "trust_boundary_description" text NULL, "data_flow_summary" text NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "physical_location" character varying NULL, "region" character varying NULL, "contains_pii" boolean NULL DEFAULT false, "source_type" character varying NOT NULL DEFAULT 'MANUAL', "source_identifier" character varying NULL, "cost_center" character varying NULL, "estimated_monthly_cost" double precision NULL, "purchase_date" timestamptz NULL, "external_reference_id" character varying NULL, "metadata" jsonb NULL, "custom_type_enum_platforms" character varying NULL, "identity_holder_access_platforms" character varying NULL, "owner_id" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "business_owner_user_id" character varying NULL, "business_owner_group_id" character varying NULL, "technical_owner_user_id" character varying NULL, "technical_owner_group_id" character varying NULL, "security_owner_user_id" character varying NULL, "security_owner_group_id" character varying NULL, "platform_kind_id" character varying NULL, "platform_data_classification_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "access_model_id" character varying NULL, "encryption_status_id" character varying NULL, "security_tier_id" character varying NULL, "criticality_id" character varying NULL, "platform_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "platform_display_id_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_display_id_owner_id" ON "platforms" ("display_id", "owner_id");
-- create index "platform_external_uuid_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_external_uuid_owner_id" ON "platforms" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "platform_name_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_name_owner_id" ON "platforms" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_owner_id_idx" ON "platforms" ("owner_id");
-- create index "platform_platform_owner_id_idx" to table: "platforms"
CREATE INDEX "platform_platform_owner_id_idx" ON "platforms" ("platform_owner_id");
-- create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED', "details" text NULL, "details_json" jsonb NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "summary" character varying NULL, "tag_suggestions" jsonb NULL, "dismissed_tag_suggestions" jsonb NULL, "control_suggestions" jsonb NULL, "dismissed_control_suggestions" jsonb NULL, "improvement_suggestions" jsonb NULL, "dismissed_improvement_suggestions" jsonb NULL, "url" character varying NULL, "external_file_id" character varying NULL, "external_contents" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "procedure_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "control_objective_procedures" character varying NULL, "custom_type_enum_procedures" character varying NULL, "owner_id" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "procedure_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "file_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "procedure_display_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_display_id_owner_id" ON "procedures" ("display_id", "owner_id");
-- create index "procedure_file_id_idx" to table: "procedures"
CREATE INDEX "procedure_file_id_idx" ON "procedures" ("file_id");
-- create index "procedure_owner_id_idx" to table: "procedures"
CREATE INDEX "procedure_owner_id_idx" ON "procedures" ("owner_id");
-- create "programs" table
CREATE TABLE "programs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "program_kind_name" character varying NULL, "external_uuid" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "framework_name" character varying NULL, "start_date" timestamptz NULL, "end_date" timestamptz NULL, "observation_period_start_date" timestamptz NULL, "observation_period_end_date" timestamptz NULL, "fieldwork_start_date" timestamptz NULL, "fieldwork_end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, "audit_firm" character varying NULL, "auditor" character varying NULL, "auditor_email" character varying NULL, "custom_type_enum_programs" character varying NULL, "owner_id" character varying NULL, "program_kind_id" character varying NULL, "program_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "program_display_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_display_id_owner_id" ON "programs" ("display_id", "owner_id");
-- create index "program_external_uuid_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_external_uuid_owner_id" ON "programs" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "program_owner_id_idx" to table: "programs"
CREATE INDEX "program_owner_id_idx" ON "programs" ("owner_id");
-- create index "program_program_owner_id_idx" to table: "programs"
CREATE INDEX "program_program_owner_id_idx" ON "programs" ("program_owner_id");
-- create "program_memberships" table
CREATE TABLE "program_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, "program_membership_org_membership" character varying NULL, PRIMARY KEY ("id"));
-- create index "program_membership_program_id_idx" to table: "program_memberships"
CREATE INDEX "program_membership_program_id_idx" ON "program_memberships" ("program_id");
-- create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id");
-- create "remediations" table
CREATE TABLE "remediations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NULL, "status" character varying NULL DEFAULT 'IN_PROGRESS', "state" character varying NULL, "intent" character varying NULL, "summary" text NULL, "explanation" text NULL, "instructions" text NULL, "owner_reference" character varying NULL, "repository_uri" character varying NULL, "pull_request_uri" character varying NULL, "ticket_reference" character varying NULL, "due_at" timestamptz NULL, "completed_at" timestamptz NULL, "pr_generated_at" timestamptz NULL, "error" text NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "owner_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "remediation_display_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_display_id_owner_id" ON "remediations" ("display_id", "owner_id");
-- create index "remediation_external_id_external_owner_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_external_id_external_owner_id_owner_id" ON "remediations" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "remediation_owner_id_idx" to table: "remediations"
CREATE INDEX "remediation_owner_id_idx" ON "remediations" ("owner_id");
-- create "reviews" table
CREATE TABLE "reviews" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NOT NULL, "state" character varying NULL, "status" character varying NULL DEFAULT 'OPEN', "category" character varying NULL, "classification" character varying NULL, "summary" text NULL, "details" text NULL, "reporter" character varying NULL, "approved" boolean NULL DEFAULT false, "reviewed_at" timestamptz NULL, "reported_at" timestamptz NULL, "approved_at" timestamptz NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "owner_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "reviewer_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "review_external_id_external_owner_id_owner_id" to table: "reviews"
CREATE UNIQUE INDEX "review_external_id_external_owner_id_owner_id" ON "reviews" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "review_owner_id_idx" to table: "reviews"
CREATE INDEX "review_owner_id_idx" ON "reviews" ("owner_id");
-- create index "review_reviewer_id_idx" to table: "reviews"
CREATE INDEX "review_reviewer_id_idx" ON "reviews" ("reviewer_id");
-- create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "risk_kind_name" character varying NULL, "risk_category_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_id" character varying NULL, "integration_id" character varying NULL, "observed_at" timestamptz NULL, "external_uuid" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL, "impact" character varying NULL, "likelihood" character varying NULL DEFAULT 'LIKELY', "score" bigint NULL, "mitigation" text NULL, "mitigation_json" jsonb NULL, "details" text NULL, "details_json" jsonb NULL, "business_costs" text NULL, "business_costs_json" jsonb NULL, "mitigated_at" timestamptz NULL, "review_required" boolean NULL DEFAULT true, "last_reviewed_at" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "due_date" timestamptz NULL, "next_review_due_at" timestamptz NULL, "residual_score" bigint NULL, "risk_decision" character varying NULL DEFAULT 'NONE', "control_objective_risks" character varying NULL, "custom_type_enum_risks" character varying NULL, "custom_type_enum_risk_categories" character varying NULL, "owner_id" character varying NULL, "risk_kind_id" character varying NULL, "risk_category_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "stakeholder_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "risk_delegate_id_idx" to table: "risks"
CREATE INDEX "risk_delegate_id_idx" ON "risks" ("delegate_id");
-- create index "risk_display_id_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_display_id_owner_id" ON "risks" ("display_id", "owner_id");
-- create index "risk_external_uuid_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_external_uuid_owner_id" ON "risks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "risk_owner_id_idx" to table: "risks"
CREATE INDEX "risk_owner_id_idx" ON "risks" ("owner_id");
-- create index "risk_stakeholder_id_idx" to table: "risks"
CREATE INDEX "risk_stakeholder_id_idx" ON "risks" ("stakeholder_id");
-- create "sla_definitions" table
CREATE TABLE "sla_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "sla_days" bigint NOT NULL, "security_level" character varying NOT NULL DEFAULT 'NONE', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "sla_definition_owner_id_idx" to table: "sla_definitions"
CREATE INDEX "sla_definition_owner_id_idx" ON "sla_definitions" ("owner_id");
-- create index "sladefinition_display_id_owner_id" to table: "sla_definitions"
CREATE UNIQUE INDEX "sladefinition_display_id_owner_id" ON "sla_definitions" ("display_id", "owner_id");
-- create index "sladefinition_security_level_owner_id" to table: "sla_definitions"
CREATE UNIQUE INDEX "sladefinition_security_level_owner_id" ON "sla_definitions" ("security_level", "owner_id") WHERE (deleted_at IS NULL);
-- create "scans" table
CREATE TABLE "scans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "reviewed_by" character varying NULL, "assigned_to" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "target" character varying NOT NULL, "scan_type" character varying NOT NULL DEFAULT 'DOMAIN', "metadata" jsonb NULL, "scan_date" timestamptz NULL, "scan_schedule" character varying NULL, "next_scan_run_at" timestamptz NULL, "performed_by" character varying NULL, "discovered_vulnerability_ids" jsonb NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "owner_id" character varying NULL, "generated_by_platform_id" character varying NULL, "risk_scans" character varying NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "assigned_to_user_id" character varying NULL, "assigned_to_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "performed_by_user_id" character varying NULL, "performed_by_group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "scan_generated_by_platform_id_idx" to table: "scans"
CREATE INDEX "scan_generated_by_platform_id_idx" ON "scans" ("generated_by_platform_id");
-- create index "scan_owner_id_idx" to table: "scans"
CREATE INDEX "scan_owner_id_idx" ON "scans" ("owner_id");
-- create index "scan_performed_by_group_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_group_id_idx" ON "scans" ("performed_by_group_id");
-- create index "scan_performed_by_user_id_idx" to table: "scans"
CREATE INDEX "scan_performed_by_user_id_idx" ON "scans" ("performed_by_user_id");
-- create "scheduled_jobs" table
CREATE TABLE "scheduled_jobs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "active" boolean NOT NULL DEFAULT true, "configuration" jsonb NULL, "cron" character varying NULL, "job_id" character varying NOT NULL, "owner_id" character varying NULL, "job_runner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "scheduled_job_job_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_id_idx" ON "scheduled_jobs" ("job_id");
-- create index "scheduled_job_job_runner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_job_runner_id_idx" ON "scheduled_jobs" ("job_runner_id");
-- create index "scheduled_job_owner_id_idx" to table: "scheduled_jobs"
CREATE INDEX "scheduled_job_owner_id_idx" ON "scheduled_jobs" ("owner_id");
-- create index "scheduledjob_display_id_owner_id" to table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_display_id_owner_id" ON "scheduled_jobs" ("display_id", "owner_id");
-- create "scheduled_job_runs" table
CREATE TABLE "scheduled_job_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "expected_execution_time" timestamptz NOT NULL, "script" character varying NOT NULL, "owner_id" character varying NULL, "scheduled_job_id" character varying NOT NULL, "job_runner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "scheduled_job_run_job_runner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_job_runner_id_idx" ON "scheduled_job_runs" ("job_runner_id");
-- create index "scheduled_job_run_owner_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_owner_id_idx" ON "scheduled_job_runs" ("owner_id");
-- create index "scheduled_job_run_scheduled_job_id_idx" to table: "scheduled_job_runs"
CREATE INDEX "scheduled_job_run_scheduled_job_id_idx" ON "scheduled_job_runs" ("scheduled_job_id");
-- create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "short_name" character varying NULL, "framework" text NULL, "description" text NULL, "governing_body_logo_url" character varying NULL, "governing_body" character varying NULL, "domains" jsonb NULL, "link" character varying NULL, "status" character varying NULL DEFAULT 'ACTIVE', "is_public" boolean NULL DEFAULT false, "free_to_use" boolean NULL DEFAULT false, "standard_type" character varying NULL, "version" character varying NULL, "owner_id" character varying NULL, "logo_file_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "standard_logo_file_id_idx" to table: "standards"
CREATE INDEX "standard_logo_file_id_idx" ON "standards" ("logo_file_id");
-- create index "standard_owner_id_idx" to table: "standards"
CREATE INDEX "standard_owner_id_idx" ON "standards" ("owner_id");
-- create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "external_uuid" character varying NULL, "title" character varying NULL, "description" text NULL, "description_json" jsonb NULL, "aliases" jsonb NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NOT_IMPLEMENTED', "implementation_status" character varying NULL DEFAULT 'PLANNED', "implementation_description" text NULL, "public_representation" text NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "source_name" character varying NULL, "reference_framework" character varying NULL, "reference_framework_revision" character varying NULL, "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "testing_procedures" jsonb NULL, "evidence_requests" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "subcontrol_kind_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "ref_code" character varying NOT NULL, "control_id" character varying NOT NULL, "custom_type_enum_subcontrols" character varying NULL, "owner_id" character varying NULL, "program_subcontrols" character varying NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "responsible_party_id" character varying NULL, "subcontrol_kind_id" character varying NULL, "user_subcontrols" character varying NULL, PRIMARY KEY ("id"));
-- create index "subcontrol_auditor_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_auditor_reference_id_deleted_at_owner_id" ON "subcontrols" ("auditor_reference_id", "deleted_at", "owner_id");
-- create index "subcontrol_control_id_ref_code" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_control_id_ref_code" ON "subcontrols" ("control_id", "ref_code") WHERE (deleted_at IS NULL);
-- create index "subcontrol_control_id_ref_code_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_control_id_ref_code_owner_id" ON "subcontrols" ("control_id", "ref_code", "owner_id") WHERE (deleted_at IS NULL);
-- create index "subcontrol_display_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_display_id_owner_id" ON "subcontrols" ("display_id", "owner_id");
-- create index "subcontrol_external_uuid_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_external_uuid_owner_id" ON "subcontrols" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "subcontrol_owner_id_idx" to table: "subcontrols"
CREATE INDEX "subcontrol_owner_id_idx" ON "subcontrols" ("owner_id");
-- create index "subcontrol_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_reference_id_deleted_at_owner_id" ON "subcontrols" ("reference_id", "deleted_at", "owner_id");
-- create "subprocessors" table
CREATE TABLE "subprocessors" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "logo_remote_url" character varying NULL, "owner_id" character varying NULL, "logo_file_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "subprocessor_logo_file_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_logo_file_id_idx" ON "subprocessors" ("logo_file_id");
-- create index "subprocessor_name_owner_id" to table: "subprocessors"
CREATE UNIQUE INDEX "subprocessor_name_owner_id" ON "subprocessors" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "subprocessor_owner_id_idx" to table: "subprocessors"
CREATE INDEX "subprocessor_owner_id_idx" ON "subprocessors" ("owner_id");
-- create "subscribers" table
CREATE TABLE "subscribers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "phone_number" character varying NULL, "verified_email" boolean NOT NULL DEFAULT false, "verified_phone" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT false, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "secret" bytea NOT NULL, "unsubscribed" boolean NOT NULL DEFAULT false, "send_attempts" bigint NOT NULL DEFAULT 1, "contact_id" character varying NULL, "owner_id" character varying NULL, "trust_center_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "subscriber_contact_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_contact_id_idx" ON "subscribers" ("contact_id");
-- create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NULL));
-- create index "subscriber_email_trust_center_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_trust_center_id" ON "subscribers" ("email", "trust_center_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false) AND (trust_center_id IS NOT NULL));
-- create index "subscriber_owner_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_owner_id_idx" ON "subscribers" ("owner_id");
-- create index "subscriber_trust_center_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_trust_center_id_idx" ON "subscribers" ("trust_center_id");
-- create index "subscriber_user_id_idx" to table: "subscribers"
CREATE INDEX "subscriber_user_id_idx" ON "subscribers" ("user_id");
-- create index "subscribers_token_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_token_key" ON "subscribers" ("token");
-- create "system_details" table
CREATE TABLE "system_details" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_name" character varying NOT NULL, "version" character varying NULL, "description" text NULL, "authorization_boundary" text NULL, "sensitivity_level" character varying NULL DEFAULT 'UNKNOWN', "last_reviewed" timestamptz NULL, "revision_history" jsonb NULL, "oscal_metadata_json" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "system_detail_owner_id_idx" to table: "system_details"
CREATE INDEX "system_detail_owner_id_idx" ON "system_details" ("owner_id");
-- create index "systemdetail_display_id_owner_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_display_id_owner_id" ON "system_details" ("display_id", "owner_id");
-- create "tfa_settings" table
CREATE TABLE "tfa_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tfa_secret" character varying NULL, "verified" boolean NOT NULL DEFAULT false, "recovery_codes" jsonb NULL, "phone_otp_allowed" boolean NULL DEFAULT false, "email_otp_allowed" boolean NULL DEFAULT false, "totp_allowed" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "tfa_settings_owner_id_fk" to table: "tfa_settings"
CREATE INDEX "tfa_settings_owner_id_fk" ON "tfa_settings" ("owner_id");
-- create index "tfasetting_owner_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_owner_id" ON "tfa_settings" ("owner_id") WHERE (deleted_at IS NULL);
-- create "tag_definitions" table
CREATE TABLE "tag_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" citext NOT NULL, "aliases" jsonb NULL, "slug" citext NULL, "description" character varying NULL, "color" character varying NULL, "owner_id" character varying NULL, "workflow_definition_tag_definitions" character varying NULL, PRIMARY KEY ("id"));
-- create index "tag_definition_owner_id_idx" to table: "tag_definitions"
CREATE INDEX "tag_definition_owner_id_idx" ON "tag_definitions" ("owner_id");
-- create index "tagdefinition_name_owner_id" to table: "tag_definitions"
CREATE UNIQUE INDEX "tagdefinition_name_owner_id" ON "tag_definitions" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "tagdefinition_slug_owner_id" to table: "tag_definitions"
CREATE UNIQUE INDEX "tagdefinition_slug_owner_id" ON "tag_definitions" ("slug", "owner_id") WHERE (deleted_at IS NULL);
-- create "tasks" table
CREATE TABLE "tasks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "task_kind_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_uuid" character varying NULL, "title" character varying NOT NULL, "details" text NULL, "details_json" jsonb NULL, "metadata" jsonb NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "due" timestamptz NULL, "completed" timestamptz NULL, "system_generated" boolean NOT NULL DEFAULT false, "is_template" boolean NOT NULL DEFAULT false, "is_suggested" boolean NOT NULL DEFAULT false, "priority" bigint NOT NULL DEFAULT 0, "source" character varying NULL, "source_key" character varying NULL, "idempotency_key" character varying NULL, "external_reference_url" jsonb NULL, "custom_type_enum_tasks" character varying NULL, "integration_tasks" character varying NULL, "owner_id" character varying NULL, "remediation_tasks" character varying NULL, "review_tasks" character varying NULL, "task_kind_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "parent_task_id" character varying NULL, "assigner_id" character varying NULL, "assignee_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "tasks_tasks_tasks" FOREIGN KEY ("parent_task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "task_assignee_id_idx" to table: "tasks"
CREATE INDEX "task_assignee_id_idx" ON "tasks" ("assignee_id");
-- create index "task_assigner_id_idx" to table: "tasks"
CREATE INDEX "task_assigner_id_idx" ON "tasks" ("assigner_id");
-- create index "task_display_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_display_id_owner_id" ON "tasks" ("display_id", "owner_id");
-- create index "task_external_uuid_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_external_uuid_owner_id" ON "tasks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- create index "task_owner_id_idempotency_key" to table: "tasks"
CREATE UNIQUE INDEX "task_owner_id_idempotency_key" ON "tasks" ("owner_id", "idempotency_key") WHERE ((deleted_at IS NULL) AND (idempotency_key IS NOT NULL));
-- create index "task_owner_id_idx" to table: "tasks"
CREATE INDEX "task_owner_id_idx" ON "tasks" ("owner_id");
-- create index "task_owner_id_is_suggested_priority" to table: "tasks"
CREATE INDEX "task_owner_id_is_suggested_priority" ON "tasks" ("owner_id", "is_suggested", "priority") WHERE (deleted_at IS NULL);
-- create index "task_parent_task_id_idx" to table: "tasks"
CREATE INDEX "task_parent_task_id_idx" ON "tasks" ("parent_task_id");
-- create "templates" table
CREATE TABLE "templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "kind" character varying NULL DEFAULT 'QUESTIONNAIRE', "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, "transform_configuration" jsonb NULL, "owner_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "templates_check" CHECK ((trust_center_id IS NOT NULL) OR ((kind)::text <> 'TRUSTCENTER_NDA'::text)));
-- create index "template_name_owner_id_template_type" to table: "templates"
CREATE UNIQUE INDEX "template_name_owner_id_template_type" ON "templates" ("name", "owner_id", "template_type") WHERE (deleted_at IS NULL);
-- create index "template_owner_id_idx" to table: "templates"
CREATE INDEX "template_owner_id_idx" ON "templates" ("owner_id");
-- create index "template_trust_center_id" to table: "templates"
CREATE UNIQUE INDEX "template_trust_center_id" ON "templates" ("trust_center_id") WHERE ((deleted_at IS NULL) AND ((kind)::text = 'TRUSTCENTER_NDA'::text));
-- create "trust_centers" table
CREATE TABLE "trust_centers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "slug" character varying NULL, "pirsch_domain_id" character varying NULL, "pirsch_identification_code" character varying NULL, "pirsch_access_link" character varying NULL, "preview_status" character varying NULL DEFAULT 'NONE', "subprocessor_url" character varying NULL, "owner_id" character varying NULL, "custom_domain_id" character varying NULL, "preview_domain_id" character varying NULL, "trust_center_setting" character varying NULL, "trust_center_preview_setting" character varying NULL, "trust_center_watermark_config" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_custom_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_custom_domain_id_idx" ON "trust_centers" ("custom_domain_id");
-- create index "trust_center_owner_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_owner_id_idx" ON "trust_centers" ("owner_id");
-- create index "trust_center_preview_domain_id_idx" to table: "trust_centers"
CREATE INDEX "trust_center_preview_domain_id_idx" ON "trust_centers" ("preview_domain_id");
-- create index "trustcenter_slug" to table: "trust_centers"
CREATE UNIQUE INDEX "trustcenter_slug" ON "trust_centers" ("slug") WHERE (deleted_at IS NULL);
-- create "trust_center_compliances" table
CREATE TABLE "trust_center_compliances" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "standard_id" character varying NOT NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_compliance_trust_center_id_idx" to table: "trust_center_compliances"
CREATE INDEX "trust_center_compliance_trust_center_id_idx" ON "trust_center_compliances" ("trust_center_id");
-- create index "trustcentercompliance_standard_id_trust_center_id" to table: "trust_center_compliances"
CREATE UNIQUE INDEX "trustcentercompliance_standard_id_trust_center_id" ON "trust_center_compliances" ("standard_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- create "trust_center_docs" table
CREATE TABLE "trust_center_docs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "trust_center_doc_kind_name" character varying NULL, "title" character varying NOT NULL, "watermarking_enabled" boolean NULL, "watermark_status" character varying NULL DEFAULT 'PENDING', "visibility" character varying NULL DEFAULT 'NOT_VISIBLE', "standard_id" character varying NULL, "trust_center_id" character varying NULL, "trust_center_doc_kind_id" character varying NULL, "file_id" character varying NULL, "original_file_id" character varying NULL, "trust_center_nda_request_trust_center_docs" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_doc_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_file_id_idx" ON "trust_center_docs" ("file_id");
-- create index "trust_center_doc_original_file_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_original_file_id_idx" ON "trust_center_docs" ("original_file_id");
-- create index "trust_center_doc_standard_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_standard_id_idx" ON "trust_center_docs" ("standard_id");
-- create index "trust_center_doc_trust_center_id_idx" to table: "trust_center_docs"
CREATE INDEX "trust_center_doc_trust_center_id_idx" ON "trust_center_docs" ("trust_center_id");
-- create "trust_center_entities" table
CREATE TABLE "trust_center_entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "url" character varying NULL, "name" character varying NOT NULL, "file_trust_center_entities" character varying NULL, "trust_center_id" character varying NULL, "logo_file_id" character varying NULL, "entity_type_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_entity_entity_type_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_entity_type_id_idx" ON "trust_center_entities" ("entity_type_id");
-- create index "trust_center_entity_logo_file_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_logo_file_id_idx" ON "trust_center_entities" ("logo_file_id");
-- create index "trust_center_entity_trust_center_id_idx" to table: "trust_center_entities"
CREATE INDEX "trust_center_entity_trust_center_id_idx" ON "trust_center_entities" ("trust_center_id");
-- create "trust_center_faqs" table
CREATE TABLE "trust_center_faqs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_faq_kind_name" character varying NULL, "reference_link" character varying NULL, "display_order" bigint NULL DEFAULT 0, "note_id" character varying NOT NULL, "trust_center_id" character varying NULL, "trust_center_faq_kind_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_faq_trust_center_id_idx" to table: "trust_center_faqs"
CREATE INDEX "trust_center_faq_trust_center_id_idx" ON "trust_center_faqs" ("trust_center_id");
-- create index "trustcenterfaq_note_id_trust_center_id" to table: "trust_center_faqs"
CREATE UNIQUE INDEX "trustcenterfaq_note_id_trust_center_id" ON "trust_center_faqs" ("note_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- create "trust_center_nda_requests" table
CREATE TABLE "trust_center_nda_requests" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "first_name" character varying NOT NULL, "last_name" character varying NOT NULL, "email" character varying NOT NULL, "company_name" character varying NULL, "reason" character varying NULL, "access_level" character varying NULL DEFAULT 'FULL', "status" character varying NULL DEFAULT 'REQUESTED', "approved_at" timestamptz NULL, "signed_at" timestamptz NULL, "trust_center_id" character varying NULL, "document_data_id" character varying NULL, "file_id" character varying NULL, "approved_by_user_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_nda_request_approved_by_user_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_approved_by_user_id_idx" ON "trust_center_nda_requests" ("approved_by_user_id");
-- create index "trust_center_nda_request_document_data_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_document_data_id_idx" ON "trust_center_nda_requests" ("document_data_id");
-- create index "trust_center_nda_request_file_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_file_id_idx" ON "trust_center_nda_requests" ("file_id");
-- create index "trust_center_nda_request_trust_center_id_idx" to table: "trust_center_nda_requests"
CREATE INDEX "trust_center_nda_request_trust_center_id_idx" ON "trust_center_nda_requests" ("trust_center_id");
-- create "trust_center_settings" table
CREATE TABLE "trust_center_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "title" character varying NULL, "company_name" character varying NULL, "company_description" character varying NULL, "overview" text NULL, "logo_remote_url" character varying NULL, "favicon_remote_url" character varying NULL, "theme_mode" character varying NULL DEFAULT 'EASY', "primary_color" character varying NULL, "font" character varying NULL, "foreground_color" character varying NULL, "background_color" character varying NULL, "accent_color" character varying NULL, "secondary_background_color" character varying NULL, "secondary_foreground_color" character varying NULL, "environment" character varying NULL DEFAULT 'LIVE', "remove_branding" boolean NULL DEFAULT false, "company_domain" character varying NULL, "security_contact" character varying NULL, "nda_approval_required" boolean NULL DEFAULT false, "allow_subscribers" boolean NULL DEFAULT true, "notify_subscribers_on_subprocessor_change" boolean NULL DEFAULT false, "subprocessors_notified_at" timestamptz NULL, "status_page_url" character varying NULL, "logo_local_file_id" character varying NULL, "favicon_local_file_id" character varying NULL, "hero_image_local_file_id" character varying NULL, "nda_approver_group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_setting_favicon_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_favicon_local_file_id_idx" ON "trust_center_settings" ("favicon_local_file_id");
-- create index "trust_center_setting_hero_image_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_hero_image_local_file_id_idx" ON "trust_center_settings" ("hero_image_local_file_id");
-- create index "trust_center_setting_logo_local_file_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_logo_local_file_id_idx" ON "trust_center_settings" ("logo_local_file_id");
-- create index "trust_center_setting_nda_approver_group_id_idx" to table: "trust_center_settings"
CREATE INDEX "trust_center_setting_nda_approver_group_id_idx" ON "trust_center_settings" ("nda_approver_group_id");
-- create index "trustcentersetting_trust_center_id_environment" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_trust_center_id_environment" ON "trust_center_settings" ("trust_center_id", "environment") WHERE (deleted_at IS NULL);
-- create "trust_center_subprocessors" table
CREATE TABLE "trust_center_subprocessors" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_subprocessor_kind_name" character varying NULL, "countries" jsonb NULL, "subprocessor_id" character varying NOT NULL, "trust_center_id" character varying NULL, "trust_center_subprocessor_kind_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trust_center_subprocessor_trust_center_id_idx" to table: "trust_center_subprocessors"
CREATE INDEX "trust_center_subprocessor_trust_center_id_idx" ON "trust_center_subprocessors" ("trust_center_id");
-- create index "trustcentersubprocessor_subprocessor_id_trust_center_id" to table: "trust_center_subprocessors"
CREATE UNIQUE INDEX "trustcentersubprocessor_subprocessor_id_trust_center_id" ON "trust_center_subprocessors" ("subprocessor_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- create "trust_center_watermark_configs" table
CREATE TABLE "trust_center_watermark_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "is_enabled" boolean NULL DEFAULT true, "text" character varying NULL, "font_size" double precision NULL DEFAULT 48, "opacity" double precision NULL DEFAULT 0.3, "rotation" double precision NULL DEFAULT 45, "color" character varying NULL DEFAULT '#808080', "font" character varying NULL DEFAULT 'HELVETICA', "owner_id" character varying NULL, "logo_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "text_or_logo_id_not_null" CHECK ((text IS NOT NULL) OR (logo_id IS NOT NULL)));
-- create index "trust_center_watermark_config_logo_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_logo_id_idx" ON "trust_center_watermark_configs" ("logo_id");
-- create index "trust_center_watermark_config_owner_id_idx" to table: "trust_center_watermark_configs"
CREATE INDEX "trust_center_watermark_config_owner_id_idx" ON "trust_center_watermark_configs" ("owner_id");
-- create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
CREATE UNIQUE INDEX "trustcenterwatermarkconfig_trust_center_id" ON "trust_center_watermark_configs" ("trust_center_id") WHERE (deleted_at IS NULL);
-- create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "last_login_provider" character varying NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', "scim_external_id" character varying NULL, "scim_username" character varying NULL, "scim_active" boolean NULL DEFAULT true, "scim_preferred_language" character varying NULL, "scim_locale" character varying NULL, "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "user_email" to table: "users"
CREATE UNIQUE INDEX "user_email" ON "users" ("email") WHERE (deleted_at IS NULL);
-- create index "users_display_id_key" to table: "users"
CREATE UNIQUE INDEX "users_display_id_key" ON "users" ("display_id");
-- create index "users_sub_key" to table: "users"
CREATE UNIQUE INDEX "users_sub_key" ON "users" ("sub");
-- create "user_settings" table
CREATE TABLE "user_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "delegate_user_id" character varying NULL, "delegate_start_at" timestamptz NULL, "delegate_end_at" timestamptz NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, "user_id" character varying NULL, "user_setting_default_org" character varying NULL, PRIMARY KEY ("id"));
-- create index "user_setting_user_id_idx" to table: "user_settings"
CREATE INDEX "user_setting_user_id_idx" ON "user_settings" ("user_id");
-- create index "user_settings_user_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_user_id_key" ON "user_settings" ("user_id");
-- create "vendor_risk_scores" table
CREATE TABLE "vendor_risk_scores" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "question_key" character varying NOT NULL, "question_name" character varying NOT NULL, "question_description" character varying NULL, "question_category" character varying NOT NULL, "answer_type" character varying NOT NULL, "impact" character varying NOT NULL, "likelihood" character varying NOT NULL, "score" double precision NOT NULL DEFAULT 0, "answer" character varying NULL, "notes" character varying NULL, "assessment_response_vendor_risk_scores" character varying NULL, "entity_vendor_risk_scores" character varying NULL, "owner_id" character varying NULL, "vendor_scoring_config_id" character varying NULL, "entity_id" character varying NOT NULL, "assessment_response_id" character varying NULL, "vendor_scoring_config_vendor_risk_scores" character varying NULL, PRIMARY KEY ("id"));
-- create index "vendor_risk_score_assessment_response_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_assessment_response_id_idx" ON "vendor_risk_scores" ("assessment_response_id");
-- create index "vendor_risk_score_entity_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_entity_id_idx" ON "vendor_risk_scores" ("entity_id");
-- create index "vendor_risk_score_owner_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_owner_id_idx" ON "vendor_risk_scores" ("owner_id");
-- create index "vendor_risk_score_vendor_scoring_config_id_idx" to table: "vendor_risk_scores"
CREATE INDEX "vendor_risk_score_vendor_scoring_config_id_idx" ON "vendor_risk_scores" ("vendor_scoring_config_id");
-- create "vendor_scoring_configs" table
CREATE TABLE "vendor_scoring_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "questions" jsonb NOT NULL, "scoring_mode" character varying NOT NULL DEFAULT 'ANSWERED_ONLY', "risk_thresholds" jsonb NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "vendor_scoring_config_owner_id_idx" to table: "vendor_scoring_configs"
CREATE INDEX "vendor_scoring_config_owner_id_idx" ON "vendor_scoring_configs" ("owner_id");
-- create "vulnerabilities" table
CREATE TABLE "vulnerabilities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "reviewed_by" character varying NULL, "assigned_to" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "vulnerability_status_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "external_owner_id" character varying NULL, "security_level" character varying NULL DEFAULT 'NONE', "external_id" character varying NOT NULL, "cve_id" character varying NULL, "source" character varying NULL, "display_name" character varying NULL, "category" character varying NULL, "severity" character varying NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "summary" text NULL, "description" text NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "open" boolean NULL DEFAULT true, "blocking" boolean NULL DEFAULT false, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "references" jsonb NULL, "impacts" jsonb NULL, "cwe_ids" jsonb NULL, "vulnerable_version_range" character varying NULL, "first_patched_version" character varying NULL, "fix_available" boolean NULL, "package_name" character varying NULL, "package_ecosystem" character varying NULL, "manifest_path" character varying NULL, "dependency_scope" character varying NULL, "published_at" timestamptz NULL, "discovered_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "dismissed_at" timestamptz NULL, "dismissed_reason" character varying NULL, "dismissed_comment" text NULL, "fixed_at" timestamptz NULL, "auto_dismissed_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "owner_id" character varying NULL, "reviewed_by_user_id" character varying NULL, "reviewed_by_group_id" character varying NULL, "assigned_to_user_id" character varying NULL, "assigned_to_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "vulnerability_status_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_cve_id_owner_id" ON "vulnerabilities" ("cve_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "vulnerability_display_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_display_id_owner_id" ON "vulnerabilities" ("display_id", "owner_id");
-- create index "vulnerability_external_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_external_id_owner_id" ON "vulnerabilities" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "vulnerability_owner_id_idx" to table: "vulnerabilities"
CREATE INDEX "vulnerability_owner_id_idx" ON "vulnerabilities" ("owner_id");
-- create "webauthns" table
CREATE TABLE "webauthns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "credential_id" bytea NULL, "public_key" bytea NULL, "attestation_type" character varying NULL, "aaguid" bytea NOT NULL, "sign_count" integer NOT NULL, "transports" jsonb NOT NULL, "backup_eligible" boolean NOT NULL DEFAULT false, "backup_state" boolean NOT NULL DEFAULT false, "user_present" boolean NOT NULL DEFAULT false, "user_verified" boolean NOT NULL DEFAULT false, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "webauthns_credential_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_credential_id_key" ON "webauthns" ("credential_id");
-- create index "webauthns_owner_id_fk" to table: "webauthns"
CREATE INDEX "webauthns_owner_id_fk" ON "webauthns" ("owner_id");
-- create "workflow_assignments" table
CREATE TABLE "workflow_assignments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "assignment_key" character varying NOT NULL, "role" character varying NOT NULL DEFAULT 'APPROVER', "label" character varying NULL, "required" boolean NOT NULL DEFAULT true, "status" character varying NOT NULL DEFAULT 'PENDING', "metadata" jsonb NULL, "approval_metadata" jsonb NULL, "rejection_metadata" jsonb NULL, "invalidation_metadata" jsonb NULL, "outcome_metadata" jsonb NULL, "decided_at" timestamptz NULL, "notes" text NULL, "due_at" timestamptz NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "actor_user_id" character varying NULL, "actor_group_id" character varying NULL, "workflow_instance_workflow_assignments" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflow_assignment_actor_group_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_group_id_idx" ON "workflow_assignments" ("actor_group_id");
-- create index "workflow_assignment_actor_user_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_actor_user_id_idx" ON "workflow_assignments" ("actor_user_id");
-- create index "workflow_assignment_owner_id_idx" to table: "workflow_assignments"
CREATE INDEX "workflow_assignment_owner_id_idx" ON "workflow_assignments" ("owner_id");
-- create index "workflowassignment_display_id_owner_id" to table: "workflow_assignments"
CREATE UNIQUE INDEX "workflowassignment_display_id_owner_id" ON "workflow_assignments" ("display_id", "owner_id");
-- create index "workflowassignment_workflow_instance_id_assignment_key" to table: "workflow_assignments"
CREATE UNIQUE INDEX "workflowassignment_workflow_instance_id_assignment_key" ON "workflow_assignments" ("workflow_instance_id", "assignment_key");
-- create "workflow_assignment_targets" table
CREATE TABLE "workflow_assignment_targets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "target_type" character varying NOT NULL, "resolver_key" character varying NULL, "owner_id" character varying NULL, "workflow_assignment_workflow_assignment_targets" character varying NULL, "workflow_assignment_id" character varying NOT NULL, "target_user_id" character varying NULL, "target_group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflow_assignment_target_owner_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_owner_id_idx" ON "workflow_assignment_targets" ("owner_id");
-- create index "workflow_assignment_target_target_group_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_group_id_idx" ON "workflow_assignment_targets" ("target_group_id");
-- create index "workflow_assignment_target_target_user_id_idx" to table: "workflow_assignment_targets"
CREATE INDEX "workflow_assignment_target_target_user_id_idx" ON "workflow_assignment_targets" ("target_user_id");
-- create index "workflowassignmenttarget_display_id_owner_id" to table: "workflow_assignment_targets"
CREATE UNIQUE INDEX "workflowassignmenttarget_display_id_owner_id" ON "workflow_assignment_targets" ("display_id", "owner_id");
-- create index "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" to table: "workflow_assignment_targets"
CREATE UNIQUE INDEX "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" ON "workflow_assignment_targets" ("workflow_assignment_id", "target_type", "target_user_id", "target_group_id", "resolver_key") WHERE (deleted_at IS NULL);
-- create index "workflowassignmenttarget_workflow_assignment_id" to table: "workflow_assignment_targets"
CREATE INDEX "workflowassignmenttarget_workflow_assignment_id" ON "workflow_assignment_targets" ("workflow_assignment_id") WHERE (deleted_at IS NULL);
-- create "workflow_definitions" table
CREATE TABLE "workflow_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "workflow_kind" character varying NOT NULL, "schema_type" character varying NOT NULL, "revision" bigint NOT NULL DEFAULT 1, "draft" boolean NOT NULL DEFAULT true, "published_at" timestamptz NULL, "cooldown_seconds" bigint NOT NULL DEFAULT 0, "is_default" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT true, "trigger_operations" jsonb NULL, "trigger_fields" jsonb NULL, "approval_fields" jsonb NULL, "approval_edges" jsonb NULL, "approval_submission_mode" character varying NULL DEFAULT 'AUTO_SUBMIT', "definition_json" jsonb NULL, "tracked_fields" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflow_definition_owner_id_idx" to table: "workflow_definitions"
CREATE INDEX "workflow_definition_owner_id_idx" ON "workflow_definitions" ("owner_id");
-- create index "workflowdefinition_display_id_owner_id" to table: "workflow_definitions"
CREATE UNIQUE INDEX "workflowdefinition_display_id_owner_id" ON "workflow_definitions" ("display_id", "owner_id");
-- create "workflow_events" table
CREATE TABLE "workflow_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "event_type" character varying NOT NULL, "payload" jsonb NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "workflow_instance_workflow_events" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflow_event_owner_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_owner_id_idx" ON "workflow_events" ("owner_id");
-- create index "workflow_event_workflow_instance_id_idx" to table: "workflow_events"
CREATE INDEX "workflow_event_workflow_instance_id_idx" ON "workflow_events" ("workflow_instance_id");
-- create index "workflowevent_display_id_owner_id" to table: "workflow_events"
CREATE UNIQUE INDEX "workflowevent_display_id_owner_id" ON "workflow_events" ("display_id", "owner_id");
-- create "workflow_instances" table
CREATE TABLE "workflow_instances" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "state" character varying NOT NULL DEFAULT 'RUNNING', "context" jsonb NULL, "last_evaluated_at" timestamptz NULL, "definition_snapshot" jsonb NULL, "current_action_index" bigint NOT NULL DEFAULT 0, "owner_id" character varying NULL, "workflow_definition_id" character varying NOT NULL, "control_id" character varying NULL, "internal_policy_id" character varying NULL, "evidence_id" character varying NULL, "subcontrol_id" character varying NULL, "action_plan_id" character varying NULL, "procedure_id" character varying NULL, "campaign_id" character varying NULL, "campaign_target_id" character varying NULL, "identity_holder_id" character varying NULL, "platform_id" character varying NULL, "assessment_id" character varying NULL, "assessment_response_id" character varying NULL, "finding_id" character varying NULL, "integration_id" character varying NULL, "remediation_id" character varying NULL, "risk_id" character varying NULL, "task_id" character varying NULL, "vulnerability_id" character varying NULL, "workflow_proposal_id" character varying NULL, PRIMARY KEY ("id"));
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
-- create index "workflowinstance_display_id_owner_id" to table: "workflow_instances"
CREATE UNIQUE INDEX "workflowinstance_display_id_owner_id" ON "workflow_instances" ("display_id", "owner_id");
-- create index "workflowinstance_workflow_definition_id" to table: "workflow_instances"
CREATE INDEX "workflowinstance_workflow_definition_id" ON "workflow_instances" ("workflow_definition_id") WHERE (deleted_at IS NULL);
-- create "workflow_object_refs" table
CREATE TABLE "workflow_object_refs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "workflow_instance_workflow_object_refs" character varying NULL, "workflow_instance_id" character varying NOT NULL, "control_id" character varying NULL, "task_id" character varying NULL, "internal_policy_id" character varying NULL, "finding_id" character varying NULL, "directory_account_id" character varying NULL, "directory_group_id" character varying NULL, "directory_membership_id" character varying NULL, "evidence_id" character varying NULL, "subcontrol_id" character varying NULL, "action_plan_id" character varying NULL, "procedure_id" character varying NULL, "campaign_id" character varying NULL, "campaign_target_id" character varying NULL, "identity_holder_id" character varying NULL, "platform_id" character varying NULL, "vulnerability_id" character varying NULL, "risk_id" character varying NULL, "assessment_id" character varying NULL, "assessment_response_id" character varying NULL, "remediation_id" character varying NULL, PRIMARY KEY ("id"));
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
-- create index "workflowobjectref_display_id_owner_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_display_id_owner_id" ON "workflow_object_refs" ("display_id", "owner_id");
-- create index "workflowobjectref_workflow_instance_id_action_plan_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_action_plan_id" ON "workflow_object_refs" ("workflow_instance_id", "action_plan_id");
-- create index "workflowobjectref_workflow_instance_id_assessment_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_id");
-- create index "workflowobjectref_workflow_instance_id_assessment_response_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_response_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_response_id");
-- create index "workflowobjectref_workflow_instance_id_campaign_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_campaign_id" ON "workflow_object_refs" ("workflow_instance_id", "campaign_id");
-- create index "workflowobjectref_workflow_instance_id_campaign_target_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_campaign_target_id" ON "workflow_object_refs" ("workflow_instance_id", "campaign_target_id");
-- create index "workflowobjectref_workflow_instance_id_control_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_control_id" ON "workflow_object_refs" ("workflow_instance_id", "control_id");
-- create index "workflowobjectref_workflow_instance_id_directory_account_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_account_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_account_id");
-- create index "workflowobjectref_workflow_instance_id_directory_group_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_group_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_group_id");
-- create index "workflowobjectref_workflow_instance_id_directory_membership_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_membership_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_membership_id");
-- create index "workflowobjectref_workflow_instance_id_evidence_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_evidence_id" ON "workflow_object_refs" ("workflow_instance_id", "evidence_id");
-- create index "workflowobjectref_workflow_instance_id_finding_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_finding_id" ON "workflow_object_refs" ("workflow_instance_id", "finding_id");
-- create index "workflowobjectref_workflow_instance_id_identity_holder_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_identity_holder_id" ON "workflow_object_refs" ("workflow_instance_id", "identity_holder_id");
-- create index "workflowobjectref_workflow_instance_id_internal_policy_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_internal_policy_id" ON "workflow_object_refs" ("workflow_instance_id", "internal_policy_id");
-- create index "workflowobjectref_workflow_instance_id_platform_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_platform_id" ON "workflow_object_refs" ("workflow_instance_id", "platform_id");
-- create index "workflowobjectref_workflow_instance_id_procedure_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_procedure_id" ON "workflow_object_refs" ("workflow_instance_id", "procedure_id");
-- create index "workflowobjectref_workflow_instance_id_remediation_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_remediation_id" ON "workflow_object_refs" ("workflow_instance_id", "remediation_id");
-- create index "workflowobjectref_workflow_instance_id_risk_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_risk_id" ON "workflow_object_refs" ("workflow_instance_id", "risk_id");
-- create index "workflowobjectref_workflow_instance_id_subcontrol_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_subcontrol_id" ON "workflow_object_refs" ("workflow_instance_id", "subcontrol_id");
-- create index "workflowobjectref_workflow_instance_id_task_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_task_id" ON "workflow_object_refs" ("workflow_instance_id", "task_id");
-- create index "workflowobjectref_workflow_instance_id_vulnerability_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_vulnerability_id" ON "workflow_object_refs" ("workflow_instance_id", "vulnerability_id");
-- create "workflow_proposals" table
CREATE TABLE "workflow_proposals" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "tags" jsonb NULL, "domain_key" character varying NOT NULL, "state" character varying NOT NULL DEFAULT 'DRAFT', "revision" bigint NOT NULL DEFAULT 1, "changes" jsonb NULL, "proposed_changes" jsonb NULL, "proposed_hash" character varying NULL, "approved_hash" character varying NULL, "submitted_at" timestamptz NULL, "owner_id" character varying NULL, "workflow_object_ref_id" character varying NOT NULL, "submitted_by_user_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflow_proposal_owner_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_owner_id_idx" ON "workflow_proposals" ("owner_id");
-- create index "workflow_proposal_submitted_by_user_id_idx" to table: "workflow_proposals"
CREATE INDEX "workflow_proposal_submitted_by_user_id_idx" ON "workflow_proposals" ("submitted_by_user_id");
-- create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY (ARRAY[('DRAFT'::character varying)::text, ('SUBMITTED'::character varying)::text]));
-- create "action_plan_blocked_groups" table
CREATE TABLE "action_plan_blocked_groups" ("action_plan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "group_id"));
-- create index "action_plan_blocked_groups_group_id_idx" to table: "action_plan_blocked_groups"
CREATE INDEX "action_plan_blocked_groups_group_id_idx" ON "action_plan_blocked_groups" ("group_id");
-- create "action_plan_editors" table
CREATE TABLE "action_plan_editors" ("action_plan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "group_id"));
-- create index "action_plan_editors_group_id_idx" to table: "action_plan_editors"
CREATE INDEX "action_plan_editors_group_id_idx" ON "action_plan_editors" ("group_id");
-- create "action_plan_viewers" table
CREATE TABLE "action_plan_viewers" ("action_plan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "group_id"));
-- create index "action_plan_viewers_group_id_idx" to table: "action_plan_viewers"
CREATE INDEX "action_plan_viewers_group_id_idx" ON "action_plan_viewers" ("group_id");
-- create "action_plan_tasks" table
CREATE TABLE "action_plan_tasks" ("action_plan_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "task_id"));
-- create index "action_plan_tasks_task_id_idx" to table: "action_plan_tasks"
CREATE INDEX "action_plan_tasks_task_id_idx" ON "action_plan_tasks" ("task_id");
-- create "asset_connected_assets" table
CREATE TABLE "asset_connected_assets" ("asset_id" character varying NOT NULL, "connected_from_id" character varying NOT NULL, PRIMARY KEY ("asset_id", "connected_from_id"));
-- create index "asset_connected_assets_connected_from_id_idx" to table: "asset_connected_assets"
CREATE INDEX "asset_connected_assets_connected_from_id_idx" ON "asset_connected_assets" ("connected_from_id");
-- create "campaign_blocked_groups" table
CREATE TABLE "campaign_blocked_groups" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- create index "campaign_blocked_groups_group_id_idx" to table: "campaign_blocked_groups"
CREATE INDEX "campaign_blocked_groups_group_id_idx" ON "campaign_blocked_groups" ("group_id");
-- create "campaign_editors" table
CREATE TABLE "campaign_editors" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- create index "campaign_editors_group_id_idx" to table: "campaign_editors"
CREATE INDEX "campaign_editors_group_id_idx" ON "campaign_editors" ("group_id");
-- create "campaign_viewers" table
CREATE TABLE "campaign_viewers" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- create index "campaign_viewers_group_id_idx" to table: "campaign_viewers"
CREATE INDEX "campaign_viewers_group_id_idx" ON "campaign_viewers" ("group_id");
-- create "campaign_contacts" table
CREATE TABLE "campaign_contacts" ("campaign_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "contact_id"));
-- create index "campaign_contacts_contact_id_idx" to table: "campaign_contacts"
CREATE INDEX "campaign_contacts_contact_id_idx" ON "campaign_contacts" ("contact_id");
-- create "campaign_users" table
CREATE TABLE "campaign_users" ("campaign_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "user_id"));
-- create index "campaign_users_user_id_idx" to table: "campaign_users"
CREATE INDEX "campaign_users_user_id_idx" ON "campaign_users" ("user_id");
-- create "campaign_groups" table
CREATE TABLE "campaign_groups" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- create index "campaign_groups_group_id_idx" to table: "campaign_groups"
CREATE INDEX "campaign_groups_group_id_idx" ON "campaign_groups" ("group_id");
-- create "campaign_identity_holders" table
CREATE TABLE "campaign_identity_holders" ("campaign_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "identity_holder_id"));
-- create index "campaign_identity_holders_identity_holder_id_idx" to table: "campaign_identity_holders"
CREATE INDEX "campaign_identity_holders_identity_holder_id_idx" ON "campaign_identity_holders" ("identity_holder_id");
-- create "check_result_controls" table
CREATE TABLE "check_result_controls" ("check_result_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("check_result_id", "control_id"));
-- create index "check_result_controls_control_id_idx" to table: "check_result_controls"
CREATE INDEX "check_result_controls_control_id_idx" ON "check_result_controls" ("control_id");
-- create "contact_files" table
CREATE TABLE "contact_files" ("contact_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "file_id"));
-- create index "contact_files_file_id_idx" to table: "contact_files"
CREATE INDEX "contact_files_file_id_idx" ON "contact_files" ("file_id");
-- create "control_control_objectives" table
CREATE TABLE "control_control_objectives" ("control_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_objective_id"));
-- create index "control_control_objectives_control_objective_id_idx" to table: "control_control_objectives"
CREATE INDEX "control_control_objectives_control_objective_id_idx" ON "control_control_objectives" ("control_objective_id");
-- create "control_tasks" table
CREATE TABLE "control_tasks" ("control_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_id", "task_id"));
-- create index "control_tasks_task_id_idx" to table: "control_tasks"
CREATE INDEX "control_tasks_task_id_idx" ON "control_tasks" ("task_id");
-- create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- create index "control_narratives_narrative_id_idx" to table: "control_narratives"
CREATE INDEX "control_narratives_narrative_id_idx" ON "control_narratives" ("narrative_id");
-- create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- create index "control_risks_risk_id_idx" to table: "control_risks"
CREATE INDEX "control_risks_risk_id_idx" ON "control_risks" ("risk_id");
-- create "control_action_plans" table
CREATE TABLE "control_action_plans" ("control_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "action_plan_id"));
-- create index "control_action_plans_action_plan_id_idx" to table: "control_action_plans"
CREATE INDEX "control_action_plans_action_plan_id_idx" ON "control_action_plans" ("action_plan_id");
-- create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- create index "control_procedures_procedure_id_idx" to table: "control_procedures"
CREATE INDEX "control_procedures_procedure_id_idx" ON "control_procedures" ("procedure_id");
-- create "control_scans" table
CREATE TABLE "control_scans" ("control_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "scan_id"));
-- create index "control_scans_scan_id_idx" to table: "control_scans"
CREATE INDEX "control_scans_scan_id_idx" ON "control_scans" ("scan_id");
-- create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create index "control_blocked_groups_group_id_idx" to table: "control_blocked_groups"
CREATE INDEX "control_blocked_groups_group_id_idx" ON "control_blocked_groups" ("group_id");
-- create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create index "control_editors_group_id_idx" to table: "control_editors"
CREATE INDEX "control_editors_group_id_idx" ON "control_editors" ("group_id");
-- create "control_assets" table
CREATE TABLE "control_assets" ("control_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("control_id", "asset_id"));
-- create index "control_assets_asset_id_idx" to table: "control_assets"
CREATE INDEX "control_assets_asset_id_idx" ON "control_assets" ("asset_id");
-- create "control_entities" table
CREATE TABLE "control_entities" ("control_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("control_id", "entity_id"));
-- create index "control_entities_entity_id_idx" to table: "control_entities"
CREATE INDEX "control_entities_entity_id_idx" ON "control_entities" ("entity_id");
-- create "control_identity_holders" table
CREATE TABLE "control_identity_holders" ("control_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("control_id", "identity_holder_id"));
-- create index "control_identity_holders_identity_holder_id_idx" to table: "control_identity_holders"
CREATE INDEX "control_identity_holders_identity_holder_id_idx" ON "control_identity_holders" ("identity_holder_id");
-- create "control_campaigns" table
CREATE TABLE "control_campaigns" ("control_id" character varying NOT NULL, "campaign_id" character varying NOT NULL, PRIMARY KEY ("control_id", "campaign_id"));
-- create index "control_campaigns_campaign_id_idx" to table: "control_campaigns"
CREATE INDEX "control_campaigns_campaign_id_idx" ON "control_campaigns" ("campaign_id");
-- create "control_control_implementations" table
CREATE TABLE "control_control_implementations" ("control_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_implementation_id"));
-- create index "control_control_implementations_control_implementation_id_idx" to table: "control_control_implementations"
CREATE INDEX "control_control_implementations_control_implementation_id_idx" ON "control_control_implementations" ("control_implementation_id");
-- create "control_implementation_blocked_groups" table
CREATE TABLE "control_implementation_blocked_groups" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"));
-- create index "control_implementation_blocked_groups_group_id_idx" to table: "control_implementation_blocked_groups"
CREATE INDEX "control_implementation_blocked_groups_group_id_idx" ON "control_implementation_blocked_groups" ("group_id");
-- create "control_implementation_editors" table
CREATE TABLE "control_implementation_editors" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"));
-- create index "control_implementation_editors_group_id_idx" to table: "control_implementation_editors"
CREATE INDEX "control_implementation_editors_group_id_idx" ON "control_implementation_editors" ("group_id");
-- create "control_implementation_viewers" table
CREATE TABLE "control_implementation_viewers" ("control_implementation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "group_id"));
-- create index "control_implementation_viewers_group_id_idx" to table: "control_implementation_viewers"
CREATE INDEX "control_implementation_viewers_group_id_idx" ON "control_implementation_viewers" ("group_id");
-- create "control_implementation_tasks" table
CREATE TABLE "control_implementation_tasks" ("control_implementation_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_implementation_id", "task_id"));
-- create index "control_implementation_tasks_task_id_idx" to table: "control_implementation_tasks"
CREATE INDEX "control_implementation_tasks_task_id_idx" ON "control_implementation_tasks" ("task_id");
-- create "control_objective_blocked_groups" table
CREATE TABLE "control_objective_blocked_groups" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create index "control_objective_blocked_groups_group_id_idx" to table: "control_objective_blocked_groups"
CREATE INDEX "control_objective_blocked_groups_group_id_idx" ON "control_objective_blocked_groups" ("group_id");
-- create "control_objective_editors" table
CREATE TABLE "control_objective_editors" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create index "control_objective_editors_group_id_idx" to table: "control_objective_editors"
CREATE INDEX "control_objective_editors_group_id_idx" ON "control_objective_editors" ("group_id");
-- create "control_objective_viewers" table
CREATE TABLE "control_objective_viewers" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create index "control_objective_viewers_group_id_idx" to table: "control_objective_viewers"
CREATE INDEX "control_objective_viewers_group_id_idx" ON "control_objective_viewers" ("group_id");
-- create "control_objective_tasks" table
CREATE TABLE "control_objective_tasks" ("control_objective_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "task_id"));
-- create index "control_objective_tasks_task_id_idx" to table: "control_objective_tasks"
CREATE INDEX "control_objective_tasks_task_id_idx" ON "control_objective_tasks" ("task_id");
-- create "document_data_files" table
CREATE TABLE "document_data_files" ("document_data_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("document_data_id", "file_id"));
-- create index "document_data_files_file_id_idx" to table: "document_data_files"
CREATE INDEX "document_data_files_file_id_idx" ON "document_data_files" ("file_id");
-- create "entity_blocked_groups" table
CREATE TABLE "entity_blocked_groups" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"));
-- create index "entity_blocked_groups_group_id_idx" to table: "entity_blocked_groups"
CREATE INDEX "entity_blocked_groups_group_id_idx" ON "entity_blocked_groups" ("group_id");
-- create "entity_editors" table
CREATE TABLE "entity_editors" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"));
-- create index "entity_editors_group_id_idx" to table: "entity_editors"
CREATE INDEX "entity_editors_group_id_idx" ON "entity_editors" ("group_id");
-- create "entity_contacts" table
CREATE TABLE "entity_contacts" ("entity_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "contact_id"));
-- create index "entity_contacts_contact_id_idx" to table: "entity_contacts"
CREATE INDEX "entity_contacts_contact_id_idx" ON "entity_contacts" ("contact_id");
-- create "entity_documents" table
CREATE TABLE "entity_documents" ("entity_id" character varying NOT NULL, "document_data_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "document_data_id"));
-- create index "entity_documents_document_data_id_idx" to table: "entity_documents"
CREATE INDEX "entity_documents_document_data_id_idx" ON "entity_documents" ("document_data_id");
-- create "entity_files" table
CREATE TABLE "entity_files" ("entity_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "file_id"));
-- create index "entity_files_file_id_idx" to table: "entity_files"
CREATE INDEX "entity_files_file_id_idx" ON "entity_files" ("file_id");
-- create "entity_assets" table
CREATE TABLE "entity_assets" ("entity_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "asset_id"));
-- create index "entity_assets_asset_id_idx" to table: "entity_assets"
CREATE INDEX "entity_assets_asset_id_idx" ON "entity_assets" ("asset_id");
-- create "entity_system_details" table
CREATE TABLE "entity_system_details" ("entity_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "system_detail_id"));
-- create index "entity_system_details_system_detail_id_idx" to table: "entity_system_details"
CREATE INDEX "entity_system_details_system_detail_id_idx" ON "entity_system_details" ("system_detail_id");
-- create "entity_integrations" table
CREATE TABLE "entity_integrations" ("entity_id" character varying NOT NULL, "integration_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "integration_id"));
-- create index "entity_integrations_integration_id_idx" to table: "entity_integrations"
CREATE INDEX "entity_integrations_integration_id_idx" ON "entity_integrations" ("integration_id");
-- create "entity_subprocessors" table
CREATE TABLE "entity_subprocessors" ("entity_id" character varying NOT NULL, "subprocessor_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "subprocessor_id"));
-- create index "entity_subprocessors_subprocessor_id_idx" to table: "entity_subprocessors"
CREATE INDEX "entity_subprocessors_subprocessor_id_idx" ON "entity_subprocessors" ("subprocessor_id");
-- create "evidence_controls" table
CREATE TABLE "evidence_controls" ("evidence_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_id"));
-- create index "evidence_controls_control_id_idx" to table: "evidence_controls"
CREATE INDEX "evidence_controls_control_id_idx" ON "evidence_controls" ("control_id");
-- create "evidence_subcontrols" table
CREATE TABLE "evidence_subcontrols" ("evidence_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "subcontrol_id"));
-- create index "evidence_subcontrols_subcontrol_id_idx" to table: "evidence_subcontrols"
CREATE INDEX "evidence_subcontrols_subcontrol_id_idx" ON "evidence_subcontrols" ("subcontrol_id");
-- create "evidence_control_objectives" table
CREATE TABLE "evidence_control_objectives" ("evidence_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_objective_id"));
-- create index "evidence_control_objectives_control_objective_id_idx" to table: "evidence_control_objectives"
CREATE INDEX "evidence_control_objectives_control_objective_id_idx" ON "evidence_control_objectives" ("control_objective_id");
-- create "evidence_files" table
CREATE TABLE "evidence_files" ("evidence_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "file_id"));
-- create index "evidence_files_file_id_idx" to table: "evidence_files"
CREATE INDEX "evidence_files_file_id_idx" ON "evidence_files" ("file_id");
-- create "file_events" table
CREATE TABLE "file_events" ("file_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("file_id", "event_id"));
-- create index "file_events_event_id_idx" to table: "file_events"
CREATE INDEX "file_events_event_id_idx" ON "file_events" ("event_id");
-- create "file_secrets" table
CREATE TABLE "file_secrets" ("file_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("file_id", "hush_id"));
-- create index "file_secrets_hush_id_idx" to table: "file_secrets"
CREATE INDEX "file_secrets_hush_id_idx" ON "file_secrets" ("hush_id");
-- create "finding_blocked_groups" table
CREATE TABLE "finding_blocked_groups" ("finding_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "group_id"));
-- create index "finding_blocked_groups_group_id_idx" to table: "finding_blocked_groups"
CREATE INDEX "finding_blocked_groups_group_id_idx" ON "finding_blocked_groups" ("group_id");
-- create "finding_editors" table
CREATE TABLE "finding_editors" ("finding_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "group_id"));
-- create index "finding_editors_group_id_idx" to table: "finding_editors"
CREATE INDEX "finding_editors_group_id_idx" ON "finding_editors" ("group_id");
-- create "finding_vulnerabilities" table
CREATE TABLE "finding_vulnerabilities" ("finding_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "vulnerability_id"));
-- create index "finding_vulnerabilities_vulnerability_id_idx" to table: "finding_vulnerabilities"
CREATE INDEX "finding_vulnerabilities_vulnerability_id_idx" ON "finding_vulnerabilities" ("vulnerability_id");
-- create "finding_action_plans" table
CREATE TABLE "finding_action_plans" ("finding_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "action_plan_id"));
-- create index "finding_action_plans_action_plan_id_idx" to table: "finding_action_plans"
CREATE INDEX "finding_action_plans_action_plan_id_idx" ON "finding_action_plans" ("action_plan_id");
-- create "finding_subcontrols" table
CREATE TABLE "finding_subcontrols" ("finding_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "subcontrol_id"));
-- create index "finding_subcontrols_subcontrol_id_idx" to table: "finding_subcontrols"
CREATE INDEX "finding_subcontrols_subcontrol_id_idx" ON "finding_subcontrols" ("subcontrol_id");
-- create "finding_risks" table
CREATE TABLE "finding_risks" ("finding_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "risk_id"));
-- create index "finding_risks_risk_id_idx" to table: "finding_risks"
CREATE INDEX "finding_risks_risk_id_idx" ON "finding_risks" ("risk_id");
-- create "finding_programs" table
CREATE TABLE "finding_programs" ("finding_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "program_id"));
-- create index "finding_programs_program_id_idx" to table: "finding_programs"
CREATE INDEX "finding_programs_program_id_idx" ON "finding_programs" ("program_id");
-- create "finding_assets" table
CREATE TABLE "finding_assets" ("finding_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "asset_id"));
-- create index "finding_assets_asset_id_idx" to table: "finding_assets"
CREATE INDEX "finding_assets_asset_id_idx" ON "finding_assets" ("asset_id");
-- create "finding_entities" table
CREATE TABLE "finding_entities" ("finding_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "entity_id"));
-- create index "finding_entities_entity_id_idx" to table: "finding_entities"
CREATE INDEX "finding_entities_entity_id_idx" ON "finding_entities" ("entity_id");
-- create "finding_scans" table
CREATE TABLE "finding_scans" ("finding_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "scan_id"));
-- create index "finding_scans_scan_id_idx" to table: "finding_scans"
CREATE INDEX "finding_scans_scan_id_idx" ON "finding_scans" ("scan_id");
-- create "finding_tasks" table
CREATE TABLE "finding_tasks" ("finding_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "task_id"));
-- create index "finding_tasks_task_id_idx" to table: "finding_tasks"
CREATE INDEX "finding_tasks_task_id_idx" ON "finding_tasks" ("task_id");
-- create "finding_directory_accounts" table
CREATE TABLE "finding_directory_accounts" ("finding_id" character varying NOT NULL, "directory_account_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "directory_account_id"));
-- create index "finding_directory_accounts_directory_account_id_idx" to table: "finding_directory_accounts"
CREATE INDEX "finding_directory_accounts_directory_account_id_idx" ON "finding_directory_accounts" ("directory_account_id");
-- create "finding_identity_holders" table
CREATE TABLE "finding_identity_holders" ("finding_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "identity_holder_id"));
-- create index "finding_identity_holders_identity_holder_id_idx" to table: "finding_identity_holders"
CREATE INDEX "finding_identity_holders_identity_holder_id_idx" ON "finding_identity_holders" ("identity_holder_id");
-- create "finding_check_results" table
CREATE TABLE "finding_check_results" ("finding_id" character varying NOT NULL, "check_result_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "check_result_id"));
-- create index "finding_check_results_check_result_id_idx" to table: "finding_check_results"
CREATE INDEX "finding_check_results_check_result_id_idx" ON "finding_check_results" ("check_result_id");
-- create "group_events" table
CREATE TABLE "group_events" ("group_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_id", "event_id"));
-- create index "group_events_event_id_idx" to table: "group_events"
CREATE INDEX "group_events_event_id_idx" ON "group_events" ("event_id");
-- create "group_files" table
CREATE TABLE "group_files" ("group_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("group_id", "file_id"));
-- create index "group_files_file_id_idx" to table: "group_files"
CREATE INDEX "group_files_file_id_idx" ON "group_files" ("file_id");
-- create "group_tasks" table
CREATE TABLE "group_tasks" ("group_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("group_id", "task_id"));
-- create index "group_tasks_task_id_idx" to table: "group_tasks"
CREATE INDEX "group_tasks_task_id_idx" ON "group_tasks" ("task_id");
-- create "group_membership_events" table
CREATE TABLE "group_membership_events" ("group_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_membership_id", "event_id"));
-- create index "group_membership_events_event_id_idx" to table: "group_membership_events"
CREATE INDEX "group_membership_events_event_id_idx" ON "group_membership_events" ("event_id");
-- create "hush_events" table
CREATE TABLE "hush_events" ("hush_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("hush_id", "event_id"));
-- create index "hush_events_event_id_idx" to table: "hush_events"
CREATE INDEX "hush_events_event_id_idx" ON "hush_events" ("event_id");
-- create "identity_holder_assessments" table
CREATE TABLE "identity_holder_assessments" ("identity_holder_id" character varying NOT NULL, "assessment_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "assessment_id"));
-- create index "identity_holder_assessments_assessment_id_idx" to table: "identity_holder_assessments"
CREATE INDEX "identity_holder_assessments_assessment_id_idx" ON "identity_holder_assessments" ("assessment_id");
-- create "identity_holder_templates" table
CREATE TABLE "identity_holder_templates" ("identity_holder_id" character varying NOT NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "template_id"));
-- create index "identity_holder_templates_template_id_idx" to table: "identity_holder_templates"
CREATE INDEX "identity_holder_templates_template_id_idx" ON "identity_holder_templates" ("template_id");
-- create "identity_holder_assets" table
CREATE TABLE "identity_holder_assets" ("identity_holder_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "asset_id"));
-- create index "identity_holder_assets_asset_id_idx" to table: "identity_holder_assets"
CREATE INDEX "identity_holder_assets_asset_id_idx" ON "identity_holder_assets" ("asset_id");
-- create "identity_holder_entities" table
CREATE TABLE "identity_holder_entities" ("identity_holder_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "entity_id"));
-- create index "identity_holder_entities_entity_id_idx" to table: "identity_holder_entities"
CREATE INDEX "identity_holder_entities_entity_id_idx" ON "identity_holder_entities" ("entity_id");
-- create "identity_holder_tasks" table
CREATE TABLE "identity_holder_tasks" ("identity_holder_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "task_id"));
-- create index "identity_holder_tasks_task_id_idx" to table: "identity_holder_tasks"
CREATE INDEX "identity_holder_tasks_task_id_idx" ON "identity_holder_tasks" ("task_id");
-- create "identity_holder_files" table
CREATE TABLE "identity_holder_files" ("identity_holder_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "file_id"));
-- create index "identity_holder_files_file_id_idx" to table: "identity_holder_files"
CREATE INDEX "identity_holder_files_file_id_idx" ON "identity_holder_files" ("file_id");
-- create "integration_secrets" table
CREATE TABLE "integration_secrets" ("integration_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "hush_id"));
-- create index "integration_secrets_hush_id_idx" to table: "integration_secrets"
CREATE INDEX "integration_secrets_hush_id_idx" ON "integration_secrets" ("hush_id");
-- create "integration_events" table
CREATE TABLE "integration_events" ("integration_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "event_id"));
-- create index "integration_events_event_id_idx" to table: "integration_events"
CREATE INDEX "integration_events_event_id_idx" ON "integration_events" ("event_id");
-- create "integration_findings" table
CREATE TABLE "integration_findings" ("integration_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "finding_id"));
-- create index "integration_findings_finding_id_idx" to table: "integration_findings"
CREATE INDEX "integration_findings_finding_id_idx" ON "integration_findings" ("finding_id");
-- create "integration_vulnerabilities" table
CREATE TABLE "integration_vulnerabilities" ("integration_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "vulnerability_id"));
-- create index "integration_vulnerabilities_vulnerability_id_idx" to table: "integration_vulnerabilities"
CREATE INDEX "integration_vulnerabilities_vulnerability_id_idx" ON "integration_vulnerabilities" ("vulnerability_id");
-- create "integration_internal_policies" table
CREATE TABLE "integration_internal_policies" ("integration_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "internal_policy_id"));
-- create index "integration_internal_policies_internal_policy_id_idx" to table: "integration_internal_policies"
CREATE INDEX "integration_internal_policies_internal_policy_id_idx" ON "integration_internal_policies" ("internal_policy_id");
-- create "integration_reviews" table
CREATE TABLE "integration_reviews" ("integration_id" character varying NOT NULL, "review_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "review_id"));
-- create index "integration_reviews_review_id_idx" to table: "integration_reviews"
CREATE INDEX "integration_reviews_review_id_idx" ON "integration_reviews" ("review_id");
-- create "integration_remediations" table
CREATE TABLE "integration_remediations" ("integration_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "remediation_id"));
-- create index "integration_remediations_remediation_id_idx" to table: "integration_remediations"
CREATE INDEX "integration_remediations_remediation_id_idx" ON "integration_remediations" ("remediation_id");
-- create "integration_action_plans" table
CREATE TABLE "integration_action_plans" ("integration_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "action_plan_id"));
-- create index "integration_action_plans_action_plan_id_idx" to table: "integration_action_plans"
CREATE INDEX "integration_action_plans_action_plan_id_idx" ON "integration_action_plans" ("action_plan_id");
-- create "internal_policy_blocked_groups" table
CREATE TABLE "internal_policy_blocked_groups" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- create index "internal_policy_blocked_groups_group_id_idx" to table: "internal_policy_blocked_groups"
CREATE INDEX "internal_policy_blocked_groups_group_id_idx" ON "internal_policy_blocked_groups" ("group_id");
-- create "internal_policy_editors" table
CREATE TABLE "internal_policy_editors" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- create index "internal_policy_editors_group_id_idx" to table: "internal_policy_editors"
CREATE INDEX "internal_policy_editors_group_id_idx" ON "internal_policy_editors" ("group_id");
-- create "internal_policy_control_objectives" table
CREATE TABLE "internal_policy_control_objectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- create index "internal_policy_control_objectives_control_objective_id_idx" to table: "internal_policy_control_objectives"
CREATE INDEX "internal_policy_control_objectives_control_objective_id_idx" ON "internal_policy_control_objectives" ("control_objective_id");
-- create "internal_policy_controls" table
CREATE TABLE "internal_policy_controls" ("internal_policy_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_id"));
-- create index "internal_policy_controls_control_id_idx" to table: "internal_policy_controls"
CREATE INDEX "internal_policy_controls_control_id_idx" ON "internal_policy_controls" ("control_id");
-- create "internal_policy_subcontrols" table
CREATE TABLE "internal_policy_subcontrols" ("internal_policy_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "subcontrol_id"));
-- create index "internal_policy_subcontrols_subcontrol_id_idx" to table: "internal_policy_subcontrols"
CREATE INDEX "internal_policy_subcontrols_subcontrol_id_idx" ON "internal_policy_subcontrols" ("subcontrol_id");
-- create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- create index "internal_policy_procedures_procedure_id_idx" to table: "internal_policy_procedures"
CREATE INDEX "internal_policy_procedures_procedure_id_idx" ON "internal_policy_procedures" ("procedure_id");
-- create "internal_policy_narratives" table
CREATE TABLE "internal_policy_narratives" ("internal_policy_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "narrative_id"));
-- create index "internal_policy_narratives_narrative_id_idx" to table: "internal_policy_narratives"
CREATE INDEX "internal_policy_narratives_narrative_id_idx" ON "internal_policy_narratives" ("narrative_id");
-- create "internal_policy_tasks" table
CREATE TABLE "internal_policy_tasks" ("internal_policy_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "task_id"));
-- create index "internal_policy_tasks_task_id_idx" to table: "internal_policy_tasks"
CREATE INDEX "internal_policy_tasks_task_id_idx" ON "internal_policy_tasks" ("task_id");
-- create "internal_policy_risks" table
CREATE TABLE "internal_policy_risks" ("internal_policy_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "risk_id"));
-- create index "internal_policy_risks_risk_id_idx" to table: "internal_policy_risks"
CREATE INDEX "internal_policy_risks_risk_id_idx" ON "internal_policy_risks" ("risk_id");
-- create "internal_policy_assets" table
CREATE TABLE "internal_policy_assets" ("internal_policy_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "asset_id"));
-- create index "internal_policy_assets_asset_id_idx" to table: "internal_policy_assets"
CREATE INDEX "internal_policy_assets_asset_id_idx" ON "internal_policy_assets" ("asset_id");
-- create "internal_policy_entities" table
CREATE TABLE "internal_policy_entities" ("internal_policy_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "entity_id"));
-- create index "internal_policy_entities_entity_id_idx" to table: "internal_policy_entities"
CREATE INDEX "internal_policy_entities_entity_id_idx" ON "internal_policy_entities" ("entity_id");
-- create "internal_policy_identity_holders" table
CREATE TABLE "internal_policy_identity_holders" ("internal_policy_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "identity_holder_id"));
-- create index "internal_policy_identity_holders_identity_holder_id_idx" to table: "internal_policy_identity_holders"
CREATE INDEX "internal_policy_identity_holders_identity_holder_id_idx" ON "internal_policy_identity_holders" ("identity_holder_id");
-- create "invite_events" table
CREATE TABLE "invite_events" ("invite_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "event_id"));
-- create index "invite_events_event_id_idx" to table: "invite_events"
CREATE INDEX "invite_events_event_id_idx" ON "invite_events" ("event_id");
-- create "invite_groups" table
CREATE TABLE "invite_groups" ("invite_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "group_id"));
-- create index "invite_groups_group_id_idx" to table: "invite_groups"
CREATE INDEX "invite_groups_group_id_idx" ON "invite_groups" ("group_id");
-- create "job_runner_job_runner_tokens" table
CREATE TABLE "job_runner_job_runner_tokens" ("job_runner_id" character varying NOT NULL, "job_runner_token_id" character varying NOT NULL, PRIMARY KEY ("job_runner_id", "job_runner_token_id"));
-- create index "job_runner_job_runner_tokens_job_runner_token_id_idx" to table: "job_runner_job_runner_tokens"
CREATE INDEX "job_runner_job_runner_tokens_job_runner_token_id_idx" ON "job_runner_job_runner_tokens" ("job_runner_token_id");
-- create "mapped_control_blocked_groups" table
CREATE TABLE "mapped_control_blocked_groups" ("mapped_control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "group_id"));
-- create index "mapped_control_blocked_groups_group_id_idx" to table: "mapped_control_blocked_groups"
CREATE INDEX "mapped_control_blocked_groups_group_id_idx" ON "mapped_control_blocked_groups" ("group_id");
-- create "mapped_control_editors" table
CREATE TABLE "mapped_control_editors" ("mapped_control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "group_id"));
-- create index "mapped_control_editors_group_id_idx" to table: "mapped_control_editors"
CREATE INDEX "mapped_control_editors_group_id_idx" ON "mapped_control_editors" ("group_id");
-- create "mapped_control_from_controls" table
CREATE TABLE "mapped_control_from_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"));
-- create index "mapped_control_from_controls_control_id_idx" to table: "mapped_control_from_controls"
CREATE INDEX "mapped_control_from_controls_control_id_idx" ON "mapped_control_from_controls" ("control_id");
-- create "mapped_control_to_controls" table
CREATE TABLE "mapped_control_to_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"));
-- create index "mapped_control_to_controls_control_id_idx" to table: "mapped_control_to_controls"
CREATE INDEX "mapped_control_to_controls_control_id_idx" ON "mapped_control_to_controls" ("control_id");
-- create "mapped_control_from_subcontrols" table
CREATE TABLE "mapped_control_from_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"));
-- create index "mapped_control_from_subcontrols_subcontrol_id_idx" to table: "mapped_control_from_subcontrols"
CREATE INDEX "mapped_control_from_subcontrols_subcontrol_id_idx" ON "mapped_control_from_subcontrols" ("subcontrol_id");
-- create "mapped_control_to_subcontrols" table
CREATE TABLE "mapped_control_to_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"));
-- create index "mapped_control_to_subcontrols_subcontrol_id_idx" to table: "mapped_control_to_subcontrols"
CREATE INDEX "mapped_control_to_subcontrols_subcontrol_id_idx" ON "mapped_control_to_subcontrols" ("subcontrol_id");
-- create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create index "narrative_blocked_groups_group_id_idx" to table: "narrative_blocked_groups"
CREATE INDEX "narrative_blocked_groups_group_id_idx" ON "narrative_blocked_groups" ("group_id");
-- create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create index "narrative_editors_group_id_idx" to table: "narrative_editors"
CREATE INDEX "narrative_editors_group_id_idx" ON "narrative_editors" ("group_id");
-- create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create index "narrative_viewers_group_id_idx" to table: "narrative_viewers"
CREATE INDEX "narrative_viewers_group_id_idx" ON "narrative_viewers" ("group_id");
-- create "org_membership_events" table
CREATE TABLE "org_membership_events" ("org_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_membership_id", "event_id"));
-- create index "org_membership_events_event_id_idx" to table: "org_membership_events"
CREATE INDEX "org_membership_events_event_id_idx" ON "org_membership_events" ("event_id");
-- create "org_module_org_prices" table
CREATE TABLE "org_module_org_prices" ("org_module_id" character varying NOT NULL, "org_price_id" character varying NOT NULL, PRIMARY KEY ("org_module_id", "org_price_id"));
-- create index "org_module_org_prices_org_price_id_idx" to table: "org_module_org_prices"
CREATE INDEX "org_module_org_prices_org_price_id_idx" ON "org_module_org_prices" ("org_price_id");
-- create "org_product_org_prices" table
CREATE TABLE "org_product_org_prices" ("org_product_id" character varying NOT NULL, "org_price_id" character varying NOT NULL, PRIMARY KEY ("org_product_id", "org_price_id"));
-- create index "org_product_org_prices_org_price_id_idx" to table: "org_product_org_prices"
CREATE INDEX "org_product_org_prices_org_price_id_idx" ON "org_product_org_prices" ("org_price_id");
-- create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"));
-- create index "org_subscription_events_event_id_idx" to table: "org_subscription_events"
CREATE INDEX "org_subscription_events_event_id_idx" ON "org_subscription_events" ("event_id");
-- create "organization_personal_access_tokens" table
CREATE TABLE "organization_personal_access_tokens" ("organization_id" character varying NOT NULL, "personal_access_token_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "personal_access_token_id"));
-- create index "organization_personal_access_tokens_personal_access_token_id_id" to table: "organization_personal_access_tokens"
CREATE INDEX "organization_personal_access_tokens_personal_access_token_id_id" ON "organization_personal_access_tokens" ("personal_access_token_id");
-- create "organization_files" table
CREATE TABLE "organization_files" ("organization_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "file_id"));
-- create index "organization_files_file_id_idx" to table: "organization_files"
CREATE INDEX "organization_files_file_id_idx" ON "organization_files" ("file_id");
-- create "organization_events" table
CREATE TABLE "organization_events" ("organization_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "event_id"));
-- create index "organization_events_event_id_idx" to table: "organization_events"
CREATE INDEX "organization_events_event_id_idx" ON "organization_events" ("event_id");
-- create "organization_setting_files" table
CREATE TABLE "organization_setting_files" ("organization_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_setting_id", "file_id"));
-- create index "organization_setting_files_file_id_idx" to table: "organization_setting_files"
CREATE INDEX "organization_setting_files_file_id_idx" ON "organization_setting_files" ("file_id");
-- create "personal_access_token_events" table
CREATE TABLE "personal_access_token_events" ("personal_access_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("personal_access_token_id", "event_id"));
-- create index "personal_access_token_events_event_id_idx" to table: "personal_access_token_events"
CREATE INDEX "personal_access_token_events_event_id_idx" ON "personal_access_token_events" ("event_id");
-- create "platform_blocked_groups" table
CREATE TABLE "platform_blocked_groups" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- create index "platform_blocked_groups_group_id_idx" to table: "platform_blocked_groups"
CREATE INDEX "platform_blocked_groups_group_id_idx" ON "platform_blocked_groups" ("group_id");
-- create "platform_editors" table
CREATE TABLE "platform_editors" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- create index "platform_editors_group_id_idx" to table: "platform_editors"
CREATE INDEX "platform_editors_group_id_idx" ON "platform_editors" ("group_id");
-- create "platform_viewers" table
CREATE TABLE "platform_viewers" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- create index "platform_viewers_group_id_idx" to table: "platform_viewers"
CREATE INDEX "platform_viewers_group_id_idx" ON "platform_viewers" ("group_id");
-- create "platform_assets" table
CREATE TABLE "platform_assets" ("platform_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "asset_id"));
-- create index "platform_assets_asset_id_idx" to table: "platform_assets"
CREATE INDEX "platform_assets_asset_id_idx" ON "platform_assets" ("asset_id");
-- create "platform_entities" table
CREATE TABLE "platform_entities" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- create index "platform_entities_entity_id_idx" to table: "platform_entities"
CREATE INDEX "platform_entities_entity_id_idx" ON "platform_entities" ("entity_id");
-- create "platform_evidence" table
CREATE TABLE "platform_evidence" ("platform_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "evidence_id"));
-- create index "platform_evidence_evidence_id_idx" to table: "platform_evidence"
CREATE INDEX "platform_evidence_evidence_id_idx" ON "platform_evidence" ("evidence_id");
-- create "platform_files" table
CREATE TABLE "platform_files" ("platform_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "file_id"));
-- create index "platform_files_file_id_idx" to table: "platform_files"
CREATE INDEX "platform_files_file_id_idx" ON "platform_files" ("file_id");
-- create "platform_risks" table
CREATE TABLE "platform_risks" ("platform_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "risk_id"));
-- create index "platform_risks_risk_id_idx" to table: "platform_risks"
CREATE INDEX "platform_risks_risk_id_idx" ON "platform_risks" ("risk_id");
-- create "platform_controls" table
CREATE TABLE "platform_controls" ("platform_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "control_id"));
-- create index "platform_controls_control_id_idx" to table: "platform_controls"
CREATE INDEX "platform_controls_control_id_idx" ON "platform_controls" ("control_id");
-- create "platform_assessments" table
CREATE TABLE "platform_assessments" ("platform_id" character varying NOT NULL, "assessment_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "assessment_id"));
-- create index "platform_assessments_assessment_id_idx" to table: "platform_assessments"
CREATE INDEX "platform_assessments_assessment_id_idx" ON "platform_assessments" ("assessment_id");
-- create "platform_scans" table
CREATE TABLE "platform_scans" ("platform_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "scan_id"));
-- create index "platform_scans_scan_id_idx" to table: "platform_scans"
CREATE INDEX "platform_scans_scan_id_idx" ON "platform_scans" ("scan_id");
-- create "platform_tasks" table
CREATE TABLE "platform_tasks" ("platform_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "task_id"));
-- create index "platform_tasks_task_id_idx" to table: "platform_tasks"
CREATE INDEX "platform_tasks_task_id_idx" ON "platform_tasks" ("task_id");
-- create "platform_identity_holders" table
CREATE TABLE "platform_identity_holders" ("platform_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "identity_holder_id"));
-- create index "platform_identity_holders_identity_holder_id_idx" to table: "platform_identity_holders"
CREATE INDEX "platform_identity_holders_identity_holder_id_idx" ON "platform_identity_holders" ("identity_holder_id");
-- create "platform_source_entities" table
CREATE TABLE "platform_source_entities" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- create index "platform_source_entities_entity_id_idx" to table: "platform_source_entities"
CREATE INDEX "platform_source_entities_entity_id_idx" ON "platform_source_entities" ("entity_id");
-- create "platform_out_of_scope_assets" table
CREATE TABLE "platform_out_of_scope_assets" ("platform_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "asset_id"));
-- create index "platform_out_of_scope_assets_asset_id_idx" to table: "platform_out_of_scope_assets"
CREATE INDEX "platform_out_of_scope_assets_asset_id_idx" ON "platform_out_of_scope_assets" ("asset_id");
-- create "platform_out_of_scope_vendors" table
CREATE TABLE "platform_out_of_scope_vendors" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- create index "platform_out_of_scope_vendors_entity_id_idx" to table: "platform_out_of_scope_vendors"
CREATE INDEX "platform_out_of_scope_vendors_entity_id_idx" ON "platform_out_of_scope_vendors" ("entity_id");
-- create "platform_applicable_frameworks" table
CREATE TABLE "platform_applicable_frameworks" ("platform_id" character varying NOT NULL, "standard_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "standard_id"));
-- create index "platform_applicable_frameworks_standard_id_idx" to table: "platform_applicable_frameworks"
CREATE INDEX "platform_applicable_frameworks_standard_id_idx" ON "platform_applicable_frameworks" ("standard_id");
-- create "platform_system_details" table
CREATE TABLE "platform_system_details" ("platform_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "system_detail_id"));
-- create index "platform_system_details_system_detail_id_idx" to table: "platform_system_details"
CREATE INDEX "platform_system_details_system_detail_id_idx" ON "platform_system_details" ("system_detail_id");
-- create "procedure_blocked_groups" table
CREATE TABLE "procedure_blocked_groups" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- create index "procedure_blocked_groups_group_id_idx" to table: "procedure_blocked_groups"
CREATE INDEX "procedure_blocked_groups_group_id_idx" ON "procedure_blocked_groups" ("group_id");
-- create "procedure_editors" table
CREATE TABLE "procedure_editors" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- create index "procedure_editors_group_id_idx" to table: "procedure_editors"
CREATE INDEX "procedure_editors_group_id_idx" ON "procedure_editors" ("group_id");
-- create "procedure_narratives" table
CREATE TABLE "procedure_narratives" ("procedure_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "narrative_id"));
-- create index "procedure_narratives_narrative_id_idx" to table: "procedure_narratives"
CREATE INDEX "procedure_narratives_narrative_id_idx" ON "procedure_narratives" ("narrative_id");
-- create "procedure_risks" table
CREATE TABLE "procedure_risks" ("procedure_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "risk_id"));
-- create index "procedure_risks_risk_id_idx" to table: "procedure_risks"
CREATE INDEX "procedure_risks_risk_id_idx" ON "procedure_risks" ("risk_id");
-- create "procedure_tasks" table
CREATE TABLE "procedure_tasks" ("procedure_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "task_id"));
-- create index "procedure_tasks_task_id_idx" to table: "procedure_tasks"
CREATE INDEX "procedure_tasks_task_id_idx" ON "procedure_tasks" ("task_id");
-- create "program_blocked_groups" table
CREATE TABLE "program_blocked_groups" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- create index "program_blocked_groups_group_id_idx" to table: "program_blocked_groups"
CREATE INDEX "program_blocked_groups_group_id_idx" ON "program_blocked_groups" ("group_id");
-- create "program_editors" table
CREATE TABLE "program_editors" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- create index "program_editors_group_id_idx" to table: "program_editors"
CREATE INDEX "program_editors_group_id_idx" ON "program_editors" ("group_id");
-- create "program_viewers" table
CREATE TABLE "program_viewers" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- create index "program_viewers_group_id_idx" to table: "program_viewers"
CREATE INDEX "program_viewers_group_id_idx" ON "program_viewers" ("group_id");
-- create "program_controls" table
CREATE TABLE "program_controls" ("program_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_id"));
-- create index "program_controls_control_id_idx" to table: "program_controls"
CREATE INDEX "program_controls_control_id_idx" ON "program_controls" ("control_id");
-- create "program_control_objectives" table
CREATE TABLE "program_control_objectives" ("program_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_objective_id"));
-- create index "program_control_objectives_control_objective_id_idx" to table: "program_control_objectives"
CREATE INDEX "program_control_objectives_control_objective_id_idx" ON "program_control_objectives" ("control_objective_id");
-- create "program_internal_policies" table
CREATE TABLE "program_internal_policies" ("program_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("program_id", "internal_policy_id"));
-- create index "program_internal_policies_internal_policy_id_idx" to table: "program_internal_policies"
CREATE INDEX "program_internal_policies_internal_policy_id_idx" ON "program_internal_policies" ("internal_policy_id");
-- create "program_procedures" table
CREATE TABLE "program_procedures" ("program_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("program_id", "procedure_id"));
-- create index "program_procedures_procedure_id_idx" to table: "program_procedures"
CREATE INDEX "program_procedures_procedure_id_idx" ON "program_procedures" ("procedure_id");
-- create "program_risks" table
CREATE TABLE "program_risks" ("program_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("program_id", "risk_id"));
-- create index "program_risks_risk_id_idx" to table: "program_risks"
CREATE INDEX "program_risks_risk_id_idx" ON "program_risks" ("risk_id");
-- create "program_tasks" table
CREATE TABLE "program_tasks" ("program_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("program_id", "task_id"));
-- create index "program_tasks_task_id_idx" to table: "program_tasks"
CREATE INDEX "program_tasks_task_id_idx" ON "program_tasks" ("task_id");
-- create "program_files" table
CREATE TABLE "program_files" ("program_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("program_id", "file_id"));
-- create index "program_files_file_id_idx" to table: "program_files"
CREATE INDEX "program_files_file_id_idx" ON "program_files" ("file_id");
-- create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"));
-- create index "program_evidence_evidence_id_idx" to table: "program_evidence"
CREATE INDEX "program_evidence_evidence_id_idx" ON "program_evidence" ("evidence_id");
-- create "program_narratives" table
CREATE TABLE "program_narratives" ("program_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("program_id", "narrative_id"));
-- create index "program_narratives_narrative_id_idx" to table: "program_narratives"
CREATE INDEX "program_narratives_narrative_id_idx" ON "program_narratives" ("narrative_id");
-- create "program_action_plans" table
CREATE TABLE "program_action_plans" ("program_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("program_id", "action_plan_id"));
-- create index "program_action_plans_action_plan_id_idx" to table: "program_action_plans"
CREATE INDEX "program_action_plans_action_plan_id_idx" ON "program_action_plans" ("action_plan_id");
-- create "program_system_details" table
CREATE TABLE "program_system_details" ("program_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("program_id", "system_detail_id"));
-- create index "program_system_details_system_detail_id_idx" to table: "program_system_details"
CREATE INDEX "program_system_details_system_detail_id_idx" ON "program_system_details" ("system_detail_id");
-- create "remediation_blocked_groups" table
CREATE TABLE "remediation_blocked_groups" ("remediation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "group_id"));
-- create index "remediation_blocked_groups_group_id_idx" to table: "remediation_blocked_groups"
CREATE INDEX "remediation_blocked_groups_group_id_idx" ON "remediation_blocked_groups" ("group_id");
-- create "remediation_editors" table
CREATE TABLE "remediation_editors" ("remediation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "group_id"));
-- create index "remediation_editors_group_id_idx" to table: "remediation_editors"
CREATE INDEX "remediation_editors_group_id_idx" ON "remediation_editors" ("group_id");
-- create "remediation_findings" table
CREATE TABLE "remediation_findings" ("remediation_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "finding_id"));
-- create index "remediation_findings_finding_id_idx" to table: "remediation_findings"
CREATE INDEX "remediation_findings_finding_id_idx" ON "remediation_findings" ("finding_id");
-- create "remediation_vulnerabilities" table
CREATE TABLE "remediation_vulnerabilities" ("remediation_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "vulnerability_id"));
-- create index "remediation_vulnerabilities_vulnerability_id_idx" to table: "remediation_vulnerabilities"
CREATE INDEX "remediation_vulnerabilities_vulnerability_id_idx" ON "remediation_vulnerabilities" ("vulnerability_id");
-- create "remediation_action_plans" table
CREATE TABLE "remediation_action_plans" ("remediation_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "action_plan_id"));
-- create index "remediation_action_plans_action_plan_id_idx" to table: "remediation_action_plans"
CREATE INDEX "remediation_action_plans_action_plan_id_idx" ON "remediation_action_plans" ("action_plan_id");
-- create "remediation_controls" table
CREATE TABLE "remediation_controls" ("remediation_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "control_id"));
-- create index "remediation_controls_control_id_idx" to table: "remediation_controls"
CREATE INDEX "remediation_controls_control_id_idx" ON "remediation_controls" ("control_id");
-- create "remediation_subcontrols" table
CREATE TABLE "remediation_subcontrols" ("remediation_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "subcontrol_id"));
-- create index "remediation_subcontrols_subcontrol_id_idx" to table: "remediation_subcontrols"
CREATE INDEX "remediation_subcontrols_subcontrol_id_idx" ON "remediation_subcontrols" ("subcontrol_id");
-- create "remediation_risks" table
CREATE TABLE "remediation_risks" ("remediation_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "risk_id"));
-- create index "remediation_risks_risk_id_idx" to table: "remediation_risks"
CREATE INDEX "remediation_risks_risk_id_idx" ON "remediation_risks" ("risk_id");
-- create "remediation_programs" table
CREATE TABLE "remediation_programs" ("remediation_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "program_id"));
-- create index "remediation_programs_program_id_idx" to table: "remediation_programs"
CREATE INDEX "remediation_programs_program_id_idx" ON "remediation_programs" ("program_id");
-- create "remediation_assets" table
CREATE TABLE "remediation_assets" ("remediation_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "asset_id"));
-- create index "remediation_assets_asset_id_idx" to table: "remediation_assets"
CREATE INDEX "remediation_assets_asset_id_idx" ON "remediation_assets" ("asset_id");
-- create "remediation_entities" table
CREATE TABLE "remediation_entities" ("remediation_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "entity_id"));
-- create index "remediation_entities_entity_id_idx" to table: "remediation_entities"
CREATE INDEX "remediation_entities_entity_id_idx" ON "remediation_entities" ("entity_id");
-- create "review_blocked_groups" table
CREATE TABLE "review_blocked_groups" ("review_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("review_id", "group_id"));
-- create index "review_blocked_groups_group_id_idx" to table: "review_blocked_groups"
CREATE INDEX "review_blocked_groups_group_id_idx" ON "review_blocked_groups" ("group_id");
-- create "review_editors" table
CREATE TABLE "review_editors" ("review_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("review_id", "group_id"));
-- create index "review_editors_group_id_idx" to table: "review_editors"
CREATE INDEX "review_editors_group_id_idx" ON "review_editors" ("group_id");
-- create "review_findings" table
CREATE TABLE "review_findings" ("review_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("review_id", "finding_id"));
-- create index "review_findings_finding_id_idx" to table: "review_findings"
CREATE INDEX "review_findings_finding_id_idx" ON "review_findings" ("finding_id");
-- create "review_vulnerabilities" table
CREATE TABLE "review_vulnerabilities" ("review_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("review_id", "vulnerability_id"));
-- create index "review_vulnerabilities_vulnerability_id_idx" to table: "review_vulnerabilities"
CREATE INDEX "review_vulnerabilities_vulnerability_id_idx" ON "review_vulnerabilities" ("vulnerability_id");
-- create "review_action_plans" table
CREATE TABLE "review_action_plans" ("review_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("review_id", "action_plan_id"));
-- create index "review_action_plans_action_plan_id_idx" to table: "review_action_plans"
CREATE INDEX "review_action_plans_action_plan_id_idx" ON "review_action_plans" ("action_plan_id");
-- create "review_remediations" table
CREATE TABLE "review_remediations" ("review_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("review_id", "remediation_id"));
-- create index "review_remediations_remediation_id_idx" to table: "review_remediations"
CREATE INDEX "review_remediations_remediation_id_idx" ON "review_remediations" ("remediation_id");
-- create "review_controls" table
CREATE TABLE "review_controls" ("review_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("review_id", "control_id"));
-- create index "review_controls_control_id_idx" to table: "review_controls"
CREATE INDEX "review_controls_control_id_idx" ON "review_controls" ("control_id");
-- create "review_subcontrols" table
CREATE TABLE "review_subcontrols" ("review_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("review_id", "subcontrol_id"));
-- create index "review_subcontrols_subcontrol_id_idx" to table: "review_subcontrols"
CREATE INDEX "review_subcontrols_subcontrol_id_idx" ON "review_subcontrols" ("subcontrol_id");
-- create "review_risks" table
CREATE TABLE "review_risks" ("review_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("review_id", "risk_id"));
-- create index "review_risks_risk_id_idx" to table: "review_risks"
CREATE INDEX "review_risks_risk_id_idx" ON "review_risks" ("risk_id");
-- create "review_programs" table
CREATE TABLE "review_programs" ("review_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("review_id", "program_id"));
-- create index "review_programs_program_id_idx" to table: "review_programs"
CREATE INDEX "review_programs_program_id_idx" ON "review_programs" ("program_id");
-- create "review_assets" table
CREATE TABLE "review_assets" ("review_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("review_id", "asset_id"));
-- create index "review_assets_asset_id_idx" to table: "review_assets"
CREATE INDEX "review_assets_asset_id_idx" ON "review_assets" ("asset_id");
-- create "review_entities" table
CREATE TABLE "review_entities" ("review_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("review_id", "entity_id"));
-- create index "review_entities_entity_id_idx" to table: "review_entities"
CREATE INDEX "review_entities_entity_id_idx" ON "review_entities" ("entity_id");
-- create "review_internal_policies" table
CREATE TABLE "review_internal_policies" ("review_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("review_id", "internal_policy_id"));
-- create index "review_internal_policies_internal_policy_id_idx" to table: "review_internal_policies"
CREATE INDEX "review_internal_policies_internal_policy_id_idx" ON "review_internal_policies" ("internal_policy_id");
-- create "risk_blocked_groups" table
CREATE TABLE "risk_blocked_groups" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create index "risk_blocked_groups_group_id_idx" to table: "risk_blocked_groups"
CREATE INDEX "risk_blocked_groups_group_id_idx" ON "risk_blocked_groups" ("group_id");
-- create "risk_editors" table
CREATE TABLE "risk_editors" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create index "risk_editors_group_id_idx" to table: "risk_editors"
CREATE INDEX "risk_editors_group_id_idx" ON "risk_editors" ("group_id");
-- create "risk_viewers" table
CREATE TABLE "risk_viewers" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create index "risk_viewers_group_id_idx" to table: "risk_viewers"
CREATE INDEX "risk_viewers_group_id_idx" ON "risk_viewers" ("group_id");
-- create "risk_action_plans" table
CREATE TABLE "risk_action_plans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- create index "risk_action_plans_action_plan_id_idx" to table: "risk_action_plans"
CREATE INDEX "risk_action_plans_action_plan_id_idx" ON "risk_action_plans" ("action_plan_id");
-- create "risk_tasks" table
CREATE TABLE "risk_tasks" ("risk_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "task_id"));
-- create index "risk_tasks_task_id_idx" to table: "risk_tasks"
CREATE INDEX "risk_tasks_task_id_idx" ON "risk_tasks" ("task_id");
-- create "scan_blocked_groups" table
CREATE TABLE "scan_blocked_groups" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- create index "scan_blocked_groups_group_id_idx" to table: "scan_blocked_groups"
CREATE INDEX "scan_blocked_groups_group_id_idx" ON "scan_blocked_groups" ("group_id");
-- create "scan_editors" table
CREATE TABLE "scan_editors" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- create index "scan_editors_group_id_idx" to table: "scan_editors"
CREATE INDEX "scan_editors_group_id_idx" ON "scan_editors" ("group_id");
-- create "scan_assets" table
CREATE TABLE "scan_assets" ("scan_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "asset_id"));
-- create index "scan_assets_asset_id_idx" to table: "scan_assets"
CREATE INDEX "scan_assets_asset_id_idx" ON "scan_assets" ("asset_id");
-- create "scan_entities" table
CREATE TABLE "scan_entities" ("scan_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "entity_id"));
-- create index "scan_entities_entity_id_idx" to table: "scan_entities"
CREATE INDEX "scan_entities_entity_id_idx" ON "scan_entities" ("entity_id");
-- create "scan_evidence" table
CREATE TABLE "scan_evidence" ("scan_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "evidence_id"));
-- create index "scan_evidence_evidence_id_idx" to table: "scan_evidence"
CREATE INDEX "scan_evidence_evidence_id_idx" ON "scan_evidence" ("evidence_id");
-- create "scan_files" table
CREATE TABLE "scan_files" ("scan_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "file_id"));
-- create index "scan_files_file_id_idx" to table: "scan_files"
CREATE INDEX "scan_files_file_id_idx" ON "scan_files" ("file_id");
-- create "scan_remediations" table
CREATE TABLE "scan_remediations" ("scan_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "remediation_id"));
-- create index "scan_remediations_remediation_id_idx" to table: "scan_remediations"
CREATE INDEX "scan_remediations_remediation_id_idx" ON "scan_remediations" ("remediation_id");
-- create "scan_action_plans" table
CREATE TABLE "scan_action_plans" ("scan_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "action_plan_id"));
-- create index "scan_action_plans_action_plan_id_idx" to table: "scan_action_plans"
CREATE INDEX "scan_action_plans_action_plan_id_idx" ON "scan_action_plans" ("action_plan_id");
-- create "scan_tasks" table
CREATE TABLE "scan_tasks" ("scan_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "task_id"));
-- create index "scan_tasks_task_id_idx" to table: "scan_tasks"
CREATE INDEX "scan_tasks_task_id_idx" ON "scan_tasks" ("task_id");
-- create "scheduled_job_controls" table
CREATE TABLE "scheduled_job_controls" ("scheduled_job_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("scheduled_job_id", "control_id"));
-- create index "scheduled_job_controls_control_id_idx" to table: "scheduled_job_controls"
CREATE INDEX "scheduled_job_controls_control_id_idx" ON "scheduled_job_controls" ("control_id");
-- create "scheduled_job_subcontrols" table
CREATE TABLE "scheduled_job_subcontrols" ("scheduled_job_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("scheduled_job_id", "subcontrol_id"));
-- create index "scheduled_job_subcontrols_subcontrol_id_idx" to table: "scheduled_job_subcontrols"
CREATE INDEX "scheduled_job_subcontrols_subcontrol_id_idx" ON "scheduled_job_subcontrols" ("subcontrol_id");
-- create "subcontrol_control_objectives" table
CREATE TABLE "subcontrol_control_objectives" ("subcontrol_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_objective_id"));
-- create index "subcontrol_control_objectives_control_objective_id_idx" to table: "subcontrol_control_objectives"
CREATE INDEX "subcontrol_control_objectives_control_objective_id_idx" ON "subcontrol_control_objectives" ("control_objective_id");
-- create "subcontrol_tasks" table
CREATE TABLE "subcontrol_tasks" ("subcontrol_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "task_id"));
-- create index "subcontrol_tasks_task_id_idx" to table: "subcontrol_tasks"
CREATE INDEX "subcontrol_tasks_task_id_idx" ON "subcontrol_tasks" ("task_id");
-- create "subcontrol_risks" table
CREATE TABLE "subcontrol_risks" ("subcontrol_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "risk_id"));
-- create index "subcontrol_risks_risk_id_idx" to table: "subcontrol_risks"
CREATE INDEX "subcontrol_risks_risk_id_idx" ON "subcontrol_risks" ("risk_id");
-- create "subcontrol_procedures" table
CREATE TABLE "subcontrol_procedures" ("subcontrol_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "procedure_id"));
-- create index "subcontrol_procedures_procedure_id_idx" to table: "subcontrol_procedures"
CREATE INDEX "subcontrol_procedures_procedure_id_idx" ON "subcontrol_procedures" ("procedure_id");
-- create "subcontrol_scans" table
CREATE TABLE "subcontrol_scans" ("subcontrol_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "scan_id"));
-- create index "subcontrol_scans_scan_id_idx" to table: "subcontrol_scans"
CREATE INDEX "subcontrol_scans_scan_id_idx" ON "subcontrol_scans" ("scan_id");
-- create "subcontrol_control_implementations" table
CREATE TABLE "subcontrol_control_implementations" ("subcontrol_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_implementation_id"));
-- create index "subcontrol_control_implementations_control_implementation_id_id" to table: "subcontrol_control_implementations"
CREATE INDEX "subcontrol_control_implementations_control_implementation_id_id" ON "subcontrol_control_implementations" ("control_implementation_id");
-- create "subcontrol_assets" table
CREATE TABLE "subcontrol_assets" ("subcontrol_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "asset_id"));
-- create index "subcontrol_assets_asset_id_idx" to table: "subcontrol_assets"
CREATE INDEX "subcontrol_assets_asset_id_idx" ON "subcontrol_assets" ("asset_id");
-- create "subcontrol_entities" table
CREATE TABLE "subcontrol_entities" ("subcontrol_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "entity_id"));
-- create index "subcontrol_entities_entity_id_idx" to table: "subcontrol_entities"
CREATE INDEX "subcontrol_entities_entity_id_idx" ON "subcontrol_entities" ("entity_id");
-- create "subcontrol_identity_holders" table
CREATE TABLE "subcontrol_identity_holders" ("subcontrol_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "identity_holder_id"));
-- create index "subcontrol_identity_holders_identity_holder_id_idx" to table: "subcontrol_identity_holders"
CREATE INDEX "subcontrol_identity_holders_identity_holder_id_idx" ON "subcontrol_identity_holders" ("identity_holder_id");
-- create "subscriber_events" table
CREATE TABLE "subscriber_events" ("subscriber_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("subscriber_id", "event_id"));
-- create index "subscriber_events_event_id_idx" to table: "subscriber_events"
CREATE INDEX "subscriber_events_event_id_idx" ON "subscriber_events" ("event_id");
-- create "system_detail_assets" table
CREATE TABLE "system_detail_assets" ("system_detail_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("system_detail_id", "asset_id"));
-- create index "system_detail_assets_asset_id_idx" to table: "system_detail_assets"
CREATE INDEX "system_detail_assets_asset_id_idx" ON "system_detail_assets" ("asset_id");
-- create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"));
-- create index "task_evidence_evidence_id_idx" to table: "task_evidence"
CREATE INDEX "task_evidence_evidence_id_idx" ON "task_evidence" ("evidence_id");
-- create "template_files" table
CREATE TABLE "template_files" ("template_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("template_id", "file_id"));
-- create index "template_files_file_id_idx" to table: "template_files"
CREATE INDEX "template_files_file_id_idx" ON "template_files" ("file_id");
-- create "user_events" table
CREATE TABLE "user_events" ("user_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("user_id", "event_id"));
-- create index "user_events_event_id_idx" to table: "user_events"
CREATE INDEX "user_events_event_id_idx" ON "user_events" ("event_id");
-- create "vulnerability_action_plans" table
CREATE TABLE "vulnerability_action_plans" ("vulnerability_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "action_plan_id"));
-- create index "vulnerability_action_plans_action_plan_id_idx" to table: "vulnerability_action_plans"
CREATE INDEX "vulnerability_action_plans_action_plan_id_idx" ON "vulnerability_action_plans" ("action_plan_id");
-- create "vulnerability_controls" table
CREATE TABLE "vulnerability_controls" ("vulnerability_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "control_id"));
-- create index "vulnerability_controls_control_id_idx" to table: "vulnerability_controls"
CREATE INDEX "vulnerability_controls_control_id_idx" ON "vulnerability_controls" ("control_id");
-- create "vulnerability_subcontrols" table
CREATE TABLE "vulnerability_subcontrols" ("vulnerability_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "subcontrol_id"));
-- create index "vulnerability_subcontrols_subcontrol_id_idx" to table: "vulnerability_subcontrols"
CREATE INDEX "vulnerability_subcontrols_subcontrol_id_idx" ON "vulnerability_subcontrols" ("subcontrol_id");
-- create "vulnerability_risks" table
CREATE TABLE "vulnerability_risks" ("vulnerability_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "risk_id"));
-- create index "vulnerability_risks_risk_id_idx" to table: "vulnerability_risks"
CREATE INDEX "vulnerability_risks_risk_id_idx" ON "vulnerability_risks" ("risk_id");
-- create "vulnerability_programs" table
CREATE TABLE "vulnerability_programs" ("vulnerability_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "program_id"));
-- create index "vulnerability_programs_program_id_idx" to table: "vulnerability_programs"
CREATE INDEX "vulnerability_programs_program_id_idx" ON "vulnerability_programs" ("program_id");
-- create "vulnerability_assets" table
CREATE TABLE "vulnerability_assets" ("vulnerability_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "asset_id"));
-- create index "vulnerability_assets_asset_id_idx" to table: "vulnerability_assets"
CREATE INDEX "vulnerability_assets_asset_id_idx" ON "vulnerability_assets" ("asset_id");
-- create "vulnerability_entities" table
CREATE TABLE "vulnerability_entities" ("vulnerability_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "entity_id"));
-- create index "vulnerability_entities_entity_id_idx" to table: "vulnerability_entities"
CREATE INDEX "vulnerability_entities_entity_id_idx" ON "vulnerability_entities" ("entity_id");
-- create "vulnerability_scans" table
CREATE TABLE "vulnerability_scans" ("vulnerability_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "scan_id"));
-- create index "vulnerability_scans_scan_id_idx" to table: "vulnerability_scans"
CREATE INDEX "vulnerability_scans_scan_id_idx" ON "vulnerability_scans" ("scan_id");
-- create "vulnerability_tasks" table
CREATE TABLE "vulnerability_tasks" ("vulnerability_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "task_id"));
-- create index "vulnerability_tasks_task_id_idx" to table: "vulnerability_tasks"
CREATE INDEX "vulnerability_tasks_task_id_idx" ON "vulnerability_tasks" ("task_id");
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD CONSTRAINT "api_tokens_organizations_api_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD CONSTRAINT "action_plans_custom_type_enums_action_plan_kind" FOREIGN KEY ("action_plan_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_custom_type_enums_action_plans" FOREIGN KEY ("custom_type_enum_action_plans") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_organizations_action_plans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_subcontrols_action_plans" FOREIGN KEY ("subcontrol_action_plans") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_users_action_plans" FOREIGN KEY ("user_action_plans") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "assessments" table
ALTER TABLE "assessments" ADD CONSTRAINT "assessments_organizations_assessments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD CONSTRAINT "assessment_responses_assessments_assessment_responses" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "assessment_responses_campaigns_assessment_responses" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_entities_assessment_responses" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_identity_holders_assessment_responses" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_organizations_assessment_responses" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "assets" table
ALTER TABLE "assets" ADD CONSTRAINT "assets_custom_type_enums_access_model" FOREIGN KEY ("access_model_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_asset_data_classification" FOREIGN KEY ("asset_data_classification_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_asset_subtype" FOREIGN KEY ("asset_subtype_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_criticality" FOREIGN KEY ("criticality_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_encryption_status" FOREIGN KEY ("encryption_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_security_tier" FOREIGN KEY ("security_tier_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_integrations_assets" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_organizations_assets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_platforms_source_assets" FOREIGN KEY ("source_platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_risks_assets" FOREIGN KEY ("risk_assets") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD CONSTRAINT "campaigns_assessments_campaigns" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_email_templates_campaigns" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_entities_campaigns" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_integrations_campaigns" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_organizations_campaigns" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_templates_campaigns" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_trust_centers_campaigns" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD CONSTRAINT "campaign_targets_campaigns_campaign_targets" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_contacts_campaign_targets" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_groups_campaign_targets" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_organizations_campaign_targets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_subscribers_campaign_targets" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_users_campaign_targets" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "check_results" table
ALTER TABLE "check_results" ADD CONSTRAINT "check_results_integrations_check_results" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ADD CONSTRAINT "contacts_organizations_contacts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_custom_type_enums_control_kind" FOREIGN KEY ("control_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_controls" FOREIGN KEY ("custom_type_enum_controls") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_entities_responsible_party" FOREIGN KEY ("responsible_party_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_standards_controls" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD CONSTRAINT "control_implementations_evidences_control_implementations" FOREIGN KEY ("evidence_control_implementations") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_implementations_intern_78a7d74302db6f99776c0594111f170b" FOREIGN KEY ("internal_policy_control_implementations") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_implementations_organizations_control_implementations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_organizations_control_objectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD CONSTRAINT "custom_domains_dns_verifications_custom_domains" FOREIGN KEY ("dns_verification_custom_domains") REFERENCES "dns_verifications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_dns_verifications_dns_verification" FOREIGN KEY ("dns_verification_id") REFERENCES "dns_verifications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_mappable_domains_custom_domains" FOREIGN KEY ("mappable_domain_custom_domains") REFERENCES "mappable_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_mappable_domains_mappable_domain" FOREIGN KEY ("mappable_domain_id") REFERENCES "mappable_domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "custom_domains_organizations_custom_domains" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD CONSTRAINT "custom_type_enums_entities_auth_methods" FOREIGN KEY ("entity_auth_methods") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_type_enums_organizations_custom_type_enums" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "dns_verifications" table
ALTER TABLE "dns_verifications" ADD CONSTRAINT "dns_verifications_organizations_dns_verifications" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD CONSTRAINT "directory_accounts_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_directory_sync_runs_directory_accounts" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_identity_holders_directory_accounts" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_integrations_directory_accounts" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_organizations_directory_accounts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_platforms_directory_accounts" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "directory_groups" table
ALTER TABLE "directory_groups" ADD CONSTRAINT "directory_groups_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_directory_sync_runs_directory_groups" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_platforms_directory_groups" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD CONSTRAINT "directory_memberships_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_directory_accounts_directory_account" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_directory_groups_directory_group" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_directory_sync_runs_directory_memberships" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_integrations_directory_memberships" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_memberships_organizations_directory_memberships" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_platforms_directory_memberships" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD CONSTRAINT "directory_sync_runs_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_integrations_directory_sync_runs" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_sync_runs_organizations_directory_sync_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_platforms_directory_sync_runs" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "discussions" table
ALTER TABLE "discussions" ADD CONSTRAINT "discussions_controls_discussions" FOREIGN KEY ("control_discussions") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_internal_policies_discussions" FOREIGN KEY ("internal_policy_discussions") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_organizations_discussions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_procedures_discussions" FOREIGN KEY ("procedure_discussions") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_risks_discussions" FOREIGN KEY ("risk_discussions") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "discussions_subcontrols_discussions" FOREIGN KEY ("subcontrol_discussions") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "document_data" table
ALTER TABLE "document_data" ADD CONSTRAINT "document_data_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_organizations_documents" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "email_templates" table
ALTER TABLE "email_templates" ADD CONSTRAINT "email_templates_integrations_email_templates" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_organizations_email_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_trust_centers_email_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_workflow_definitions_email_templates" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_templates_workflow_instances_email_templates" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD CONSTRAINT "email_verification_tokens_users_email_verification_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_custom_type_enums_entity_relationship_state" FOREIGN KEY ("entity_relationship_state_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_entity_security_questionnaire_status" FOREIGN KEY ("entity_security_questionnaire_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_entity_source_type" FOREIGN KEY ("entity_source_type_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_organizations_entities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_risks_entities" FOREIGN KEY ("risk_entities") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" ADD CONSTRAINT "entity_types_organizations_entity_types" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "events" table
ALTER TABLE "events" ADD CONSTRAINT "events_directory_memberships_events" FOREIGN KEY ("directory_membership_events") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "events_exports_events" FOREIGN KEY ("export_events") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "evidences" table
ALTER TABLE "evidences" ADD CONSTRAINT "evidences_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "exports" table
ALTER TABLE "exports" ADD CONSTRAINT "exports_organizations_exports" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_custom_type_enums_category" FOREIGN KEY ("category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_email_templates_files" FOREIGN KEY ("email_template_files") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_exports_files" FOREIGN KEY ("export_files") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_findings_files" FOREIGN KEY ("finding_files") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_integrations_files" FOREIGN KEY ("integration_files") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_notes_files" FOREIGN KEY ("note_files") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_architecture_diagrams" FOREIGN KEY ("platform_architecture_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_data_flow_diagrams" FOREIGN KEY ("platform_data_flow_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_platforms_trust_boundary_diagrams" FOREIGN KEY ("platform_trust_boundary_diagrams") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_remediations_files" FOREIGN KEY ("remediation_files") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_reviews_files" FOREIGN KEY ("review_files") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_vulnerabilities_files" FOREIGN KEY ("vulnerability_files") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "file_download_tokens" table
ALTER TABLE "file_download_tokens" ADD CONSTRAINT "file_download_tokens_users_file_download_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "findings" table
ALTER TABLE "findings" ADD CONSTRAINT "findings_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_custom_type_enums_finding_status" FOREIGN KEY ("finding_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_organizations_findings" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "finding_controls" table
ALTER TABLE "finding_controls" ADD CONSTRAINT "finding_controls_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "finding_controls_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "finding_controls_organizations_finding_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "finding_controls_standards_standard" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_assessments_blocked_groups" FOREIGN KEY ("assessment_blocked_groups") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_editors" FOREIGN KEY ("assessment_editors") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_viewers" FOREIGN KEY ("assessment_viewers") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_blocked_groups" FOREIGN KEY ("asset_blocked_groups") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_editors" FOREIGN KEY ("asset_editors") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_viewers" FOREIGN KEY ("asset_viewers") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_blocked_groups" FOREIGN KEY ("check_result_blocked_groups") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_editors" FOREIGN KEY ("check_result_editors") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_viewers" FOREIGN KEY ("check_result_viewers") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_blocked_groups" FOREIGN KEY ("email_template_blocked_groups") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_editors" FOREIGN KEY ("email_template_editors") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_viewers" FOREIGN KEY ("email_template_viewers") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_blocked_groups" FOREIGN KEY ("identity_holder_blocked_groups") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_editors" FOREIGN KEY ("identity_holder_editors") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_viewers" FOREIGN KEY ("identity_holder_viewers") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_action_plan_creators" FOREIGN KEY ("organization_action_plan_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_api_token_creators" FOREIGN KEY ("organization_api_token_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_assessment_creators" FOREIGN KEY ("organization_assessment_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_asset_creators" FOREIGN KEY ("organization_asset_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_campaign_creators" FOREIGN KEY ("organization_campaign_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_campaign_target_creators" FOREIGN KEY ("organization_campaign_target_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_campaigns_manager" FOREIGN KEY ("organization_campaigns_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_check_result_creators" FOREIGN KEY ("organization_check_result_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_compliance_manager" FOREIGN KEY ("organization_compliance_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_contact_creators" FOREIGN KEY ("organization_contact_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_creators" FOREIGN KEY ("organization_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_implementation_creators" FOREIGN KEY ("organization_control_implementation_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_objective_creators" FOREIGN KEY ("organization_control_objective_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_custom_domain_creators" FOREIGN KEY ("organization_custom_domain_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_custom_type_enum_creators" FOREIGN KEY ("organization_custom_type_enum_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_account_creators" FOREIGN KEY ("organization_directory_account_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_group_creators" FOREIGN KEY ("organization_directory_group_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_membership_creators" FOREIGN KEY ("organization_directory_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_directory_sync_run_creators" FOREIGN KEY ("organization_directory_sync_run_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_discussion_creators" FOREIGN KEY ("organization_discussion_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_document_data_creators" FOREIGN KEY ("organization_document_data_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_email_template_creators" FOREIGN KEY ("organization_email_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_entity_creators" FOREIGN KEY ("organization_entity_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_entity_type_creators" FOREIGN KEY ("organization_entity_type_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_evidence_creators" FOREIGN KEY ("organization_evidence_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_file_creators" FOREIGN KEY ("organization_file_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_finding_control_creators" FOREIGN KEY ("organization_finding_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_finding_creators" FOREIGN KEY ("organization_finding_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_creators" FOREIGN KEY ("organization_group_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_manager" FOREIGN KEY ("organization_group_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_membership_creators" FOREIGN KEY ("organization_group_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_setting_creators" FOREIGN KEY ("organization_group_setting_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_hush_creators" FOREIGN KEY ("organization_hush_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_identity_holder_creators" FOREIGN KEY ("organization_identity_holder_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_internal_policy_creators" FOREIGN KEY ("organization_internal_policy_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_invite_creators" FOREIGN KEY ("organization_invite_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_runner_creators" FOREIGN KEY ("organization_job_runner_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_runner_registration_token_creators" FOREIGN KEY ("organization_job_runner_registration_token_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_runner_token_creators" FOREIGN KEY ("organization_job_runner_token_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_job_template_creators" FOREIGN KEY ("organization_job_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_mapped_control_creators" FOREIGN KEY ("organization_mapped_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_narrative_creators" FOREIGN KEY ("organization_narrative_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_note_creators" FOREIGN KEY ("organization_note_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_notification_template_creators" FOREIGN KEY ("organization_notification_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_org_membership_creators" FOREIGN KEY ("organization_org_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_platform_creators" FOREIGN KEY ("organization_platform_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_policies_manager" FOREIGN KEY ("organization_policies_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_procedure_creators" FOREIGN KEY ("organization_procedure_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_program_creators" FOREIGN KEY ("organization_program_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_program_membership_creators" FOREIGN KEY ("organization_program_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_registry_manager" FOREIGN KEY ("organization_registry_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_remediation_creators" FOREIGN KEY ("organization_remediation_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_review_creators" FOREIGN KEY ("organization_review_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_risk_creators" FOREIGN KEY ("organization_risk_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_risk_manager" FOREIGN KEY ("organization_risk_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scan_creators" FOREIGN KEY ("organization_scan_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scheduled_job_creators" FOREIGN KEY ("organization_scheduled_job_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_scheduled_job_run_creators" FOREIGN KEY ("organization_scheduled_job_run_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_sla_definition_creators" FOREIGN KEY ("organization_sla_definition_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_standard_creators" FOREIGN KEY ("organization_standard_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_subcontrol_creators" FOREIGN KEY ("organization_subcontrol_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_subprocessor_creators" FOREIGN KEY ("organization_subprocessor_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_subscriber_creators" FOREIGN KEY ("organization_subscriber_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_system_detail_creators" FOREIGN KEY ("organization_system_detail_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_tag_definition_creators" FOREIGN KEY ("organization_tag_definition_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_task_creators" FOREIGN KEY ("organization_task_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_template_creators" FOREIGN KEY ("organization_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_compliance_creators" FOREIGN KEY ("organization_trust_center_compliance_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_creators" FOREIGN KEY ("organization_trust_center_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_doc_creators" FOREIGN KEY ("organization_trust_center_doc_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_entity_creators" FOREIGN KEY ("organization_trust_center_entity_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_faq_creators" FOREIGN KEY ("organization_trust_center_faq_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_manager" FOREIGN KEY ("organization_trust_center_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_nda_request_creators" FOREIGN KEY ("organization_trust_center_nda_request_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_subprocessor_creators" FOREIGN KEY ("organization_trust_center_subprocessor_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_watermark_config_creators" FOREIGN KEY ("organization_trust_center_watermark_config_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_vendor_risk_score_creators" FOREIGN KEY ("organization_vendor_risk_score_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_vendor_scoring_config_creators" FOREIGN KEY ("organization_vendor_scoring_config_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_vulnerability_creators" FOREIGN KEY ("organization_vulnerability_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_workflow_definition_creators" FOREIGN KEY ("organization_workflow_definition_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_workflows_manager" FOREIGN KEY ("organization_workflows_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_sla_definitions_blocked_groups" FOREIGN KEY ("sla_definition_blocked_groups") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_sla_definitions_editors" FOREIGN KEY ("sla_definition_editors") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_compliances_blocked_groups" FOREIGN KEY ("trust_center_compliance_blocked_groups") REFERENCES "trust_center_compliances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_compliances_editors" FOREIGN KEY ("trust_center_compliance_editors") REFERENCES "trust_center_compliances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_docs_blocked_groups" FOREIGN KEY ("trust_center_doc_blocked_groups") REFERENCES "trust_center_docs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_docs_editors" FOREIGN KEY ("trust_center_doc_editors") REFERENCES "trust_center_docs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_entities_blocked_groups" FOREIGN KEY ("trust_center_entity_blocked_groups") REFERENCES "trust_center_entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_entities_editors" FOREIGN KEY ("trust_center_entity_editors") REFERENCES "trust_center_entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_faqs_blocked_groups" FOREIGN KEY ("trust_center_faq_blocked_groups") REFERENCES "trust_center_faqs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_faqs_editors" FOREIGN KEY ("trust_center_faq_editors") REFERENCES "trust_center_faqs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_nda_requests_blocked_groups" FOREIGN KEY ("trust_center_nda_request_blocked_groups") REFERENCES "trust_center_nda_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_nda_requests_editors" FOREIGN KEY ("trust_center_nda_request_editors") REFERENCES "trust_center_nda_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_settings_blocked_groups" FOREIGN KEY ("trust_center_setting_blocked_groups") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_settings_editors" FOREIGN KEY ("trust_center_setting_editors") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_subprocessors_blocked_groups" FOREIGN KEY ("trust_center_subprocessor_blocked_groups") REFERENCES "trust_center_subprocessors" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_subprocessors_editors" FOREIGN KEY ("trust_center_subprocessor_editors") REFERENCES "trust_center_subprocessors" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_watermark_configs_blocked_groups" FOREIGN KEY ("trust_center_watermark_config_blocked_groups") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_watermark_configs_editors" FOREIGN KEY ("trust_center_watermark_config_editors") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_centers_blocked_groups" FOREIGN KEY ("trust_center_blocked_groups") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_centers_editors" FOREIGN KEY ("trust_center_editors") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_blocked_groups" FOREIGN KEY ("vulnerability_blocked_groups") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_editors" FOREIGN KEY ("vulnerability_editors") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_viewers" FOREIGN KEY ("vulnerability_viewers") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_blocked_groups" FOREIGN KEY ("workflow_definition_blocked_groups") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_editors" FOREIGN KEY ("workflow_definition_editors") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_groups" FOREIGN KEY ("workflow_definition_groups") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_workflow_definitions_viewers" FOREIGN KEY ("workflow_definition_viewers") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "group_memberships" table
ALTER TABLE "group_memberships" ADD CONSTRAINT "group_memberships_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "group_memberships_org_memberships_org_membership" FOREIGN KEY ("group_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "group_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "group_settings" table
ALTER TABLE "group_settings" ADD CONSTRAINT "group_settings_groups_setting" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "hushes" table
ALTER TABLE "hushes" ADD CONSTRAINT "hushes_organizations_secrets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "identity_holders" table
ALTER TABLE "identity_holders" ADD CONSTRAINT "identity_holders_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_entities_employer" FOREIGN KEY ("employer_entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_organizations_identity_holders" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_users_identity_holder_profiles" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "impersonation_events" table
ALTER TABLE "impersonation_events" ADD CONSTRAINT "impersonation_events_organizations_impersonation_events" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "impersonation_events_users_impersonation_events" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "impersonation_events_users_targeted_impersonations" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_files_integrations" FOREIGN KEY ("file_integrations") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_groups_integrations" FOREIGN KEY ("group_integrations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_organizations_integrations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_platforms_integrations" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD CONSTRAINT "integration_runs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_events_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_files_request_file" FOREIGN KEY ("request_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_files_response_file" FOREIGN KEY ("response_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_integrations_integration_runs" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_organizations_integration_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" ADD CONSTRAINT "integration_webhooks_integrations_integration_webhooks" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_webhooks_organizations_integration_webhooks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_internal_policies" FOREIGN KEY ("custom_type_enum_internal_policies") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_internal_policy_kind" FOREIGN KEY ("internal_policy_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_organizations_internal_policies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "invites" table
ALTER TABLE "invites" ADD CONSTRAINT "invites_organizations_invites" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "job_results" table
ALTER TABLE "job_results" ADD CONSTRAINT "job_results_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "job_results_organizations_job_results" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "job_results_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "job_runners" table
ALTER TABLE "job_runners" ADD CONSTRAINT "job_runners_organizations_job_runners" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "job_runner_registration_tokens" table
ALTER TABLE "job_runner_registration_tokens" ADD CONSTRAINT "job_runner_registration_tokens_daddf3e078805108b2d174df258ddb4b" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "job_runner_registration_tokens_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "job_runner_tokens" table
ALTER TABLE "job_runner_tokens" ADD CONSTRAINT "job_runner_tokens_organizations_job_runner_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "job_templates" table
ALTER TABLE "job_templates" ADD CONSTRAINT "job_templates_organizations_job_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD CONSTRAINT "mapped_controls_organizations_mapped_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD CONSTRAINT "narratives_control_objectives_narratives" FOREIGN KEY ("control_objective_narratives") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_subcontrols_narratives" FOREIGN KEY ("subcontrol_narratives") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD CONSTRAINT "notes_controls_comments" FOREIGN KEY ("control_comments") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_discussions_comments" FOREIGN KEY ("discussion_id") REFERENCES "discussions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_entities_notes" FOREIGN KEY ("entity_notes") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_evidences_comments" FOREIGN KEY ("evidence_comments") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_findings_comments" FOREIGN KEY ("finding_comments") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_internal_policies_comments" FOREIGN KEY ("internal_policy_comments") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_organizations_notes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_procedures_comments" FOREIGN KEY ("procedure_comments") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_programs_notes" FOREIGN KEY ("program_notes") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_remediations_comments" FOREIGN KEY ("remediation_comments") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_reviews_comments" FOREIGN KEY ("review_comments") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_risks_comments" FOREIGN KEY ("risk_comments") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_subcontrols_comments" FOREIGN KEY ("subcontrol_comments") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_tasks_comments" FOREIGN KEY ("task_comments") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_trust_centers_posts" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_vulnerabilities_comments" FOREIGN KEY ("vulnerability_comments") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notifications" table
ALTER TABLE "notifications" ADD CONSTRAINT "notifications_notification_templates_notifications" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notifications_organizations_notifications" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notification_preferences" table
ALTER TABLE "notification_preferences" ADD CONSTRAINT "notification_preferences_notif_aabd0a3ca9e335110ce7c2348e4f4cf0" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_preferences_organizations_notification_preferences" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_preferences_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "notification_templates" table
ALTER TABLE "notification_templates" ADD CONSTRAINT "notification_templates_email_templates_notification_templates" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_templates_integrations_notification_templates" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_templates_organizations_notification_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notification_templates_workflo_439a17f2830fbf868eeb61d3d3fdac37" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "onboardings" table
ALTER TABLE "onboardings" ADD CONSTRAINT "onboardings_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_memberships" table
ALTER TABLE "org_memberships" ADD CONSTRAINT "org_memberships_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "org_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "org_modules" table
ALTER TABLE "org_modules" ADD CONSTRAINT "org_modules_org_products_org_modules" FOREIGN KEY ("org_product_org_modules") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_modules_org_subscriptions_modules" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_modules_organizations_org_modules" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_prices" table
ALTER TABLE "org_prices" ADD CONSTRAINT "org_prices_org_subscriptions_prices" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_prices_organizations_org_prices" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_products" table
ALTER TABLE "org_products" ADD CONSTRAINT "org_products_org_modules_org_products" FOREIGN KEY ("org_module_org_products") REFERENCES "org_modules" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_products_org_subscriptions_products" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_products_organizations_org_products" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD CONSTRAINT "org_subscriptions_organizations_org_subscriptions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "organizations" table
ALTER TABLE "organizations" ADD CONSTRAINT "organizations_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD CONSTRAINT "organization_settings_organizations_setting" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" ADD CONSTRAINT "password_reset_tokens_users_password_reset_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD CONSTRAINT "personal_access_tokens_users_personal_access_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "platforms" table
ALTER TABLE "platforms" ADD CONSTRAINT "platforms_custom_type_enums_access_model" FOREIGN KEY ("access_model_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_criticality" FOREIGN KEY ("criticality_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_encryption_status" FOREIGN KEY ("encryption_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platform_data_classification" FOREIGN KEY ("platform_data_classification_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platform_kind" FOREIGN KEY ("platform_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platforms" FOREIGN KEY ("custom_type_enum_platforms") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_security_tier" FOREIGN KEY ("security_tier_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_business_owner_group" FOREIGN KEY ("business_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_security_owner_group" FOREIGN KEY ("security_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_technical_owner_group" FOREIGN KEY ("technical_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_identity_holders_access_platforms" FOREIGN KEY ("identity_holder_access_platforms") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_organizations_platforms" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_business_owner_user" FOREIGN KEY ("business_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_platforms_owned" FOREIGN KEY ("platform_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_security_owner_user" FOREIGN KEY ("security_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_technical_owner_user" FOREIGN KEY ("technical_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_procedure_kind" FOREIGN KEY ("procedure_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_procedures" FOREIGN KEY ("custom_type_enum_procedures") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_organizations_procedures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD CONSTRAINT "programs_custom_type_enums_program_kind" FOREIGN KEY ("program_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_custom_type_enums_programs" FOREIGN KEY ("custom_type_enum_programs") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_organizations_programs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_users_programs_owned" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" ADD CONSTRAINT "program_memberships_org_memberships_org_membership" FOREIGN KEY ("program_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "program_memberships_programs_program" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "program_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "remediations" table
ALTER TABLE "remediations" ADD CONSTRAINT "remediations_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_organizations_remediations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "reviews" table
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_organizations_reviews" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_users_reviewer" FOREIGN KEY ("reviewer_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risk_categories" FOREIGN KEY ("custom_type_enum_risk_categories") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risk_category" FOREIGN KEY ("risk_category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risk_kind" FOREIGN KEY ("risk_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_risks" FOREIGN KEY ("custom_type_enum_risks") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("stakeholder_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "sla_definitions" table
ALTER TABLE "sla_definitions" ADD CONSTRAINT "sla_definitions_organizations_sla_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD CONSTRAINT "scans_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_performed_by_group" FOREIGN KEY ("performed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_organizations_scans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_platforms_generated_scans" FOREIGN KEY ("generated_by_platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_risks_scans" FOREIGN KEY ("risk_scans") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_performed_by_user" FOREIGN KEY ("performed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD CONSTRAINT "scheduled_jobs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs" FOREIGN KEY ("job_id") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "scheduled_jobs_organizations_scheduled_jobs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ADD CONSTRAINT "scheduled_job_runs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "scheduled_job_runs_organizations_scheduled_job_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scheduled_job_runs_scheduled_jobs_scheduled_job" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "standards" table
ALTER TABLE "standards" ADD CONSTRAINT "standards_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "standards_organizations_standards" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_controls_subcontrols" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "subcontrols_custom_type_enums_subcontrol_kind" FOREIGN KEY ("subcontrol_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_custom_type_enums_subcontrols" FOREIGN KEY ("custom_type_enum_subcontrols") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_entities_responsible_party" FOREIGN KEY ("responsible_party_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_programs_subcontrols" FOREIGN KEY ("program_subcontrols") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_subcontrols" FOREIGN KEY ("user_subcontrols") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subprocessors" table
ALTER TABLE "subprocessors" ADD CONSTRAINT "subprocessors_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subprocessors_organizations_subprocessors" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD CONSTRAINT "subscribers_contacts_subscribers" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_organizations_subscribers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_trust_centers_subscribers" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_users_subscribers" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "system_details" table
ALTER TABLE "system_details" ADD CONSTRAINT "system_details_organizations_system_details" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD CONSTRAINT "tfa_settings_users_tfa_settings" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD CONSTRAINT "tag_definitions_organizations_tag_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tag_definitions_workflow_definitions_tag_definitions" FOREIGN KEY ("workflow_definition_tag_definitions") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_task_kind" FOREIGN KEY ("task_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_tasks" FOREIGN KEY ("custom_type_enum_tasks") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_integrations_tasks" FOREIGN KEY ("integration_tasks") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_remediations_tasks" FOREIGN KEY ("remediation_tasks") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_reviews_tasks" FOREIGN KEY ("review_tasks") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_organizations_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_trust_centers_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD CONSTRAINT "trust_centers_custom_domains_custom_domain" FOREIGN KEY ("custom_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_custom_domains_preview_domain" FOREIGN KEY ("preview_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_organizations_trust_centers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_preview_setting" FOREIGN KEY ("trust_center_preview_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_setting" FOREIGN KEY ("trust_center_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_watermark_configs_watermark_config" FOREIGN KEY ("trust_center_watermark_config") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" ADD CONSTRAINT "trust_center_compliances_standards_trust_center_compliances" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_compliances_trust_centers_trust_center_compliances" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD CONSTRAINT "trust_center_docs_custom_type_enums_trust_center_doc_kind" FOREIGN KEY ("trust_center_doc_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_files_original_file" FOREIGN KEY ("original_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_standards_trust_center_docs" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_trust_center_nda_requests_trust_center_docs" FOREIGN KEY ("trust_center_nda_request_trust_center_docs") REFERENCES "trust_center_nda_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_trust_centers_trust_center_docs" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_entities" table
ALTER TABLE "trust_center_entities" ADD CONSTRAINT "trust_center_entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_entities_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_entities_files_trust_center_entities" FOREIGN KEY ("file_trust_center_entities") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_entities_trust_centers_trust_center_entities" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" ADD CONSTRAINT "trust_center_faqs_custom_type_enums_trust_center_faq_kind" FOREIGN KEY ("trust_center_faq_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_faqs_notes_trust_center_faqs" FOREIGN KEY ("note_id") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_faqs_trust_centers_trust_center_faqs" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD CONSTRAINT "trust_center_nda_requests_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_nda_requests_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_nda_requests_trus_166c4573710ee5957bac7d4b99111f81" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_nda_requests_users_approved_by_user" FOREIGN KEY ("approved_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD CONSTRAINT "trust_center_settings_files_favicon_file" FOREIGN KEY ("favicon_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_settings_files_hero_image_file" FOREIGN KEY ("hero_image_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_settings_files_logo_file" FOREIGN KEY ("logo_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_settings_groups_nda_approver_group" FOREIGN KEY ("nda_approver_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" ADD CONSTRAINT "trust_center_subprocessors_cus_d5ebb915269b07a0bf77b5b0ec180583" FOREIGN KEY ("trust_center_subprocessor_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_subprocessors_sub_24055b695e9bd0e49b3edea05d355a0b" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_subprocessors_tru_bb0fd7936579c86ecda7d42ebfe60199" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ADD CONSTRAINT "trust_center_watermark_configs_e2f038ca8412a7e2b03e1fad46be2f7f" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_watermark_configs_files_file" FOREIGN KEY ("logo_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "user_settings" table
ALTER TABLE "user_settings" ADD CONSTRAINT "user_settings_organizations_default_org" FOREIGN KEY ("user_setting_default_org") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_settings_users_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" ADD CONSTRAINT "vendor_risk_scores_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_assessment_responses_vendor_risk_scores" FOREIGN KEY ("assessment_response_vendor_risk_scores") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_entities_entity" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "vendor_risk_scores_entities_vendor_risk_scores" FOREIGN KEY ("entity_vendor_risk_scores") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_organizations_vendor_risk_scores" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_risk_scores" FOREIGN KEY ("vendor_scoring_config_vendor_risk_scores") REFERENCES "vendor_scoring_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_scoring_config" FOREIGN KEY ("vendor_scoring_config_id") REFERENCES "vendor_scoring_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vendor_scoring_configs" table
ALTER TABLE "vendor_scoring_configs" ADD CONSTRAINT "vendor_scoring_configs_organizations_vendor_scoring_configs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD CONSTRAINT "vulnerabilities_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_custom_type_enums_vulnerability_status" FOREIGN KEY ("vulnerability_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_organizations_vulnerabilities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "webauthns" table
ALTER TABLE "webauthns" ADD CONSTRAINT "webauthns_users_webauthns" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD CONSTRAINT "workflow_assignments_groups_group" FOREIGN KEY ("actor_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_organizations_workflow_assignments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_users_user" FOREIGN KEY ("actor_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_workflow_instances_workflow_assignments" FOREIGN KEY ("workflow_instance_workflow_assignments") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignments_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "workflow_assignment_targets" table
ALTER TABLE "workflow_assignment_targets" ADD CONSTRAINT "workflow_assignment_targets_groups_group" FOREIGN KEY ("target_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_or_8bb74468c70e1b9fcce1d5b038516f9a" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_users_user" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_wo_35919ebc89c62ef82cb5889ff40ce351" FOREIGN KEY ("workflow_assignment_workflow_assignment_targets") REFERENCES "workflow_assignments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_assignment_targets_wo_6077e6f4bf744947c345bb2733c1c240" FOREIGN KEY ("workflow_assignment_id") REFERENCES "workflow_assignments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ADD CONSTRAINT "workflow_definitions_organizations_workflow_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "workflow_events" table
ALTER TABLE "workflow_events" ADD CONSTRAINT "workflow_events_organizations_workflow_events" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_events_workflow_instances_workflow_events" FOREIGN KEY ("workflow_instance_workflow_events") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_events_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD CONSTRAINT "workflow_instances_action_plans_action_plan" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_campaign_targets_campaign_target" FOREIGN KEY ("campaign_target_id") REFERENCES "campaign_targets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_campaigns_campaign" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_identity_holders_identity_holder" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_organizations_workflow_instances" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_platforms_platform" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_procedures_procedure" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_subcontrols_subcontrol" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_workflow_definitions_workflow_definition" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "workflow_instances_workflow_proposals_workflow_proposal" FOREIGN KEY ("workflow_proposal_id") REFERENCES "workflow_proposals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD CONSTRAINT "workflow_object_refs_action_plans_action_plan" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_campaign_targets_campaign_target" FOREIGN KEY ("campaign_target_id") REFERENCES "campaign_targets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_campaigns_campaign" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_directory_accounts_directory_account" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_directory_groups_directory_group" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_directory_memberships_directory_membership" FOREIGN KEY ("directory_membership_id") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_identity_holders_identity_holder" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_organizations_workflow_object_refs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_platforms_platform" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_procedures_procedure" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_subcontrols_subcontrol" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "workflow_object_refs_workflow_instances_workflow_object_refs" FOREIGN KEY ("workflow_instance_workflow_object_refs") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" ADD CONSTRAINT "workflow_proposals_organizations_workflow_proposals" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_proposals_users_user" FOREIGN KEY ("submitted_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_proposals_workflow_object_refs_workflow_object_ref" FOREIGN KEY ("workflow_object_ref_id") REFERENCES "workflow_object_refs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "action_plan_blocked_groups" table
ALTER TABLE "action_plan_blocked_groups" ADD CONSTRAINT "action_plan_blocked_groups_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "action_plan_editors" table
ALTER TABLE "action_plan_editors" ADD CONSTRAINT "action_plan_editors_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "action_plan_viewers" table
ALTER TABLE "action_plan_viewers" ADD CONSTRAINT "action_plan_viewers_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "action_plan_tasks" table
ALTER TABLE "action_plan_tasks" ADD CONSTRAINT "action_plan_tasks_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "asset_connected_assets" table
ALTER TABLE "asset_connected_assets" ADD CONSTRAINT "asset_connected_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "asset_connected_assets_connected_from_id" FOREIGN KEY ("connected_from_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_blocked_groups" table
ALTER TABLE "campaign_blocked_groups" ADD CONSTRAINT "campaign_blocked_groups_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_editors" table
ALTER TABLE "campaign_editors" ADD CONSTRAINT "campaign_editors_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_viewers" table
ALTER TABLE "campaign_viewers" ADD CONSTRAINT "campaign_viewers_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_contacts" table
ALTER TABLE "campaign_contacts" ADD CONSTRAINT "campaign_contacts_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_users" table
ALTER TABLE "campaign_users" ADD CONSTRAINT "campaign_users_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_groups" table
ALTER TABLE "campaign_groups" ADD CONSTRAINT "campaign_groups_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "campaign_identity_holders" table
ALTER TABLE "campaign_identity_holders" ADD CONSTRAINT "campaign_identity_holders_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "campaign_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "check_result_controls" table
ALTER TABLE "check_result_controls" ADD CONSTRAINT "check_result_controls_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "check_result_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "contact_files" table
ALTER TABLE "contact_files" ADD CONSTRAINT "contact_files_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "contact_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_control_objectives" table
ALTER TABLE "control_control_objectives" ADD CONSTRAINT "control_control_objectives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_tasks" table
ALTER TABLE "control_tasks" ADD CONSTRAINT "control_tasks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_narratives" table
ALTER TABLE "control_narratives" ADD CONSTRAINT "control_narratives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_risks" table
ALTER TABLE "control_risks" ADD CONSTRAINT "control_risks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_action_plans" table
ALTER TABLE "control_action_plans" ADD CONSTRAINT "control_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_action_plans_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_procedures" table
ALTER TABLE "control_procedures" ADD CONSTRAINT "control_procedures_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_scans" table
ALTER TABLE "control_scans" ADD CONSTRAINT "control_scans_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" ADD CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_editors" table
ALTER TABLE "control_editors" ADD CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_assets" table
ALTER TABLE "control_assets" ADD CONSTRAINT "control_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_assets_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_entities" table
ALTER TABLE "control_entities" ADD CONSTRAINT "control_entities_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_identity_holders" table
ALTER TABLE "control_identity_holders" ADD CONSTRAINT "control_identity_holders_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_campaigns" table
ALTER TABLE "control_campaigns" ADD CONSTRAINT "control_campaigns_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_campaigns_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_control_implementations" table
ALTER TABLE "control_control_implementations" ADD CONSTRAINT "control_control_implementations_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_implementation_blocked_groups" table
ALTER TABLE "control_implementation_blocked_groups" ADD CONSTRAINT "control_implementation_blocked_groups_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_implementation_editors" table
ALTER TABLE "control_implementation_editors" ADD CONSTRAINT "control_implementation_editors_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_implementation_viewers" table
ALTER TABLE "control_implementation_viewers" ADD CONSTRAINT "control_implementation_viewers_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_implementation_tasks" table
ALTER TABLE "control_implementation_tasks" ADD CONSTRAINT "control_implementation_tasks_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_implementation_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_blocked_groups" table
ALTER TABLE "control_objective_blocked_groups" ADD CONSTRAINT "control_objective_blocked_groups_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_editors" table
ALTER TABLE "control_objective_editors" ADD CONSTRAINT "control_objective_editors_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_viewers" table
ALTER TABLE "control_objective_viewers" ADD CONSTRAINT "control_objective_viewers_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_tasks" table
ALTER TABLE "control_objective_tasks" ADD CONSTRAINT "control_objective_tasks_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "document_data_files" table
ALTER TABLE "document_data_files" ADD CONSTRAINT "document_data_files_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "document_data_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_blocked_groups" table
ALTER TABLE "entity_blocked_groups" ADD CONSTRAINT "entity_blocked_groups_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_editors" table
ALTER TABLE "entity_editors" ADD CONSTRAINT "entity_editors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_contacts" table
ALTER TABLE "entity_contacts" ADD CONSTRAINT "entity_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_contacts_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_documents" table
ALTER TABLE "entity_documents" ADD CONSTRAINT "entity_documents_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_documents_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_files" table
ALTER TABLE "entity_files" ADD CONSTRAINT "entity_files_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_assets" table
ALTER TABLE "entity_assets" ADD CONSTRAINT "entity_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_assets_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_system_details" table
ALTER TABLE "entity_system_details" ADD CONSTRAINT "entity_system_details_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_integrations" table
ALTER TABLE "entity_integrations" ADD CONSTRAINT "entity_integrations_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_integrations_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_subprocessors" table
ALTER TABLE "entity_subprocessors" ADD CONSTRAINT "entity_subprocessors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_subprocessors_subprocessor_id" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_controls" table
ALTER TABLE "evidence_controls" ADD CONSTRAINT "evidence_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_controls_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_subcontrols" table
ALTER TABLE "evidence_subcontrols" ADD CONSTRAINT "evidence_subcontrols_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_control_objectives" table
ALTER TABLE "evidence_control_objectives" ADD CONSTRAINT "evidence_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_control_objectives_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_files" table
ALTER TABLE "evidence_files" ADD CONSTRAINT "evidence_files_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "file_events" table
ALTER TABLE "file_events" ADD CONSTRAINT "file_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_events_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "file_secrets" table
ALTER TABLE "file_secrets" ADD CONSTRAINT "file_secrets_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_blocked_groups" table
ALTER TABLE "finding_blocked_groups" ADD CONSTRAINT "finding_blocked_groups_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_editors" table
ALTER TABLE "finding_editors" ADD CONSTRAINT "finding_editors_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_vulnerabilities" table
ALTER TABLE "finding_vulnerabilities" ADD CONSTRAINT "finding_vulnerabilities_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_action_plans" table
ALTER TABLE "finding_action_plans" ADD CONSTRAINT "finding_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_action_plans_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_subcontrols" table
ALTER TABLE "finding_subcontrols" ADD CONSTRAINT "finding_subcontrols_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_risks" table
ALTER TABLE "finding_risks" ADD CONSTRAINT "finding_risks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_programs" table
ALTER TABLE "finding_programs" ADD CONSTRAINT "finding_programs_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_assets" table
ALTER TABLE "finding_assets" ADD CONSTRAINT "finding_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_assets_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_entities" table
ALTER TABLE "finding_entities" ADD CONSTRAINT "finding_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_entities_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_scans" table
ALTER TABLE "finding_scans" ADD CONSTRAINT "finding_scans_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_tasks" table
ALTER TABLE "finding_tasks" ADD CONSTRAINT "finding_tasks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_directory_accounts" table
ALTER TABLE "finding_directory_accounts" ADD CONSTRAINT "finding_directory_accounts_directory_account_id" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_directory_accounts_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_identity_holders" table
ALTER TABLE "finding_identity_holders" ADD CONSTRAINT "finding_identity_holders_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_check_results" table
ALTER TABLE "finding_check_results" ADD CONSTRAINT "finding_check_results_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_check_results_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "group_events" table
ALTER TABLE "group_events" ADD CONSTRAINT "group_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_events_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "group_files" table
ALTER TABLE "group_files" ADD CONSTRAINT "group_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_files_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "group_tasks" table
ALTER TABLE "group_tasks" ADD CONSTRAINT "group_tasks_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "group_membership_events" table
ALTER TABLE "group_membership_events" ADD CONSTRAINT "group_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "group_membership_events_group_membership_id" FOREIGN KEY ("group_membership_id") REFERENCES "group_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "hush_events" table
ALTER TABLE "hush_events" ADD CONSTRAINT "hush_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "hush_events_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "identity_holder_assessments" table
ALTER TABLE "identity_holder_assessments" ADD CONSTRAINT "identity_holder_assessments_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_assessments_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "identity_holder_templates" table
ALTER TABLE "identity_holder_templates" ADD CONSTRAINT "identity_holder_templates_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_templates_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "identity_holder_assets" table
ALTER TABLE "identity_holder_assets" ADD CONSTRAINT "identity_holder_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_assets_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "identity_holder_entities" table
ALTER TABLE "identity_holder_entities" ADD CONSTRAINT "identity_holder_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_entities_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "identity_holder_tasks" table
ALTER TABLE "identity_holder_tasks" ADD CONSTRAINT "identity_holder_tasks_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "identity_holder_files" table
ALTER TABLE "identity_holder_files" ADD CONSTRAINT "identity_holder_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "identity_holder_files_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_secrets" table
ALTER TABLE "integration_secrets" ADD CONSTRAINT "integration_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_secrets_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_events" table
ALTER TABLE "integration_events" ADD CONSTRAINT "integration_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_events_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_findings" table
ALTER TABLE "integration_findings" ADD CONSTRAINT "integration_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_findings_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_vulnerabilities" table
ALTER TABLE "integration_vulnerabilities" ADD CONSTRAINT "integration_vulnerabilities_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_internal_policies" table
ALTER TABLE "integration_internal_policies" ADD CONSTRAINT "integration_internal_policies_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_reviews" table
ALTER TABLE "integration_reviews" ADD CONSTRAINT "integration_reviews_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_reviews_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_remediations" table
ALTER TABLE "integration_remediations" ADD CONSTRAINT "integration_remediations_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_action_plans" table
ALTER TABLE "integration_action_plans" ADD CONSTRAINT "integration_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_action_plans_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_blocked_groups" table
ALTER TABLE "internal_policy_blocked_groups" ADD CONSTRAINT "internal_policy_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_blocked_groups_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_editors" table
ALTER TABLE "internal_policy_editors" ADD CONSTRAINT "internal_policy_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_editors_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_control_objectives" table
ALTER TABLE "internal_policy_control_objectives" ADD CONSTRAINT "internal_policy_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_control_objectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_controls" table
ALTER TABLE "internal_policy_controls" ADD CONSTRAINT "internal_policy_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_controls_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_subcontrols" table
ALTER TABLE "internal_policy_subcontrols" ADD CONSTRAINT "internal_policy_subcontrols_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" ADD CONSTRAINT "internal_policy_narratives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" ADD CONSTRAINT "internal_policy_tasks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_risks" table
ALTER TABLE "internal_policy_risks" ADD CONSTRAINT "internal_policy_risks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_assets" table
ALTER TABLE "internal_policy_assets" ADD CONSTRAINT "internal_policy_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_assets_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_entities" table
ALTER TABLE "internal_policy_entities" ADD CONSTRAINT "internal_policy_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_entities_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_identity_holders" table
ALTER TABLE "internal_policy_identity_holders" ADD CONSTRAINT "internal_policy_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_identity_holders_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "invite_events" table
ALTER TABLE "invite_events" ADD CONSTRAINT "invite_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_events_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "invite_groups" table
ALTER TABLE "invite_groups" ADD CONSTRAINT "invite_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_groups_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "job_runner_job_runner_tokens" table
ALTER TABLE "job_runner_job_runner_tokens" ADD CONSTRAINT "job_runner_job_runner_tokens_job_runner_id" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "job_runner_job_runner_tokens_job_runner_token_id" FOREIGN KEY ("job_runner_token_id") REFERENCES "job_runner_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_blocked_groups" table
ALTER TABLE "mapped_control_blocked_groups" ADD CONSTRAINT "mapped_control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_blocked_groups_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_editors" table
ALTER TABLE "mapped_control_editors" ADD CONSTRAINT "mapped_control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_editors_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_from_controls" table
ALTER TABLE "mapped_control_from_controls" ADD CONSTRAINT "mapped_control_from_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_from_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_to_controls" table
ALTER TABLE "mapped_control_to_controls" ADD CONSTRAINT "mapped_control_to_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_to_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_from_subcontrols" table
ALTER TABLE "mapped_control_from_subcontrols" ADD CONSTRAINT "mapped_control_from_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_from_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_to_subcontrols" table
ALTER TABLE "mapped_control_to_subcontrols" ADD CONSTRAINT "mapped_control_to_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_to_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" ADD CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_editors" table
ALTER TABLE "narrative_editors" ADD CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" ADD CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_membership_events" table
ALTER TABLE "org_membership_events" ADD CONSTRAINT "org_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_membership_events_org_membership_id" FOREIGN KEY ("org_membership_id") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_module_org_prices" table
ALTER TABLE "org_module_org_prices" ADD CONSTRAINT "org_module_org_prices_org_module_id" FOREIGN KEY ("org_module_id") REFERENCES "org_modules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_module_org_prices_org_price_id" FOREIGN KEY ("org_price_id") REFERENCES "org_prices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_product_org_prices" table
ALTER TABLE "org_product_org_prices" ADD CONSTRAINT "org_product_org_prices_org_price_id" FOREIGN KEY ("org_price_id") REFERENCES "org_prices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_product_org_prices_org_product_id" FOREIGN KEY ("org_product_id") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_subscription_events" table
ALTER TABLE "org_subscription_events" ADD CONSTRAINT "org_subscription_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_subscription_events_org_subscription_id" FOREIGN KEY ("org_subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_personal_access_tokens" table
ALTER TABLE "organization_personal_access_tokens" ADD CONSTRAINT "organization_personal_access_tokens_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_personal_access_tokens_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_files" table
ALTER TABLE "organization_files" ADD CONSTRAINT "organization_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_files_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_events" table
ALTER TABLE "organization_events" ADD CONSTRAINT "organization_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_events_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_setting_files" table
ALTER TABLE "organization_setting_files" ADD CONSTRAINT "organization_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_setting_files_organization_setting_id" FOREIGN KEY ("organization_setting_id") REFERENCES "organization_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "personal_access_token_events" table
ALTER TABLE "personal_access_token_events" ADD CONSTRAINT "personal_access_token_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "personal_access_token_events_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_blocked_groups" table
ALTER TABLE "platform_blocked_groups" ADD CONSTRAINT "platform_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_blocked_groups_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_editors" table
ALTER TABLE "platform_editors" ADD CONSTRAINT "platform_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_editors_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_viewers" table
ALTER TABLE "platform_viewers" ADD CONSTRAINT "platform_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_viewers_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_assets" table
ALTER TABLE "platform_assets" ADD CONSTRAINT "platform_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_assets_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_entities" table
ALTER TABLE "platform_entities" ADD CONSTRAINT "platform_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_entities_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_evidence" table
ALTER TABLE "platform_evidence" ADD CONSTRAINT "platform_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_evidence_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_files" table
ALTER TABLE "platform_files" ADD CONSTRAINT "platform_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_files_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_risks" table
ALTER TABLE "platform_risks" ADD CONSTRAINT "platform_risks_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_controls" table
ALTER TABLE "platform_controls" ADD CONSTRAINT "platform_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_controls_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_assessments" table
ALTER TABLE "platform_assessments" ADD CONSTRAINT "platform_assessments_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_assessments_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_scans" table
ALTER TABLE "platform_scans" ADD CONSTRAINT "platform_scans_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_tasks" table
ALTER TABLE "platform_tasks" ADD CONSTRAINT "platform_tasks_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_identity_holders" table
ALTER TABLE "platform_identity_holders" ADD CONSTRAINT "platform_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_identity_holders_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_source_entities" table
ALTER TABLE "platform_source_entities" ADD CONSTRAINT "platform_source_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_source_entities_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_out_of_scope_assets" table
ALTER TABLE "platform_out_of_scope_assets" ADD CONSTRAINT "platform_out_of_scope_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_out_of_scope_assets_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_out_of_scope_vendors" table
ALTER TABLE "platform_out_of_scope_vendors" ADD CONSTRAINT "platform_out_of_scope_vendors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_out_of_scope_vendors_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_applicable_frameworks" table
ALTER TABLE "platform_applicable_frameworks" ADD CONSTRAINT "platform_applicable_frameworks_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_applicable_frameworks_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "platform_system_details" table
ALTER TABLE "platform_system_details" ADD CONSTRAINT "platform_system_details_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "platform_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_blocked_groups" table
ALTER TABLE "procedure_blocked_groups" ADD CONSTRAINT "procedure_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_blocked_groups_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_editors" table
ALTER TABLE "procedure_editors" ADD CONSTRAINT "procedure_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_editors_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_narratives" table
ALTER TABLE "procedure_narratives" ADD CONSTRAINT "procedure_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_narratives_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_risks" table
ALTER TABLE "procedure_risks" ADD CONSTRAINT "procedure_risks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_tasks" table
ALTER TABLE "procedure_tasks" ADD CONSTRAINT "procedure_tasks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_blocked_groups" table
ALTER TABLE "program_blocked_groups" ADD CONSTRAINT "program_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_blocked_groups_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_editors" table
ALTER TABLE "program_editors" ADD CONSTRAINT "program_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_editors_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_viewers" table
ALTER TABLE "program_viewers" ADD CONSTRAINT "program_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_viewers_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_controls" table
ALTER TABLE "program_controls" ADD CONSTRAINT "program_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_controls_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_control_objectives" table
ALTER TABLE "program_control_objectives" ADD CONSTRAINT "program_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_control_objectives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_internal_policies" table
ALTER TABLE "program_internal_policies" ADD CONSTRAINT "program_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_internal_policies_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_procedures" table
ALTER TABLE "program_procedures" ADD CONSTRAINT "program_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_procedures_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_risks" table
ALTER TABLE "program_risks" ADD CONSTRAINT "program_risks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_tasks" table
ALTER TABLE "program_tasks" ADD CONSTRAINT "program_tasks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_files" table
ALTER TABLE "program_files" ADD CONSTRAINT "program_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_files_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_evidence" table
ALTER TABLE "program_evidence" ADD CONSTRAINT "program_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_evidence_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_narratives" table
ALTER TABLE "program_narratives" ADD CONSTRAINT "program_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_narratives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_action_plans" table
ALTER TABLE "program_action_plans" ADD CONSTRAINT "program_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_action_plans_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_system_details" table
ALTER TABLE "program_system_details" ADD CONSTRAINT "program_system_details_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_blocked_groups" table
ALTER TABLE "remediation_blocked_groups" ADD CONSTRAINT "remediation_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_blocked_groups_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_editors" table
ALTER TABLE "remediation_editors" ADD CONSTRAINT "remediation_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_editors_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_findings" table
ALTER TABLE "remediation_findings" ADD CONSTRAINT "remediation_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_findings_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_vulnerabilities" table
ALTER TABLE "remediation_vulnerabilities" ADD CONSTRAINT "remediation_vulnerabilities_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_action_plans" table
ALTER TABLE "remediation_action_plans" ADD CONSTRAINT "remediation_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_action_plans_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_controls" table
ALTER TABLE "remediation_controls" ADD CONSTRAINT "remediation_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_controls_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_subcontrols" table
ALTER TABLE "remediation_subcontrols" ADD CONSTRAINT "remediation_subcontrols_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_risks" table
ALTER TABLE "remediation_risks" ADD CONSTRAINT "remediation_risks_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_programs" table
ALTER TABLE "remediation_programs" ADD CONSTRAINT "remediation_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_programs_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_assets" table
ALTER TABLE "remediation_assets" ADD CONSTRAINT "remediation_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_assets_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_entities" table
ALTER TABLE "remediation_entities" ADD CONSTRAINT "remediation_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_entities_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_blocked_groups" table
ALTER TABLE "review_blocked_groups" ADD CONSTRAINT "review_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_blocked_groups_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_editors" table
ALTER TABLE "review_editors" ADD CONSTRAINT "review_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_editors_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_findings" table
ALTER TABLE "review_findings" ADD CONSTRAINT "review_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_findings_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_vulnerabilities" table
ALTER TABLE "review_vulnerabilities" ADD CONSTRAINT "review_vulnerabilities_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_action_plans" table
ALTER TABLE "review_action_plans" ADD CONSTRAINT "review_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_action_plans_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_remediations" table
ALTER TABLE "review_remediations" ADD CONSTRAINT "review_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_remediations_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_controls" table
ALTER TABLE "review_controls" ADD CONSTRAINT "review_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_controls_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_subcontrols" table
ALTER TABLE "review_subcontrols" ADD CONSTRAINT "review_subcontrols_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_risks" table
ALTER TABLE "review_risks" ADD CONSTRAINT "review_risks_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_programs" table
ALTER TABLE "review_programs" ADD CONSTRAINT "review_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_programs_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_assets" table
ALTER TABLE "review_assets" ADD CONSTRAINT "review_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_assets_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_entities" table
ALTER TABLE "review_entities" ADD CONSTRAINT "review_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_entities_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_internal_policies" table
ALTER TABLE "review_internal_policies" ADD CONSTRAINT "review_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_internal_policies_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" ADD CONSTRAINT "risk_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_blocked_groups_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_editors" table
ALTER TABLE "risk_editors" ADD CONSTRAINT "risk_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_editors_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_viewers" table
ALTER TABLE "risk_viewers" ADD CONSTRAINT "risk_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_viewers_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_action_plans" table
ALTER TABLE "risk_action_plans" ADD CONSTRAINT "risk_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_action_plans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_tasks" table
ALTER TABLE "risk_tasks" ADD CONSTRAINT "risk_tasks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_blocked_groups" table
ALTER TABLE "scan_blocked_groups" ADD CONSTRAINT "scan_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_blocked_groups_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_editors" table
ALTER TABLE "scan_editors" ADD CONSTRAINT "scan_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_editors_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_assets" table
ALTER TABLE "scan_assets" ADD CONSTRAINT "scan_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_assets_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_entities" table
ALTER TABLE "scan_entities" ADD CONSTRAINT "scan_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_entities_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_evidence" table
ALTER TABLE "scan_evidence" ADD CONSTRAINT "scan_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_evidence_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_files" table
ALTER TABLE "scan_files" ADD CONSTRAINT "scan_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_files_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_remediations" table
ALTER TABLE "scan_remediations" ADD CONSTRAINT "scan_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_remediations_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_action_plans" table
ALTER TABLE "scan_action_plans" ADD CONSTRAINT "scan_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_action_plans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_tasks" table
ALTER TABLE "scan_tasks" ADD CONSTRAINT "scan_tasks_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scheduled_job_controls" table
ALTER TABLE "scheduled_job_controls" ADD CONSTRAINT "scheduled_job_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scheduled_job_controls_scheduled_job_id" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scheduled_job_subcontrols" table
ALTER TABLE "scheduled_job_subcontrols" ADD CONSTRAINT "scheduled_job_subcontrols_scheduled_job_id" FOREIGN KEY ("scheduled_job_id") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scheduled_job_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" ADD CONSTRAINT "subcontrol_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_objectives_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" ADD CONSTRAINT "subcontrol_tasks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_risks" table
ALTER TABLE "subcontrol_risks" ADD CONSTRAINT "subcontrol_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_risks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_procedures" table
ALTER TABLE "subcontrol_procedures" ADD CONSTRAINT "subcontrol_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_procedures_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_scans" table
ALTER TABLE "subcontrol_scans" ADD CONSTRAINT "subcontrol_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_scans_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_control_implementations" table
ALTER TABLE "subcontrol_control_implementations" ADD CONSTRAINT "subcontrol_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_implementations_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_assets" table
ALTER TABLE "subcontrol_assets" ADD CONSTRAINT "subcontrol_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_assets_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_entities" table
ALTER TABLE "subcontrol_entities" ADD CONSTRAINT "subcontrol_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_entities_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_identity_holders" table
ALTER TABLE "subcontrol_identity_holders" ADD CONSTRAINT "subcontrol_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_identity_holders_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subscriber_events" table
ALTER TABLE "subscriber_events" ADD CONSTRAINT "subscriber_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subscriber_events_subscriber_id" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "system_detail_assets" table
ALTER TABLE "system_detail_assets" ADD CONSTRAINT "system_detail_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "system_detail_assets_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "task_evidence" table
ALTER TABLE "task_evidence" ADD CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "template_files" table
ALTER TABLE "template_files" ADD CONSTRAINT "template_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "template_files_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_events" table
ALTER TABLE "user_events" ADD CONSTRAINT "user_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_action_plans" table
ALTER TABLE "vulnerability_action_plans" ADD CONSTRAINT "vulnerability_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_action_plans_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_controls" table
ALTER TABLE "vulnerability_controls" ADD CONSTRAINT "vulnerability_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_controls_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_subcontrols" table
ALTER TABLE "vulnerability_subcontrols" ADD CONSTRAINT "vulnerability_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_subcontrols_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_risks" table
ALTER TABLE "vulnerability_risks" ADD CONSTRAINT "vulnerability_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_risks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_programs" table
ALTER TABLE "vulnerability_programs" ADD CONSTRAINT "vulnerability_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_programs_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_assets" table
ALTER TABLE "vulnerability_assets" ADD CONSTRAINT "vulnerability_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_assets_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_entities" table
ALTER TABLE "vulnerability_entities" ADD CONSTRAINT "vulnerability_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_entities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_scans" table
ALTER TABLE "vulnerability_scans" ADD CONSTRAINT "vulnerability_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_scans_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_tasks" table
ALTER TABLE "vulnerability_tasks" ADD CONSTRAINT "vulnerability_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_tasks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "vulnerability_tasks" table
ALTER TABLE "vulnerability_tasks" DROP CONSTRAINT "vulnerability_tasks_vulnerability_id", DROP CONSTRAINT "vulnerability_tasks_task_id";
-- reverse: modify "vulnerability_scans" table
ALTER TABLE "vulnerability_scans" DROP CONSTRAINT "vulnerability_scans_vulnerability_id", DROP CONSTRAINT "vulnerability_scans_scan_id";
-- reverse: modify "vulnerability_entities" table
ALTER TABLE "vulnerability_entities" DROP CONSTRAINT "vulnerability_entities_vulnerability_id", DROP CONSTRAINT "vulnerability_entities_entity_id";
-- reverse: modify "vulnerability_assets" table
ALTER TABLE "vulnerability_assets" DROP CONSTRAINT "vulnerability_assets_vulnerability_id", DROP CONSTRAINT "vulnerability_assets_asset_id";
-- reverse: modify "vulnerability_programs" table
ALTER TABLE "vulnerability_programs" DROP CONSTRAINT "vulnerability_programs_vulnerability_id", DROP CONSTRAINT "vulnerability_programs_program_id";
-- reverse: modify "vulnerability_risks" table
ALTER TABLE "vulnerability_risks" DROP CONSTRAINT "vulnerability_risks_vulnerability_id", DROP CONSTRAINT "vulnerability_risks_risk_id";
-- reverse: modify "vulnerability_subcontrols" table
ALTER TABLE "vulnerability_subcontrols" DROP CONSTRAINT "vulnerability_subcontrols_vulnerability_id", DROP CONSTRAINT "vulnerability_subcontrols_subcontrol_id";
-- reverse: modify "vulnerability_controls" table
ALTER TABLE "vulnerability_controls" DROP CONSTRAINT "vulnerability_controls_vulnerability_id", DROP CONSTRAINT "vulnerability_controls_control_id";
-- reverse: modify "vulnerability_action_plans" table
ALTER TABLE "vulnerability_action_plans" DROP CONSTRAINT "vulnerability_action_plans_vulnerability_id", DROP CONSTRAINT "vulnerability_action_plans_action_plan_id";
-- reverse: modify "user_events" table
ALTER TABLE "user_events" DROP CONSTRAINT "user_events_user_id", DROP CONSTRAINT "user_events_event_id";
-- reverse: modify "template_files" table
ALTER TABLE "template_files" DROP CONSTRAINT "template_files_template_id", DROP CONSTRAINT "template_files_file_id";
-- reverse: modify "task_evidence" table
ALTER TABLE "task_evidence" DROP CONSTRAINT "task_evidence_task_id", DROP CONSTRAINT "task_evidence_evidence_id";
-- reverse: modify "system_detail_assets" table
ALTER TABLE "system_detail_assets" DROP CONSTRAINT "system_detail_assets_system_detail_id", DROP CONSTRAINT "system_detail_assets_asset_id";
-- reverse: modify "subscriber_events" table
ALTER TABLE "subscriber_events" DROP CONSTRAINT "subscriber_events_subscriber_id", DROP CONSTRAINT "subscriber_events_event_id";
-- reverse: modify "subcontrol_identity_holders" table
ALTER TABLE "subcontrol_identity_holders" DROP CONSTRAINT "subcontrol_identity_holders_subcontrol_id", DROP CONSTRAINT "subcontrol_identity_holders_identity_holder_id";
-- reverse: modify "subcontrol_entities" table
ALTER TABLE "subcontrol_entities" DROP CONSTRAINT "subcontrol_entities_subcontrol_id", DROP CONSTRAINT "subcontrol_entities_entity_id";
-- reverse: modify "subcontrol_assets" table
ALTER TABLE "subcontrol_assets" DROP CONSTRAINT "subcontrol_assets_subcontrol_id", DROP CONSTRAINT "subcontrol_assets_asset_id";
-- reverse: modify "subcontrol_control_implementations" table
ALTER TABLE "subcontrol_control_implementations" DROP CONSTRAINT "subcontrol_control_implementations_subcontrol_id", DROP CONSTRAINT "subcontrol_control_implementations_control_implementation_id";
-- reverse: modify "subcontrol_scans" table
ALTER TABLE "subcontrol_scans" DROP CONSTRAINT "subcontrol_scans_subcontrol_id", DROP CONSTRAINT "subcontrol_scans_scan_id";
-- reverse: modify "subcontrol_procedures" table
ALTER TABLE "subcontrol_procedures" DROP CONSTRAINT "subcontrol_procedures_subcontrol_id", DROP CONSTRAINT "subcontrol_procedures_procedure_id";
-- reverse: modify "subcontrol_risks" table
ALTER TABLE "subcontrol_risks" DROP CONSTRAINT "subcontrol_risks_subcontrol_id", DROP CONSTRAINT "subcontrol_risks_risk_id";
-- reverse: modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" DROP CONSTRAINT "subcontrol_tasks_task_id", DROP CONSTRAINT "subcontrol_tasks_subcontrol_id";
-- reverse: modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" DROP CONSTRAINT "subcontrol_control_objectives_subcontrol_id", DROP CONSTRAINT "subcontrol_control_objectives_control_objective_id";
-- reverse: modify "scheduled_job_subcontrols" table
ALTER TABLE "scheduled_job_subcontrols" DROP CONSTRAINT "scheduled_job_subcontrols_subcontrol_id", DROP CONSTRAINT "scheduled_job_subcontrols_scheduled_job_id";
-- reverse: modify "scheduled_job_controls" table
ALTER TABLE "scheduled_job_controls" DROP CONSTRAINT "scheduled_job_controls_scheduled_job_id", DROP CONSTRAINT "scheduled_job_controls_control_id";
-- reverse: modify "scan_tasks" table
ALTER TABLE "scan_tasks" DROP CONSTRAINT "scan_tasks_task_id", DROP CONSTRAINT "scan_tasks_scan_id";
-- reverse: modify "scan_action_plans" table
ALTER TABLE "scan_action_plans" DROP CONSTRAINT "scan_action_plans_scan_id", DROP CONSTRAINT "scan_action_plans_action_plan_id";
-- reverse: modify "scan_remediations" table
ALTER TABLE "scan_remediations" DROP CONSTRAINT "scan_remediations_scan_id", DROP CONSTRAINT "scan_remediations_remediation_id";
-- reverse: modify "scan_files" table
ALTER TABLE "scan_files" DROP CONSTRAINT "scan_files_scan_id", DROP CONSTRAINT "scan_files_file_id";
-- reverse: modify "scan_evidence" table
ALTER TABLE "scan_evidence" DROP CONSTRAINT "scan_evidence_scan_id", DROP CONSTRAINT "scan_evidence_evidence_id";
-- reverse: modify "scan_entities" table
ALTER TABLE "scan_entities" DROP CONSTRAINT "scan_entities_scan_id", DROP CONSTRAINT "scan_entities_entity_id";
-- reverse: modify "scan_assets" table
ALTER TABLE "scan_assets" DROP CONSTRAINT "scan_assets_scan_id", DROP CONSTRAINT "scan_assets_asset_id";
-- reverse: modify "scan_editors" table
ALTER TABLE "scan_editors" DROP CONSTRAINT "scan_editors_scan_id", DROP CONSTRAINT "scan_editors_group_id";
-- reverse: modify "scan_blocked_groups" table
ALTER TABLE "scan_blocked_groups" DROP CONSTRAINT "scan_blocked_groups_scan_id", DROP CONSTRAINT "scan_blocked_groups_group_id";
-- reverse: modify "risk_tasks" table
ALTER TABLE "risk_tasks" DROP CONSTRAINT "risk_tasks_task_id", DROP CONSTRAINT "risk_tasks_risk_id";
-- reverse: modify "risk_action_plans" table
ALTER TABLE "risk_action_plans" DROP CONSTRAINT "risk_action_plans_risk_id", DROP CONSTRAINT "risk_action_plans_action_plan_id";
-- reverse: modify "risk_viewers" table
ALTER TABLE "risk_viewers" DROP CONSTRAINT "risk_viewers_risk_id", DROP CONSTRAINT "risk_viewers_group_id";
-- reverse: modify "risk_editors" table
ALTER TABLE "risk_editors" DROP CONSTRAINT "risk_editors_risk_id", DROP CONSTRAINT "risk_editors_group_id";
-- reverse: modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" DROP CONSTRAINT "risk_blocked_groups_risk_id", DROP CONSTRAINT "risk_blocked_groups_group_id";
-- reverse: modify "review_internal_policies" table
ALTER TABLE "review_internal_policies" DROP CONSTRAINT "review_internal_policies_review_id", DROP CONSTRAINT "review_internal_policies_internal_policy_id";
-- reverse: modify "review_entities" table
ALTER TABLE "review_entities" DROP CONSTRAINT "review_entities_review_id", DROP CONSTRAINT "review_entities_entity_id";
-- reverse: modify "review_assets" table
ALTER TABLE "review_assets" DROP CONSTRAINT "review_assets_review_id", DROP CONSTRAINT "review_assets_asset_id";
-- reverse: modify "review_programs" table
ALTER TABLE "review_programs" DROP CONSTRAINT "review_programs_review_id", DROP CONSTRAINT "review_programs_program_id";
-- reverse: modify "review_risks" table
ALTER TABLE "review_risks" DROP CONSTRAINT "review_risks_risk_id", DROP CONSTRAINT "review_risks_review_id";
-- reverse: modify "review_subcontrols" table
ALTER TABLE "review_subcontrols" DROP CONSTRAINT "review_subcontrols_subcontrol_id", DROP CONSTRAINT "review_subcontrols_review_id";
-- reverse: modify "review_controls" table
ALTER TABLE "review_controls" DROP CONSTRAINT "review_controls_review_id", DROP CONSTRAINT "review_controls_control_id";
-- reverse: modify "review_remediations" table
ALTER TABLE "review_remediations" DROP CONSTRAINT "review_remediations_review_id", DROP CONSTRAINT "review_remediations_remediation_id";
-- reverse: modify "review_action_plans" table
ALTER TABLE "review_action_plans" DROP CONSTRAINT "review_action_plans_review_id", DROP CONSTRAINT "review_action_plans_action_plan_id";
-- reverse: modify "review_vulnerabilities" table
ALTER TABLE "review_vulnerabilities" DROP CONSTRAINT "review_vulnerabilities_vulnerability_id", DROP CONSTRAINT "review_vulnerabilities_review_id";
-- reverse: modify "review_findings" table
ALTER TABLE "review_findings" DROP CONSTRAINT "review_findings_review_id", DROP CONSTRAINT "review_findings_finding_id";
-- reverse: modify "review_editors" table
ALTER TABLE "review_editors" DROP CONSTRAINT "review_editors_review_id", DROP CONSTRAINT "review_editors_group_id";
-- reverse: modify "review_blocked_groups" table
ALTER TABLE "review_blocked_groups" DROP CONSTRAINT "review_blocked_groups_review_id", DROP CONSTRAINT "review_blocked_groups_group_id";
-- reverse: modify "remediation_entities" table
ALTER TABLE "remediation_entities" DROP CONSTRAINT "remediation_entities_remediation_id", DROP CONSTRAINT "remediation_entities_entity_id";
-- reverse: modify "remediation_assets" table
ALTER TABLE "remediation_assets" DROP CONSTRAINT "remediation_assets_remediation_id", DROP CONSTRAINT "remediation_assets_asset_id";
-- reverse: modify "remediation_programs" table
ALTER TABLE "remediation_programs" DROP CONSTRAINT "remediation_programs_remediation_id", DROP CONSTRAINT "remediation_programs_program_id";
-- reverse: modify "remediation_risks" table
ALTER TABLE "remediation_risks" DROP CONSTRAINT "remediation_risks_risk_id", DROP CONSTRAINT "remediation_risks_remediation_id";
-- reverse: modify "remediation_subcontrols" table
ALTER TABLE "remediation_subcontrols" DROP CONSTRAINT "remediation_subcontrols_subcontrol_id", DROP CONSTRAINT "remediation_subcontrols_remediation_id";
-- reverse: modify "remediation_controls" table
ALTER TABLE "remediation_controls" DROP CONSTRAINT "remediation_controls_remediation_id", DROP CONSTRAINT "remediation_controls_control_id";
-- reverse: modify "remediation_action_plans" table
ALTER TABLE "remediation_action_plans" DROP CONSTRAINT "remediation_action_plans_remediation_id", DROP CONSTRAINT "remediation_action_plans_action_plan_id";
-- reverse: modify "remediation_vulnerabilities" table
ALTER TABLE "remediation_vulnerabilities" DROP CONSTRAINT "remediation_vulnerabilities_vulnerability_id", DROP CONSTRAINT "remediation_vulnerabilities_remediation_id";
-- reverse: modify "remediation_findings" table
ALTER TABLE "remediation_findings" DROP CONSTRAINT "remediation_findings_remediation_id", DROP CONSTRAINT "remediation_findings_finding_id";
-- reverse: modify "remediation_editors" table
ALTER TABLE "remediation_editors" DROP CONSTRAINT "remediation_editors_remediation_id", DROP CONSTRAINT "remediation_editors_group_id";
-- reverse: modify "remediation_blocked_groups" table
ALTER TABLE "remediation_blocked_groups" DROP CONSTRAINT "remediation_blocked_groups_remediation_id", DROP CONSTRAINT "remediation_blocked_groups_group_id";
-- reverse: modify "program_system_details" table
ALTER TABLE "program_system_details" DROP CONSTRAINT "program_system_details_system_detail_id", DROP CONSTRAINT "program_system_details_program_id";
-- reverse: modify "program_action_plans" table
ALTER TABLE "program_action_plans" DROP CONSTRAINT "program_action_plans_program_id", DROP CONSTRAINT "program_action_plans_action_plan_id";
-- reverse: modify "program_narratives" table
ALTER TABLE "program_narratives" DROP CONSTRAINT "program_narratives_program_id", DROP CONSTRAINT "program_narratives_narrative_id";
-- reverse: modify "program_evidence" table
ALTER TABLE "program_evidence" DROP CONSTRAINT "program_evidence_program_id", DROP CONSTRAINT "program_evidence_evidence_id";
-- reverse: modify "program_files" table
ALTER TABLE "program_files" DROP CONSTRAINT "program_files_program_id", DROP CONSTRAINT "program_files_file_id";
-- reverse: modify "program_tasks" table
ALTER TABLE "program_tasks" DROP CONSTRAINT "program_tasks_task_id", DROP CONSTRAINT "program_tasks_program_id";
-- reverse: modify "program_risks" table
ALTER TABLE "program_risks" DROP CONSTRAINT "program_risks_risk_id", DROP CONSTRAINT "program_risks_program_id";
-- reverse: modify "program_procedures" table
ALTER TABLE "program_procedures" DROP CONSTRAINT "program_procedures_program_id", DROP CONSTRAINT "program_procedures_procedure_id";
-- reverse: modify "program_internal_policies" table
ALTER TABLE "program_internal_policies" DROP CONSTRAINT "program_internal_policies_program_id", DROP CONSTRAINT "program_internal_policies_internal_policy_id";
-- reverse: modify "program_control_objectives" table
ALTER TABLE "program_control_objectives" DROP CONSTRAINT "program_control_objectives_program_id", DROP CONSTRAINT "program_control_objectives_control_objective_id";
-- reverse: modify "program_controls" table
ALTER TABLE "program_controls" DROP CONSTRAINT "program_controls_program_id", DROP CONSTRAINT "program_controls_control_id";
-- reverse: modify "program_viewers" table
ALTER TABLE "program_viewers" DROP CONSTRAINT "program_viewers_program_id", DROP CONSTRAINT "program_viewers_group_id";
-- reverse: modify "program_editors" table
ALTER TABLE "program_editors" DROP CONSTRAINT "program_editors_program_id", DROP CONSTRAINT "program_editors_group_id";
-- reverse: modify "program_blocked_groups" table
ALTER TABLE "program_blocked_groups" DROP CONSTRAINT "program_blocked_groups_program_id", DROP CONSTRAINT "program_blocked_groups_group_id";
-- reverse: modify "procedure_tasks" table
ALTER TABLE "procedure_tasks" DROP CONSTRAINT "procedure_tasks_task_id", DROP CONSTRAINT "procedure_tasks_procedure_id";
-- reverse: modify "procedure_risks" table
ALTER TABLE "procedure_risks" DROP CONSTRAINT "procedure_risks_risk_id", DROP CONSTRAINT "procedure_risks_procedure_id";
-- reverse: modify "procedure_narratives" table
ALTER TABLE "procedure_narratives" DROP CONSTRAINT "procedure_narratives_procedure_id", DROP CONSTRAINT "procedure_narratives_narrative_id";
-- reverse: modify "procedure_editors" table
ALTER TABLE "procedure_editors" DROP CONSTRAINT "procedure_editors_procedure_id", DROP CONSTRAINT "procedure_editors_group_id";
-- reverse: modify "procedure_blocked_groups" table
ALTER TABLE "procedure_blocked_groups" DROP CONSTRAINT "procedure_blocked_groups_procedure_id", DROP CONSTRAINT "procedure_blocked_groups_group_id";
-- reverse: modify "platform_system_details" table
ALTER TABLE "platform_system_details" DROP CONSTRAINT "platform_system_details_system_detail_id", DROP CONSTRAINT "platform_system_details_platform_id";
-- reverse: modify "platform_applicable_frameworks" table
ALTER TABLE "platform_applicable_frameworks" DROP CONSTRAINT "platform_applicable_frameworks_standard_id", DROP CONSTRAINT "platform_applicable_frameworks_platform_id";
-- reverse: modify "platform_out_of_scope_vendors" table
ALTER TABLE "platform_out_of_scope_vendors" DROP CONSTRAINT "platform_out_of_scope_vendors_platform_id", DROP CONSTRAINT "platform_out_of_scope_vendors_entity_id";
-- reverse: modify "platform_out_of_scope_assets" table
ALTER TABLE "platform_out_of_scope_assets" DROP CONSTRAINT "platform_out_of_scope_assets_platform_id", DROP CONSTRAINT "platform_out_of_scope_assets_asset_id";
-- reverse: modify "platform_source_entities" table
ALTER TABLE "platform_source_entities" DROP CONSTRAINT "platform_source_entities_platform_id", DROP CONSTRAINT "platform_source_entities_entity_id";
-- reverse: modify "platform_identity_holders" table
ALTER TABLE "platform_identity_holders" DROP CONSTRAINT "platform_identity_holders_platform_id", DROP CONSTRAINT "platform_identity_holders_identity_holder_id";
-- reverse: modify "platform_tasks" table
ALTER TABLE "platform_tasks" DROP CONSTRAINT "platform_tasks_task_id", DROP CONSTRAINT "platform_tasks_platform_id";
-- reverse: modify "platform_scans" table
ALTER TABLE "platform_scans" DROP CONSTRAINT "platform_scans_scan_id", DROP CONSTRAINT "platform_scans_platform_id";
-- reverse: modify "platform_assessments" table
ALTER TABLE "platform_assessments" DROP CONSTRAINT "platform_assessments_platform_id", DROP CONSTRAINT "platform_assessments_assessment_id";
-- reverse: modify "platform_controls" table
ALTER TABLE "platform_controls" DROP CONSTRAINT "platform_controls_platform_id", DROP CONSTRAINT "platform_controls_control_id";
-- reverse: modify "platform_risks" table
ALTER TABLE "platform_risks" DROP CONSTRAINT "platform_risks_risk_id", DROP CONSTRAINT "platform_risks_platform_id";
-- reverse: modify "platform_files" table
ALTER TABLE "platform_files" DROP CONSTRAINT "platform_files_platform_id", DROP CONSTRAINT "platform_files_file_id";
-- reverse: modify "platform_evidence" table
ALTER TABLE "platform_evidence" DROP CONSTRAINT "platform_evidence_platform_id", DROP CONSTRAINT "platform_evidence_evidence_id";
-- reverse: modify "platform_entities" table
ALTER TABLE "platform_entities" DROP CONSTRAINT "platform_entities_platform_id", DROP CONSTRAINT "platform_entities_entity_id";
-- reverse: modify "platform_assets" table
ALTER TABLE "platform_assets" DROP CONSTRAINT "platform_assets_platform_id", DROP CONSTRAINT "platform_assets_asset_id";
-- reverse: modify "platform_viewers" table
ALTER TABLE "platform_viewers" DROP CONSTRAINT "platform_viewers_platform_id", DROP CONSTRAINT "platform_viewers_group_id";
-- reverse: modify "platform_editors" table
ALTER TABLE "platform_editors" DROP CONSTRAINT "platform_editors_platform_id", DROP CONSTRAINT "platform_editors_group_id";
-- reverse: modify "platform_blocked_groups" table
ALTER TABLE "platform_blocked_groups" DROP CONSTRAINT "platform_blocked_groups_platform_id", DROP CONSTRAINT "platform_blocked_groups_group_id";
-- reverse: modify "personal_access_token_events" table
ALTER TABLE "personal_access_token_events" DROP CONSTRAINT "personal_access_token_events_personal_access_token_id", DROP CONSTRAINT "personal_access_token_events_event_id";
-- reverse: modify "organization_setting_files" table
ALTER TABLE "organization_setting_files" DROP CONSTRAINT "organization_setting_files_organization_setting_id", DROP CONSTRAINT "organization_setting_files_file_id";
-- reverse: modify "organization_events" table
ALTER TABLE "organization_events" DROP CONSTRAINT "organization_events_organization_id", DROP CONSTRAINT "organization_events_event_id";
-- reverse: modify "organization_files" table
ALTER TABLE "organization_files" DROP CONSTRAINT "organization_files_organization_id", DROP CONSTRAINT "organization_files_file_id";
-- reverse: modify "organization_personal_access_tokens" table
ALTER TABLE "organization_personal_access_tokens" DROP CONSTRAINT "organization_personal_access_tokens_personal_access_token_id", DROP CONSTRAINT "organization_personal_access_tokens_organization_id";
-- reverse: modify "org_subscription_events" table
ALTER TABLE "org_subscription_events" DROP CONSTRAINT "org_subscription_events_org_subscription_id", DROP CONSTRAINT "org_subscription_events_event_id";
-- reverse: modify "org_product_org_prices" table
ALTER TABLE "org_product_org_prices" DROP CONSTRAINT "org_product_org_prices_org_product_id", DROP CONSTRAINT "org_product_org_prices_org_price_id";
-- reverse: modify "org_module_org_prices" table
ALTER TABLE "org_module_org_prices" DROP CONSTRAINT "org_module_org_prices_org_price_id", DROP CONSTRAINT "org_module_org_prices_org_module_id";
-- reverse: modify "org_membership_events" table
ALTER TABLE "org_membership_events" DROP CONSTRAINT "org_membership_events_org_membership_id", DROP CONSTRAINT "org_membership_events_event_id";
-- reverse: modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" DROP CONSTRAINT "narrative_viewers_narrative_id", DROP CONSTRAINT "narrative_viewers_group_id";
-- reverse: modify "narrative_editors" table
ALTER TABLE "narrative_editors" DROP CONSTRAINT "narrative_editors_narrative_id", DROP CONSTRAINT "narrative_editors_group_id";
-- reverse: modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" DROP CONSTRAINT "narrative_blocked_groups_narrative_id", DROP CONSTRAINT "narrative_blocked_groups_group_id";
-- reverse: modify "mapped_control_to_subcontrols" table
ALTER TABLE "mapped_control_to_subcontrols" DROP CONSTRAINT "mapped_control_to_subcontrols_subcontrol_id", DROP CONSTRAINT "mapped_control_to_subcontrols_mapped_control_id";
-- reverse: modify "mapped_control_from_subcontrols" table
ALTER TABLE "mapped_control_from_subcontrols" DROP CONSTRAINT "mapped_control_from_subcontrols_subcontrol_id", DROP CONSTRAINT "mapped_control_from_subcontrols_mapped_control_id";
-- reverse: modify "mapped_control_to_controls" table
ALTER TABLE "mapped_control_to_controls" DROP CONSTRAINT "mapped_control_to_controls_mapped_control_id", DROP CONSTRAINT "mapped_control_to_controls_control_id";
-- reverse: modify "mapped_control_from_controls" table
ALTER TABLE "mapped_control_from_controls" DROP CONSTRAINT "mapped_control_from_controls_mapped_control_id", DROP CONSTRAINT "mapped_control_from_controls_control_id";
-- reverse: modify "mapped_control_editors" table
ALTER TABLE "mapped_control_editors" DROP CONSTRAINT "mapped_control_editors_mapped_control_id", DROP CONSTRAINT "mapped_control_editors_group_id";
-- reverse: modify "mapped_control_blocked_groups" table
ALTER TABLE "mapped_control_blocked_groups" DROP CONSTRAINT "mapped_control_blocked_groups_mapped_control_id", DROP CONSTRAINT "mapped_control_blocked_groups_group_id";
-- reverse: modify "job_runner_job_runner_tokens" table
ALTER TABLE "job_runner_job_runner_tokens" DROP CONSTRAINT "job_runner_job_runner_tokens_job_runner_token_id", DROP CONSTRAINT "job_runner_job_runner_tokens_job_runner_id";
-- reverse: modify "invite_groups" table
ALTER TABLE "invite_groups" DROP CONSTRAINT "invite_groups_invite_id", DROP CONSTRAINT "invite_groups_group_id";
-- reverse: modify "invite_events" table
ALTER TABLE "invite_events" DROP CONSTRAINT "invite_events_invite_id", DROP CONSTRAINT "invite_events_event_id";
-- reverse: modify "internal_policy_identity_holders" table
ALTER TABLE "internal_policy_identity_holders" DROP CONSTRAINT "internal_policy_identity_holders_internal_policy_id", DROP CONSTRAINT "internal_policy_identity_holders_identity_holder_id";
-- reverse: modify "internal_policy_entities" table
ALTER TABLE "internal_policy_entities" DROP CONSTRAINT "internal_policy_entities_internal_policy_id", DROP CONSTRAINT "internal_policy_entities_entity_id";
-- reverse: modify "internal_policy_assets" table
ALTER TABLE "internal_policy_assets" DROP CONSTRAINT "internal_policy_assets_internal_policy_id", DROP CONSTRAINT "internal_policy_assets_asset_id";
-- reverse: modify "internal_policy_risks" table
ALTER TABLE "internal_policy_risks" DROP CONSTRAINT "internal_policy_risks_risk_id", DROP CONSTRAINT "internal_policy_risks_internal_policy_id";
-- reverse: modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" DROP CONSTRAINT "internal_policy_tasks_task_id", DROP CONSTRAINT "internal_policy_tasks_internal_policy_id";
-- reverse: modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" DROP CONSTRAINT "internal_policy_narratives_narrative_id", DROP CONSTRAINT "internal_policy_narratives_internal_policy_id";
-- reverse: modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" DROP CONSTRAINT "internal_policy_procedures_procedure_id", DROP CONSTRAINT "internal_policy_procedures_internal_policy_id";
-- reverse: modify "internal_policy_subcontrols" table
ALTER TABLE "internal_policy_subcontrols" DROP CONSTRAINT "internal_policy_subcontrols_subcontrol_id", DROP CONSTRAINT "internal_policy_subcontrols_internal_policy_id";
-- reverse: modify "internal_policy_controls" table
ALTER TABLE "internal_policy_controls" DROP CONSTRAINT "internal_policy_controls_internal_policy_id", DROP CONSTRAINT "internal_policy_controls_control_id";
-- reverse: modify "internal_policy_control_objectives" table
ALTER TABLE "internal_policy_control_objectives" DROP CONSTRAINT "internal_policy_control_objectives_internal_policy_id", DROP CONSTRAINT "internal_policy_control_objectives_control_objective_id";
-- reverse: modify "internal_policy_editors" table
ALTER TABLE "internal_policy_editors" DROP CONSTRAINT "internal_policy_editors_internal_policy_id", DROP CONSTRAINT "internal_policy_editors_group_id";
-- reverse: modify "internal_policy_blocked_groups" table
ALTER TABLE "internal_policy_blocked_groups" DROP CONSTRAINT "internal_policy_blocked_groups_internal_policy_id", DROP CONSTRAINT "internal_policy_blocked_groups_group_id";
-- reverse: modify "integration_action_plans" table
ALTER TABLE "integration_action_plans" DROP CONSTRAINT "integration_action_plans_integration_id", DROP CONSTRAINT "integration_action_plans_action_plan_id";
-- reverse: modify "integration_remediations" table
ALTER TABLE "integration_remediations" DROP CONSTRAINT "integration_remediations_remediation_id", DROP CONSTRAINT "integration_remediations_integration_id";
-- reverse: modify "integration_reviews" table
ALTER TABLE "integration_reviews" DROP CONSTRAINT "integration_reviews_review_id", DROP CONSTRAINT "integration_reviews_integration_id";
-- reverse: modify "integration_internal_policies" table
ALTER TABLE "integration_internal_policies" DROP CONSTRAINT "integration_internal_policies_internal_policy_id", DROP CONSTRAINT "integration_internal_policies_integration_id";
-- reverse: modify "integration_vulnerabilities" table
ALTER TABLE "integration_vulnerabilities" DROP CONSTRAINT "integration_vulnerabilities_vulnerability_id", DROP CONSTRAINT "integration_vulnerabilities_integration_id";
-- reverse: modify "integration_findings" table
ALTER TABLE "integration_findings" DROP CONSTRAINT "integration_findings_integration_id", DROP CONSTRAINT "integration_findings_finding_id";
-- reverse: modify "integration_events" table
ALTER TABLE "integration_events" DROP CONSTRAINT "integration_events_integration_id", DROP CONSTRAINT "integration_events_event_id";
-- reverse: modify "integration_secrets" table
ALTER TABLE "integration_secrets" DROP CONSTRAINT "integration_secrets_integration_id", DROP CONSTRAINT "integration_secrets_hush_id";
-- reverse: modify "identity_holder_files" table
ALTER TABLE "identity_holder_files" DROP CONSTRAINT "identity_holder_files_identity_holder_id", DROP CONSTRAINT "identity_holder_files_file_id";
-- reverse: modify "identity_holder_tasks" table
ALTER TABLE "identity_holder_tasks" DROP CONSTRAINT "identity_holder_tasks_task_id", DROP CONSTRAINT "identity_holder_tasks_identity_holder_id";
-- reverse: modify "identity_holder_entities" table
ALTER TABLE "identity_holder_entities" DROP CONSTRAINT "identity_holder_entities_identity_holder_id", DROP CONSTRAINT "identity_holder_entities_entity_id";
-- reverse: modify "identity_holder_assets" table
ALTER TABLE "identity_holder_assets" DROP CONSTRAINT "identity_holder_assets_identity_holder_id", DROP CONSTRAINT "identity_holder_assets_asset_id";
-- reverse: modify "identity_holder_templates" table
ALTER TABLE "identity_holder_templates" DROP CONSTRAINT "identity_holder_templates_template_id", DROP CONSTRAINT "identity_holder_templates_identity_holder_id";
-- reverse: modify "identity_holder_assessments" table
ALTER TABLE "identity_holder_assessments" DROP CONSTRAINT "identity_holder_assessments_identity_holder_id", DROP CONSTRAINT "identity_holder_assessments_assessment_id";
-- reverse: modify "hush_events" table
ALTER TABLE "hush_events" DROP CONSTRAINT "hush_events_hush_id", DROP CONSTRAINT "hush_events_event_id";
-- reverse: modify "group_membership_events" table
ALTER TABLE "group_membership_events" DROP CONSTRAINT "group_membership_events_group_membership_id", DROP CONSTRAINT "group_membership_events_event_id";
-- reverse: modify "group_tasks" table
ALTER TABLE "group_tasks" DROP CONSTRAINT "group_tasks_task_id", DROP CONSTRAINT "group_tasks_group_id";
-- reverse: modify "group_files" table
ALTER TABLE "group_files" DROP CONSTRAINT "group_files_group_id", DROP CONSTRAINT "group_files_file_id";
-- reverse: modify "group_events" table
ALTER TABLE "group_events" DROP CONSTRAINT "group_events_group_id", DROP CONSTRAINT "group_events_event_id";
-- reverse: modify "finding_check_results" table
ALTER TABLE "finding_check_results" DROP CONSTRAINT "finding_check_results_finding_id", DROP CONSTRAINT "finding_check_results_check_result_id";
-- reverse: modify "finding_identity_holders" table
ALTER TABLE "finding_identity_holders" DROP CONSTRAINT "finding_identity_holders_identity_holder_id", DROP CONSTRAINT "finding_identity_holders_finding_id";
-- reverse: modify "finding_directory_accounts" table
ALTER TABLE "finding_directory_accounts" DROP CONSTRAINT "finding_directory_accounts_finding_id", DROP CONSTRAINT "finding_directory_accounts_directory_account_id";
-- reverse: modify "finding_tasks" table
ALTER TABLE "finding_tasks" DROP CONSTRAINT "finding_tasks_task_id", DROP CONSTRAINT "finding_tasks_finding_id";
-- reverse: modify "finding_scans" table
ALTER TABLE "finding_scans" DROP CONSTRAINT "finding_scans_scan_id", DROP CONSTRAINT "finding_scans_finding_id";
-- reverse: modify "finding_entities" table
ALTER TABLE "finding_entities" DROP CONSTRAINT "finding_entities_finding_id", DROP CONSTRAINT "finding_entities_entity_id";
-- reverse: modify "finding_assets" table
ALTER TABLE "finding_assets" DROP CONSTRAINT "finding_assets_finding_id", DROP CONSTRAINT "finding_assets_asset_id";
-- reverse: modify "finding_programs" table
ALTER TABLE "finding_programs" DROP CONSTRAINT "finding_programs_program_id", DROP CONSTRAINT "finding_programs_finding_id";
-- reverse: modify "finding_risks" table
ALTER TABLE "finding_risks" DROP CONSTRAINT "finding_risks_risk_id", DROP CONSTRAINT "finding_risks_finding_id";
-- reverse: modify "finding_subcontrols" table
ALTER TABLE "finding_subcontrols" DROP CONSTRAINT "finding_subcontrols_subcontrol_id", DROP CONSTRAINT "finding_subcontrols_finding_id";
-- reverse: modify "finding_action_plans" table
ALTER TABLE "finding_action_plans" DROP CONSTRAINT "finding_action_plans_finding_id", DROP CONSTRAINT "finding_action_plans_action_plan_id";
-- reverse: modify "finding_vulnerabilities" table
ALTER TABLE "finding_vulnerabilities" DROP CONSTRAINT "finding_vulnerabilities_vulnerability_id", DROP CONSTRAINT "finding_vulnerabilities_finding_id";
-- reverse: modify "finding_editors" table
ALTER TABLE "finding_editors" DROP CONSTRAINT "finding_editors_group_id", DROP CONSTRAINT "finding_editors_finding_id";
-- reverse: modify "finding_blocked_groups" table
ALTER TABLE "finding_blocked_groups" DROP CONSTRAINT "finding_blocked_groups_group_id", DROP CONSTRAINT "finding_blocked_groups_finding_id";
-- reverse: modify "file_secrets" table
ALTER TABLE "file_secrets" DROP CONSTRAINT "file_secrets_hush_id", DROP CONSTRAINT "file_secrets_file_id";
-- reverse: modify "file_events" table
ALTER TABLE "file_events" DROP CONSTRAINT "file_events_file_id", DROP CONSTRAINT "file_events_event_id";
-- reverse: modify "evidence_files" table
ALTER TABLE "evidence_files" DROP CONSTRAINT "evidence_files_file_id", DROP CONSTRAINT "evidence_files_evidence_id";
-- reverse: modify "evidence_control_objectives" table
ALTER TABLE "evidence_control_objectives" DROP CONSTRAINT "evidence_control_objectives_evidence_id", DROP CONSTRAINT "evidence_control_objectives_control_objective_id";
-- reverse: modify "evidence_subcontrols" table
ALTER TABLE "evidence_subcontrols" DROP CONSTRAINT "evidence_subcontrols_subcontrol_id", DROP CONSTRAINT "evidence_subcontrols_evidence_id";
-- reverse: modify "evidence_controls" table
ALTER TABLE "evidence_controls" DROP CONSTRAINT "evidence_controls_evidence_id", DROP CONSTRAINT "evidence_controls_control_id";
-- reverse: modify "entity_subprocessors" table
ALTER TABLE "entity_subprocessors" DROP CONSTRAINT "entity_subprocessors_subprocessor_id", DROP CONSTRAINT "entity_subprocessors_entity_id";
-- reverse: modify "entity_integrations" table
ALTER TABLE "entity_integrations" DROP CONSTRAINT "entity_integrations_integration_id", DROP CONSTRAINT "entity_integrations_entity_id";
-- reverse: modify "entity_system_details" table
ALTER TABLE "entity_system_details" DROP CONSTRAINT "entity_system_details_system_detail_id", DROP CONSTRAINT "entity_system_details_entity_id";
-- reverse: modify "entity_assets" table
ALTER TABLE "entity_assets" DROP CONSTRAINT "entity_assets_entity_id", DROP CONSTRAINT "entity_assets_asset_id";
-- reverse: modify "entity_files" table
ALTER TABLE "entity_files" DROP CONSTRAINT "entity_files_file_id", DROP CONSTRAINT "entity_files_entity_id";
-- reverse: modify "entity_documents" table
ALTER TABLE "entity_documents" DROP CONSTRAINT "entity_documents_entity_id", DROP CONSTRAINT "entity_documents_document_data_id";
-- reverse: modify "entity_contacts" table
ALTER TABLE "entity_contacts" DROP CONSTRAINT "entity_contacts_entity_id", DROP CONSTRAINT "entity_contacts_contact_id";
-- reverse: modify "entity_editors" table
ALTER TABLE "entity_editors" DROP CONSTRAINT "entity_editors_group_id", DROP CONSTRAINT "entity_editors_entity_id";
-- reverse: modify "entity_blocked_groups" table
ALTER TABLE "entity_blocked_groups" DROP CONSTRAINT "entity_blocked_groups_group_id", DROP CONSTRAINT "entity_blocked_groups_entity_id";
-- reverse: modify "document_data_files" table
ALTER TABLE "document_data_files" DROP CONSTRAINT "document_data_files_file_id", DROP CONSTRAINT "document_data_files_document_data_id";
-- reverse: modify "control_objective_tasks" table
ALTER TABLE "control_objective_tasks" DROP CONSTRAINT "control_objective_tasks_task_id", DROP CONSTRAINT "control_objective_tasks_control_objective_id";
-- reverse: modify "control_objective_viewers" table
ALTER TABLE "control_objective_viewers" DROP CONSTRAINT "control_objective_viewers_group_id", DROP CONSTRAINT "control_objective_viewers_control_objective_id";
-- reverse: modify "control_objective_editors" table
ALTER TABLE "control_objective_editors" DROP CONSTRAINT "control_objective_editors_group_id", DROP CONSTRAINT "control_objective_editors_control_objective_id";
-- reverse: modify "control_objective_blocked_groups" table
ALTER TABLE "control_objective_blocked_groups" DROP CONSTRAINT "control_objective_blocked_groups_group_id", DROP CONSTRAINT "control_objective_blocked_groups_control_objective_id";
-- reverse: modify "control_implementation_tasks" table
ALTER TABLE "control_implementation_tasks" DROP CONSTRAINT "control_implementation_tasks_task_id", DROP CONSTRAINT "control_implementation_tasks_control_implementation_id";
-- reverse: modify "control_implementation_viewers" table
ALTER TABLE "control_implementation_viewers" DROP CONSTRAINT "control_implementation_viewers_group_id", DROP CONSTRAINT "control_implementation_viewers_control_implementation_id";
-- reverse: modify "control_implementation_editors" table
ALTER TABLE "control_implementation_editors" DROP CONSTRAINT "control_implementation_editors_group_id", DROP CONSTRAINT "control_implementation_editors_control_implementation_id";
-- reverse: modify "control_implementation_blocked_groups" table
ALTER TABLE "control_implementation_blocked_groups" DROP CONSTRAINT "control_implementation_blocked_groups_group_id", DROP CONSTRAINT "control_implementation_blocked_groups_control_implementation_id";
-- reverse: modify "control_control_implementations" table
ALTER TABLE "control_control_implementations" DROP CONSTRAINT "control_control_implementations_control_implementation_id", DROP CONSTRAINT "control_control_implementations_control_id";
-- reverse: modify "control_campaigns" table
ALTER TABLE "control_campaigns" DROP CONSTRAINT "control_campaigns_control_id", DROP CONSTRAINT "control_campaigns_campaign_id";
-- reverse: modify "control_identity_holders" table
ALTER TABLE "control_identity_holders" DROP CONSTRAINT "control_identity_holders_identity_holder_id", DROP CONSTRAINT "control_identity_holders_control_id";
-- reverse: modify "control_entities" table
ALTER TABLE "control_entities" DROP CONSTRAINT "control_entities_entity_id", DROP CONSTRAINT "control_entities_control_id";
-- reverse: modify "control_assets" table
ALTER TABLE "control_assets" DROP CONSTRAINT "control_assets_control_id", DROP CONSTRAINT "control_assets_asset_id";
-- reverse: modify "control_editors" table
ALTER TABLE "control_editors" DROP CONSTRAINT "control_editors_group_id", DROP CONSTRAINT "control_editors_control_id";
-- reverse: modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" DROP CONSTRAINT "control_blocked_groups_group_id", DROP CONSTRAINT "control_blocked_groups_control_id";
-- reverse: modify "control_scans" table
ALTER TABLE "control_scans" DROP CONSTRAINT "control_scans_scan_id", DROP CONSTRAINT "control_scans_control_id";
-- reverse: modify "control_procedures" table
ALTER TABLE "control_procedures" DROP CONSTRAINT "control_procedures_procedure_id", DROP CONSTRAINT "control_procedures_control_id";
-- reverse: modify "control_action_plans" table
ALTER TABLE "control_action_plans" DROP CONSTRAINT "control_action_plans_control_id", DROP CONSTRAINT "control_action_plans_action_plan_id";
-- reverse: modify "control_risks" table
ALTER TABLE "control_risks" DROP CONSTRAINT "control_risks_risk_id", DROP CONSTRAINT "control_risks_control_id";
-- reverse: modify "control_narratives" table
ALTER TABLE "control_narratives" DROP CONSTRAINT "control_narratives_narrative_id", DROP CONSTRAINT "control_narratives_control_id";
-- reverse: modify "control_tasks" table
ALTER TABLE "control_tasks" DROP CONSTRAINT "control_tasks_task_id", DROP CONSTRAINT "control_tasks_control_id";
-- reverse: modify "control_control_objectives" table
ALTER TABLE "control_control_objectives" DROP CONSTRAINT "control_control_objectives_control_objective_id", DROP CONSTRAINT "control_control_objectives_control_id";
-- reverse: modify "contact_files" table
ALTER TABLE "contact_files" DROP CONSTRAINT "contact_files_file_id", DROP CONSTRAINT "contact_files_contact_id";
-- reverse: modify "check_result_controls" table
ALTER TABLE "check_result_controls" DROP CONSTRAINT "check_result_controls_control_id", DROP CONSTRAINT "check_result_controls_check_result_id";
-- reverse: modify "campaign_identity_holders" table
ALTER TABLE "campaign_identity_holders" DROP CONSTRAINT "campaign_identity_holders_identity_holder_id", DROP CONSTRAINT "campaign_identity_holders_campaign_id";
-- reverse: modify "campaign_groups" table
ALTER TABLE "campaign_groups" DROP CONSTRAINT "campaign_groups_group_id", DROP CONSTRAINT "campaign_groups_campaign_id";
-- reverse: modify "campaign_users" table
ALTER TABLE "campaign_users" DROP CONSTRAINT "campaign_users_user_id", DROP CONSTRAINT "campaign_users_campaign_id";
-- reverse: modify "campaign_contacts" table
ALTER TABLE "campaign_contacts" DROP CONSTRAINT "campaign_contacts_contact_id", DROP CONSTRAINT "campaign_contacts_campaign_id";
-- reverse: modify "campaign_viewers" table
ALTER TABLE "campaign_viewers" DROP CONSTRAINT "campaign_viewers_group_id", DROP CONSTRAINT "campaign_viewers_campaign_id";
-- reverse: modify "campaign_editors" table
ALTER TABLE "campaign_editors" DROP CONSTRAINT "campaign_editors_group_id", DROP CONSTRAINT "campaign_editors_campaign_id";
-- reverse: modify "campaign_blocked_groups" table
ALTER TABLE "campaign_blocked_groups" DROP CONSTRAINT "campaign_blocked_groups_group_id", DROP CONSTRAINT "campaign_blocked_groups_campaign_id";
-- reverse: modify "asset_connected_assets" table
ALTER TABLE "asset_connected_assets" DROP CONSTRAINT "asset_connected_assets_connected_from_id", DROP CONSTRAINT "asset_connected_assets_asset_id";
-- reverse: modify "action_plan_tasks" table
ALTER TABLE "action_plan_tasks" DROP CONSTRAINT "action_plan_tasks_task_id", DROP CONSTRAINT "action_plan_tasks_action_plan_id";
-- reverse: modify "action_plan_viewers" table
ALTER TABLE "action_plan_viewers" DROP CONSTRAINT "action_plan_viewers_group_id", DROP CONSTRAINT "action_plan_viewers_action_plan_id";
-- reverse: modify "action_plan_editors" table
ALTER TABLE "action_plan_editors" DROP CONSTRAINT "action_plan_editors_group_id", DROP CONSTRAINT "action_plan_editors_action_plan_id";
-- reverse: modify "action_plan_blocked_groups" table
ALTER TABLE "action_plan_blocked_groups" DROP CONSTRAINT "action_plan_blocked_groups_group_id", DROP CONSTRAINT "action_plan_blocked_groups_action_plan_id";
-- reverse: modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" DROP CONSTRAINT "workflow_proposals_workflow_object_refs_workflow_object_ref", DROP CONSTRAINT "workflow_proposals_users_user", DROP CONSTRAINT "workflow_proposals_organizations_workflow_proposals";
-- reverse: modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" DROP CONSTRAINT "workflow_object_refs_workflow_instances_workflow_object_refs", DROP CONSTRAINT "workflow_object_refs_workflow_instances_workflow_instance", DROP CONSTRAINT "workflow_object_refs_vulnerabilities_vulnerability", DROP CONSTRAINT "workflow_object_refs_tasks_task", DROP CONSTRAINT "workflow_object_refs_subcontrols_subcontrol", DROP CONSTRAINT "workflow_object_refs_risks_risk", DROP CONSTRAINT "workflow_object_refs_remediations_remediation", DROP CONSTRAINT "workflow_object_refs_procedures_procedure", DROP CONSTRAINT "workflow_object_refs_platforms_platform", DROP CONSTRAINT "workflow_object_refs_organizations_workflow_object_refs", DROP CONSTRAINT "workflow_object_refs_internal_policies_internal_policy", DROP CONSTRAINT "workflow_object_refs_identity_holders_identity_holder", DROP CONSTRAINT "workflow_object_refs_findings_finding", DROP CONSTRAINT "workflow_object_refs_evidences_evidence", DROP CONSTRAINT "workflow_object_refs_directory_memberships_directory_membership", DROP CONSTRAINT "workflow_object_refs_directory_groups_directory_group", DROP CONSTRAINT "workflow_object_refs_directory_accounts_directory_account", DROP CONSTRAINT "workflow_object_refs_controls_control", DROP CONSTRAINT "workflow_object_refs_campaigns_campaign", DROP CONSTRAINT "workflow_object_refs_campaign_targets_campaign_target", DROP CONSTRAINT "workflow_object_refs_assessments_assessment", DROP CONSTRAINT "workflow_object_refs_assessment_responses_assessment_response", DROP CONSTRAINT "workflow_object_refs_action_plans_action_plan";
-- reverse: modify "workflow_instances" table
ALTER TABLE "workflow_instances" DROP CONSTRAINT "workflow_instances_workflow_proposals_workflow_proposal", DROP CONSTRAINT "workflow_instances_workflow_definitions_workflow_definition", DROP CONSTRAINT "workflow_instances_vulnerabilities_vulnerability", DROP CONSTRAINT "workflow_instances_tasks_task", DROP CONSTRAINT "workflow_instances_subcontrols_subcontrol", DROP CONSTRAINT "workflow_instances_risks_risk", DROP CONSTRAINT "workflow_instances_remediations_remediation", DROP CONSTRAINT "workflow_instances_procedures_procedure", DROP CONSTRAINT "workflow_instances_platforms_platform", DROP CONSTRAINT "workflow_instances_organizations_workflow_instances", DROP CONSTRAINT "workflow_instances_internal_policies_internal_policy", DROP CONSTRAINT "workflow_instances_integrations_integration", DROP CONSTRAINT "workflow_instances_identity_holders_identity_holder", DROP CONSTRAINT "workflow_instances_findings_finding", DROP CONSTRAINT "workflow_instances_evidences_evidence", DROP CONSTRAINT "workflow_instances_controls_control", DROP CONSTRAINT "workflow_instances_campaigns_campaign", DROP CONSTRAINT "workflow_instances_campaign_targets_campaign_target", DROP CONSTRAINT "workflow_instances_assessments_assessment", DROP CONSTRAINT "workflow_instances_assessment_responses_assessment_response", DROP CONSTRAINT "workflow_instances_action_plans_action_plan";
-- reverse: modify "workflow_events" table
ALTER TABLE "workflow_events" DROP CONSTRAINT "workflow_events_workflow_instances_workflow_instance", DROP CONSTRAINT "workflow_events_workflow_instances_workflow_events", DROP CONSTRAINT "workflow_events_organizations_workflow_events";
-- reverse: modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" DROP CONSTRAINT "workflow_definitions_organizations_workflow_definitions";
-- reverse: modify "workflow_assignment_targets" table
ALTER TABLE "workflow_assignment_targets" DROP CONSTRAINT "workflow_assignment_targets_wo_6077e6f4bf744947c345bb2733c1c240", DROP CONSTRAINT "workflow_assignment_targets_wo_35919ebc89c62ef82cb5889ff40ce351", DROP CONSTRAINT "workflow_assignment_targets_users_user", DROP CONSTRAINT "workflow_assignment_targets_or_8bb74468c70e1b9fcce1d5b038516f9a", DROP CONSTRAINT "workflow_assignment_targets_groups_group";
-- reverse: modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" DROP CONSTRAINT "workflow_assignments_workflow_instances_workflow_instance", DROP CONSTRAINT "workflow_assignments_workflow_instances_workflow_assignments", DROP CONSTRAINT "workflow_assignments_users_user", DROP CONSTRAINT "workflow_assignments_organizations_workflow_assignments", DROP CONSTRAINT "workflow_assignments_groups_group";
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_webauthns";
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP CONSTRAINT "vulnerabilities_users_reviewed_by_user", DROP CONSTRAINT "vulnerabilities_users_assigned_to_user", DROP CONSTRAINT "vulnerabilities_organizations_vulnerabilities", DROP CONSTRAINT "vulnerabilities_groups_reviewed_by_group", DROP CONSTRAINT "vulnerabilities_groups_assigned_to_group", DROP CONSTRAINT "vulnerabilities_custom_type_enums_vulnerability_status", DROP CONSTRAINT "vulnerabilities_custom_type_enums_scope", DROP CONSTRAINT "vulnerabilities_custom_type_enums_environment";
-- reverse: modify "vendor_scoring_configs" table
ALTER TABLE "vendor_scoring_configs" DROP CONSTRAINT "vendor_scoring_configs_organizations_vendor_scoring_configs";
-- reverse: modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" DROP CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_scoring_config", DROP CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_risk_scores", DROP CONSTRAINT "vendor_risk_scores_organizations_vendor_risk_scores", DROP CONSTRAINT "vendor_risk_scores_entities_vendor_risk_scores", DROP CONSTRAINT "vendor_risk_scores_entities_entity", DROP CONSTRAINT "vendor_risk_scores_assessment_responses_vendor_risk_scores", DROP CONSTRAINT "vendor_risk_scores_assessment_responses_assessment_response";
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP CONSTRAINT "user_settings_users_setting", DROP CONSTRAINT "user_settings_organizations_default_org";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_avatar_file";
-- reverse: modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" DROP CONSTRAINT "trust_center_watermark_configs_files_file", DROP CONSTRAINT "trust_center_watermark_configs_e2f038ca8412a7e2b03e1fad46be2f7f";
-- reverse: modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" DROP CONSTRAINT "trust_center_subprocessors_tru_bb0fd7936579c86ecda7d42ebfe60199", DROP CONSTRAINT "trust_center_subprocessors_sub_24055b695e9bd0e49b3edea05d355a0b", DROP CONSTRAINT "trust_center_subprocessors_cus_d5ebb915269b07a0bf77b5b0ec180583";
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP CONSTRAINT "trust_center_settings_groups_nda_approver_group", DROP CONSTRAINT "trust_center_settings_files_logo_file", DROP CONSTRAINT "trust_center_settings_files_hero_image_file", DROP CONSTRAINT "trust_center_settings_files_favicon_file";
-- reverse: modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" DROP CONSTRAINT "trust_center_nda_requests_users_approved_by_user", DROP CONSTRAINT "trust_center_nda_requests_trus_166c4573710ee5957bac7d4b99111f81", DROP CONSTRAINT "trust_center_nda_requests_files_file", DROP CONSTRAINT "trust_center_nda_requests_document_data_document";
-- reverse: modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" DROP CONSTRAINT "trust_center_faqs_trust_centers_trust_center_faqs", DROP CONSTRAINT "trust_center_faqs_notes_trust_center_faqs", DROP CONSTRAINT "trust_center_faqs_custom_type_enums_trust_center_faq_kind";
-- reverse: modify "trust_center_entities" table
ALTER TABLE "trust_center_entities" DROP CONSTRAINT "trust_center_entities_trust_centers_trust_center_entities", DROP CONSTRAINT "trust_center_entities_files_trust_center_entities", DROP CONSTRAINT "trust_center_entities_files_logo_file", DROP CONSTRAINT "trust_center_entities_entity_types_entity_type";
-- reverse: modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" DROP CONSTRAINT "trust_center_docs_trust_centers_trust_center_docs", DROP CONSTRAINT "trust_center_docs_trust_center_nda_requests_trust_center_docs", DROP CONSTRAINT "trust_center_docs_standards_trust_center_docs", DROP CONSTRAINT "trust_center_docs_files_original_file", DROP CONSTRAINT "trust_center_docs_files_file", DROP CONSTRAINT "trust_center_docs_custom_type_enums_trust_center_doc_kind";
-- reverse: modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" DROP CONSTRAINT "trust_center_compliances_trust_centers_trust_center_compliances", DROP CONSTRAINT "trust_center_compliances_standards_trust_center_compliances";
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP CONSTRAINT "trust_centers_trust_center_watermark_configs_watermark_config", DROP CONSTRAINT "trust_centers_trust_center_settings_setting", DROP CONSTRAINT "trust_centers_trust_center_settings_preview_setting", DROP CONSTRAINT "trust_centers_organizations_trust_centers", DROP CONSTRAINT "trust_centers_custom_domains_preview_domain", DROP CONSTRAINT "trust_centers_custom_domains_custom_domain";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP CONSTRAINT "templates_trust_centers_templates", DROP CONSTRAINT "templates_organizations_templates", DROP CONSTRAINT "templates_custom_type_enums_scope", DROP CONSTRAINT "templates_custom_type_enums_environment";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_assigner_tasks", DROP CONSTRAINT "tasks_users_assignee_tasks", DROP CONSTRAINT "tasks_reviews_tasks", DROP CONSTRAINT "tasks_remediations_tasks", DROP CONSTRAINT "tasks_organizations_tasks", DROP CONSTRAINT "tasks_integrations_tasks", DROP CONSTRAINT "tasks_custom_type_enums_tasks", DROP CONSTRAINT "tasks_custom_type_enums_task_kind", DROP CONSTRAINT "tasks_custom_type_enums_scope", DROP CONSTRAINT "tasks_custom_type_enums_environment";
-- reverse: modify "tag_definitions" table
ALTER TABLE "tag_definitions" DROP CONSTRAINT "tag_definitions_workflow_definitions_tag_definitions", DROP CONSTRAINT "tag_definitions_organizations_tag_definitions";
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP CONSTRAINT "tfa_settings_users_tfa_settings";
-- reverse: modify "system_details" table
ALTER TABLE "system_details" DROP CONSTRAINT "system_details_organizations_system_details";
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP CONSTRAINT "subscribers_users_subscribers", DROP CONSTRAINT "subscribers_trust_centers_subscribers", DROP CONSTRAINT "subscribers_organizations_subscribers", DROP CONSTRAINT "subscribers_contacts_subscribers";
-- reverse: modify "subprocessors" table
ALTER TABLE "subprocessors" DROP CONSTRAINT "subprocessors_organizations_subprocessors", DROP CONSTRAINT "subprocessors_files_logo_file";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_users_subcontrols", DROP CONSTRAINT "subcontrols_programs_subcontrols", DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP CONSTRAINT "subcontrols_groups_delegate", DROP CONSTRAINT "subcontrols_groups_control_owner", DROP CONSTRAINT "subcontrols_entities_responsible_party", DROP CONSTRAINT "subcontrols_custom_type_enums_subcontrols", DROP CONSTRAINT "subcontrols_custom_type_enums_subcontrol_kind", DROP CONSTRAINT "subcontrols_controls_subcontrols";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP CONSTRAINT "standards_organizations_standards", DROP CONSTRAINT "standards_files_logo_file";
-- reverse: modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP CONSTRAINT "scheduled_job_runs_scheduled_jobs_scheduled_job", DROP CONSTRAINT "scheduled_job_runs_organizations_scheduled_job_runs", DROP CONSTRAINT "scheduled_job_runs_job_runners_job_runner";
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP CONSTRAINT "scheduled_jobs_organizations_scheduled_jobs", DROP CONSTRAINT "scheduled_jobs_job_templates_scheduled_jobs", DROP CONSTRAINT "scheduled_jobs_job_runners_job_runner";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP CONSTRAINT "scans_users_reviewed_by_user", DROP CONSTRAINT "scans_users_performed_by_user", DROP CONSTRAINT "scans_users_assigned_to_user", DROP CONSTRAINT "scans_risks_scans", DROP CONSTRAINT "scans_platforms_generated_scans", DROP CONSTRAINT "scans_organizations_scans", DROP CONSTRAINT "scans_groups_reviewed_by_group", DROP CONSTRAINT "scans_groups_performed_by_group", DROP CONSTRAINT "scans_groups_assigned_to_group", DROP CONSTRAINT "scans_custom_type_enums_scope", DROP CONSTRAINT "scans_custom_type_enums_environment";
-- reverse: modify "sla_definitions" table
ALTER TABLE "sla_definitions" DROP CONSTRAINT "sla_definitions_organizations_sla_definitions";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_organizations_risks", DROP CONSTRAINT "risks_groups_stakeholder", DROP CONSTRAINT "risks_groups_delegate", DROP CONSTRAINT "risks_custom_type_enums_scope", DROP CONSTRAINT "risks_custom_type_enums_risks", DROP CONSTRAINT "risks_custom_type_enums_risk_kind", DROP CONSTRAINT "risks_custom_type_enums_risk_category", DROP CONSTRAINT "risks_custom_type_enums_risk_categories", DROP CONSTRAINT "risks_custom_type_enums_environment", DROP CONSTRAINT "risks_control_objectives_risks";
-- reverse: modify "reviews" table
ALTER TABLE "reviews" DROP CONSTRAINT "reviews_users_reviewer", DROP CONSTRAINT "reviews_organizations_reviews", DROP CONSTRAINT "reviews_custom_type_enums_scope", DROP CONSTRAINT "reviews_custom_type_enums_environment";
-- reverse: modify "remediations" table
ALTER TABLE "remediations" DROP CONSTRAINT "remediations_organizations_remediations", DROP CONSTRAINT "remediations_custom_type_enums_scope", DROP CONSTRAINT "remediations_custom_type_enums_environment";
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP CONSTRAINT "program_memberships_users_user", DROP CONSTRAINT "program_memberships_programs_program", DROP CONSTRAINT "program_memberships_org_memberships_org_membership";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_users_programs_owned", DROP CONSTRAINT "programs_organizations_programs", DROP CONSTRAINT "programs_custom_type_enums_programs", DROP CONSTRAINT "programs_custom_type_enums_program_kind";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_organizations_procedures", DROP CONSTRAINT "procedures_groups_delegate", DROP CONSTRAINT "procedures_groups_approver", DROP CONSTRAINT "procedures_files_file", DROP CONSTRAINT "procedures_custom_type_enums_scope", DROP CONSTRAINT "procedures_custom_type_enums_procedures", DROP CONSTRAINT "procedures_custom_type_enums_procedure_kind", DROP CONSTRAINT "procedures_custom_type_enums_environment", DROP CONSTRAINT "procedures_control_objectives_procedures";
-- reverse: modify "platforms" table
ALTER TABLE "platforms" DROP CONSTRAINT "platforms_users_technical_owner_user", DROP CONSTRAINT "platforms_users_security_owner_user", DROP CONSTRAINT "platforms_users_platforms_owned", DROP CONSTRAINT "platforms_users_internal_owner_user", DROP CONSTRAINT "platforms_users_business_owner_user", DROP CONSTRAINT "platforms_organizations_platforms", DROP CONSTRAINT "platforms_identity_holders_access_platforms", DROP CONSTRAINT "platforms_groups_technical_owner_group", DROP CONSTRAINT "platforms_groups_security_owner_group", DROP CONSTRAINT "platforms_groups_internal_owner_group", DROP CONSTRAINT "platforms_groups_business_owner_group", DROP CONSTRAINT "platforms_custom_type_enums_security_tier", DROP CONSTRAINT "platforms_custom_type_enums_scope", DROP CONSTRAINT "platforms_custom_type_enums_platforms", DROP CONSTRAINT "platforms_custom_type_enums_platform_kind", DROP CONSTRAINT "platforms_custom_type_enums_platform_data_classification", DROP CONSTRAINT "platforms_custom_type_enums_environment", DROP CONSTRAINT "platforms_custom_type_enums_encryption_status", DROP CONSTRAINT "platforms_custom_type_enums_criticality", DROP CONSTRAINT "platforms_custom_type_enums_access_model";
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP CONSTRAINT "personal_access_tokens_users_personal_access_tokens";
-- reverse: modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP CONSTRAINT "password_reset_tokens_users_password_reset_tokens";
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP CONSTRAINT "organization_settings_organizations_setting";
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP CONSTRAINT "organizations_files_avatar_file";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP CONSTRAINT "org_subscriptions_organizations_org_subscriptions";
-- reverse: modify "org_products" table
ALTER TABLE "org_products" DROP CONSTRAINT "org_products_organizations_org_products", DROP CONSTRAINT "org_products_org_subscriptions_products", DROP CONSTRAINT "org_products_org_modules_org_products";
-- reverse: modify "org_prices" table
ALTER TABLE "org_prices" DROP CONSTRAINT "org_prices_organizations_org_prices", DROP CONSTRAINT "org_prices_org_subscriptions_prices";
-- reverse: modify "org_modules" table
ALTER TABLE "org_modules" DROP CONSTRAINT "org_modules_organizations_org_modules", DROP CONSTRAINT "org_modules_org_subscriptions_modules", DROP CONSTRAINT "org_modules_org_products_org_modules";
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP CONSTRAINT "org_memberships_users_user", DROP CONSTRAINT "org_memberships_organizations_organization";
-- reverse: modify "onboardings" table
ALTER TABLE "onboardings" DROP CONSTRAINT "onboardings_organizations_organization";
-- reverse: modify "notification_templates" table
ALTER TABLE "notification_templates" DROP CONSTRAINT "notification_templates_workflo_439a17f2830fbf868eeb61d3d3fdac37", DROP CONSTRAINT "notification_templates_organizations_notification_templates", DROP CONSTRAINT "notification_templates_integrations_notification_templates", DROP CONSTRAINT "notification_templates_email_templates_notification_templates";
-- reverse: modify "notification_preferences" table
ALTER TABLE "notification_preferences" DROP CONSTRAINT "notification_preferences_users_user", DROP CONSTRAINT "notification_preferences_organizations_notification_preferences", DROP CONSTRAINT "notification_preferences_notif_aabd0a3ca9e335110ce7c2348e4f4cf0";
-- reverse: modify "notifications" table
ALTER TABLE "notifications" DROP CONSTRAINT "notifications_organizations_notifications", DROP CONSTRAINT "notifications_notification_templates_notifications";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_vulnerabilities_comments", DROP CONSTRAINT "notes_trust_centers_posts", DROP CONSTRAINT "notes_tasks_comments", DROP CONSTRAINT "notes_subcontrols_comments", DROP CONSTRAINT "notes_risks_comments", DROP CONSTRAINT "notes_reviews_comments", DROP CONSTRAINT "notes_remediations_comments", DROP CONSTRAINT "notes_programs_notes", DROP CONSTRAINT "notes_procedures_comments", DROP CONSTRAINT "notes_organizations_notes", DROP CONSTRAINT "notes_internal_policies_comments", DROP CONSTRAINT "notes_findings_comments", DROP CONSTRAINT "notes_evidences_comments", DROP CONSTRAINT "notes_entities_notes", DROP CONSTRAINT "notes_discussions_comments", DROP CONSTRAINT "notes_controls_comments";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_subcontrols_narratives", DROP CONSTRAINT "narratives_organizations_narratives", DROP CONSTRAINT "narratives_control_objectives_narratives";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP CONSTRAINT "mapped_controls_organizations_mapped_controls";
-- reverse: modify "job_templates" table
ALTER TABLE "job_templates" DROP CONSTRAINT "job_templates_organizations_job_templates";
-- reverse: modify "job_runner_tokens" table
ALTER TABLE "job_runner_tokens" DROP CONSTRAINT "job_runner_tokens_organizations_job_runner_tokens";
-- reverse: modify "job_runner_registration_tokens" table
ALTER TABLE "job_runner_registration_tokens" DROP CONSTRAINT "job_runner_registration_tokens_job_runners_job_runner", DROP CONSTRAINT "job_runner_registration_tokens_daddf3e078805108b2d174df258ddb4b";
-- reverse: modify "job_runners" table
ALTER TABLE "job_runners" DROP CONSTRAINT "job_runners_organizations_job_runners";
-- reverse: modify "job_results" table
ALTER TABLE "job_results" DROP CONSTRAINT "job_results_scheduled_jobs_scheduled_job", DROP CONSTRAINT "job_results_organizations_job_results", DROP CONSTRAINT "job_results_files_file";
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP CONSTRAINT "invites_organizations_invites";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_organizations_internal_policies", DROP CONSTRAINT "internal_policies_groups_delegate", DROP CONSTRAINT "internal_policies_groups_approver", DROP CONSTRAINT "internal_policies_files_file", DROP CONSTRAINT "internal_policies_custom_type_enums_scope", DROP CONSTRAINT "internal_policies_custom_type_enums_internal_policy_kind", DROP CONSTRAINT "internal_policies_custom_type_enums_internal_policies", DROP CONSTRAINT "internal_policies_custom_type_enums_environment";
-- reverse: modify "integration_webhooks" table
ALTER TABLE "integration_webhooks" DROP CONSTRAINT "integration_webhooks_organizations_integration_webhooks", DROP CONSTRAINT "integration_webhooks_integrations_integration_webhooks";
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP CONSTRAINT "integration_runs_organizations_integration_runs", DROP CONSTRAINT "integration_runs_integrations_integration_runs", DROP CONSTRAINT "integration_runs_files_response_file", DROP CONSTRAINT "integration_runs_files_request_file", DROP CONSTRAINT "integration_runs_events_event", DROP CONSTRAINT "integration_runs_assessment_responses_assessment_response";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_platforms_integrations", DROP CONSTRAINT "integrations_organizations_integrations", DROP CONSTRAINT "integrations_groups_integrations", DROP CONSTRAINT "integrations_files_integrations", DROP CONSTRAINT "integrations_custom_type_enums_scope", DROP CONSTRAINT "integrations_custom_type_enums_environment";
-- reverse: modify "impersonation_events" table
ALTER TABLE "impersonation_events" DROP CONSTRAINT "impersonation_events_users_targeted_impersonations", DROP CONSTRAINT "impersonation_events_users_impersonation_events", DROP CONSTRAINT "impersonation_events_organizations_impersonation_events";
-- reverse: modify "identity_holders" table
ALTER TABLE "identity_holders" DROP CONSTRAINT "identity_holders_users_internal_owner_user", DROP CONSTRAINT "identity_holders_users_identity_holder_profiles", DROP CONSTRAINT "identity_holders_organizations_identity_holders", DROP CONSTRAINT "identity_holders_groups_internal_owner_group", DROP CONSTRAINT "identity_holders_entities_employer", DROP CONSTRAINT "identity_holders_custom_type_enums_scope", DROP CONSTRAINT "identity_holders_custom_type_enums_environment";
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP CONSTRAINT "hushes_organizations_secrets";
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" DROP CONSTRAINT "group_settings_groups_setting";
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP CONSTRAINT "group_memberships_users_user", DROP CONSTRAINT "group_memberships_org_memberships_org_membership", DROP CONSTRAINT "group_memberships_groups_group";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_workflow_definitions_viewers", DROP CONSTRAINT "groups_workflow_definitions_groups", DROP CONSTRAINT "groups_workflow_definitions_editors", DROP CONSTRAINT "groups_workflow_definitions_blocked_groups", DROP CONSTRAINT "groups_vulnerabilities_viewers", DROP CONSTRAINT "groups_vulnerabilities_editors", DROP CONSTRAINT "groups_vulnerabilities_blocked_groups", DROP CONSTRAINT "groups_trust_centers_editors", DROP CONSTRAINT "groups_trust_centers_blocked_groups", DROP CONSTRAINT "groups_trust_center_watermark_configs_editors", DROP CONSTRAINT "groups_trust_center_watermark_configs_blocked_groups", DROP CONSTRAINT "groups_trust_center_subprocessors_editors", DROP CONSTRAINT "groups_trust_center_subprocessors_blocked_groups", DROP CONSTRAINT "groups_trust_center_settings_editors", DROP CONSTRAINT "groups_trust_center_settings_blocked_groups", DROP CONSTRAINT "groups_trust_center_nda_requests_editors", DROP CONSTRAINT "groups_trust_center_nda_requests_blocked_groups", DROP CONSTRAINT "groups_trust_center_faqs_editors", DROP CONSTRAINT "groups_trust_center_faqs_blocked_groups", DROP CONSTRAINT "groups_trust_center_entities_editors", DROP CONSTRAINT "groups_trust_center_entities_blocked_groups", DROP CONSTRAINT "groups_trust_center_docs_editors", DROP CONSTRAINT "groups_trust_center_docs_blocked_groups", DROP CONSTRAINT "groups_trust_center_compliances_editors", DROP CONSTRAINT "groups_trust_center_compliances_blocked_groups", DROP CONSTRAINT "groups_sla_definitions_editors", DROP CONSTRAINT "groups_sla_definitions_blocked_groups", DROP CONSTRAINT "groups_organizations_workflows_manager", DROP CONSTRAINT "groups_organizations_workflow_definition_creators", DROP CONSTRAINT "groups_organizations_vulnerability_creators", DROP CONSTRAINT "groups_organizations_vendor_scoring_config_creators", DROP CONSTRAINT "groups_organizations_vendor_risk_score_creators", DROP CONSTRAINT "groups_organizations_trust_center_watermark_config_creators", DROP CONSTRAINT "groups_organizations_trust_center_subprocessor_creators", DROP CONSTRAINT "groups_organizations_trust_center_nda_request_creators", DROP CONSTRAINT "groups_organizations_trust_center_manager", DROP CONSTRAINT "groups_organizations_trust_center_faq_creators", DROP CONSTRAINT "groups_organizations_trust_center_entity_creators", DROP CONSTRAINT "groups_organizations_trust_center_doc_creators", DROP CONSTRAINT "groups_organizations_trust_center_creators", DROP CONSTRAINT "groups_organizations_trust_center_compliance_creators", DROP CONSTRAINT "groups_organizations_template_creators", DROP CONSTRAINT "groups_organizations_task_creators", DROP CONSTRAINT "groups_organizations_tag_definition_creators", DROP CONSTRAINT "groups_organizations_system_detail_creators", DROP CONSTRAINT "groups_organizations_subscriber_creators", DROP CONSTRAINT "groups_organizations_subprocessor_creators", DROP CONSTRAINT "groups_organizations_subcontrol_creators", DROP CONSTRAINT "groups_organizations_standard_creators", DROP CONSTRAINT "groups_organizations_sla_definition_creators", DROP CONSTRAINT "groups_organizations_scheduled_job_run_creators", DROP CONSTRAINT "groups_organizations_scheduled_job_creators", DROP CONSTRAINT "groups_organizations_scan_creators", DROP CONSTRAINT "groups_organizations_risk_manager", DROP CONSTRAINT "groups_organizations_risk_creators", DROP CONSTRAINT "groups_organizations_review_creators", DROP CONSTRAINT "groups_organizations_remediation_creators", DROP CONSTRAINT "groups_organizations_registry_manager", DROP CONSTRAINT "groups_organizations_program_membership_creators", DROP CONSTRAINT "groups_organizations_program_creators", DROP CONSTRAINT "groups_organizations_procedure_creators", DROP CONSTRAINT "groups_organizations_policies_manager", DROP CONSTRAINT "groups_organizations_platform_creators", DROP CONSTRAINT "groups_organizations_org_membership_creators", DROP CONSTRAINT "groups_organizations_notification_template_creators", DROP CONSTRAINT "groups_organizations_note_creators", DROP CONSTRAINT "groups_organizations_narrative_creators", DROP CONSTRAINT "groups_organizations_mapped_control_creators", DROP CONSTRAINT "groups_organizations_job_template_creators", DROP CONSTRAINT "groups_organizations_job_runner_token_creators", DROP CONSTRAINT "groups_organizations_job_runner_registration_token_creators", DROP CONSTRAINT "groups_organizations_job_runner_creators", DROP CONSTRAINT "groups_organizations_invite_creators", DROP CONSTRAINT "groups_organizations_internal_policy_creators", DROP CONSTRAINT "groups_organizations_identity_holder_creators", DROP CONSTRAINT "groups_organizations_hush_creators", DROP CONSTRAINT "groups_organizations_groups", DROP CONSTRAINT "groups_organizations_group_setting_creators", DROP CONSTRAINT "groups_organizations_group_membership_creators", DROP CONSTRAINT "groups_organizations_group_manager", DROP CONSTRAINT "groups_organizations_group_creators", DROP CONSTRAINT "groups_organizations_finding_creators", DROP CONSTRAINT "groups_organizations_finding_control_creators", DROP CONSTRAINT "groups_organizations_file_creators", DROP CONSTRAINT "groups_organizations_evidence_creators", DROP CONSTRAINT "groups_organizations_entity_type_creators", DROP CONSTRAINT "groups_organizations_entity_creators", DROP CONSTRAINT "groups_organizations_email_template_creators", DROP CONSTRAINT "groups_organizations_document_data_creators", DROP CONSTRAINT "groups_organizations_discussion_creators", DROP CONSTRAINT "groups_organizations_directory_sync_run_creators", DROP CONSTRAINT "groups_organizations_directory_membership_creators", DROP CONSTRAINT "groups_organizations_directory_group_creators", DROP CONSTRAINT "groups_organizations_directory_account_creators", DROP CONSTRAINT "groups_organizations_custom_type_enum_creators", DROP CONSTRAINT "groups_organizations_custom_domain_creators", DROP CONSTRAINT "groups_organizations_control_objective_creators", DROP CONSTRAINT "groups_organizations_control_implementation_creators", DROP CONSTRAINT "groups_organizations_control_creators", DROP CONSTRAINT "groups_organizations_contact_creators", DROP CONSTRAINT "groups_organizations_compliance_manager", DROP CONSTRAINT "groups_organizations_check_result_creators", DROP CONSTRAINT "groups_organizations_campaigns_manager", DROP CONSTRAINT "groups_organizations_campaign_target_creators", DROP CONSTRAINT "groups_organizations_campaign_creators", DROP CONSTRAINT "groups_organizations_asset_creators", DROP CONSTRAINT "groups_organizations_assessment_creators", DROP CONSTRAINT "groups_organizations_api_token_creators", DROP CONSTRAINT "groups_organizations_action_plan_creators", DROP CONSTRAINT "groups_identity_holders_viewers", DROP CONSTRAINT "groups_identity_holders_editors", DROP CONSTRAINT "groups_identity_holders_blocked_groups", DROP CONSTRAINT "groups_files_avatar_file", DROP CONSTRAINT "groups_email_templates_viewers", DROP CONSTRAINT "groups_email_templates_editors", DROP CONSTRAINT "groups_email_templates_blocked_groups", DROP CONSTRAINT "groups_check_results_viewers", DROP CONSTRAINT "groups_check_results_editors", DROP CONSTRAINT "groups_check_results_blocked_groups", DROP CONSTRAINT "groups_assets_viewers", DROP CONSTRAINT "groups_assets_editors", DROP CONSTRAINT "groups_assets_blocked_groups", DROP CONSTRAINT "groups_assessments_viewers", DROP CONSTRAINT "groups_assessments_editors", DROP CONSTRAINT "groups_assessments_blocked_groups";
-- reverse: modify "finding_controls" table
ALTER TABLE "finding_controls" DROP CONSTRAINT "finding_controls_standards_standard", DROP CONSTRAINT "finding_controls_organizations_finding_controls", DROP CONSTRAINT "finding_controls_findings_finding", DROP CONSTRAINT "finding_controls_controls_control";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP CONSTRAINT "findings_users_reviewed_by_user", DROP CONSTRAINT "findings_users_assigned_to_user", DROP CONSTRAINT "findings_organizations_findings", DROP CONSTRAINT "findings_groups_reviewed_by_group", DROP CONSTRAINT "findings_groups_assigned_to_group", DROP CONSTRAINT "findings_custom_type_enums_scope", DROP CONSTRAINT "findings_custom_type_enums_finding_status", DROP CONSTRAINT "findings_custom_type_enums_environment";
-- reverse: modify "file_download_tokens" table
ALTER TABLE "file_download_tokens" DROP CONSTRAINT "file_download_tokens_users_file_download_tokens";
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_vulnerabilities_files", DROP CONSTRAINT "files_reviews_files", DROP CONSTRAINT "files_remediations_files", DROP CONSTRAINT "files_platforms_trust_boundary_diagrams", DROP CONSTRAINT "files_platforms_data_flow_diagrams", DROP CONSTRAINT "files_platforms_architecture_diagrams", DROP CONSTRAINT "files_notes_files", DROP CONSTRAINT "files_integrations_files", DROP CONSTRAINT "files_findings_files", DROP CONSTRAINT "files_exports_files", DROP CONSTRAINT "files_email_templates_files", DROP CONSTRAINT "files_custom_type_enums_scope", DROP CONSTRAINT "files_custom_type_enums_environment", DROP CONSTRAINT "files_custom_type_enums_category";
-- reverse: modify "exports" table
ALTER TABLE "exports" DROP CONSTRAINT "exports_organizations_exports";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP CONSTRAINT "evidences_organizations_evidence", DROP CONSTRAINT "evidences_custom_type_enums_scope", DROP CONSTRAINT "evidences_custom_type_enums_environment";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "events_exports_events", DROP CONSTRAINT "events_directory_memberships_events";
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP CONSTRAINT "entity_types_organizations_entity_types";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_users_reviewed_by_user", DROP CONSTRAINT "entities_users_internal_owner_user", DROP CONSTRAINT "entities_risks_entities", DROP CONSTRAINT "entities_organizations_entities", DROP CONSTRAINT "entities_groups_reviewed_by_group", DROP CONSTRAINT "entities_groups_internal_owner_group", DROP CONSTRAINT "entities_files_logo_file", DROP CONSTRAINT "entities_entity_types_entity_type", DROP CONSTRAINT "entities_entity_types_entities", DROP CONSTRAINT "entities_custom_type_enums_scope", DROP CONSTRAINT "entities_custom_type_enums_environment", DROP CONSTRAINT "entities_custom_type_enums_entity_source_type", DROP CONSTRAINT "entities_custom_type_enums_entity_security_questionnaire_status", DROP CONSTRAINT "entities_custom_type_enums_entity_relationship_state";
-- reverse: modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP CONSTRAINT "email_verification_tokens_users_email_verification_tokens";
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" DROP CONSTRAINT "email_templates_workflow_instances_email_templates", DROP CONSTRAINT "email_templates_workflow_definitions_email_templates", DROP CONSTRAINT "email_templates_trust_centers_email_templates", DROP CONSTRAINT "email_templates_organizations_email_templates", DROP CONSTRAINT "email_templates_integrations_email_templates";
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_templates_documents", DROP CONSTRAINT "document_data_organizations_documents", DROP CONSTRAINT "document_data_custom_type_enums_scope", DROP CONSTRAINT "document_data_custom_type_enums_environment";
-- reverse: modify "discussions" table
ALTER TABLE "discussions" DROP CONSTRAINT "discussions_subcontrols_discussions", DROP CONSTRAINT "discussions_risks_discussions", DROP CONSTRAINT "discussions_procedures_discussions", DROP CONSTRAINT "discussions_organizations_discussions", DROP CONSTRAINT "discussions_internal_policies_discussions", DROP CONSTRAINT "discussions_controls_discussions";
-- reverse: modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" DROP CONSTRAINT "directory_sync_runs_platforms_directory_sync_runs", DROP CONSTRAINT "directory_sync_runs_organizations_directory_sync_runs", DROP CONSTRAINT "directory_sync_runs_integrations_directory_sync_runs", DROP CONSTRAINT "directory_sync_runs_custom_type_enums_scope", DROP CONSTRAINT "directory_sync_runs_custom_type_enums_environment";
-- reverse: modify "directory_memberships" table
ALTER TABLE "directory_memberships" DROP CONSTRAINT "directory_memberships_platforms_directory_memberships", DROP CONSTRAINT "directory_memberships_organizations_directory_memberships", DROP CONSTRAINT "directory_memberships_integrations_directory_memberships", DROP CONSTRAINT "directory_memberships_directory_sync_runs_directory_memberships", DROP CONSTRAINT "directory_memberships_directory_groups_directory_group", DROP CONSTRAINT "directory_memberships_directory_accounts_directory_account", DROP CONSTRAINT "directory_memberships_custom_type_enums_scope", DROP CONSTRAINT "directory_memberships_custom_type_enums_environment";
-- reverse: modify "directory_groups" table
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_platforms_directory_groups", DROP CONSTRAINT "directory_groups_organizations_directory_groups", DROP CONSTRAINT "directory_groups_integrations_directory_groups", DROP CONSTRAINT "directory_groups_directory_sync_runs_directory_groups", DROP CONSTRAINT "directory_groups_custom_type_enums_scope", DROP CONSTRAINT "directory_groups_custom_type_enums_environment";
-- reverse: modify "directory_accounts" table
ALTER TABLE "directory_accounts" DROP CONSTRAINT "directory_accounts_platforms_directory_accounts", DROP CONSTRAINT "directory_accounts_organizations_directory_accounts", DROP CONSTRAINT "directory_accounts_integrations_directory_accounts", DROP CONSTRAINT "directory_accounts_identity_holders_directory_accounts", DROP CONSTRAINT "directory_accounts_files_avatar_file", DROP CONSTRAINT "directory_accounts_directory_sync_runs_directory_accounts", DROP CONSTRAINT "directory_accounts_custom_type_enums_scope", DROP CONSTRAINT "directory_accounts_custom_type_enums_environment";
-- reverse: modify "dns_verifications" table
ALTER TABLE "dns_verifications" DROP CONSTRAINT "dns_verifications_organizations_dns_verifications";
-- reverse: modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" DROP CONSTRAINT "custom_type_enums_organizations_custom_type_enums", DROP CONSTRAINT "custom_type_enums_entities_auth_methods";
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP CONSTRAINT "custom_domains_organizations_custom_domains", DROP CONSTRAINT "custom_domains_mappable_domains_mappable_domain", DROP CONSTRAINT "custom_domains_mappable_domains_custom_domains", DROP CONSTRAINT "custom_domains_dns_verifications_dns_verification", DROP CONSTRAINT "custom_domains_dns_verifications_custom_domains";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_organizations_control_objectives";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP CONSTRAINT "control_implementations_organizations_control_implementations", DROP CONSTRAINT "control_implementations_intern_78a7d74302db6f99776c0594111f170b", DROP CONSTRAINT "control_implementations_evidences_control_implementations";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_standards_controls", DROP CONSTRAINT "controls_organizations_controls", DROP CONSTRAINT "controls_groups_delegate", DROP CONSTRAINT "controls_groups_control_owner", DROP CONSTRAINT "controls_entities_responsible_party", DROP CONSTRAINT "controls_custom_type_enums_scope", DROP CONSTRAINT "controls_custom_type_enums_environment", DROP CONSTRAINT "controls_custom_type_enums_controls", DROP CONSTRAINT "controls_custom_type_enums_control_kind";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP CONSTRAINT "contacts_organizations_contacts";
-- reverse: modify "check_results" table
ALTER TABLE "check_results" DROP CONSTRAINT "check_results_integrations_check_results";
-- reverse: modify "campaign_targets" table
ALTER TABLE "campaign_targets" DROP CONSTRAINT "campaign_targets_users_campaign_targets", DROP CONSTRAINT "campaign_targets_subscribers_campaign_targets", DROP CONSTRAINT "campaign_targets_organizations_campaign_targets", DROP CONSTRAINT "campaign_targets_groups_campaign_targets", DROP CONSTRAINT "campaign_targets_contacts_campaign_targets", DROP CONSTRAINT "campaign_targets_campaigns_campaign_targets";
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_users_internal_owner_user", DROP CONSTRAINT "campaigns_trust_centers_campaigns", DROP CONSTRAINT "campaigns_templates_campaigns", DROP CONSTRAINT "campaigns_organizations_campaigns", DROP CONSTRAINT "campaigns_integrations_campaigns", DROP CONSTRAINT "campaigns_groups_internal_owner_group", DROP CONSTRAINT "campaigns_entities_campaigns", DROP CONSTRAINT "campaigns_email_templates_campaigns", DROP CONSTRAINT "campaigns_assessments_campaigns";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP CONSTRAINT "assets_users_internal_owner_user", DROP CONSTRAINT "assets_risks_assets", DROP CONSTRAINT "assets_platforms_source_assets", DROP CONSTRAINT "assets_organizations_assets", DROP CONSTRAINT "assets_integrations_assets", DROP CONSTRAINT "assets_groups_internal_owner_group", DROP CONSTRAINT "assets_custom_type_enums_security_tier", DROP CONSTRAINT "assets_custom_type_enums_scope", DROP CONSTRAINT "assets_custom_type_enums_environment", DROP CONSTRAINT "assets_custom_type_enums_encryption_status", DROP CONSTRAINT "assets_custom_type_enums_criticality", DROP CONSTRAINT "assets_custom_type_enums_asset_subtype", DROP CONSTRAINT "assets_custom_type_enums_asset_data_classification", DROP CONSTRAINT "assets_custom_type_enums_access_model";
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP CONSTRAINT "assessment_responses_organizations_assessment_responses", DROP CONSTRAINT "assessment_responses_identity_holders_assessment_responses", DROP CONSTRAINT "assessment_responses_entities_assessment_responses", DROP CONSTRAINT "assessment_responses_document_data_document", DROP CONSTRAINT "assessment_responses_campaigns_assessment_responses", DROP CONSTRAINT "assessment_responses_assessments_assessment_responses";
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_assessments", DROP CONSTRAINT "assessments_organizations_assessments";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_users_action_plans", DROP CONSTRAINT "action_plans_subcontrols_action_plans", DROP CONSTRAINT "action_plans_organizations_action_plans", DROP CONSTRAINT "action_plans_groups_delegate", DROP CONSTRAINT "action_plans_groups_approver", DROP CONSTRAINT "action_plans_files_file", DROP CONSTRAINT "action_plans_custom_type_enums_action_plans", DROP CONSTRAINT "action_plans_custom_type_enums_action_plan_kind";
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP CONSTRAINT "api_tokens_organizations_api_tokens";
-- reverse: create index "vulnerability_tasks_task_id_idx" to table: "vulnerability_tasks"
DROP INDEX "vulnerability_tasks_task_id_idx";
-- reverse: create "vulnerability_tasks" table
DROP TABLE "vulnerability_tasks";
-- reverse: create index "vulnerability_scans_scan_id_idx" to table: "vulnerability_scans"
DROP INDEX "vulnerability_scans_scan_id_idx";
-- reverse: create "vulnerability_scans" table
DROP TABLE "vulnerability_scans";
-- reverse: create index "vulnerability_entities_entity_id_idx" to table: "vulnerability_entities"
DROP INDEX "vulnerability_entities_entity_id_idx";
-- reverse: create "vulnerability_entities" table
DROP TABLE "vulnerability_entities";
-- reverse: create index "vulnerability_assets_asset_id_idx" to table: "vulnerability_assets"
DROP INDEX "vulnerability_assets_asset_id_idx";
-- reverse: create "vulnerability_assets" table
DROP TABLE "vulnerability_assets";
-- reverse: create index "vulnerability_programs_program_id_idx" to table: "vulnerability_programs"
DROP INDEX "vulnerability_programs_program_id_idx";
-- reverse: create "vulnerability_programs" table
DROP TABLE "vulnerability_programs";
-- reverse: create index "vulnerability_risks_risk_id_idx" to table: "vulnerability_risks"
DROP INDEX "vulnerability_risks_risk_id_idx";
-- reverse: create "vulnerability_risks" table
DROP TABLE "vulnerability_risks";
-- reverse: create index "vulnerability_subcontrols_subcontrol_id_idx" to table: "vulnerability_subcontrols"
DROP INDEX "vulnerability_subcontrols_subcontrol_id_idx";
-- reverse: create "vulnerability_subcontrols" table
DROP TABLE "vulnerability_subcontrols";
-- reverse: create index "vulnerability_controls_control_id_idx" to table: "vulnerability_controls"
DROP INDEX "vulnerability_controls_control_id_idx";
-- reverse: create "vulnerability_controls" table
DROP TABLE "vulnerability_controls";
-- reverse: create index "vulnerability_action_plans_action_plan_id_idx" to table: "vulnerability_action_plans"
DROP INDEX "vulnerability_action_plans_action_plan_id_idx";
-- reverse: create "vulnerability_action_plans" table
DROP TABLE "vulnerability_action_plans";
-- reverse: create index "user_events_event_id_idx" to table: "user_events"
DROP INDEX "user_events_event_id_idx";
-- reverse: create "user_events" table
DROP TABLE "user_events";
-- reverse: create index "template_files_file_id_idx" to table: "template_files"
DROP INDEX "template_files_file_id_idx";
-- reverse: create "template_files" table
DROP TABLE "template_files";
-- reverse: create index "task_evidence_evidence_id_idx" to table: "task_evidence"
DROP INDEX "task_evidence_evidence_id_idx";
-- reverse: create "task_evidence" table
DROP TABLE "task_evidence";
-- reverse: create index "system_detail_assets_asset_id_idx" to table: "system_detail_assets"
DROP INDEX "system_detail_assets_asset_id_idx";
-- reverse: create "system_detail_assets" table
DROP TABLE "system_detail_assets";
-- reverse: create index "subscriber_events_event_id_idx" to table: "subscriber_events"
DROP INDEX "subscriber_events_event_id_idx";
-- reverse: create "subscriber_events" table
DROP TABLE "subscriber_events";
-- reverse: create index "subcontrol_identity_holders_identity_holder_id_idx" to table: "subcontrol_identity_holders"
DROP INDEX "subcontrol_identity_holders_identity_holder_id_idx";
-- reverse: create "subcontrol_identity_holders" table
DROP TABLE "subcontrol_identity_holders";
-- reverse: create index "subcontrol_entities_entity_id_idx" to table: "subcontrol_entities"
DROP INDEX "subcontrol_entities_entity_id_idx";
-- reverse: create "subcontrol_entities" table
DROP TABLE "subcontrol_entities";
-- reverse: create index "subcontrol_assets_asset_id_idx" to table: "subcontrol_assets"
DROP INDEX "subcontrol_assets_asset_id_idx";
-- reverse: create "subcontrol_assets" table
DROP TABLE "subcontrol_assets";
-- reverse: create index "subcontrol_control_implementations_control_implementation_id_id" to table: "subcontrol_control_implementations"
DROP INDEX "subcontrol_control_implementations_control_implementation_id_id";
-- reverse: create "subcontrol_control_implementations" table
DROP TABLE "subcontrol_control_implementations";
-- reverse: create index "subcontrol_scans_scan_id_idx" to table: "subcontrol_scans"
DROP INDEX "subcontrol_scans_scan_id_idx";
-- reverse: create "subcontrol_scans" table
DROP TABLE "subcontrol_scans";
-- reverse: create index "subcontrol_procedures_procedure_id_idx" to table: "subcontrol_procedures"
DROP INDEX "subcontrol_procedures_procedure_id_idx";
-- reverse: create "subcontrol_procedures" table
DROP TABLE "subcontrol_procedures";
-- reverse: create index "subcontrol_risks_risk_id_idx" to table: "subcontrol_risks"
DROP INDEX "subcontrol_risks_risk_id_idx";
-- reverse: create "subcontrol_risks" table
DROP TABLE "subcontrol_risks";
-- reverse: create index "subcontrol_tasks_task_id_idx" to table: "subcontrol_tasks"
DROP INDEX "subcontrol_tasks_task_id_idx";
-- reverse: create "subcontrol_tasks" table
DROP TABLE "subcontrol_tasks";
-- reverse: create index "subcontrol_control_objectives_control_objective_id_idx" to table: "subcontrol_control_objectives"
DROP INDEX "subcontrol_control_objectives_control_objective_id_idx";
-- reverse: create "subcontrol_control_objectives" table
DROP TABLE "subcontrol_control_objectives";
-- reverse: create index "scheduled_job_subcontrols_subcontrol_id_idx" to table: "scheduled_job_subcontrols"
DROP INDEX "scheduled_job_subcontrols_subcontrol_id_idx";
-- reverse: create "scheduled_job_subcontrols" table
DROP TABLE "scheduled_job_subcontrols";
-- reverse: create index "scheduled_job_controls_control_id_idx" to table: "scheduled_job_controls"
DROP INDEX "scheduled_job_controls_control_id_idx";
-- reverse: create "scheduled_job_controls" table
DROP TABLE "scheduled_job_controls";
-- reverse: create index "scan_tasks_task_id_idx" to table: "scan_tasks"
DROP INDEX "scan_tasks_task_id_idx";
-- reverse: create "scan_tasks" table
DROP TABLE "scan_tasks";
-- reverse: create index "scan_action_plans_action_plan_id_idx" to table: "scan_action_plans"
DROP INDEX "scan_action_plans_action_plan_id_idx";
-- reverse: create "scan_action_plans" table
DROP TABLE "scan_action_plans";
-- reverse: create index "scan_remediations_remediation_id_idx" to table: "scan_remediations"
DROP INDEX "scan_remediations_remediation_id_idx";
-- reverse: create "scan_remediations" table
DROP TABLE "scan_remediations";
-- reverse: create index "scan_files_file_id_idx" to table: "scan_files"
DROP INDEX "scan_files_file_id_idx";
-- reverse: create "scan_files" table
DROP TABLE "scan_files";
-- reverse: create index "scan_evidence_evidence_id_idx" to table: "scan_evidence"
DROP INDEX "scan_evidence_evidence_id_idx";
-- reverse: create "scan_evidence" table
DROP TABLE "scan_evidence";
-- reverse: create index "scan_entities_entity_id_idx" to table: "scan_entities"
DROP INDEX "scan_entities_entity_id_idx";
-- reverse: create "scan_entities" table
DROP TABLE "scan_entities";
-- reverse: create index "scan_assets_asset_id_idx" to table: "scan_assets"
DROP INDEX "scan_assets_asset_id_idx";
-- reverse: create "scan_assets" table
DROP TABLE "scan_assets";
-- reverse: create index "scan_editors_group_id_idx" to table: "scan_editors"
DROP INDEX "scan_editors_group_id_idx";
-- reverse: create "scan_editors" table
DROP TABLE "scan_editors";
-- reverse: create index "scan_blocked_groups_group_id_idx" to table: "scan_blocked_groups"
DROP INDEX "scan_blocked_groups_group_id_idx";
-- reverse: create "scan_blocked_groups" table
DROP TABLE "scan_blocked_groups";
-- reverse: create index "risk_tasks_task_id_idx" to table: "risk_tasks"
DROP INDEX "risk_tasks_task_id_idx";
-- reverse: create "risk_tasks" table
DROP TABLE "risk_tasks";
-- reverse: create index "risk_action_plans_action_plan_id_idx" to table: "risk_action_plans"
DROP INDEX "risk_action_plans_action_plan_id_idx";
-- reverse: create "risk_action_plans" table
DROP TABLE "risk_action_plans";
-- reverse: create index "risk_viewers_group_id_idx" to table: "risk_viewers"
DROP INDEX "risk_viewers_group_id_idx";
-- reverse: create "risk_viewers" table
DROP TABLE "risk_viewers";
-- reverse: create index "risk_editors_group_id_idx" to table: "risk_editors"
DROP INDEX "risk_editors_group_id_idx";
-- reverse: create "risk_editors" table
DROP TABLE "risk_editors";
-- reverse: create index "risk_blocked_groups_group_id_idx" to table: "risk_blocked_groups"
DROP INDEX "risk_blocked_groups_group_id_idx";
-- reverse: create "risk_blocked_groups" table
DROP TABLE "risk_blocked_groups";
-- reverse: create index "review_internal_policies_internal_policy_id_idx" to table: "review_internal_policies"
DROP INDEX "review_internal_policies_internal_policy_id_idx";
-- reverse: create "review_internal_policies" table
DROP TABLE "review_internal_policies";
-- reverse: create index "review_entities_entity_id_idx" to table: "review_entities"
DROP INDEX "review_entities_entity_id_idx";
-- reverse: create "review_entities" table
DROP TABLE "review_entities";
-- reverse: create index "review_assets_asset_id_idx" to table: "review_assets"
DROP INDEX "review_assets_asset_id_idx";
-- reverse: create "review_assets" table
DROP TABLE "review_assets";
-- reverse: create index "review_programs_program_id_idx" to table: "review_programs"
DROP INDEX "review_programs_program_id_idx";
-- reverse: create "review_programs" table
DROP TABLE "review_programs";
-- reverse: create index "review_risks_risk_id_idx" to table: "review_risks"
DROP INDEX "review_risks_risk_id_idx";
-- reverse: create "review_risks" table
DROP TABLE "review_risks";
-- reverse: create index "review_subcontrols_subcontrol_id_idx" to table: "review_subcontrols"
DROP INDEX "review_subcontrols_subcontrol_id_idx";
-- reverse: create "review_subcontrols" table
DROP TABLE "review_subcontrols";
-- reverse: create index "review_controls_control_id_idx" to table: "review_controls"
DROP INDEX "review_controls_control_id_idx";
-- reverse: create "review_controls" table
DROP TABLE "review_controls";
-- reverse: create index "review_remediations_remediation_id_idx" to table: "review_remediations"
DROP INDEX "review_remediations_remediation_id_idx";
-- reverse: create "review_remediations" table
DROP TABLE "review_remediations";
-- reverse: create index "review_action_plans_action_plan_id_idx" to table: "review_action_plans"
DROP INDEX "review_action_plans_action_plan_id_idx";
-- reverse: create "review_action_plans" table
DROP TABLE "review_action_plans";
-- reverse: create index "review_vulnerabilities_vulnerability_id_idx" to table: "review_vulnerabilities"
DROP INDEX "review_vulnerabilities_vulnerability_id_idx";
-- reverse: create "review_vulnerabilities" table
DROP TABLE "review_vulnerabilities";
-- reverse: create index "review_findings_finding_id_idx" to table: "review_findings"
DROP INDEX "review_findings_finding_id_idx";
-- reverse: create "review_findings" table
DROP TABLE "review_findings";
-- reverse: create index "review_editors_group_id_idx" to table: "review_editors"
DROP INDEX "review_editors_group_id_idx";
-- reverse: create "review_editors" table
DROP TABLE "review_editors";
-- reverse: create index "review_blocked_groups_group_id_idx" to table: "review_blocked_groups"
DROP INDEX "review_blocked_groups_group_id_idx";
-- reverse: create "review_blocked_groups" table
DROP TABLE "review_blocked_groups";
-- reverse: create index "remediation_entities_entity_id_idx" to table: "remediation_entities"
DROP INDEX "remediation_entities_entity_id_idx";
-- reverse: create "remediation_entities" table
DROP TABLE "remediation_entities";
-- reverse: create index "remediation_assets_asset_id_idx" to table: "remediation_assets"
DROP INDEX "remediation_assets_asset_id_idx";
-- reverse: create "remediation_assets" table
DROP TABLE "remediation_assets";
-- reverse: create index "remediation_programs_program_id_idx" to table: "remediation_programs"
DROP INDEX "remediation_programs_program_id_idx";
-- reverse: create "remediation_programs" table
DROP TABLE "remediation_programs";
-- reverse: create index "remediation_risks_risk_id_idx" to table: "remediation_risks"
DROP INDEX "remediation_risks_risk_id_idx";
-- reverse: create "remediation_risks" table
DROP TABLE "remediation_risks";
-- reverse: create index "remediation_subcontrols_subcontrol_id_idx" to table: "remediation_subcontrols"
DROP INDEX "remediation_subcontrols_subcontrol_id_idx";
-- reverse: create "remediation_subcontrols" table
DROP TABLE "remediation_subcontrols";
-- reverse: create index "remediation_controls_control_id_idx" to table: "remediation_controls"
DROP INDEX "remediation_controls_control_id_idx";
-- reverse: create "remediation_controls" table
DROP TABLE "remediation_controls";
-- reverse: create index "remediation_action_plans_action_plan_id_idx" to table: "remediation_action_plans"
DROP INDEX "remediation_action_plans_action_plan_id_idx";
-- reverse: create "remediation_action_plans" table
DROP TABLE "remediation_action_plans";
-- reverse: create index "remediation_vulnerabilities_vulnerability_id_idx" to table: "remediation_vulnerabilities"
DROP INDEX "remediation_vulnerabilities_vulnerability_id_idx";
-- reverse: create "remediation_vulnerabilities" table
DROP TABLE "remediation_vulnerabilities";
-- reverse: create index "remediation_findings_finding_id_idx" to table: "remediation_findings"
DROP INDEX "remediation_findings_finding_id_idx";
-- reverse: create "remediation_findings" table
DROP TABLE "remediation_findings";
-- reverse: create index "remediation_editors_group_id_idx" to table: "remediation_editors"
DROP INDEX "remediation_editors_group_id_idx";
-- reverse: create "remediation_editors" table
DROP TABLE "remediation_editors";
-- reverse: create index "remediation_blocked_groups_group_id_idx" to table: "remediation_blocked_groups"
DROP INDEX "remediation_blocked_groups_group_id_idx";
-- reverse: create "remediation_blocked_groups" table
DROP TABLE "remediation_blocked_groups";
-- reverse: create index "program_system_details_system_detail_id_idx" to table: "program_system_details"
DROP INDEX "program_system_details_system_detail_id_idx";
-- reverse: create "program_system_details" table
DROP TABLE "program_system_details";
-- reverse: create index "program_action_plans_action_plan_id_idx" to table: "program_action_plans"
DROP INDEX "program_action_plans_action_plan_id_idx";
-- reverse: create "program_action_plans" table
DROP TABLE "program_action_plans";
-- reverse: create index "program_narratives_narrative_id_idx" to table: "program_narratives"
DROP INDEX "program_narratives_narrative_id_idx";
-- reverse: create "program_narratives" table
DROP TABLE "program_narratives";
-- reverse: create index "program_evidence_evidence_id_idx" to table: "program_evidence"
DROP INDEX "program_evidence_evidence_id_idx";
-- reverse: create "program_evidence" table
DROP TABLE "program_evidence";
-- reverse: create index "program_files_file_id_idx" to table: "program_files"
DROP INDEX "program_files_file_id_idx";
-- reverse: create "program_files" table
DROP TABLE "program_files";
-- reverse: create index "program_tasks_task_id_idx" to table: "program_tasks"
DROP INDEX "program_tasks_task_id_idx";
-- reverse: create "program_tasks" table
DROP TABLE "program_tasks";
-- reverse: create index "program_risks_risk_id_idx" to table: "program_risks"
DROP INDEX "program_risks_risk_id_idx";
-- reverse: create "program_risks" table
DROP TABLE "program_risks";
-- reverse: create index "program_procedures_procedure_id_idx" to table: "program_procedures"
DROP INDEX "program_procedures_procedure_id_idx";
-- reverse: create "program_procedures" table
DROP TABLE "program_procedures";
-- reverse: create index "program_internal_policies_internal_policy_id_idx" to table: "program_internal_policies"
DROP INDEX "program_internal_policies_internal_policy_id_idx";
-- reverse: create "program_internal_policies" table
DROP TABLE "program_internal_policies";
-- reverse: create index "program_control_objectives_control_objective_id_idx" to table: "program_control_objectives"
DROP INDEX "program_control_objectives_control_objective_id_idx";
-- reverse: create "program_control_objectives" table
DROP TABLE "program_control_objectives";
-- reverse: create index "program_controls_control_id_idx" to table: "program_controls"
DROP INDEX "program_controls_control_id_idx";
-- reverse: create "program_controls" table
DROP TABLE "program_controls";
-- reverse: create index "program_viewers_group_id_idx" to table: "program_viewers"
DROP INDEX "program_viewers_group_id_idx";
-- reverse: create "program_viewers" table
DROP TABLE "program_viewers";
-- reverse: create index "program_editors_group_id_idx" to table: "program_editors"
DROP INDEX "program_editors_group_id_idx";
-- reverse: create "program_editors" table
DROP TABLE "program_editors";
-- reverse: create index "program_blocked_groups_group_id_idx" to table: "program_blocked_groups"
DROP INDEX "program_blocked_groups_group_id_idx";
-- reverse: create "program_blocked_groups" table
DROP TABLE "program_blocked_groups";
-- reverse: create index "procedure_tasks_task_id_idx" to table: "procedure_tasks"
DROP INDEX "procedure_tasks_task_id_idx";
-- reverse: create "procedure_tasks" table
DROP TABLE "procedure_tasks";
-- reverse: create index "procedure_risks_risk_id_idx" to table: "procedure_risks"
DROP INDEX "procedure_risks_risk_id_idx";
-- reverse: create "procedure_risks" table
DROP TABLE "procedure_risks";
-- reverse: create index "procedure_narratives_narrative_id_idx" to table: "procedure_narratives"
DROP INDEX "procedure_narratives_narrative_id_idx";
-- reverse: create "procedure_narratives" table
DROP TABLE "procedure_narratives";
-- reverse: create index "procedure_editors_group_id_idx" to table: "procedure_editors"
DROP INDEX "procedure_editors_group_id_idx";
-- reverse: create "procedure_editors" table
DROP TABLE "procedure_editors";
-- reverse: create index "procedure_blocked_groups_group_id_idx" to table: "procedure_blocked_groups"
DROP INDEX "procedure_blocked_groups_group_id_idx";
-- reverse: create "procedure_blocked_groups" table
DROP TABLE "procedure_blocked_groups";
-- reverse: create index "platform_system_details_system_detail_id_idx" to table: "platform_system_details"
DROP INDEX "platform_system_details_system_detail_id_idx";
-- reverse: create "platform_system_details" table
DROP TABLE "platform_system_details";
-- reverse: create index "platform_applicable_frameworks_standard_id_idx" to table: "platform_applicable_frameworks"
DROP INDEX "platform_applicable_frameworks_standard_id_idx";
-- reverse: create "platform_applicable_frameworks" table
DROP TABLE "platform_applicable_frameworks";
-- reverse: create index "platform_out_of_scope_vendors_entity_id_idx" to table: "platform_out_of_scope_vendors"
DROP INDEX "platform_out_of_scope_vendors_entity_id_idx";
-- reverse: create "platform_out_of_scope_vendors" table
DROP TABLE "platform_out_of_scope_vendors";
-- reverse: create index "platform_out_of_scope_assets_asset_id_idx" to table: "platform_out_of_scope_assets"
DROP INDEX "platform_out_of_scope_assets_asset_id_idx";
-- reverse: create "platform_out_of_scope_assets" table
DROP TABLE "platform_out_of_scope_assets";
-- reverse: create index "platform_source_entities_entity_id_idx" to table: "platform_source_entities"
DROP INDEX "platform_source_entities_entity_id_idx";
-- reverse: create "platform_source_entities" table
DROP TABLE "platform_source_entities";
-- reverse: create index "platform_identity_holders_identity_holder_id_idx" to table: "platform_identity_holders"
DROP INDEX "platform_identity_holders_identity_holder_id_idx";
-- reverse: create "platform_identity_holders" table
DROP TABLE "platform_identity_holders";
-- reverse: create index "platform_tasks_task_id_idx" to table: "platform_tasks"
DROP INDEX "platform_tasks_task_id_idx";
-- reverse: create "platform_tasks" table
DROP TABLE "platform_tasks";
-- reverse: create index "platform_scans_scan_id_idx" to table: "platform_scans"
DROP INDEX "platform_scans_scan_id_idx";
-- reverse: create "platform_scans" table
DROP TABLE "platform_scans";
-- reverse: create index "platform_assessments_assessment_id_idx" to table: "platform_assessments"
DROP INDEX "platform_assessments_assessment_id_idx";
-- reverse: create "platform_assessments" table
DROP TABLE "platform_assessments";
-- reverse: create index "platform_controls_control_id_idx" to table: "platform_controls"
DROP INDEX "platform_controls_control_id_idx";
-- reverse: create "platform_controls" table
DROP TABLE "platform_controls";
-- reverse: create index "platform_risks_risk_id_idx" to table: "platform_risks"
DROP INDEX "platform_risks_risk_id_idx";
-- reverse: create "platform_risks" table
DROP TABLE "platform_risks";
-- reverse: create index "platform_files_file_id_idx" to table: "platform_files"
DROP INDEX "platform_files_file_id_idx";
-- reverse: create "platform_files" table
DROP TABLE "platform_files";
-- reverse: create index "platform_evidence_evidence_id_idx" to table: "platform_evidence"
DROP INDEX "platform_evidence_evidence_id_idx";
-- reverse: create "platform_evidence" table
DROP TABLE "platform_evidence";
-- reverse: create index "platform_entities_entity_id_idx" to table: "platform_entities"
DROP INDEX "platform_entities_entity_id_idx";
-- reverse: create "platform_entities" table
DROP TABLE "platform_entities";
-- reverse: create index "platform_assets_asset_id_idx" to table: "platform_assets"
DROP INDEX "platform_assets_asset_id_idx";
-- reverse: create "platform_assets" table
DROP TABLE "platform_assets";
-- reverse: create index "platform_viewers_group_id_idx" to table: "platform_viewers"
DROP INDEX "platform_viewers_group_id_idx";
-- reverse: create "platform_viewers" table
DROP TABLE "platform_viewers";
-- reverse: create index "platform_editors_group_id_idx" to table: "platform_editors"
DROP INDEX "platform_editors_group_id_idx";
-- reverse: create "platform_editors" table
DROP TABLE "platform_editors";
-- reverse: create index "platform_blocked_groups_group_id_idx" to table: "platform_blocked_groups"
DROP INDEX "platform_blocked_groups_group_id_idx";
-- reverse: create "platform_blocked_groups" table
DROP TABLE "platform_blocked_groups";
-- reverse: create index "personal_access_token_events_event_id_idx" to table: "personal_access_token_events"
DROP INDEX "personal_access_token_events_event_id_idx";
-- reverse: create "personal_access_token_events" table
DROP TABLE "personal_access_token_events";
-- reverse: create index "organization_setting_files_file_id_idx" to table: "organization_setting_files"
DROP INDEX "organization_setting_files_file_id_idx";
-- reverse: create "organization_setting_files" table
DROP TABLE "organization_setting_files";
-- reverse: create index "organization_events_event_id_idx" to table: "organization_events"
DROP INDEX "organization_events_event_id_idx";
-- reverse: create "organization_events" table
DROP TABLE "organization_events";
-- reverse: create index "organization_files_file_id_idx" to table: "organization_files"
DROP INDEX "organization_files_file_id_idx";
-- reverse: create "organization_files" table
DROP TABLE "organization_files";
-- reverse: create index "organization_personal_access_tokens_personal_access_token_id_id" to table: "organization_personal_access_tokens"
DROP INDEX "organization_personal_access_tokens_personal_access_token_id_id";
-- reverse: create "organization_personal_access_tokens" table
DROP TABLE "organization_personal_access_tokens";
-- reverse: create index "org_subscription_events_event_id_idx" to table: "org_subscription_events"
DROP INDEX "org_subscription_events_event_id_idx";
-- reverse: create "org_subscription_events" table
DROP TABLE "org_subscription_events";
-- reverse: create index "org_product_org_prices_org_price_id_idx" to table: "org_product_org_prices"
DROP INDEX "org_product_org_prices_org_price_id_idx";
-- reverse: create "org_product_org_prices" table
DROP TABLE "org_product_org_prices";
-- reverse: create index "org_module_org_prices_org_price_id_idx" to table: "org_module_org_prices"
DROP INDEX "org_module_org_prices_org_price_id_idx";
-- reverse: create "org_module_org_prices" table
DROP TABLE "org_module_org_prices";
-- reverse: create index "org_membership_events_event_id_idx" to table: "org_membership_events"
DROP INDEX "org_membership_events_event_id_idx";
-- reverse: create "org_membership_events" table
DROP TABLE "org_membership_events";
-- reverse: create index "narrative_viewers_group_id_idx" to table: "narrative_viewers"
DROP INDEX "narrative_viewers_group_id_idx";
-- reverse: create "narrative_viewers" table
DROP TABLE "narrative_viewers";
-- reverse: create index "narrative_editors_group_id_idx" to table: "narrative_editors"
DROP INDEX "narrative_editors_group_id_idx";
-- reverse: create "narrative_editors" table
DROP TABLE "narrative_editors";
-- reverse: create index "narrative_blocked_groups_group_id_idx" to table: "narrative_blocked_groups"
DROP INDEX "narrative_blocked_groups_group_id_idx";
-- reverse: create "narrative_blocked_groups" table
DROP TABLE "narrative_blocked_groups";
-- reverse: create index "mapped_control_to_subcontrols_subcontrol_id_idx" to table: "mapped_control_to_subcontrols"
DROP INDEX "mapped_control_to_subcontrols_subcontrol_id_idx";
-- reverse: create "mapped_control_to_subcontrols" table
DROP TABLE "mapped_control_to_subcontrols";
-- reverse: create index "mapped_control_from_subcontrols_subcontrol_id_idx" to table: "mapped_control_from_subcontrols"
DROP INDEX "mapped_control_from_subcontrols_subcontrol_id_idx";
-- reverse: create "mapped_control_from_subcontrols" table
DROP TABLE "mapped_control_from_subcontrols";
-- reverse: create index "mapped_control_to_controls_control_id_idx" to table: "mapped_control_to_controls"
DROP INDEX "mapped_control_to_controls_control_id_idx";
-- reverse: create "mapped_control_to_controls" table
DROP TABLE "mapped_control_to_controls";
-- reverse: create index "mapped_control_from_controls_control_id_idx" to table: "mapped_control_from_controls"
DROP INDEX "mapped_control_from_controls_control_id_idx";
-- reverse: create "mapped_control_from_controls" table
DROP TABLE "mapped_control_from_controls";
-- reverse: create index "mapped_control_editors_group_id_idx" to table: "mapped_control_editors"
DROP INDEX "mapped_control_editors_group_id_idx";
-- reverse: create "mapped_control_editors" table
DROP TABLE "mapped_control_editors";
-- reverse: create index "mapped_control_blocked_groups_group_id_idx" to table: "mapped_control_blocked_groups"
DROP INDEX "mapped_control_blocked_groups_group_id_idx";
-- reverse: create "mapped_control_blocked_groups" table
DROP TABLE "mapped_control_blocked_groups";
-- reverse: create index "job_runner_job_runner_tokens_job_runner_token_id_idx" to table: "job_runner_job_runner_tokens"
DROP INDEX "job_runner_job_runner_tokens_job_runner_token_id_idx";
-- reverse: create "job_runner_job_runner_tokens" table
DROP TABLE "job_runner_job_runner_tokens";
-- reverse: create index "invite_groups_group_id_idx" to table: "invite_groups"
DROP INDEX "invite_groups_group_id_idx";
-- reverse: create "invite_groups" table
DROP TABLE "invite_groups";
-- reverse: create index "invite_events_event_id_idx" to table: "invite_events"
DROP INDEX "invite_events_event_id_idx";
-- reverse: create "invite_events" table
DROP TABLE "invite_events";
-- reverse: create index "internal_policy_identity_holders_identity_holder_id_idx" to table: "internal_policy_identity_holders"
DROP INDEX "internal_policy_identity_holders_identity_holder_id_idx";
-- reverse: create "internal_policy_identity_holders" table
DROP TABLE "internal_policy_identity_holders";
-- reverse: create index "internal_policy_entities_entity_id_idx" to table: "internal_policy_entities"
DROP INDEX "internal_policy_entities_entity_id_idx";
-- reverse: create "internal_policy_entities" table
DROP TABLE "internal_policy_entities";
-- reverse: create index "internal_policy_assets_asset_id_idx" to table: "internal_policy_assets"
DROP INDEX "internal_policy_assets_asset_id_idx";
-- reverse: create "internal_policy_assets" table
DROP TABLE "internal_policy_assets";
-- reverse: create index "internal_policy_risks_risk_id_idx" to table: "internal_policy_risks"
DROP INDEX "internal_policy_risks_risk_id_idx";
-- reverse: create "internal_policy_risks" table
DROP TABLE "internal_policy_risks";
-- reverse: create index "internal_policy_tasks_task_id_idx" to table: "internal_policy_tasks"
DROP INDEX "internal_policy_tasks_task_id_idx";
-- reverse: create "internal_policy_tasks" table
DROP TABLE "internal_policy_tasks";
-- reverse: create index "internal_policy_narratives_narrative_id_idx" to table: "internal_policy_narratives"
DROP INDEX "internal_policy_narratives_narrative_id_idx";
-- reverse: create "internal_policy_narratives" table
DROP TABLE "internal_policy_narratives";
-- reverse: create index "internal_policy_procedures_procedure_id_idx" to table: "internal_policy_procedures"
DROP INDEX "internal_policy_procedures_procedure_id_idx";
-- reverse: create "internal_policy_procedures" table
DROP TABLE "internal_policy_procedures";
-- reverse: create index "internal_policy_subcontrols_subcontrol_id_idx" to table: "internal_policy_subcontrols"
DROP INDEX "internal_policy_subcontrols_subcontrol_id_idx";
-- reverse: create "internal_policy_subcontrols" table
DROP TABLE "internal_policy_subcontrols";
-- reverse: create index "internal_policy_controls_control_id_idx" to table: "internal_policy_controls"
DROP INDEX "internal_policy_controls_control_id_idx";
-- reverse: create "internal_policy_controls" table
DROP TABLE "internal_policy_controls";
-- reverse: create index "internal_policy_control_objectives_control_objective_id_idx" to table: "internal_policy_control_objectives"
DROP INDEX "internal_policy_control_objectives_control_objective_id_idx";
-- reverse: create "internal_policy_control_objectives" table
DROP TABLE "internal_policy_control_objectives";
-- reverse: create index "internal_policy_editors_group_id_idx" to table: "internal_policy_editors"
DROP INDEX "internal_policy_editors_group_id_idx";
-- reverse: create "internal_policy_editors" table
DROP TABLE "internal_policy_editors";
-- reverse: create index "internal_policy_blocked_groups_group_id_idx" to table: "internal_policy_blocked_groups"
DROP INDEX "internal_policy_blocked_groups_group_id_idx";
-- reverse: create "internal_policy_blocked_groups" table
DROP TABLE "internal_policy_blocked_groups";
-- reverse: create index "integration_action_plans_action_plan_id_idx" to table: "integration_action_plans"
DROP INDEX "integration_action_plans_action_plan_id_idx";
-- reverse: create "integration_action_plans" table
DROP TABLE "integration_action_plans";
-- reverse: create index "integration_remediations_remediation_id_idx" to table: "integration_remediations"
DROP INDEX "integration_remediations_remediation_id_idx";
-- reverse: create "integration_remediations" table
DROP TABLE "integration_remediations";
-- reverse: create index "integration_reviews_review_id_idx" to table: "integration_reviews"
DROP INDEX "integration_reviews_review_id_idx";
-- reverse: create "integration_reviews" table
DROP TABLE "integration_reviews";
-- reverse: create index "integration_internal_policies_internal_policy_id_idx" to table: "integration_internal_policies"
DROP INDEX "integration_internal_policies_internal_policy_id_idx";
-- reverse: create "integration_internal_policies" table
DROP TABLE "integration_internal_policies";
-- reverse: create index "integration_vulnerabilities_vulnerability_id_idx" to table: "integration_vulnerabilities"
DROP INDEX "integration_vulnerabilities_vulnerability_id_idx";
-- reverse: create "integration_vulnerabilities" table
DROP TABLE "integration_vulnerabilities";
-- reverse: create index "integration_findings_finding_id_idx" to table: "integration_findings"
DROP INDEX "integration_findings_finding_id_idx";
-- reverse: create "integration_findings" table
DROP TABLE "integration_findings";
-- reverse: create index "integration_events_event_id_idx" to table: "integration_events"
DROP INDEX "integration_events_event_id_idx";
-- reverse: create "integration_events" table
DROP TABLE "integration_events";
-- reverse: create index "integration_secrets_hush_id_idx" to table: "integration_secrets"
DROP INDEX "integration_secrets_hush_id_idx";
-- reverse: create "integration_secrets" table
DROP TABLE "integration_secrets";
-- reverse: create index "identity_holder_files_file_id_idx" to table: "identity_holder_files"
DROP INDEX "identity_holder_files_file_id_idx";
-- reverse: create "identity_holder_files" table
DROP TABLE "identity_holder_files";
-- reverse: create index "identity_holder_tasks_task_id_idx" to table: "identity_holder_tasks"
DROP INDEX "identity_holder_tasks_task_id_idx";
-- reverse: create "identity_holder_tasks" table
DROP TABLE "identity_holder_tasks";
-- reverse: create index "identity_holder_entities_entity_id_idx" to table: "identity_holder_entities"
DROP INDEX "identity_holder_entities_entity_id_idx";
-- reverse: create "identity_holder_entities" table
DROP TABLE "identity_holder_entities";
-- reverse: create index "identity_holder_assets_asset_id_idx" to table: "identity_holder_assets"
DROP INDEX "identity_holder_assets_asset_id_idx";
-- reverse: create "identity_holder_assets" table
DROP TABLE "identity_holder_assets";
-- reverse: create index "identity_holder_templates_template_id_idx" to table: "identity_holder_templates"
DROP INDEX "identity_holder_templates_template_id_idx";
-- reverse: create "identity_holder_templates" table
DROP TABLE "identity_holder_templates";
-- reverse: create index "identity_holder_assessments_assessment_id_idx" to table: "identity_holder_assessments"
DROP INDEX "identity_holder_assessments_assessment_id_idx";
-- reverse: create "identity_holder_assessments" table
DROP TABLE "identity_holder_assessments";
-- reverse: create index "hush_events_event_id_idx" to table: "hush_events"
DROP INDEX "hush_events_event_id_idx";
-- reverse: create "hush_events" table
DROP TABLE "hush_events";
-- reverse: create index "group_membership_events_event_id_idx" to table: "group_membership_events"
DROP INDEX "group_membership_events_event_id_idx";
-- reverse: create "group_membership_events" table
DROP TABLE "group_membership_events";
-- reverse: create index "group_tasks_task_id_idx" to table: "group_tasks"
DROP INDEX "group_tasks_task_id_idx";
-- reverse: create "group_tasks" table
DROP TABLE "group_tasks";
-- reverse: create index "group_files_file_id_idx" to table: "group_files"
DROP INDEX "group_files_file_id_idx";
-- reverse: create "group_files" table
DROP TABLE "group_files";
-- reverse: create index "group_events_event_id_idx" to table: "group_events"
DROP INDEX "group_events_event_id_idx";
-- reverse: create "group_events" table
DROP TABLE "group_events";
-- reverse: create index "finding_check_results_check_result_id_idx" to table: "finding_check_results"
DROP INDEX "finding_check_results_check_result_id_idx";
-- reverse: create "finding_check_results" table
DROP TABLE "finding_check_results";
-- reverse: create index "finding_identity_holders_identity_holder_id_idx" to table: "finding_identity_holders"
DROP INDEX "finding_identity_holders_identity_holder_id_idx";
-- reverse: create "finding_identity_holders" table
DROP TABLE "finding_identity_holders";
-- reverse: create index "finding_directory_accounts_directory_account_id_idx" to table: "finding_directory_accounts"
DROP INDEX "finding_directory_accounts_directory_account_id_idx";
-- reverse: create "finding_directory_accounts" table
DROP TABLE "finding_directory_accounts";
-- reverse: create index "finding_tasks_task_id_idx" to table: "finding_tasks"
DROP INDEX "finding_tasks_task_id_idx";
-- reverse: create "finding_tasks" table
DROP TABLE "finding_tasks";
-- reverse: create index "finding_scans_scan_id_idx" to table: "finding_scans"
DROP INDEX "finding_scans_scan_id_idx";
-- reverse: create "finding_scans" table
DROP TABLE "finding_scans";
-- reverse: create index "finding_entities_entity_id_idx" to table: "finding_entities"
DROP INDEX "finding_entities_entity_id_idx";
-- reverse: create "finding_entities" table
DROP TABLE "finding_entities";
-- reverse: create index "finding_assets_asset_id_idx" to table: "finding_assets"
DROP INDEX "finding_assets_asset_id_idx";
-- reverse: create "finding_assets" table
DROP TABLE "finding_assets";
-- reverse: create index "finding_programs_program_id_idx" to table: "finding_programs"
DROP INDEX "finding_programs_program_id_idx";
-- reverse: create "finding_programs" table
DROP TABLE "finding_programs";
-- reverse: create index "finding_risks_risk_id_idx" to table: "finding_risks"
DROP INDEX "finding_risks_risk_id_idx";
-- reverse: create "finding_risks" table
DROP TABLE "finding_risks";
-- reverse: create index "finding_subcontrols_subcontrol_id_idx" to table: "finding_subcontrols"
DROP INDEX "finding_subcontrols_subcontrol_id_idx";
-- reverse: create "finding_subcontrols" table
DROP TABLE "finding_subcontrols";
-- reverse: create index "finding_action_plans_action_plan_id_idx" to table: "finding_action_plans"
DROP INDEX "finding_action_plans_action_plan_id_idx";
-- reverse: create "finding_action_plans" table
DROP TABLE "finding_action_plans";
-- reverse: create index "finding_vulnerabilities_vulnerability_id_idx" to table: "finding_vulnerabilities"
DROP INDEX "finding_vulnerabilities_vulnerability_id_idx";
-- reverse: create "finding_vulnerabilities" table
DROP TABLE "finding_vulnerabilities";
-- reverse: create index "finding_editors_group_id_idx" to table: "finding_editors"
DROP INDEX "finding_editors_group_id_idx";
-- reverse: create "finding_editors" table
DROP TABLE "finding_editors";
-- reverse: create index "finding_blocked_groups_group_id_idx" to table: "finding_blocked_groups"
DROP INDEX "finding_blocked_groups_group_id_idx";
-- reverse: create "finding_blocked_groups" table
DROP TABLE "finding_blocked_groups";
-- reverse: create index "file_secrets_hush_id_idx" to table: "file_secrets"
DROP INDEX "file_secrets_hush_id_idx";
-- reverse: create "file_secrets" table
DROP TABLE "file_secrets";
-- reverse: create index "file_events_event_id_idx" to table: "file_events"
DROP INDEX "file_events_event_id_idx";
-- reverse: create "file_events" table
DROP TABLE "file_events";
-- reverse: create index "evidence_files_file_id_idx" to table: "evidence_files"
DROP INDEX "evidence_files_file_id_idx";
-- reverse: create "evidence_files" table
DROP TABLE "evidence_files";
-- reverse: create index "evidence_control_objectives_control_objective_id_idx" to table: "evidence_control_objectives"
DROP INDEX "evidence_control_objectives_control_objective_id_idx";
-- reverse: create "evidence_control_objectives" table
DROP TABLE "evidence_control_objectives";
-- reverse: create index "evidence_subcontrols_subcontrol_id_idx" to table: "evidence_subcontrols"
DROP INDEX "evidence_subcontrols_subcontrol_id_idx";
-- reverse: create "evidence_subcontrols" table
DROP TABLE "evidence_subcontrols";
-- reverse: create index "evidence_controls_control_id_idx" to table: "evidence_controls"
DROP INDEX "evidence_controls_control_id_idx";
-- reverse: create "evidence_controls" table
DROP TABLE "evidence_controls";
-- reverse: create index "entity_subprocessors_subprocessor_id_idx" to table: "entity_subprocessors"
DROP INDEX "entity_subprocessors_subprocessor_id_idx";
-- reverse: create "entity_subprocessors" table
DROP TABLE "entity_subprocessors";
-- reverse: create index "entity_integrations_integration_id_idx" to table: "entity_integrations"
DROP INDEX "entity_integrations_integration_id_idx";
-- reverse: create "entity_integrations" table
DROP TABLE "entity_integrations";
-- reverse: create index "entity_system_details_system_detail_id_idx" to table: "entity_system_details"
DROP INDEX "entity_system_details_system_detail_id_idx";
-- reverse: create "entity_system_details" table
DROP TABLE "entity_system_details";
-- reverse: create index "entity_assets_asset_id_idx" to table: "entity_assets"
DROP INDEX "entity_assets_asset_id_idx";
-- reverse: create "entity_assets" table
DROP TABLE "entity_assets";
-- reverse: create index "entity_files_file_id_idx" to table: "entity_files"
DROP INDEX "entity_files_file_id_idx";
-- reverse: create "entity_files" table
DROP TABLE "entity_files";
-- reverse: create index "entity_documents_document_data_id_idx" to table: "entity_documents"
DROP INDEX "entity_documents_document_data_id_idx";
-- reverse: create "entity_documents" table
DROP TABLE "entity_documents";
-- reverse: create index "entity_contacts_contact_id_idx" to table: "entity_contacts"
DROP INDEX "entity_contacts_contact_id_idx";
-- reverse: create "entity_contacts" table
DROP TABLE "entity_contacts";
-- reverse: create index "entity_editors_group_id_idx" to table: "entity_editors"
DROP INDEX "entity_editors_group_id_idx";
-- reverse: create "entity_editors" table
DROP TABLE "entity_editors";
-- reverse: create index "entity_blocked_groups_group_id_idx" to table: "entity_blocked_groups"
DROP INDEX "entity_blocked_groups_group_id_idx";
-- reverse: create "entity_blocked_groups" table
DROP TABLE "entity_blocked_groups";
-- reverse: create index "document_data_files_file_id_idx" to table: "document_data_files"
DROP INDEX "document_data_files_file_id_idx";
-- reverse: create "document_data_files" table
DROP TABLE "document_data_files";
-- reverse: create index "control_objective_tasks_task_id_idx" to table: "control_objective_tasks"
DROP INDEX "control_objective_tasks_task_id_idx";
-- reverse: create "control_objective_tasks" table
DROP TABLE "control_objective_tasks";
-- reverse: create index "control_objective_viewers_group_id_idx" to table: "control_objective_viewers"
DROP INDEX "control_objective_viewers_group_id_idx";
-- reverse: create "control_objective_viewers" table
DROP TABLE "control_objective_viewers";
-- reverse: create index "control_objective_editors_group_id_idx" to table: "control_objective_editors"
DROP INDEX "control_objective_editors_group_id_idx";
-- reverse: create "control_objective_editors" table
DROP TABLE "control_objective_editors";
-- reverse: create index "control_objective_blocked_groups_group_id_idx" to table: "control_objective_blocked_groups"
DROP INDEX "control_objective_blocked_groups_group_id_idx";
-- reverse: create "control_objective_blocked_groups" table
DROP TABLE "control_objective_blocked_groups";
-- reverse: create index "control_implementation_tasks_task_id_idx" to table: "control_implementation_tasks"
DROP INDEX "control_implementation_tasks_task_id_idx";
-- reverse: create "control_implementation_tasks" table
DROP TABLE "control_implementation_tasks";
-- reverse: create index "control_implementation_viewers_group_id_idx" to table: "control_implementation_viewers"
DROP INDEX "control_implementation_viewers_group_id_idx";
-- reverse: create "control_implementation_viewers" table
DROP TABLE "control_implementation_viewers";
-- reverse: create index "control_implementation_editors_group_id_idx" to table: "control_implementation_editors"
DROP INDEX "control_implementation_editors_group_id_idx";
-- reverse: create "control_implementation_editors" table
DROP TABLE "control_implementation_editors";
-- reverse: create index "control_implementation_blocked_groups_group_id_idx" to table: "control_implementation_blocked_groups"
DROP INDEX "control_implementation_blocked_groups_group_id_idx";
-- reverse: create "control_implementation_blocked_groups" table
DROP TABLE "control_implementation_blocked_groups";
-- reverse: create index "control_control_implementations_control_implementation_id_idx" to table: "control_control_implementations"
DROP INDEX "control_control_implementations_control_implementation_id_idx";
-- reverse: create "control_control_implementations" table
DROP TABLE "control_control_implementations";
-- reverse: create index "control_campaigns_campaign_id_idx" to table: "control_campaigns"
DROP INDEX "control_campaigns_campaign_id_idx";
-- reverse: create "control_campaigns" table
DROP TABLE "control_campaigns";
-- reverse: create index "control_identity_holders_identity_holder_id_idx" to table: "control_identity_holders"
DROP INDEX "control_identity_holders_identity_holder_id_idx";
-- reverse: create "control_identity_holders" table
DROP TABLE "control_identity_holders";
-- reverse: create index "control_entities_entity_id_idx" to table: "control_entities"
DROP INDEX "control_entities_entity_id_idx";
-- reverse: create "control_entities" table
DROP TABLE "control_entities";
-- reverse: create index "control_assets_asset_id_idx" to table: "control_assets"
DROP INDEX "control_assets_asset_id_idx";
-- reverse: create "control_assets" table
DROP TABLE "control_assets";
-- reverse: create index "control_editors_group_id_idx" to table: "control_editors"
DROP INDEX "control_editors_group_id_idx";
-- reverse: create "control_editors" table
DROP TABLE "control_editors";
-- reverse: create index "control_blocked_groups_group_id_idx" to table: "control_blocked_groups"
DROP INDEX "control_blocked_groups_group_id_idx";
-- reverse: create "control_blocked_groups" table
DROP TABLE "control_blocked_groups";
-- reverse: create index "control_scans_scan_id_idx" to table: "control_scans"
DROP INDEX "control_scans_scan_id_idx";
-- reverse: create "control_scans" table
DROP TABLE "control_scans";
-- reverse: create index "control_procedures_procedure_id_idx" to table: "control_procedures"
DROP INDEX "control_procedures_procedure_id_idx";
-- reverse: create "control_procedures" table
DROP TABLE "control_procedures";
-- reverse: create index "control_action_plans_action_plan_id_idx" to table: "control_action_plans"
DROP INDEX "control_action_plans_action_plan_id_idx";
-- reverse: create "control_action_plans" table
DROP TABLE "control_action_plans";
-- reverse: create index "control_risks_risk_id_idx" to table: "control_risks"
DROP INDEX "control_risks_risk_id_idx";
-- reverse: create "control_risks" table
DROP TABLE "control_risks";
-- reverse: create index "control_narratives_narrative_id_idx" to table: "control_narratives"
DROP INDEX "control_narratives_narrative_id_idx";
-- reverse: create "control_narratives" table
DROP TABLE "control_narratives";
-- reverse: create index "control_tasks_task_id_idx" to table: "control_tasks"
DROP INDEX "control_tasks_task_id_idx";
-- reverse: create "control_tasks" table
DROP TABLE "control_tasks";
-- reverse: create index "control_control_objectives_control_objective_id_idx" to table: "control_control_objectives"
DROP INDEX "control_control_objectives_control_objective_id_idx";
-- reverse: create "control_control_objectives" table
DROP TABLE "control_control_objectives";
-- reverse: create index "contact_files_file_id_idx" to table: "contact_files"
DROP INDEX "contact_files_file_id_idx";
-- reverse: create "contact_files" table
DROP TABLE "contact_files";
-- reverse: create index "check_result_controls_control_id_idx" to table: "check_result_controls"
DROP INDEX "check_result_controls_control_id_idx";
-- reverse: create "check_result_controls" table
DROP TABLE "check_result_controls";
-- reverse: create index "campaign_identity_holders_identity_holder_id_idx" to table: "campaign_identity_holders"
DROP INDEX "campaign_identity_holders_identity_holder_id_idx";
-- reverse: create "campaign_identity_holders" table
DROP TABLE "campaign_identity_holders";
-- reverse: create index "campaign_groups_group_id_idx" to table: "campaign_groups"
DROP INDEX "campaign_groups_group_id_idx";
-- reverse: create "campaign_groups" table
DROP TABLE "campaign_groups";
-- reverse: create index "campaign_users_user_id_idx" to table: "campaign_users"
DROP INDEX "campaign_users_user_id_idx";
-- reverse: create "campaign_users" table
DROP TABLE "campaign_users";
-- reverse: create index "campaign_contacts_contact_id_idx" to table: "campaign_contacts"
DROP INDEX "campaign_contacts_contact_id_idx";
-- reverse: create "campaign_contacts" table
DROP TABLE "campaign_contacts";
-- reverse: create index "campaign_viewers_group_id_idx" to table: "campaign_viewers"
DROP INDEX "campaign_viewers_group_id_idx";
-- reverse: create "campaign_viewers" table
DROP TABLE "campaign_viewers";
-- reverse: create index "campaign_editors_group_id_idx" to table: "campaign_editors"
DROP INDEX "campaign_editors_group_id_idx";
-- reverse: create "campaign_editors" table
DROP TABLE "campaign_editors";
-- reverse: create index "campaign_blocked_groups_group_id_idx" to table: "campaign_blocked_groups"
DROP INDEX "campaign_blocked_groups_group_id_idx";
-- reverse: create "campaign_blocked_groups" table
DROP TABLE "campaign_blocked_groups";
-- reverse: create index "asset_connected_assets_connected_from_id_idx" to table: "asset_connected_assets"
DROP INDEX "asset_connected_assets_connected_from_id_idx";
-- reverse: create "asset_connected_assets" table
DROP TABLE "asset_connected_assets";
-- reverse: create index "action_plan_tasks_task_id_idx" to table: "action_plan_tasks"
DROP INDEX "action_plan_tasks_task_id_idx";
-- reverse: create "action_plan_tasks" table
DROP TABLE "action_plan_tasks";
-- reverse: create index "action_plan_viewers_group_id_idx" to table: "action_plan_viewers"
DROP INDEX "action_plan_viewers_group_id_idx";
-- reverse: create "action_plan_viewers" table
DROP TABLE "action_plan_viewers";
-- reverse: create index "action_plan_editors_group_id_idx" to table: "action_plan_editors"
DROP INDEX "action_plan_editors_group_id_idx";
-- reverse: create "action_plan_editors" table
DROP TABLE "action_plan_editors";
-- reverse: create index "action_plan_blocked_groups_group_id_idx" to table: "action_plan_blocked_groups"
DROP INDEX "action_plan_blocked_groups_group_id_idx";
-- reverse: create "action_plan_blocked_groups" table
DROP TABLE "action_plan_blocked_groups";
-- reverse: create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
DROP INDEX "workflowproposal_workflow_object_ref_id_domain_key";
-- reverse: create index "workflow_proposal_submitted_by_user_id_idx" to table: "workflow_proposals"
DROP INDEX "workflow_proposal_submitted_by_user_id_idx";
-- reverse: create index "workflow_proposal_owner_id_idx" to table: "workflow_proposals"
DROP INDEX "workflow_proposal_owner_id_idx";
-- reverse: create "workflow_proposals" table
DROP TABLE "workflow_proposals";
-- reverse: create index "workflowobjectref_workflow_instance_id_vulnerability_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_vulnerability_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_task_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_task_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_subcontrol_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_subcontrol_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_risk_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_risk_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_remediation_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_remediation_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_procedure_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_procedure_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_platform_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_platform_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_internal_policy_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_internal_policy_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_identity_holder_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_identity_holder_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_finding_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_finding_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_evidence_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_evidence_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_directory_membership_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_directory_membership_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_directory_group_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_directory_group_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_directory_account_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_directory_account_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_control_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_control_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_campaign_target_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_campaign_target_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_campaign_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_campaign_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_assessment_response_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_assessment_response_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_assessment_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_assessment_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_action_plan_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_action_plan_id";
-- reverse: create index "workflowobjectref_display_id_owner_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_display_id_owner_id";
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
-- reverse: create "workflow_object_refs" table
DROP TABLE "workflow_object_refs";
-- reverse: create index "workflowinstance_workflow_definition_id" to table: "workflow_instances"
DROP INDEX "workflowinstance_workflow_definition_id";
-- reverse: create index "workflowinstance_display_id_owner_id" to table: "workflow_instances"
DROP INDEX "workflowinstance_display_id_owner_id";
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
-- reverse: create "workflow_instances" table
DROP TABLE "workflow_instances";
-- reverse: create index "workflowevent_display_id_owner_id" to table: "workflow_events"
DROP INDEX "workflowevent_display_id_owner_id";
-- reverse: create index "workflow_event_workflow_instance_id_idx" to table: "workflow_events"
DROP INDEX "workflow_event_workflow_instance_id_idx";
-- reverse: create index "workflow_event_owner_id_idx" to table: "workflow_events"
DROP INDEX "workflow_event_owner_id_idx";
-- reverse: create "workflow_events" table
DROP TABLE "workflow_events";
-- reverse: create index "workflowdefinition_display_id_owner_id" to table: "workflow_definitions"
DROP INDEX "workflowdefinition_display_id_owner_id";
-- reverse: create index "workflow_definition_owner_id_idx" to table: "workflow_definitions"
DROP INDEX "workflow_definition_owner_id_idx";
-- reverse: create "workflow_definitions" table
DROP TABLE "workflow_definitions";
-- reverse: create index "workflowassignmenttarget_workflow_assignment_id" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_workflow_assignment_id";
-- reverse: create index "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc";
-- reverse: create index "workflowassignmenttarget_display_id_owner_id" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_display_id_owner_id";
-- reverse: create index "workflow_assignment_target_target_user_id_idx" to table: "workflow_assignment_targets"
DROP INDEX "workflow_assignment_target_target_user_id_idx";
-- reverse: create index "workflow_assignment_target_target_group_id_idx" to table: "workflow_assignment_targets"
DROP INDEX "workflow_assignment_target_target_group_id_idx";
-- reverse: create index "workflow_assignment_target_owner_id_idx" to table: "workflow_assignment_targets"
DROP INDEX "workflow_assignment_target_owner_id_idx";
-- reverse: create "workflow_assignment_targets" table
DROP TABLE "workflow_assignment_targets";
-- reverse: create index "workflowassignment_workflow_instance_id_assignment_key" to table: "workflow_assignments"
DROP INDEX "workflowassignment_workflow_instance_id_assignment_key";
-- reverse: create index "workflowassignment_display_id_owner_id" to table: "workflow_assignments"
DROP INDEX "workflowassignment_display_id_owner_id";
-- reverse: create index "workflow_assignment_owner_id_idx" to table: "workflow_assignments"
DROP INDEX "workflow_assignment_owner_id_idx";
-- reverse: create index "workflow_assignment_actor_user_id_idx" to table: "workflow_assignments"
DROP INDEX "workflow_assignment_actor_user_id_idx";
-- reverse: create index "workflow_assignment_actor_group_id_idx" to table: "workflow_assignments"
DROP INDEX "workflow_assignment_actor_group_id_idx";
-- reverse: create "workflow_assignments" table
DROP TABLE "workflow_assignments";
-- reverse: create index "webauthns_owner_id_fk" to table: "webauthns"
DROP INDEX "webauthns_owner_id_fk";
-- reverse: create index "webauthns_credential_id_key" to table: "webauthns"
DROP INDEX "webauthns_credential_id_key";
-- reverse: create "webauthns" table
DROP TABLE "webauthns";
-- reverse: create index "vulnerability_owner_id_idx" to table: "vulnerabilities"
DROP INDEX "vulnerability_owner_id_idx";
-- reverse: create index "vulnerability_external_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_external_id_owner_id";
-- reverse: create index "vulnerability_display_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_display_id_owner_id";
-- reverse: create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_cve_id_owner_id";
-- reverse: create "vulnerabilities" table
DROP TABLE "vulnerabilities";
-- reverse: create index "vendor_scoring_config_owner_id_idx" to table: "vendor_scoring_configs"
DROP INDEX "vendor_scoring_config_owner_id_idx";
-- reverse: create "vendor_scoring_configs" table
DROP TABLE "vendor_scoring_configs";
-- reverse: create index "vendor_risk_score_vendor_scoring_config_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_vendor_scoring_config_id_idx";
-- reverse: create index "vendor_risk_score_owner_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_owner_id_idx";
-- reverse: create index "vendor_risk_score_entity_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_entity_id_idx";
-- reverse: create index "vendor_risk_score_assessment_response_id_idx" to table: "vendor_risk_scores"
DROP INDEX "vendor_risk_score_assessment_response_id_idx";
-- reverse: create "vendor_risk_scores" table
DROP TABLE "vendor_risk_scores";
-- reverse: create index "user_settings_user_id_key" to table: "user_settings"
DROP INDEX "user_settings_user_id_key";
-- reverse: create index "user_setting_user_id_idx" to table: "user_settings"
DROP INDEX "user_setting_user_id_idx";
-- reverse: create "user_settings" table
DROP TABLE "user_settings";
-- reverse: create index "users_sub_key" to table: "users"
DROP INDEX "users_sub_key";
-- reverse: create index "users_display_id_key" to table: "users"
DROP INDEX "users_display_id_key";
-- reverse: create index "user_email" to table: "users"
DROP INDEX "user_email";
-- reverse: create "users" table
DROP TABLE "users";
-- reverse: create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
DROP INDEX "trustcenterwatermarkconfig_trust_center_id";
-- reverse: create index "trust_center_watermark_config_owner_id_idx" to table: "trust_center_watermark_configs"
DROP INDEX "trust_center_watermark_config_owner_id_idx";
-- reverse: create index "trust_center_watermark_config_logo_id_idx" to table: "trust_center_watermark_configs"
DROP INDEX "trust_center_watermark_config_logo_id_idx";
-- reverse: create "trust_center_watermark_configs" table
DROP TABLE "trust_center_watermark_configs";
-- reverse: create index "trustcentersubprocessor_subprocessor_id_trust_center_id" to table: "trust_center_subprocessors"
DROP INDEX "trustcentersubprocessor_subprocessor_id_trust_center_id";
-- reverse: create index "trust_center_subprocessor_trust_center_id_idx" to table: "trust_center_subprocessors"
DROP INDEX "trust_center_subprocessor_trust_center_id_idx";
-- reverse: create "trust_center_subprocessors" table
DROP TABLE "trust_center_subprocessors";
-- reverse: create index "trustcentersetting_trust_center_id_environment" to table: "trust_center_settings"
DROP INDEX "trustcentersetting_trust_center_id_environment";
-- reverse: create index "trust_center_setting_nda_approver_group_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_nda_approver_group_id_idx";
-- reverse: create index "trust_center_setting_logo_local_file_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_logo_local_file_id_idx";
-- reverse: create index "trust_center_setting_hero_image_local_file_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_hero_image_local_file_id_idx";
-- reverse: create index "trust_center_setting_favicon_local_file_id_idx" to table: "trust_center_settings"
DROP INDEX "trust_center_setting_favicon_local_file_id_idx";
-- reverse: create "trust_center_settings" table
DROP TABLE "trust_center_settings";
-- reverse: create index "trust_center_nda_request_trust_center_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_trust_center_id_idx";
-- reverse: create index "trust_center_nda_request_file_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_file_id_idx";
-- reverse: create index "trust_center_nda_request_document_data_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_document_data_id_idx";
-- reverse: create index "trust_center_nda_request_approved_by_user_id_idx" to table: "trust_center_nda_requests"
DROP INDEX "trust_center_nda_request_approved_by_user_id_idx";
-- reverse: create "trust_center_nda_requests" table
DROP TABLE "trust_center_nda_requests";
-- reverse: create index "trustcenterfaq_note_id_trust_center_id" to table: "trust_center_faqs"
DROP INDEX "trustcenterfaq_note_id_trust_center_id";
-- reverse: create index "trust_center_faq_trust_center_id_idx" to table: "trust_center_faqs"
DROP INDEX "trust_center_faq_trust_center_id_idx";
-- reverse: create "trust_center_faqs" table
DROP TABLE "trust_center_faqs";
-- reverse: create index "trust_center_entity_trust_center_id_idx" to table: "trust_center_entities"
DROP INDEX "trust_center_entity_trust_center_id_idx";
-- reverse: create index "trust_center_entity_logo_file_id_idx" to table: "trust_center_entities"
DROP INDEX "trust_center_entity_logo_file_id_idx";
-- reverse: create index "trust_center_entity_entity_type_id_idx" to table: "trust_center_entities"
DROP INDEX "trust_center_entity_entity_type_id_idx";
-- reverse: create "trust_center_entities" table
DROP TABLE "trust_center_entities";
-- reverse: create index "trust_center_doc_trust_center_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_trust_center_id_idx";
-- reverse: create index "trust_center_doc_standard_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_standard_id_idx";
-- reverse: create index "trust_center_doc_original_file_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_original_file_id_idx";
-- reverse: create index "trust_center_doc_file_id_idx" to table: "trust_center_docs"
DROP INDEX "trust_center_doc_file_id_idx";
-- reverse: create "trust_center_docs" table
DROP TABLE "trust_center_docs";
-- reverse: create index "trustcentercompliance_standard_id_trust_center_id" to table: "trust_center_compliances"
DROP INDEX "trustcentercompliance_standard_id_trust_center_id";
-- reverse: create index "trust_center_compliance_trust_center_id_idx" to table: "trust_center_compliances"
DROP INDEX "trust_center_compliance_trust_center_id_idx";
-- reverse: create "trust_center_compliances" table
DROP TABLE "trust_center_compliances";
-- reverse: create index "trustcenter_slug" to table: "trust_centers"
DROP INDEX "trustcenter_slug";
-- reverse: create index "trust_center_preview_domain_id_idx" to table: "trust_centers"
DROP INDEX "trust_center_preview_domain_id_idx";
-- reverse: create index "trust_center_owner_id_idx" to table: "trust_centers"
DROP INDEX "trust_center_owner_id_idx";
-- reverse: create index "trust_center_custom_domain_id_idx" to table: "trust_centers"
DROP INDEX "trust_center_custom_domain_id_idx";
-- reverse: create "trust_centers" table
DROP TABLE "trust_centers";
-- reverse: create index "template_trust_center_id" to table: "templates"
DROP INDEX "template_trust_center_id";
-- reverse: create index "template_owner_id_idx" to table: "templates"
DROP INDEX "template_owner_id_idx";
-- reverse: create index "template_name_owner_id_template_type" to table: "templates"
DROP INDEX "template_name_owner_id_template_type";
-- reverse: create "templates" table
DROP TABLE "templates";
-- reverse: create index "task_parent_task_id_idx" to table: "tasks"
DROP INDEX "task_parent_task_id_idx";
-- reverse: create index "task_owner_id_is_suggested_priority" to table: "tasks"
DROP INDEX "task_owner_id_is_suggested_priority";
-- reverse: create index "task_owner_id_idx" to table: "tasks"
DROP INDEX "task_owner_id_idx";
-- reverse: create index "task_owner_id_idempotency_key" to table: "tasks"
DROP INDEX "task_owner_id_idempotency_key";
-- reverse: create index "task_external_uuid_owner_id" to table: "tasks"
DROP INDEX "task_external_uuid_owner_id";
-- reverse: create index "task_display_id_owner_id" to table: "tasks"
DROP INDEX "task_display_id_owner_id";
-- reverse: create index "task_assigner_id_idx" to table: "tasks"
DROP INDEX "task_assigner_id_idx";
-- reverse: create index "task_assignee_id_idx" to table: "tasks"
DROP INDEX "task_assignee_id_idx";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "tagdefinition_slug_owner_id" to table: "tag_definitions"
DROP INDEX "tagdefinition_slug_owner_id";
-- reverse: create index "tagdefinition_name_owner_id" to table: "tag_definitions"
DROP INDEX "tagdefinition_name_owner_id";
-- reverse: create index "tag_definition_owner_id_idx" to table: "tag_definitions"
DROP INDEX "tag_definition_owner_id_idx";
-- reverse: create "tag_definitions" table
DROP TABLE "tag_definitions";
-- reverse: create index "tfasetting_owner_id" to table: "tfa_settings"
DROP INDEX "tfasetting_owner_id";
-- reverse: create index "tfa_settings_owner_id_fk" to table: "tfa_settings"
DROP INDEX "tfa_settings_owner_id_fk";
-- reverse: create "tfa_settings" table
DROP TABLE "tfa_settings";
-- reverse: create index "systemdetail_display_id_owner_id" to table: "system_details"
DROP INDEX "systemdetail_display_id_owner_id";
-- reverse: create index "system_detail_owner_id_idx" to table: "system_details"
DROP INDEX "system_detail_owner_id_idx";
-- reverse: create "system_details" table
DROP TABLE "system_details";
-- reverse: create index "subscribers_token_key" to table: "subscribers"
DROP INDEX "subscribers_token_key";
-- reverse: create index "subscriber_user_id_idx" to table: "subscribers"
DROP INDEX "subscriber_user_id_idx";
-- reverse: create index "subscriber_trust_center_id_idx" to table: "subscribers"
DROP INDEX "subscriber_trust_center_id_idx";
-- reverse: create index "subscriber_owner_id_idx" to table: "subscribers"
DROP INDEX "subscriber_owner_id_idx";
-- reverse: create index "subscriber_email_trust_center_id" to table: "subscribers"
DROP INDEX "subscriber_email_trust_center_id";
-- reverse: create index "subscriber_email_owner_id" to table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- reverse: create index "subscriber_contact_id_idx" to table: "subscribers"
DROP INDEX "subscriber_contact_id_idx";
-- reverse: create "subscribers" table
DROP TABLE "subscribers";
-- reverse: create index "subprocessor_owner_id_idx" to table: "subprocessors"
DROP INDEX "subprocessor_owner_id_idx";
-- reverse: create index "subprocessor_name_owner_id" to table: "subprocessors"
DROP INDEX "subprocessor_name_owner_id";
-- reverse: create index "subprocessor_logo_file_id_idx" to table: "subprocessors"
DROP INDEX "subprocessor_logo_file_id_idx";
-- reverse: create "subprocessors" table
DROP TABLE "subprocessors";
-- reverse: create index "subcontrol_reference_id_deleted_at_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_reference_id_deleted_at_owner_id";
-- reverse: create index "subcontrol_owner_id_idx" to table: "subcontrols"
DROP INDEX "subcontrol_owner_id_idx";
-- reverse: create index "subcontrol_external_uuid_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_external_uuid_owner_id";
-- reverse: create index "subcontrol_display_id_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_display_id_owner_id";
-- reverse: create index "subcontrol_control_id_ref_code_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_control_id_ref_code_owner_id";
-- reverse: create index "subcontrol_control_id_ref_code" to table: "subcontrols"
DROP INDEX "subcontrol_control_id_ref_code";
-- reverse: create index "subcontrol_auditor_reference_id_deleted_at_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_auditor_reference_id_deleted_at_owner_id";
-- reverse: create "subcontrols" table
DROP TABLE "subcontrols";
-- reverse: create index "standard_owner_id_idx" to table: "standards"
DROP INDEX "standard_owner_id_idx";
-- reverse: create index "standard_logo_file_id_idx" to table: "standards"
DROP INDEX "standard_logo_file_id_idx";
-- reverse: create "standards" table
DROP TABLE "standards";
-- reverse: create index "scheduled_job_run_scheduled_job_id_idx" to table: "scheduled_job_runs"
DROP INDEX "scheduled_job_run_scheduled_job_id_idx";
-- reverse: create index "scheduled_job_run_owner_id_idx" to table: "scheduled_job_runs"
DROP INDEX "scheduled_job_run_owner_id_idx";
-- reverse: create index "scheduled_job_run_job_runner_id_idx" to table: "scheduled_job_runs"
DROP INDEX "scheduled_job_run_job_runner_id_idx";
-- reverse: create "scheduled_job_runs" table
DROP TABLE "scheduled_job_runs";
-- reverse: create index "scheduledjob_display_id_owner_id" to table: "scheduled_jobs"
DROP INDEX "scheduledjob_display_id_owner_id";
-- reverse: create index "scheduled_job_owner_id_idx" to table: "scheduled_jobs"
DROP INDEX "scheduled_job_owner_id_idx";
-- reverse: create index "scheduled_job_job_runner_id_idx" to table: "scheduled_jobs"
DROP INDEX "scheduled_job_job_runner_id_idx";
-- reverse: create index "scheduled_job_job_id_idx" to table: "scheduled_jobs"
DROP INDEX "scheduled_job_job_id_idx";
-- reverse: create "scheduled_jobs" table
DROP TABLE "scheduled_jobs";
-- reverse: create index "scan_performed_by_user_id_idx" to table: "scans"
DROP INDEX "scan_performed_by_user_id_idx";
-- reverse: create index "scan_performed_by_group_id_idx" to table: "scans"
DROP INDEX "scan_performed_by_group_id_idx";
-- reverse: create index "scan_owner_id_idx" to table: "scans"
DROP INDEX "scan_owner_id_idx";
-- reverse: create index "scan_generated_by_platform_id_idx" to table: "scans"
DROP INDEX "scan_generated_by_platform_id_idx";
-- reverse: create "scans" table
DROP TABLE "scans";
-- reverse: create index "sladefinition_security_level_owner_id" to table: "sla_definitions"
DROP INDEX "sladefinition_security_level_owner_id";
-- reverse: create index "sladefinition_display_id_owner_id" to table: "sla_definitions"
DROP INDEX "sladefinition_display_id_owner_id";
-- reverse: create index "sla_definition_owner_id_idx" to table: "sla_definitions"
DROP INDEX "sla_definition_owner_id_idx";
-- reverse: create "sla_definitions" table
DROP TABLE "sla_definitions";
-- reverse: create index "risk_stakeholder_id_idx" to table: "risks"
DROP INDEX "risk_stakeholder_id_idx";
-- reverse: create index "risk_owner_id_idx" to table: "risks"
DROP INDEX "risk_owner_id_idx";
-- reverse: create index "risk_external_uuid_owner_id" to table: "risks"
DROP INDEX "risk_external_uuid_owner_id";
-- reverse: create index "risk_display_id_owner_id" to table: "risks"
DROP INDEX "risk_display_id_owner_id";
-- reverse: create index "risk_delegate_id_idx" to table: "risks"
DROP INDEX "risk_delegate_id_idx";
-- reverse: create "risks" table
DROP TABLE "risks";
-- reverse: create index "review_reviewer_id_idx" to table: "reviews"
DROP INDEX "review_reviewer_id_idx";
-- reverse: create index "review_owner_id_idx" to table: "reviews"
DROP INDEX "review_owner_id_idx";
-- reverse: create index "review_external_id_external_owner_id_owner_id" to table: "reviews"
DROP INDEX "review_external_id_external_owner_id_owner_id";
-- reverse: create "reviews" table
DROP TABLE "reviews";
-- reverse: create index "remediation_owner_id_idx" to table: "remediations"
DROP INDEX "remediation_owner_id_idx";
-- reverse: create index "remediation_external_id_external_owner_id_owner_id" to table: "remediations"
DROP INDEX "remediation_external_id_external_owner_id_owner_id";
-- reverse: create index "remediation_display_id_owner_id" to table: "remediations"
DROP INDEX "remediation_display_id_owner_id";
-- reverse: create "remediations" table
DROP TABLE "remediations";
-- reverse: create index "programmembership_user_id_program_id" to table: "program_memberships"
DROP INDEX "programmembership_user_id_program_id";
-- reverse: create index "program_membership_program_id_idx" to table: "program_memberships"
DROP INDEX "program_membership_program_id_idx";
-- reverse: create "program_memberships" table
DROP TABLE "program_memberships";
-- reverse: create index "program_program_owner_id_idx" to table: "programs"
DROP INDEX "program_program_owner_id_idx";
-- reverse: create index "program_owner_id_idx" to table: "programs"
DROP INDEX "program_owner_id_idx";
-- reverse: create index "program_external_uuid_owner_id" to table: "programs"
DROP INDEX "program_external_uuid_owner_id";
-- reverse: create index "program_display_id_owner_id" to table: "programs"
DROP INDEX "program_display_id_owner_id";
-- reverse: create "programs" table
DROP TABLE "programs";
-- reverse: create index "procedure_owner_id_idx" to table: "procedures"
DROP INDEX "procedure_owner_id_idx";
-- reverse: create index "procedure_file_id_idx" to table: "procedures"
DROP INDEX "procedure_file_id_idx";
-- reverse: create index "procedure_display_id_owner_id" to table: "procedures"
DROP INDEX "procedure_display_id_owner_id";
-- reverse: create "procedures" table
DROP TABLE "procedures";
-- reverse: create index "platform_platform_owner_id_idx" to table: "platforms"
DROP INDEX "platform_platform_owner_id_idx";
-- reverse: create index "platform_owner_id_idx" to table: "platforms"
DROP INDEX "platform_owner_id_idx";
-- reverse: create index "platform_name_owner_id" to table: "platforms"
DROP INDEX "platform_name_owner_id";
-- reverse: create index "platform_external_uuid_owner_id" to table: "platforms"
DROP INDEX "platform_external_uuid_owner_id";
-- reverse: create index "platform_display_id_owner_id" to table: "platforms"
DROP INDEX "platform_display_id_owner_id";
-- reverse: create "platforms" table
DROP TABLE "platforms";
-- reverse: create index "personalaccesstoken_token" to table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_token";
-- reverse: create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_token_key";
-- reverse: create index "personal_access_tokens_owner_id_fk" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_owner_id_fk";
-- reverse: create "personal_access_tokens" table
DROP TABLE "personal_access_tokens";
-- reverse: create index "passwordresettoken_token" to table: "password_reset_tokens"
DROP INDEX "passwordresettoken_token";
-- reverse: create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_token_key";
-- reverse: create index "password_reset_tokens_owner_id_fk" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_owner_id_fk";
-- reverse: create "password_reset_tokens" table
DROP TABLE "password_reset_tokens";
-- reverse: create index "organization_settings_organization_id_key" to table: "organization_settings"
DROP INDEX "organization_settings_organization_id_key";
-- reverse: create index "organization_settings_compliance_webhook_token_key" to table: "organization_settings"
DROP INDEX "organization_settings_compliance_webhook_token_key";
-- reverse: create index "organization_setting_organization_id_idx" to table: "organization_settings"
DROP INDEX "organization_setting_organization_id_idx";
-- reverse: create "organization_settings" table
DROP TABLE "organization_settings";
-- reverse: create index "organizations_stripe_customer_id_key" to table: "organizations"
DROP INDEX "organizations_stripe_customer_id_key";
-- reverse: create index "organization_parent_organization_id_idx" to table: "organizations"
DROP INDEX "organization_parent_organization_id_idx";
-- reverse: create index "organization_name" to table: "organizations"
DROP INDEX "organization_name";
-- reverse: create index "organization_avatar_local_file_id_idx" to table: "organizations"
DROP INDEX "organization_avatar_local_file_id_idx";
-- reverse: create "organizations" table
DROP TABLE "organizations";
-- reverse: create index "org_subscription_owner_id_idx" to table: "org_subscriptions"
DROP INDEX "org_subscription_owner_id_idx";
-- reverse: create "org_subscriptions" table
DROP TABLE "org_subscriptions";
-- reverse: create index "org_product_subscription_id_idx" to table: "org_products"
DROP INDEX "org_product_subscription_id_idx";
-- reverse: create index "org_product_owner_id_idx" to table: "org_products"
DROP INDEX "org_product_owner_id_idx";
-- reverse: create "org_products" table
DROP TABLE "org_products";
-- reverse: create index "org_price_subscription_id_idx" to table: "org_prices"
DROP INDEX "org_price_subscription_id_idx";
-- reverse: create index "org_price_owner_id_idx" to table: "org_prices"
DROP INDEX "org_price_owner_id_idx";
-- reverse: create "org_prices" table
DROP TABLE "org_prices";
-- reverse: create index "org_module_subscription_id_idx" to table: "org_modules"
DROP INDEX "org_module_subscription_id_idx";
-- reverse: create index "org_module_owner_id_idx" to table: "org_modules"
DROP INDEX "org_module_owner_id_idx";
-- reverse: create "org_modules" table
DROP TABLE "org_modules";
-- reverse: create index "orgmembership_user_id_organization_id" to table: "org_memberships"
DROP INDEX "orgmembership_user_id_organization_id";
-- reverse: create index "org_membership_organization_id_idx" to table: "org_memberships"
DROP INDEX "org_membership_organization_id_idx";
-- reverse: create "org_memberships" table
DROP TABLE "org_memberships";
-- reverse: create "onboardings" table
DROP TABLE "onboardings";
-- reverse: create index "notificationtemplate_owner_id_key" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_key";
-- reverse: create index "notificationtemplate_owner_id_channel_locale_topic_pattern" to table: "notification_templates"
DROP INDEX "notificationtemplate_owner_id_channel_locale_topic_pattern";
-- reverse: create index "notification_template_workflow_definition_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_workflow_definition_id_idx";
-- reverse: create index "notification_template_owner_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_owner_id_idx";
-- reverse: create index "notification_template_integration_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_integration_id_idx";
-- reverse: create index "notification_template_email_template_id_idx" to table: "notification_templates"
DROP INDEX "notification_template_email_template_id_idx";
-- reverse: create "notification_templates" table
DROP TABLE "notification_templates";
-- reverse: create index "notificationpreference_owner_id_user_id_channel" to table: "notification_preferences"
DROP INDEX "notificationpreference_owner_id_user_id_channel";
-- reverse: create index "notification_preference_user_id_idx" to table: "notification_preferences"
DROP INDEX "notification_preference_user_id_idx";
-- reverse: create index "notification_preference_template_id_idx" to table: "notification_preferences"
DROP INDEX "notification_preference_template_id_idx";
-- reverse: create index "notification_preference_owner_id_idx" to table: "notification_preferences"
DROP INDEX "notification_preference_owner_id_idx";
-- reverse: create "notification_preferences" table
DROP TABLE "notification_preferences";
-- reverse: create index "notification_user_id_read_at_owner_id" to table: "notifications"
DROP INDEX "notification_user_id_read_at_owner_id";
-- reverse: create index "notification_template_id_idx" to table: "notifications"
DROP INDEX "notification_template_id_idx";
-- reverse: create index "notification_owner_id_idx" to table: "notifications"
DROP INDEX "notification_owner_id_idx";
-- reverse: create "notifications" table
DROP TABLE "notifications";
-- reverse: create index "note_trust_center_id_idx" to table: "notes"
DROP INDEX "note_trust_center_id_idx";
-- reverse: create index "note_owner_id_idx" to table: "notes"
DROP INDEX "note_owner_id_idx";
-- reverse: create index "note_display_id_owner_id" to table: "notes"
DROP INDEX "note_display_id_owner_id";
-- reverse: create index "note_discussion_id_idx" to table: "notes"
DROP INDEX "note_discussion_id_idx";
-- reverse: create "notes" table
DROP TABLE "notes";
-- reverse: create index "narrative_owner_id_idx" to table: "narratives"
DROP INDEX "narrative_owner_id_idx";
-- reverse: create index "narrative_display_id_owner_id" to table: "narratives"
DROP INDEX "narrative_display_id_owner_id";
-- reverse: create "narratives" table
DROP TABLE "narratives";
-- reverse: create index "mapped_control_owner_id_idx" to table: "mapped_controls"
DROP INDEX "mapped_control_owner_id_idx";
-- reverse: create "mapped_controls" table
DROP TABLE "mapped_controls";
-- reverse: create index "mappabledomain_name" to table: "mappable_domains"
DROP INDEX "mappabledomain_name";
-- reverse: create "mappable_domains" table
DROP TABLE "mappable_domains";
-- reverse: create index "jobtemplate_display_id_owner_id" to table: "job_templates"
DROP INDEX "jobtemplate_display_id_owner_id";
-- reverse: create index "job_template_owner_id_idx" to table: "job_templates"
DROP INDEX "job_template_owner_id_idx";
-- reverse: create "job_templates" table
DROP TABLE "job_templates";
-- reverse: create index "jobrunnertoken_token_expires_at_is_active" to table: "job_runner_tokens"
DROP INDEX "jobrunnertoken_token_expires_at_is_active";
-- reverse: create index "job_runner_tokens_token_key" to table: "job_runner_tokens"
DROP INDEX "job_runner_tokens_token_key";
-- reverse: create index "job_runner_token_owner_id_idx" to table: "job_runner_tokens"
DROP INDEX "job_runner_token_owner_id_idx";
-- reverse: create "job_runner_tokens" table
DROP TABLE "job_runner_tokens";
-- reverse: create index "job_runner_registration_tokens_token_key" to table: "job_runner_registration_tokens"
DROP INDEX "job_runner_registration_tokens_token_key";
-- reverse: create index "job_runner_registration_token_owner_id_idx" to table: "job_runner_registration_tokens"
DROP INDEX "job_runner_registration_token_owner_id_idx";
-- reverse: create index "job_runner_registration_token_job_runner_id_idx" to table: "job_runner_registration_tokens"
DROP INDEX "job_runner_registration_token_job_runner_id_idx";
-- reverse: create "job_runner_registration_tokens" table
DROP TABLE "job_runner_registration_tokens";
-- reverse: create index "jobrunner_display_id_owner_id" to table: "job_runners"
DROP INDEX "jobrunner_display_id_owner_id";
-- reverse: create index "job_runner_owner_id_idx" to table: "job_runners"
DROP INDEX "job_runner_owner_id_idx";
-- reverse: create "job_runners" table
DROP TABLE "job_runners";
-- reverse: create index "job_result_scheduled_job_id_idx" to table: "job_results"
DROP INDEX "job_result_scheduled_job_id_idx";
-- reverse: create index "job_result_owner_id_idx" to table: "job_results"
DROP INDEX "job_result_owner_id_idx";
-- reverse: create index "job_result_file_id_idx" to table: "job_results"
DROP INDEX "job_result_file_id_idx";
-- reverse: create "job_results" table
DROP TABLE "job_results";
-- reverse: create index "invites_token_key" to table: "invites"
DROP INDEX "invites_token_key";
-- reverse: create index "invite_recipient_owner_id" to table: "invites"
DROP INDEX "invite_recipient_owner_id";
-- reverse: create index "invite_owner_id_idx" to table: "invites"
DROP INDEX "invite_owner_id_idx";
-- reverse: create "invites" table
DROP TABLE "invites";
-- reverse: create index "internalpolicy_external_uuid_owner_id" to table: "internal_policies"
DROP INDEX "internalpolicy_external_uuid_owner_id";
-- reverse: create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
DROP INDEX "internalpolicy_display_id_owner_id";
-- reverse: create index "internal_policy_owner_id_idx" to table: "internal_policies"
DROP INDEX "internal_policy_owner_id_idx";
-- reverse: create index "internal_policy_file_id_idx" to table: "internal_policies"
DROP INDEX "internal_policy_file_id_idx";
-- reverse: create "internal_policies" table
DROP TABLE "internal_policies";
-- reverse: create index "integrationwebhook_integration_id_name_external_event_id" to table: "integration_webhooks"
DROP INDEX "integrationwebhook_integration_id_name_external_event_id";
-- reverse: create index "integrationwebhook_endpoint_id" to table: "integration_webhooks"
DROP INDEX "integrationwebhook_endpoint_id";
-- reverse: create index "integration_webhook_owner_id_idx" to table: "integration_webhooks"
DROP INDEX "integration_webhook_owner_id_idx";
-- reverse: create "integration_webhooks" table
DROP TABLE "integration_webhooks";
-- reverse: create index "integrationrun_integration_id_started_at" to table: "integration_runs"
DROP INDEX "integrationrun_integration_id_started_at";
-- reverse: create index "integrationrun_assessment_response_id_started_at" to table: "integration_runs"
DROP INDEX "integrationrun_assessment_response_id_started_at";
-- reverse: create index "integrationrun_assessment_response_id_operation_name" to table: "integration_runs"
DROP INDEX "integrationrun_assessment_response_id_operation_name";
-- reverse: create index "integration_run_response_file_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_response_file_id_idx";
-- reverse: create index "integration_run_request_file_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_request_file_id_idx";
-- reverse: create index "integration_run_owner_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_owner_id_idx";
-- reverse: create index "integration_run_event_id_idx" to table: "integration_runs"
DROP INDEX "integration_run_event_id_idx";
-- reverse: create "integration_runs" table
DROP TABLE "integration_runs";
-- reverse: create index "integration_platform_id_idx" to table: "integrations"
DROP INDEX "integration_platform_id_idx";
-- reverse: create index "integration_owner_id_idx" to table: "integrations"
DROP INDEX "integration_owner_id_idx";
-- reverse: create "integrations" table
DROP TABLE "integrations";
-- reverse: create index "impersonation_event_user_id_idx" to table: "impersonation_events"
DROP INDEX "impersonation_event_user_id_idx";
-- reverse: create index "impersonation_event_target_user_id_idx" to table: "impersonation_events"
DROP INDEX "impersonation_event_target_user_id_idx";
-- reverse: create index "impersonation_event_organization_id_idx" to table: "impersonation_events"
DROP INDEX "impersonation_event_organization_id_idx";
-- reverse: create "impersonation_events" table
DROP TABLE "impersonation_events";
-- reverse: create index "identityholder_user_id" to table: "identity_holders"
DROP INDEX "identityholder_user_id";
-- reverse: create index "identityholder_external_user_id" to table: "identity_holders"
DROP INDEX "identityholder_external_user_id";
-- reverse: create index "identityholder_email_owner_id" to table: "identity_holders"
DROP INDEX "identityholder_email_owner_id";
-- reverse: create index "identityholder_display_id_owner_id" to table: "identity_holders"
DROP INDEX "identityholder_display_id_owner_id";
-- reverse: create index "identity_holder_owner_id_idx" to table: "identity_holders"
DROP INDEX "identity_holder_owner_id_idx";
-- reverse: create index "identity_holder_employer_entity_id_idx" to table: "identity_holders"
DROP INDEX "identity_holder_employer_entity_id_idx";
-- reverse: create "identity_holders" table
DROP TABLE "identity_holders";
-- reverse: create index "secret_owner_id_idx" to table: "hushes"
DROP INDEX "secret_owner_id_idx";
-- reverse: create "hushes" table
DROP TABLE "hushes";
-- reverse: create index "group_settings_group_id_key" to table: "group_settings"
DROP INDEX "group_settings_group_id_key";
-- reverse: create index "group_setting_group_id_idx" to table: "group_settings"
DROP INDEX "group_setting_group_id_idx";
-- reverse: create "group_settings" table
DROP TABLE "group_settings";
-- reverse: create index "groupmembership_user_id_group_id" to table: "group_memberships"
DROP INDEX "groupmembership_user_id_group_id";
-- reverse: create index "group_membership_group_id_idx" to table: "group_memberships"
DROP INDEX "group_membership_group_id_idx";
-- reverse: create "group_memberships" table
DROP TABLE "group_memberships";
-- reverse: create index "group_owner_id_idx" to table: "groups"
DROP INDEX "group_owner_id_idx";
-- reverse: create index "group_name_owner_id" to table: "groups"
DROP INDEX "group_name_owner_id";
-- reverse: create index "group_display_id_owner_id" to table: "groups"
DROP INDEX "group_display_id_owner_id";
-- reverse: create index "group_avatar_local_file_id_idx" to table: "groups"
DROP INDEX "group_avatar_local_file_id_idx";
-- reverse: create "groups" table
DROP TABLE "groups";
-- reverse: create index "findingcontrol_finding_id_control_id" to table: "finding_controls"
DROP INDEX "findingcontrol_finding_id_control_id";
-- reverse: create index "finding_control_standard_id_idx" to table: "finding_controls"
DROP INDEX "finding_control_standard_id_idx";
-- reverse: create index "finding_control_owner_id_idx" to table: "finding_controls"
DROP INDEX "finding_control_owner_id_idx";
-- reverse: create index "finding_control_control_id_idx" to table: "finding_controls"
DROP INDEX "finding_control_control_id_idx";
-- reverse: create "finding_controls" table
DROP TABLE "finding_controls";
-- reverse: create index "finding_owner_id_idx" to table: "findings"
DROP INDEX "finding_owner_id_idx";
-- reverse: create index "finding_external_id_external_owner_id_owner_id" to table: "findings"
DROP INDEX "finding_external_id_external_owner_id_owner_id";
-- reverse: create index "finding_display_id_owner_id" to table: "findings"
DROP INDEX "finding_display_id_owner_id";
-- reverse: create "findings" table
DROP TABLE "findings";
-- reverse: create index "filedownloadtoken_token" to table: "file_download_tokens"
DROP INDEX "filedownloadtoken_token";
-- reverse: create index "file_download_tokens_token_key" to table: "file_download_tokens"
DROP INDEX "file_download_tokens_token_key";
-- reverse: create index "file_download_tokens_owner_id_fk" to table: "file_download_tokens"
DROP INDEX "file_download_tokens_owner_id_fk";
-- reverse: create "file_download_tokens" table
DROP TABLE "file_download_tokens";
-- reverse: create "files" table
DROP TABLE "files";
-- reverse: create index "export_owner_id_idx" to table: "exports"
DROP INDEX "export_owner_id_idx";
-- reverse: create "exports" table
DROP TABLE "exports";
-- reverse: create index "evidence_owner_id_idx" to table: "evidences"
DROP INDEX "evidence_owner_id_idx";
-- reverse: create index "evidence_external_uuid_owner_id" to table: "evidences"
DROP INDEX "evidence_external_uuid_owner_id";
-- reverse: create index "evidence_display_id_owner_id" to table: "evidences"
DROP INDEX "evidence_display_id_owner_id";
-- reverse: create "evidences" table
DROP TABLE "evidences";
-- reverse: create "events" table
DROP TABLE "events";
-- reverse: create index "entitytype_name_owner_id" to table: "entity_types"
DROP INDEX "entitytype_name_owner_id";
-- reverse: create index "entity_type_owner_id_idx" to table: "entity_types"
DROP INDEX "entity_type_owner_id_idx";
-- reverse: create "entity_types" table
DROP TABLE "entity_types";
-- reverse: create index "entity_reviewed_by_user_id" to table: "entities"
DROP INDEX "entity_reviewed_by_user_id";
-- reverse: create index "entity_owner_id_idx" to table: "entities"
DROP INDEX "entity_owner_id_idx";
-- reverse: create index "entity_name_owner_id" to table: "entities"
DROP INDEX "entity_name_owner_id";
-- reverse: create index "entity_logo_file_id_idx" to table: "entities"
DROP INDEX "entity_logo_file_id_idx";
-- reverse: create index "entity_entity_type_id_idx" to table: "entities"
DROP INDEX "entity_entity_type_id_idx";
-- reverse: create "entities" table
DROP TABLE "entities";
-- reverse: create index "emailverificationtoken_token" to table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_token";
-- reverse: create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_token_key";
-- reverse: create index "email_verification_tokens_owner_id_fk" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_owner_id_fk";
-- reverse: create "email_verification_tokens" table
DROP TABLE "email_verification_tokens";
-- reverse: create index "emailtemplate_owner_id_key" to table: "email_templates"
DROP INDEX "emailtemplate_owner_id_key";
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
-- reverse: create "email_templates" table
DROP TABLE "email_templates";
-- reverse: create index "document_template_id_idx" to table: "document_data"
DROP INDEX "document_template_id_idx";
-- reverse: create index "document_owner_id_idx" to table: "document_data"
DROP INDEX "document_owner_id_idx";
-- reverse: create "document_data" table
DROP TABLE "document_data";
-- reverse: create index "discussions_external_id_key" to table: "discussions"
DROP INDEX "discussions_external_id_key";
-- reverse: create index "discussion_owner_id_idx" to table: "discussions"
DROP INDEX "discussion_owner_id_idx";
-- reverse: create "discussions" table
DROP TABLE "discussions";
-- reverse: create index "directorysyncrun_platform_id_started_at" to table: "directory_sync_runs"
DROP INDEX "directorysyncrun_platform_id_started_at";
-- reverse: create index "directorysyncrun_integration_id_started_at" to table: "directory_sync_runs"
DROP INDEX "directorysyncrun_integration_id_started_at";
-- reverse: create index "directorysyncrun_display_id_owner_id" to table: "directory_sync_runs"
DROP INDEX "directorysyncrun_display_id_owner_id";
-- reverse: create index "directorysyncrun_directory_instance_id_started_at" to table: "directory_sync_runs"
DROP INDEX "directorysyncrun_directory_instance_id_started_at";
-- reverse: create index "directory_sync_run_owner_id_idx" to table: "directory_sync_runs"
DROP INDEX "directory_sync_run_owner_id_idx";
-- reverse: create "directory_sync_runs" table
DROP TABLE "directory_sync_runs";
-- reverse: create index "directorymembership_platform_id_directory_sync_run_id" to table: "directory_memberships"
DROP INDEX "directorymembership_platform_id_directory_sync_run_id";
-- reverse: create index "directorymembership_integration_id_directory_sync_run_id" to table: "directory_memberships"
DROP INDEX "directorymembership_integration_id_directory_sync_run_id";
-- reverse: create index "directorymembership_display_id_owner_id" to table: "directory_memberships"
DROP INDEX "directorymembership_display_id_owner_id";
-- reverse: create index "directorymembership_directory_sync_run_id" to table: "directory_memberships"
DROP INDEX "directorymembership_directory_sync_run_id";
-- reverse: create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
DROP INDEX "directorymembership_directory_account_id_directory_group_id";
-- reverse: create index "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" to table: "directory_memberships"
DROP INDEX "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482";
-- reverse: create index "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125" to table: "directory_memberships"
DROP INDEX "directorymembership_directory__5b409a930567cfcdf3be9fd87b4e5125";
-- reverse: create index "directory_membership_owner_id_idx" to table: "directory_memberships"
DROP INDEX "directory_membership_owner_id_idx";
-- reverse: create index "directory_membership_directory_group_id_idx" to table: "directory_memberships"
DROP INDEX "directory_membership_directory_group_id_idx";
-- reverse: create "directory_memberships" table
DROP TABLE "directory_memberships";
-- reverse: create index "directorygroup_platform_id_external_id" to table: "directory_groups"
DROP INDEX "directorygroup_platform_id_external_id";
-- reverse: create index "directorygroup_platform_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_platform_id_email";
-- reverse: create index "directorygroup_owner_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_owner_id_email";
-- reverse: create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
DROP INDEX "directorygroup_integration_id_external_id_directory_sync_run_id";
-- reverse: create index "directorygroup_integration_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_integration_id_email";
-- reverse: create index "directorygroup_display_id_owner_id" to table: "directory_groups"
DROP INDEX "directorygroup_display_id_owner_id";
-- reverse: create index "directorygroup_directory_sync_run_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_directory_sync_run_id_email";
-- reverse: create index "directorygroup_directory_instance_id_external_id" to table: "directory_groups"
DROP INDEX "directorygroup_directory_instance_id_external_id";
-- reverse: create index "directorygroup_directory_instance_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_directory_instance_id_email";
-- reverse: create index "directory_group_owner_id_idx" to table: "directory_groups"
DROP INDEX "directory_group_owner_id_idx";
-- reverse: create "directory_groups" table
DROP TABLE "directory_groups";
-- reverse: create index "directoryaccount_platform_id_external_id" to table: "directory_accounts"
DROP INDEX "directoryaccount_platform_id_external_id";
-- reverse: create index "directoryaccount_platform_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_platform_id_canonical_email";
-- reverse: create index "directoryaccount_owner_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_owner_id_canonical_email";
-- reverse: create index "directoryaccount_integration_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_integration_id_canonical_email";
-- reverse: create index "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" to table: "directory_accounts"
DROP INDEX "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d";
-- reverse: create index "directoryaccount_identity_holder_id_directory_name" to table: "directory_accounts"
DROP INDEX "directoryaccount_identity_holder_id_directory_name";
-- reverse: create index "directoryaccount_identity_holder_id" to table: "directory_accounts"
DROP INDEX "directoryaccount_identity_holder_id";
-- reverse: create index "directoryaccount_display_id_owner_id" to table: "directory_accounts"
DROP INDEX "directoryaccount_display_id_owner_id";
-- reverse: create index "directoryaccount_directory_sync_run_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_directory_sync_run_id_canonical_email";
-- reverse: create index "directoryaccount_directory_instance_id_external_id" to table: "directory_accounts"
DROP INDEX "directoryaccount_directory_instance_id_external_id";
-- reverse: create index "directoryaccount_directory_instance_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_directory_instance_id_canonical_email";
-- reverse: create index "directory_account_owner_id_idx" to table: "directory_accounts"
DROP INDEX "directory_account_owner_id_idx";
-- reverse: create index "directory_account_avatar_local_file_id_idx" to table: "directory_accounts"
DROP INDEX "directory_account_avatar_local_file_id_idx";
-- reverse: create "directory_accounts" table
DROP TABLE "directory_accounts";
-- reverse: create index "dnsverification_cloudflare_hostname_id" to table: "dns_verifications"
DROP INDEX "dnsverification_cloudflare_hostname_id";
-- reverse: create index "dns_verification_owner_id_idx" to table: "dns_verifications"
DROP INDEX "dns_verification_owner_id_idx";
-- reverse: create "dns_verifications" table
DROP TABLE "dns_verifications";
-- reverse: create index "customtypeenum_object_type" to table: "custom_type_enums"
DROP INDEX "customtypeenum_object_type";
-- reverse: create index "customtypeenum_name_object_type_field_owner_id" to table: "custom_type_enums"
DROP INDEX "customtypeenum_name_object_type_field_owner_id";
-- reverse: create index "customtypeenum_name_field" to table: "custom_type_enums"
DROP INDEX "customtypeenum_name_field";
-- reverse: create index "custom_type_enum_owner_id_idx" to table: "custom_type_enums"
DROP INDEX "custom_type_enum_owner_id_idx";
-- reverse: create "custom_type_enums" table
DROP TABLE "custom_type_enums";
-- reverse: create index "customdomain_cname_record" to table: "custom_domains"
DROP INDEX "customdomain_cname_record";
-- reverse: create index "custom_domain_owner_id_idx" to table: "custom_domains"
DROP INDEX "custom_domain_owner_id_idx";
-- reverse: create index "custom_domain_mappable_domain_id_idx" to table: "custom_domains"
DROP INDEX "custom_domain_mappable_domain_id_idx";
-- reverse: create index "custom_domain_dns_verification_id_idx" to table: "custom_domains"
DROP INDEX "custom_domain_dns_verification_id_idx";
-- reverse: create "custom_domains" table
DROP TABLE "custom_domains";
-- reverse: create index "controlobjective_display_id_owner_id" to table: "control_objectives"
DROP INDEX "controlobjective_display_id_owner_id";
-- reverse: create index "control_objective_owner_id_idx" to table: "control_objectives"
DROP INDEX "control_objective_owner_id_idx";
-- reverse: create "control_objectives" table
DROP TABLE "control_objectives";
-- reverse: create index "control_implementation_owner_id_idx" to table: "control_implementations"
DROP INDEX "control_implementation_owner_id_idx";
-- reverse: create "control_implementations" table
DROP TABLE "control_implementations";
-- reverse: create index "control_standard_id_ref_code_owner_id" to table: "controls"
DROP INDEX "control_standard_id_ref_code_owner_id";
-- reverse: create index "control_standard_id_ref_code" to table: "controls"
DROP INDEX "control_standard_id_ref_code";
-- reverse: create index "control_reference_id" to table: "controls"
DROP INDEX "control_reference_id";
-- reverse: create index "control_ref_code_owner_id" to table: "controls"
DROP INDEX "control_ref_code_owner_id";
-- reverse: create index "control_ref_code" to table: "controls"
DROP INDEX "control_ref_code";
-- reverse: create index "control_owner_id_idx" to table: "controls"
DROP INDEX "control_owner_id_idx";
-- reverse: create index "control_external_uuid_owner_id" to table: "controls"
DROP INDEX "control_external_uuid_owner_id";
-- reverse: create index "control_display_id_owner_id" to table: "controls"
DROP INDEX "control_display_id_owner_id";
-- reverse: create index "control_auditor_reference_id" to table: "controls"
DROP INDEX "control_auditor_reference_id";
-- reverse: create "controls" table
DROP TABLE "controls";
-- reverse: create index "contact_owner_id_idx" to table: "contacts"
DROP INDEX "contact_owner_id_idx";
-- reverse: create "contacts" table
DROP TABLE "contacts";
-- reverse: create index "check_result_integration_id_idx" to table: "check_results"
DROP INDEX "check_result_integration_id_idx";
-- reverse: create "check_results" table
DROP TABLE "check_results";
-- reverse: create index "campaigntarget_user_id" to table: "campaign_targets"
DROP INDEX "campaigntarget_user_id";
-- reverse: create index "campaigntarget_subscriber_id" to table: "campaign_targets"
DROP INDEX "campaigntarget_subscriber_id";
-- reverse: create index "campaigntarget_status" to table: "campaign_targets"
DROP INDEX "campaigntarget_status";
-- reverse: create index "campaigntarget_group_id" to table: "campaign_targets"
DROP INDEX "campaigntarget_group_id";
-- reverse: create index "campaigntarget_contact_id" to table: "campaign_targets"
DROP INDEX "campaigntarget_contact_id";
-- reverse: create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
DROP INDEX "campaigntarget_campaign_id_email";
-- reverse: create index "campaign_target_owner_id_idx" to table: "campaign_targets"
DROP INDEX "campaign_target_owner_id_idx";
-- reverse: create "campaign_targets" table
DROP TABLE "campaign_targets";
-- reverse: create index "campaign_trust_center_id_idx" to table: "campaigns"
DROP INDEX "campaign_trust_center_id_idx";
-- reverse: create index "campaign_template_id_idx" to table: "campaigns"
DROP INDEX "campaign_template_id_idx";
-- reverse: create index "campaign_owner_id_idx" to table: "campaigns"
DROP INDEX "campaign_owner_id_idx";
-- reverse: create index "campaign_name_owner_id" to table: "campaigns"
DROP INDEX "campaign_name_owner_id";
-- reverse: create index "campaign_integration_id_idx" to table: "campaigns"
DROP INDEX "campaign_integration_id_idx";
-- reverse: create index "campaign_entity_id" to table: "campaigns"
DROP INDEX "campaign_entity_id";
-- reverse: create index "campaign_email_template_id_idx" to table: "campaigns"
DROP INDEX "campaign_email_template_id_idx";
-- reverse: create index "campaign_display_id_owner_id" to table: "campaigns"
DROP INDEX "campaign_display_id_owner_id";
-- reverse: create index "campaign_assessment_id_idx" to table: "campaigns"
DROP INDEX "campaign_assessment_id_idx";
-- reverse: create "campaigns" table
DROP TABLE "campaigns";
-- reverse: create index "asset_source_platform_id_idx" to table: "assets"
DROP INDEX "asset_source_platform_id_idx";
-- reverse: create index "asset_owner_id_idx" to table: "assets"
DROP INDEX "asset_owner_id_idx";
-- reverse: create index "asset_integration_id_idx" to table: "assets"
DROP INDEX "asset_integration_id_idx";
-- reverse: create "assets" table
DROP TABLE "assets";
-- reverse: create index "assessmentresponse_status" to table: "assessment_responses"
DROP INDEX "assessmentresponse_status";
-- reverse: create index "assessmentresponse_identity_holder_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_identity_holder_id";
-- reverse: create index "assessmentresponse_entity_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_entity_id";
-- reverse: create index "assessmentresponse_due_date" to table: "assessment_responses"
DROP INDEX "assessmentresponse_due_date";
-- reverse: create index "assessmentresponse_completed_at" to table: "assessment_responses"
DROP INDEX "assessmentresponse_completed_at";
-- reverse: create index "assessmentresponse_campaign_id_assessment_id_email_is_test" to table: "assessment_responses"
DROP INDEX "assessmentresponse_campaign_id_assessment_id_email_is_test";
-- reverse: create index "assessmentresponse_campaign_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_campaign_id";
-- reverse: create index "assessmentresponse_assigned_at" to table: "assessment_responses"
DROP INDEX "assessmentresponse_assigned_at";
-- reverse: create index "assessmentresponse_assessment_id_email_is_test" to table: "assessment_responses"
DROP INDEX "assessmentresponse_assessment_id_email_is_test";
-- reverse: create index "assessment_response_owner_id_idx" to table: "assessment_responses"
DROP INDEX "assessment_response_owner_id_idx";
-- reverse: create index "assessment_response_document_data_id_idx" to table: "assessment_responses"
DROP INDEX "assessment_response_document_data_id_idx";
-- reverse: create "assessment_responses" table
DROP TABLE "assessment_responses";
-- reverse: create index "assessment_template_id_idx" to table: "assessments"
DROP INDEX "assessment_template_id_idx";
-- reverse: create index "assessment_owner_id_idx" to table: "assessments"
DROP INDEX "assessment_owner_id_idx";
-- reverse: create index "assessment_name_owner_id" to table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- reverse: create "assessments" table
DROP TABLE "assessments";
-- reverse: create index "action_plan_owner_id_idx" to table: "action_plans"
DROP INDEX "action_plan_owner_id_idx";
-- reverse: create index "action_plan_file_id_idx" to table: "action_plans"
DROP INDEX "action_plan_file_id_idx";
-- reverse: create "action_plans" table
DROP TABLE "action_plans";
-- reverse: create index "apitoken_token" to table: "api_tokens"
DROP INDEX "apitoken_token";
-- reverse: create index "api_tokens_token_key" to table: "api_tokens"
DROP INDEX "api_tokens_token_key";
-- reverse: create index "api_token_owner_id_idx" to table: "api_tokens"
DROP INDEX "api_token_owner_id_idx";
-- reverse: create "api_tokens" table
DROP TABLE "api_tokens";
