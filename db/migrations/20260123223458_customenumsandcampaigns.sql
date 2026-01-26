-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "email_delivered_at" timestamptz NULL, ADD COLUMN "email_opened_at" timestamptz NULL, ADD COLUMN "email_clicked_at" timestamptz NULL, ADD COLUMN "email_open_count" bigint NULL DEFAULT 0, ADD COLUMN "email_click_count" bigint NULL DEFAULT 0, ADD COLUMN "last_email_event_at" timestamptz NULL, ADD COLUMN "email_metadata" jsonb NULL, ADD COLUMN "campaign_id" character varying NULL, ADD COLUMN "entity_id" character varying NULL, ADD COLUMN "identity_holder_id" character varying NULL;
-- Create index "assessmentresponse_campaign_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_campaign_id" ON "assessment_responses" ("campaign_id");
-- Create index "assessmentresponse_entity_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_entity_id" ON "assessment_responses" ("entity_id");
-- Create index "assessmentresponse_identity_holder_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_identity_holder_id" ON "assessment_responses" ("identity_holder_id");
-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "asset_subtype_name" character varying NULL, ADD COLUMN "asset_data_classification_name" character varying NULL, ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "access_model_name" character varying NULL, ADD COLUMN "encryption_status_name" character varying NULL, ADD COLUMN "security_tier_name" character varying NULL, ADD COLUMN "criticality_name" character varying NULL, ADD COLUMN "physical_location" character varying NULL, ADD COLUMN "region" character varying NULL, ADD COLUMN "contains_pii" boolean NULL DEFAULT false, ADD COLUMN "source_type" character varying NOT NULL DEFAULT 'MANUAL', ADD COLUMN "source_identifier" character varying NULL, ADD COLUMN "cost_center" character varying NULL, ADD COLUMN "estimated_monthly_cost" double precision NULL, ADD COLUMN "purchase_date" timestamptz NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "asset_subtype_id" character varying NULL, ADD COLUMN "asset_data_classification_id" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL, ADD COLUMN "access_model_id" character varying NULL, ADD COLUMN "encryption_status_id" character varying NULL, ADD COLUMN "security_tier_id" character varying NULL, ADD COLUMN "criticality_id" character varying NULL, ADD COLUMN "source_platform_id" character varying NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "entity_auth_methods" character varying NULL;
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "entities" table
ALTER TABLE "entities" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "reviewed_by" character varying NULL, ADD COLUMN "last_reviewed_at" timestamptz NULL, ADD COLUMN "entity_relationship_state_name" character varying NULL, ADD COLUMN "entity_security_questionnaire_status_name" character varying NULL, ADD COLUMN "entity_source_type_name" character varying NULL, ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "approved_for_use" boolean NULL DEFAULT false, ADD COLUMN "linked_asset_ids" jsonb NULL, ADD COLUMN "has_soc2" boolean NULL DEFAULT false, ADD COLUMN "soc2_period_end" timestamptz NULL, ADD COLUMN "contract_start_date" timestamptz NULL, ADD COLUMN "contract_end_date" timestamptz NULL, ADD COLUMN "auto_renews" boolean NULL DEFAULT false, ADD COLUMN "termination_notice_days" bigint NULL, ADD COLUMN "annual_spend" double precision NULL, ADD COLUMN "spend_currency" character varying NULL DEFAULT 'USD', ADD COLUMN "billing_model" character varying NULL, ADD COLUMN "renewal_risk" character varying NULL, ADD COLUMN "sso_enforced" boolean NULL DEFAULT false, ADD COLUMN "mfa_supported" boolean NULL DEFAULT false, ADD COLUMN "mfa_enforced" boolean NULL DEFAULT false, ADD COLUMN "status_page_url" character varying NULL, ADD COLUMN "provided_services" jsonb NULL, ADD COLUMN "links" jsonb NULL, ADD COLUMN "risk_rating" character varying NULL, ADD COLUMN "risk_score" bigint NULL, ADD COLUMN "tier" character varying NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "next_review_at" timestamptz NULL, ADD COLUMN "contract_renewal_at" timestamptz NULL, ADD COLUMN "vendor_metadata" jsonb NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "reviewed_by_user_id" character varying NULL, ADD COLUMN "reviewed_by_group_id" character varying NULL, ADD COLUMN "entity_relationship_state_id" character varying NULL, ADD COLUMN "entity_security_questionnaire_status_id" character varying NULL, ADD COLUMN "entity_source_type_id" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Create index "entity_reviewed_by_user_id" to table: "entities"
CREATE INDEX "entity_reviewed_by_user_id" ON "entities" ("reviewed_by_user_id");
-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "findings" table
ALTER TABLE "findings" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "identity_holder_blocked_groups" character varying NULL, ADD COLUMN "identity_holder_editors" character varying NULL, ADD COLUMN "identity_holder_viewers" character varying NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "scans" table
ALTER TABLE "scans" DROP COLUMN "control_scans", DROP COLUMN "vulnerability_scans", ADD COLUMN "reviewed_by" character varying NULL, ADD COLUMN "assigned_to" character varying NULL, ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "scan_date" timestamptz NULL, ADD COLUMN "scan_schedule" character varying NULL, ADD COLUMN "next_scan_run_at" timestamptz NULL, ADD COLUMN "performed_by" character varying NULL, ADD COLUMN "vulnerability_ids" jsonb NULL, ADD COLUMN "generated_by_platform_id" character varying NULL, ADD COLUMN "reviewed_by_user_id" character varying NULL, ADD COLUMN "reviewed_by_group_id" character varying NULL, ADD COLUMN "assigned_to_user_id" character varying NULL, ADD COLUMN "assigned_to_group_id" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL, ADD COLUMN "performed_by_user_id" character varying NULL, ADD COLUMN "performed_by_group_id" character varying NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "environment_name" character varying NULL, ADD COLUMN "scope_name" character varying NULL, ADD COLUMN "environment_id" character varying NULL, ADD COLUMN "scope_id" character varying NULL;
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "campaign_id" character varying NULL, ADD COLUMN "campaign_target_id" character varying NULL, ADD COLUMN "identity_holder_id" character varying NULL, ADD COLUMN "platform_id" character varying NULL;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "campaign_id" character varying NULL, ADD COLUMN "campaign_target_id" character varying NULL, ADD COLUMN "identity_holder_id" character varying NULL, ADD COLUMN "platform_id" character varying NULL;
-- Create index "workflowobjectref_workflow_instance_id_campaign_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_campaign_id" ON "workflow_object_refs" ("workflow_instance_id", "campaign_id");
-- Create index "workflowobjectref_workflow_instance_id_campaign_target_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_campaign_target_id" ON "workflow_object_refs" ("workflow_instance_id", "campaign_target_id");
-- Create index "workflowobjectref_workflow_instance_id_identity_holder_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_identity_holder_id" ON "workflow_object_refs" ("workflow_instance_id", "identity_holder_id");
-- Create index "workflowobjectref_workflow_instance_id_platform_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_platform_id" ON "workflow_object_refs" ("workflow_instance_id", "platform_id");
-- Create "campaigns" table
CREATE TABLE "campaigns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "name" character varying NOT NULL, "description" character varying NULL, "campaign_type" character varying NOT NULL DEFAULT 'QUESTIONNAIRE', "status" character varying NOT NULL DEFAULT 'DRAFT', "is_active" boolean NOT NULL DEFAULT false, "scheduled_at" timestamptz NULL, "launched_at" timestamptz NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "is_recurring" boolean NOT NULL DEFAULT false, "recurrence_frequency" character varying NULL, "recurrence_interval" bigint NULL DEFAULT 1, "recurrence_cron" character varying NULL, "recurrence_timezone" character varying NULL, "last_run_at" timestamptz NULL, "next_run_at" timestamptz NULL, "recurrence_end_at" timestamptz NULL, "recipient_count" bigint NULL DEFAULT 0, "resend_count" bigint NULL DEFAULT 0, "last_resent_at" timestamptz NULL, "metadata" jsonb NULL, "assessment_id" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "entity_id" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "campaign_display_id_owner_id" to table: "campaigns"
CREATE UNIQUE INDEX "campaign_display_id_owner_id" ON "campaigns" ("display_id", "owner_id");
-- Create index "campaign_entity_id" to table: "campaigns"
CREATE INDEX "campaign_entity_id" ON "campaigns" ("entity_id");
-- Create index "campaign_name_owner_id" to table: "campaigns"
CREATE UNIQUE INDEX "campaign_name_owner_id" ON "campaigns" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "campaign_owner_id" to table: "campaigns"
CREATE INDEX "campaign_owner_id" ON "campaigns" ("owner_id") WHERE (deleted_at IS NULL);
-- Create "campaign_targets" table
CREATE TABLE "campaign_targets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "email" character varying NOT NULL, "full_name" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "sent_at" timestamptz NULL, "completed_at" timestamptz NULL, "metadata" jsonb NULL, "campaign_id" character varying NOT NULL, "contact_id" character varying NULL, "group_id" character varying NULL, "owner_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "campaigntarget_campaign_id_email" to table: "campaign_targets"
CREATE UNIQUE INDEX "campaigntarget_campaign_id_email" ON "campaign_targets" ("campaign_id", "email") WHERE (deleted_at IS NULL);
-- Create index "campaigntarget_contact_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_contact_id" ON "campaign_targets" ("contact_id");
-- Create index "campaigntarget_group_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_group_id" ON "campaign_targets" ("group_id");
-- Create index "campaigntarget_owner_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_owner_id" ON "campaign_targets" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "campaigntarget_status" to table: "campaign_targets"
CREATE INDEX "campaigntarget_status" ON "campaign_targets" ("status");
-- Create index "campaigntarget_user_id" to table: "campaign_targets"
CREATE INDEX "campaigntarget_user_id" ON "campaign_targets" ("user_id");
-- Create "identity_holders" table
CREATE TABLE "identity_holders" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "full_name" character varying NOT NULL, "email" character varying NOT NULL, "alternate_email" character varying NULL, "phone_number" character varying NULL, "is_openlane_user" boolean NULL DEFAULT false, "identity_holder_type" character varying NOT NULL DEFAULT 'EMPLOYEE', "status" character varying NOT NULL DEFAULT 'ACTIVE', "is_active" boolean NOT NULL DEFAULT true, "title" character varying NULL, "department" character varying NULL, "team" character varying NULL, "location" character varying NULL, "start_date" timestamptz NULL, "end_date" timestamptz NULL, "external_user_id" character varying NULL, "external_reference_id" character varying NULL, "metadata" jsonb NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "employer_entity_id" character varying NULL, "owner_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "identityholder_display_id_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_display_id_owner_id" ON "identity_holders" ("display_id", "owner_id");
-- Create index "identityholder_email_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_email_owner_id" ON "identity_holders" ("email", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "identityholder_external_user_id" to table: "identity_holders"
CREATE INDEX "identityholder_external_user_id" ON "identity_holders" ("external_user_id");
-- Create index "identityholder_owner_id" to table: "identity_holders"
CREATE INDEX "identityholder_owner_id" ON "identity_holders" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "identityholder_user_id" to table: "identity_holders"
CREATE INDEX "identityholder_user_id" ON "identity_holders" ("user_id");
-- Create "platforms" table
CREATE TABLE "platforms" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "internal_owner" character varying NULL, "business_owner" character varying NULL, "technical_owner" character varying NULL, "security_owner" character varying NULL, "platform_kind_name" character varying NULL, "platform_data_classification_name" character varying NULL, "environment_name" character varying NULL, "scope_name" character varying NULL, "access_model_name" character varying NULL, "encryption_status_name" character varying NULL, "security_tier_name" character varying NULL, "criticality_name" character varying NULL, "workflow_eligible_marker" boolean NULL DEFAULT true, "name" character varying NOT NULL, "description" character varying NULL, "business_purpose" character varying NULL, "scope_statement" text NULL, "trust_boundary_description" text NULL, "data_flow_summary" text NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "physical_location" character varying NULL, "region" character varying NULL, "contains_pii" boolean NULL DEFAULT false, "source_type" character varying NOT NULL DEFAULT 'MANUAL', "source_identifier" character varying NULL, "cost_center" character varying NULL, "estimated_monthly_cost" double precision NULL, "purchase_date" timestamptz NULL, "external_reference_id" character varying NULL, "metadata" jsonb NULL, "custom_type_enum_platforms" character varying NULL, "identity_holder_access_platforms" character varying NULL, "owner_id" character varying NULL, "internal_owner_user_id" character varying NULL, "internal_owner_group_id" character varying NULL, "business_owner_user_id" character varying NULL, "business_owner_group_id" character varying NULL, "technical_owner_user_id" character varying NULL, "technical_owner_group_id" character varying NULL, "security_owner_user_id" character varying NULL, "security_owner_group_id" character varying NULL, "platform_kind_id" character varying NULL, "platform_data_classification_id" character varying NULL, "environment_id" character varying NULL, "scope_id" character varying NULL, "access_model_id" character varying NULL, "encryption_status_id" character varying NULL, "security_tier_id" character varying NULL, "criticality_id" character varying NULL, "platform_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "platform_display_id_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_display_id_owner_id" ON "platforms" ("display_id", "owner_id");
-- Create index "platform_name_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_name_owner_id" ON "platforms" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "platform_owner_id" to table: "platforms"
CREATE INDEX "platform_owner_id" ON "platforms" ("owner_id") WHERE (deleted_at IS NULL);
-- Create "asset_connected_assets" table
CREATE TABLE "asset_connected_assets" ("asset_id" character varying NOT NULL, "connected_from_id" character varying NOT NULL, PRIMARY KEY ("asset_id", "connected_from_id"));
-- Create "campaign_blocked_groups" table
CREATE TABLE "campaign_blocked_groups" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create "campaign_editors" table
CREATE TABLE "campaign_editors" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create "campaign_viewers" table
CREATE TABLE "campaign_viewers" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create "campaign_contacts" table
CREATE TABLE "campaign_contacts" ("campaign_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "contact_id"));
-- Create "campaign_users" table
CREATE TABLE "campaign_users" ("campaign_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "user_id"));
-- Create "campaign_groups" table
CREATE TABLE "campaign_groups" ("campaign_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "group_id"));
-- Create "campaign_identity_holders" table
CREATE TABLE "campaign_identity_holders" ("campaign_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "identity_holder_id"));
-- Create "control_scans" table
CREATE TABLE "control_scans" ("control_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "scan_id"));
-- Create "entity_integrations" table
CREATE TABLE "entity_integrations" ("entity_id" character varying NOT NULL, "integration_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "integration_id"));
-- Create "entity_subprocessors" table
CREATE TABLE "entity_subprocessors" ("entity_id" character varying NOT NULL, "subprocessor_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "subprocessor_id"));
-- Create "identity_holder_assessments" table
CREATE TABLE "identity_holder_assessments" ("identity_holder_id" character varying NOT NULL, "assessment_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "assessment_id"));
-- Create "identity_holder_templates" table
CREATE TABLE "identity_holder_templates" ("identity_holder_id" character varying NOT NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "template_id"));
-- Create "identity_holder_assets" table
CREATE TABLE "identity_holder_assets" ("identity_holder_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "asset_id"));
-- Create "identity_holder_entities" table
CREATE TABLE "identity_holder_entities" ("identity_holder_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "entity_id"));
-- Create "identity_holder_tasks" table
CREATE TABLE "identity_holder_tasks" ("identity_holder_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("identity_holder_id", "task_id"));
-- Create "platform_blocked_groups" table
CREATE TABLE "platform_blocked_groups" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- Create "platform_editors" table
CREATE TABLE "platform_editors" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- Create "platform_viewers" table
CREATE TABLE "platform_viewers" ("platform_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "group_id"));
-- Create "platform_assets" table
CREATE TABLE "platform_assets" ("platform_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "asset_id"));
-- Create "platform_entities" table
CREATE TABLE "platform_entities" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- Create "platform_evidence" table
CREATE TABLE "platform_evidence" ("platform_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "evidence_id"));
-- Create "platform_files" table
CREATE TABLE "platform_files" ("platform_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "file_id"));
-- Create "platform_risks" table
CREATE TABLE "platform_risks" ("platform_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "risk_id"));
-- Create "platform_controls" table
CREATE TABLE "platform_controls" ("platform_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "control_id"));
-- Create "platform_assessments" table
CREATE TABLE "platform_assessments" ("platform_id" character varying NOT NULL, "assessment_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "assessment_id"));
-- Create "platform_scans" table
CREATE TABLE "platform_scans" ("platform_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "scan_id"));
-- Create "platform_tasks" table
CREATE TABLE "platform_tasks" ("platform_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "task_id"));
-- Create "platform_identity_holders" table
CREATE TABLE "platform_identity_holders" ("platform_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "identity_holder_id"));
-- Create "platform_source_entities" table
CREATE TABLE "platform_source_entities" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- Create "platform_out_of_scope_assets" table
CREATE TABLE "platform_out_of_scope_assets" ("platform_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "asset_id"));
-- Create "platform_out_of_scope_vendors" table
CREATE TABLE "platform_out_of_scope_vendors" ("platform_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "entity_id"));
-- Create "platform_applicable_frameworks" table
CREATE TABLE "platform_applicable_frameworks" ("platform_id" character varying NOT NULL, "standard_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "standard_id"));
-- Create "scan_evidence" table
CREATE TABLE "scan_evidence" ("scan_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "evidence_id"));
-- Create "scan_files" table
CREATE TABLE "scan_files" ("scan_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "file_id"));
-- Create "scan_remediations" table
CREATE TABLE "scan_remediations" ("scan_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "remediation_id"));
-- Create "scan_action_plans" table
CREATE TABLE "scan_action_plans" ("scan_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "action_plan_id"));
-- Create "scan_tasks" table
CREATE TABLE "scan_tasks" ("scan_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "task_id"));
-- Create "vulnerability_scans" table
CREATE TABLE "vulnerability_scans" ("vulnerability_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "scan_id"));
-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD CONSTRAINT "assessment_responses_campaigns_assessment_responses" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_entities_assessment_responses" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_identity_holders_assessment_responses" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "assets" table
ALTER TABLE "assets" ADD CONSTRAINT "assets_custom_type_enums_access_model" FOREIGN KEY ("access_model_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_asset_data_classification" FOREIGN KEY ("asset_data_classification_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_asset_subtype" FOREIGN KEY ("asset_subtype_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_criticality" FOREIGN KEY ("criticality_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_encryption_status" FOREIGN KEY ("encryption_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_custom_type_enums_security_tier" FOREIGN KEY ("security_tier_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_platforms_source_assets" FOREIGN KEY ("source_platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD CONSTRAINT "custom_type_enums_entities_auth_methods" FOREIGN KEY ("entity_auth_methods") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD CONSTRAINT "directory_accounts_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_accounts_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD CONSTRAINT "directory_groups_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD CONSTRAINT "directory_memberships_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_memberships_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "directory_sync_runs" table
ALTER TABLE "directory_sync_runs" ADD CONSTRAINT "directory_sync_runs_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_sync_runs_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" ADD CONSTRAINT "document_data_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_custom_type_enums_entity_relationship_state" FOREIGN KEY ("entity_relationship_state_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_entity_security_questionnaire_status" FOREIGN KEY ("entity_security_questionnaire_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_entity_source_type" FOREIGN KEY ("entity_source_type_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "evidences" table
ALTER TABLE "evidences" ADD CONSTRAINT "evidences_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "findings" table
ALTER TABLE "findings" ADD CONSTRAINT "findings_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_identity_holders_blocked_groups" FOREIGN KEY ("identity_holder_blocked_groups") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_editors" FOREIGN KEY ("identity_holder_editors") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_identity_holders_viewers" FOREIGN KEY ("identity_holder_viewers") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "remediations" table
ALTER TABLE "remediations" ADD CONSTRAINT "remediations_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "reviews" table
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "scans" table
ALTER TABLE "scans" ADD CONSTRAINT "scans_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_assigned_to_group" FOREIGN KEY ("assigned_to_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_performed_by_group" FOREIGN KEY ("performed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_groups_reviewed_by_group" FOREIGN KEY ("reviewed_by_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_platforms_generated_scans" FOREIGN KEY ("generated_by_platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_assigned_to_user" FOREIGN KEY ("assigned_to_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_performed_by_user" FOREIGN KEY ("performed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_users_reviewed_by_user" FOREIGN KEY ("reviewed_by_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD CONSTRAINT "vulnerabilities_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD CONSTRAINT "workflow_instances_campaign_targets_campaign_target" FOREIGN KEY ("campaign_target_id") REFERENCES "campaign_targets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_campaigns_campaign" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_identity_holders_identity_holder" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_platforms_platform" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD CONSTRAINT "workflow_object_refs_campaign_targets_campaign_target" FOREIGN KEY ("campaign_target_id") REFERENCES "campaign_targets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_campaigns_campaign" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_identity_holders_identity_holder" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_platforms_platform" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "campaigns" table
ALTER TABLE "campaigns" ADD CONSTRAINT "campaigns_assessments_campaigns" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_entities_campaigns" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_organizations_campaigns" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_templates_campaigns" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaigns_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "campaign_targets" table
ALTER TABLE "campaign_targets" ADD CONSTRAINT "campaign_targets_campaigns_campaign_targets" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "campaign_targets_contacts_campaign_targets" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_groups_campaign_targets" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_organizations_campaign_targets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "campaign_targets_users_campaign_targets" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD CONSTRAINT "identity_holders_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_entities_employer" FOREIGN KEY ("employer_entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_organizations_identity_holders" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_users_identity_holder_profiles" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "identity_holders_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "platforms" table
ALTER TABLE "platforms" ADD CONSTRAINT "platforms_custom_type_enums_access_model" FOREIGN KEY ("access_model_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_criticality" FOREIGN KEY ("criticality_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_encryption_status" FOREIGN KEY ("encryption_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_environment" FOREIGN KEY ("environment_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platform_data_classification" FOREIGN KEY ("platform_data_classification_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platform_kind" FOREIGN KEY ("platform_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_platforms" FOREIGN KEY ("custom_type_enum_platforms") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_scope" FOREIGN KEY ("scope_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_custom_type_enums_security_tier" FOREIGN KEY ("security_tier_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_business_owner_group" FOREIGN KEY ("business_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_internal_owner_group" FOREIGN KEY ("internal_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_security_owner_group" FOREIGN KEY ("security_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_groups_technical_owner_group" FOREIGN KEY ("technical_owner_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_identity_holders_access_platforms" FOREIGN KEY ("identity_holder_access_platforms") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_organizations_platforms" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_business_owner_user" FOREIGN KEY ("business_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_internal_owner_user" FOREIGN KEY ("internal_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_platforms_owned" FOREIGN KEY ("platform_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_security_owner_user" FOREIGN KEY ("security_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_users_technical_owner_user" FOREIGN KEY ("technical_owner_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
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
-- Modify "control_scans" table
ALTER TABLE "control_scans" ADD CONSTRAINT "control_scans_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_integrations" table
ALTER TABLE "entity_integrations" ADD CONSTRAINT "entity_integrations_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_integrations_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_subprocessors" table
ALTER TABLE "entity_subprocessors" ADD CONSTRAINT "entity_subprocessors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_subprocessors_subprocessor_id" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- Modify "vulnerability_scans" table
ALTER TABLE "vulnerability_scans" ADD CONSTRAINT "vulnerability_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_scans_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
