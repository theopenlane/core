-- Create "api_tokens" table
CREATE TABLE "api_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "api_tokens_token_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_token_key" ON "api_tokens" ("token");
-- Create index "apitoken_token" to table: "api_tokens"
CREATE INDEX "apitoken_token" ON "api_tokens" ("token");
-- Create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "action_plan_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "due_date" timestamptz NULL, "priority" character varying NULL, "source" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "subcontrol_action_plans" character varying NULL, PRIMARY KEY ("id"));
-- Create "action_plan_history" table
CREATE TABLE "action_plan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "action_plan_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "approver_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "due_date" timestamptz NULL, "priority" character varying NULL, "source" character varying NULL, PRIMARY KEY ("id"));
-- Create index "actionplanhistory_history_time" to table: "action_plan_history"
CREATE INDEX "actionplanhistory_history_time" ON "action_plan_history" ("history_time");
-- Create "contacts" table
CREATE TABLE "contacts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "contact_history" table
CREATE TABLE "contact_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', PRIMARY KEY ("id"));
-- Create index "contacthistory_history_time" to table: "contact_history"
CREATE INDEX "contacthistory_history_time" ON "contact_history" ("history_time");
-- Create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "ref_code" character varying NOT NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "internal_policy_controls" character varying NULL, "owner_id" character varying NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "control_display_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_display_id_owner_id" ON "controls" ("display_id", "owner_id");
-- Create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NULL));
-- Create index "controls_auditor_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_auditor_reference_id_key" ON "controls" ("auditor_reference_id");
-- Create index "controls_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_reference_id_key" ON "controls" ("reference_id");
-- Create "control_history" table
CREATE TABLE "control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "ref_code" character varying NOT NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "controlhistory_history_time" to table: "control_history"
CREATE INDEX "controlhistory_history_time" ON "control_history" ("history_time");
-- Create "control_implementations" table
CREATE TABLE "control_implementations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NULL DEFAULT 'DRAFT', "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "control_implementation_history" table
CREATE TABLE "control_implementation_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "status" character varying NULL DEFAULT 'DRAFT', "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, PRIMARY KEY ("id"));
-- Create index "controlimplementationhistory_history_time" to table: "control_implementation_history"
CREATE INDEX "controlimplementationhistory_history_time" ON "control_implementation_history" ("history_time");
-- Create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "desired_outcome" text NULL, "status" character varying NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "control_objective_type" character varying NULL, "category" character varying NULL, "subcategory" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "controlobjective_display_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_display_id_owner_id" ON "control_objectives" ("display_id", "owner_id");
-- Create "control_objective_history" table
CREATE TABLE "control_objective_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "name" character varying NOT NULL, "desired_outcome" text NULL, "status" character varying NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "control_objective_type" character varying NULL, "category" character varying NULL, "subcategory" character varying NULL, PRIMARY KEY ("id"));
-- Create index "controlobjectivehistory_history_time" to table: "control_objective_history"
CREATE INDEX "controlobjectivehistory_history_time" ON "control_objective_history" ("history_time");
-- Create "document_data" table
CREATE TABLE "document_data" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "data" jsonb NOT NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create "document_data_history" table
CREATE TABLE "document_data_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, "data" jsonb NOT NULL, PRIMARY KEY ("id"));
-- Create index "documentdatahistory_history_time" to table: "document_data_history"
CREATE INDEX "documentdatahistory_history_time" ON "document_data_history" ("history_time");
-- Create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_token_key" ON "email_verification_tokens" ("token");
-- Create index "emailverificationtoken_token" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_token" ON "email_verification_tokens" ("token") WHERE (deleted_at IS NULL);
-- Create "entities" table
CREATE TABLE "entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "status" character varying NULL DEFAULT 'active', "entity_type_id" character varying NULL, "entity_type_entities" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "entity_name_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create "entity_history" table
CREATE TABLE "entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" citext NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "entity_type_id" character varying NULL, "status" character varying NULL DEFAULT 'active', PRIMARY KEY ("id"));
-- Create index "entityhistory_history_time" to table: "entity_history"
CREATE INDEX "entityhistory_history_time" ON "entity_history" ("history_time");
-- Create "entity_types" table
CREATE TABLE "entity_types" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "entitytype_name_owner_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_name_owner_id" ON "entity_types" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create "entity_type_history" table
CREATE TABLE "entity_type_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" citext NOT NULL, PRIMARY KEY ("id"));
-- Create index "entitytypehistory_history_time" to table: "entity_type_history"
CREATE INDEX "entitytypehistory_history_time" ON "entity_type_history" ("history_time");
-- Create "events" table
CREATE TABLE "events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- Create "event_history" table
CREATE TABLE "event_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "eventhistory_history_time" to table: "event_history"
CREATE INDEX "eventhistory_history_time" ON "event_history" ("history_time");
-- Create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "status" character varying NULL DEFAULT 'READY', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- Create "evidence_history" table
CREATE TABLE "evidence_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "status" character varying NULL DEFAULT 'READY', PRIMARY KEY ("id"));
-- Create index "evidencehistory_history_time" to table: "evidence_history"
CREATE INDEX "evidencehistory_history_time" ON "evidence_history" ("history_time");
-- Create "files" table
CREATE TABLE "files" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, "note_files" character varying NULL, PRIMARY KEY ("id"));
-- Create "file_history" table
CREATE TABLE "file_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, PRIMARY KEY ("id"));
-- Create index "filehistory_history_time" to table: "file_history"
CREATE INDEX "filehistory_history_time" ON "file_history" ("history_time");
-- Create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" citext NOT NULL, "description" character varying NULL, "is_managed" boolean NULL DEFAULT false, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', "organization_control_creators" character varying NULL, "organization_control_objective_creators" character varying NULL, "organization_group_creators" character varying NULL, "organization_internal_policy_creators" character varying NULL, "organization_narrative_creators" character varying NULL, "organization_procedure_creators" character varying NULL, "organization_program_creators" character varying NULL, "organization_risk_creators" character varying NULL, "organization_template_creators" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "group_display_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_display_id_owner_id" ON "groups" ("display_id", "owner_id");
-- Create index "group_name_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_name_owner_id" ON "groups" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create "group_history" table
CREATE TABLE "group_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" citext NOT NULL, "description" character varying NULL, "is_managed" boolean NULL DEFAULT false, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', PRIMARY KEY ("id"));
-- Create index "grouphistory_history_time" to table: "group_history"
CREATE INDEX "grouphistory_history_time" ON "group_history" ("history_time");
-- Create "group_memberships" table
CREATE TABLE "group_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, "group_membership_orgmembership" character varying NULL, PRIMARY KEY ("id"));
-- Create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id") WHERE (deleted_at IS NULL);
-- Create "group_membership_history" table
CREATE TABLE "group_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "groupmembershiphistory_history_time" to table: "group_membership_history"
CREATE INDEX "groupmembershiphistory_history_time" ON "group_membership_history" ("history_time");
-- Create "group_settings" table
CREATE TABLE "group_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "group_settings_group_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_group_id_key" ON "group_settings" ("group_id");
-- Create "group_setting_history" table
CREATE TABLE "group_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "groupsettinghistory_history_time" to table: "group_setting_history"
CREATE INDEX "groupsettinghistory_history_time" ON "group_setting_history" ("history_time");
-- Create "hushes" table
CREATE TABLE "hushes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "hush_history" table
CREATE TABLE "hush_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, PRIMARY KEY ("id"));
-- Create index "hushhistory_history_time" to table: "hush_history"
CREATE INDEX "hushhistory_history_time" ON "hush_history" ("history_time");
-- Create "integrations" table
CREATE TABLE "integrations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "group_integrations" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "integration_history" table
CREATE TABLE "integration_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, PRIMARY KEY ("id"));
-- Create index "integrationhistory_history_time" to table: "integration_history"
CREATE INDEX "integrationhistory_history_time" ON "integration_history" ("history_time");
-- Create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "policy_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "control_internal_policies" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "subcontrol_internal_policies" character varying NULL, PRIMARY KEY ("id"));
-- Create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_display_id_owner_id" ON "internal_policies" ("display_id", "owner_id");
-- Create "internal_policy_history" table
CREATE TABLE "internal_policy_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "policy_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "approver_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
CREATE INDEX "internalpolicyhistory_history_time" ON "internal_policy_history" ("history_time");
-- Create "invites" table
CREATE TABLE "invites" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "expires" timestamptz NULL, "recipient" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'INVITATION_SENT', "role" character varying NOT NULL DEFAULT 'MEMBER', "send_attempts" bigint NOT NULL DEFAULT 1, "requestor_id" character varying NULL, "secret" bytea NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "invite_recipient_owner_id" to table: "invites"
CREATE UNIQUE INDEX "invite_recipient_owner_id" ON "invites" ("recipient", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "invites_token_key" to table: "invites"
CREATE UNIQUE INDEX "invites_token_key" ON "invites" ("token");
-- Create "mapped_controls" table
CREATE TABLE "mapped_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "mapping_type" character varying NULL, "relation" character varying NULL, PRIMARY KEY ("id"));
-- Create "mapped_control_history" table
CREATE TABLE "mapped_control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "mapping_type" character varying NULL, "relation" character varying NULL, PRIMARY KEY ("id"));
-- Create index "mappedcontrolhistory_history_time" to table: "mapped_control_history"
CREATE INDEX "mappedcontrolhistory_history_time" ON "mapped_control_history" ("history_time");
-- Create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "details" text NULL, "control_objective_narratives" character varying NULL, "internal_policy_narratives" character varying NULL, "owner_id" character varying NULL, "procedure_narratives" character varying NULL, "subcontrol_narratives" character varying NULL, PRIMARY KEY ("id"));
-- Create index "narrative_display_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_display_id_owner_id" ON "narratives" ("display_id", "owner_id");
-- Create "narrative_history" table
CREATE TABLE "narrative_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "details" text NULL, PRIMARY KEY ("id"));
-- Create index "narrativehistory_history_time" to table: "narrative_history"
CREATE INDEX "narrativehistory_history_time" ON "narrative_history" ("history_time");
-- Create "notes" table
CREATE TABLE "notes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "text" text NOT NULL, "entity_notes" character varying NULL, "owner_id" character varying NULL, "program_notes" character varying NULL, "task_comments" character varying NULL, PRIMARY KEY ("id"));
-- Create index "note_display_id_owner_id" to table: "notes"
CREATE UNIQUE INDEX "note_display_id_owner_id" ON "notes" ("display_id", "owner_id");
-- Create "note_history" table
CREATE TABLE "note_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "text" text NOT NULL, PRIMARY KEY ("id"));
-- Create index "notehistory_history_time" to table: "note_history"
CREATE INDEX "notehistory_history_time" ON "note_history" ("history_time");
-- Create "onboardings" table
CREATE TABLE "onboardings" ("id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "company_name" character varying NOT NULL, "domains" jsonb NULL, "company_details" jsonb NULL, "user_details" jsonb NULL, "compliance" jsonb NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "org_memberships" table
CREATE TABLE "org_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id") WHERE (deleted_at IS NULL);
-- Create "org_membership_history" table
CREATE TABLE "org_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "orgmembershiphistory_history_time" to table: "org_membership_history"
CREATE INDEX "orgmembershiphistory_history_time" ON "org_membership_history" ("history_time");
-- Create "org_subscriptions" table
CREATE TABLE "org_subscriptions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "product_price" jsonb NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "trial_expires_at" timestamptz NULL, "days_until_due" character varying NULL, "payment_method_added" boolean NULL, "features" jsonb NULL, "feature_lookup_keys" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "org_subscriptions_stripe_customer_id_key" to table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_stripe_customer_id_key" ON "org_subscriptions" ("stripe_customer_id");
-- Create "org_subscription_history" table
CREATE TABLE "org_subscription_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "product_price" jsonb NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "trial_expires_at" timestamptz NULL, "days_until_due" character varying NULL, "payment_method_added" boolean NULL, "features" jsonb NULL, "feature_lookup_keys" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "orgsubscriptionhistory_history_time" to table: "org_subscription_history"
CREATE INDEX "orgsubscriptionhistory_history_time" ON "org_subscription_history" ("history_time");
-- Create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "dedicated_db" boolean NOT NULL DEFAULT false, "parent_organization_id" character varying NULL, "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "organization_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_name" ON "organizations" ("name") WHERE (deleted_at IS NULL);
-- Create "organization_history" table
CREATE TABLE "organization_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "parent_organization_id" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "avatar_local_file_id" character varying NULL, "avatar_updated_at" timestamptz NULL, "dedicated_db" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- Create index "organizationhistory_history_time" to table: "organization_history"
CREATE INDEX "organizationhistory_history_time" ON "organization_history" ("history_time");
-- Create "organization_settings" table
CREATE TABLE "organization_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" jsonb NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "billing_notifications_enabled" boolean NOT NULL DEFAULT true, "allowed_email_domains" jsonb NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "organization_settings_organization_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_organization_id_key" ON "organization_settings" ("organization_id");
-- Create "organization_setting_history" table
CREATE TABLE "organization_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" jsonb NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "organization_id" character varying NULL, "billing_notifications_enabled" boolean NOT NULL DEFAULT true, "allowed_email_domains" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
CREATE INDEX "organizationsettinghistory_history_time" ON "organization_setting_history" ("history_time");
-- Create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
-- Create index "passwordresettoken_token" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_token" ON "password_reset_tokens" ("token") WHERE (deleted_at IS NULL);
-- Create "personal_access_tokens" table
CREATE TABLE "personal_access_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_token_key" ON "personal_access_tokens" ("token");
-- Create index "personalaccesstoken_token" to table: "personal_access_tokens"
CREATE INDEX "personalaccesstoken_token" ON "personal_access_tokens" ("token");
-- Create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "procedure_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "control_objective_procedures" character varying NULL, "owner_id" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "subcontrol_procedures" character varying NULL, PRIMARY KEY ("id"));
-- Create index "procedure_display_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_display_id_owner_id" ON "procedures" ("display_id", "owner_id");
-- Create "procedure_history" table
CREATE TABLE "procedure_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "procedure_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "approver_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "procedurehistory_history_time" to table: "procedure_history"
CREATE INDEX "procedurehistory_history_time" ON "procedure_history" ("history_time");
-- Create "programs" table
CREATE TABLE "programs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "program_display_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_display_id_owner_id" ON "programs" ("display_id", "owner_id");
-- Create "program_history" table
CREATE TABLE "program_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- Create index "programhistory_history_time" to table: "program_history"
CREATE INDEX "programhistory_history_time" ON "program_history" ("history_time");
-- Create "program_memberships" table
CREATE TABLE "program_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, "program_membership_orgmembership" character varying NULL, PRIMARY KEY ("id"));
-- Create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id") WHERE (deleted_at IS NULL);
-- Create "program_membership_history" table
CREATE TABLE "program_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "programmembershiphistory_history_time" to table: "program_membership_history"
CREATE INDEX "programmembershiphistory_history_time" ON "program_membership_history" ("history_time");
-- Create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'OPEN', "risk_type" character varying NULL, "category" character varying NULL, "impact" character varying NULL DEFAULT 'MODERATE', "likelihood" character varying NULL DEFAULT 'LIKELY', "score" bigint NULL, "mitigation" text NULL, "details" text NULL, "business_costs" text NULL, "control_objective_risks" character varying NULL, "owner_id" character varying NULL, "stakeholder_id" character varying NULL, "delegate_id" character varying NULL, "subcontrol_risks" character varying NULL, PRIMARY KEY ("id"));
-- Create index "risk_display_id_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_display_id_owner_id" ON "risks" ("display_id", "owner_id");
-- Create "risk_history" table
CREATE TABLE "risk_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'OPEN', "risk_type" character varying NULL, "category" character varying NULL, "impact" character varying NULL DEFAULT 'MODERATE', "likelihood" character varying NULL DEFAULT 'LIKELY', "score" bigint NULL, "mitigation" text NULL, "details" text NULL, "business_costs" text NULL, "stakeholder_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "riskhistory_history_time" to table: "risk_history"
CREATE INDEX "riskhistory_history_time" ON "risk_history" ("history_time");
-- Create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "name" character varying NOT NULL, "short_name" character varying NULL, "framework" text NULL, "description" text NULL, "governing_body_logo_url" character varying NULL, "governing_body" character varying NULL, "domains" jsonb NULL, "link" character varying NULL, "status" character varying NULL DEFAULT 'ACTIVE', "is_public" boolean NULL DEFAULT false, "free_to_use" boolean NULL DEFAULT false, "standard_type" character varying NULL, "version" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create "standard_history" table
CREATE TABLE "standard_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "name" character varying NOT NULL, "short_name" character varying NULL, "framework" text NULL, "description" text NULL, "governing_body_logo_url" character varying NULL, "governing_body" character varying NULL, "domains" jsonb NULL, "link" character varying NULL, "status" character varying NULL DEFAULT 'ACTIVE', "is_public" boolean NULL DEFAULT false, "free_to_use" boolean NULL DEFAULT false, "standard_type" character varying NULL, "version" character varying NULL, PRIMARY KEY ("id"));
-- Create index "standardhistory_history_time" to table: "standard_history"
CREATE INDEX "standardhistory_history_time" ON "standard_history" ("history_time");
-- Create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "ref_code" character varying NOT NULL, "control_id" character varying NOT NULL, "owner_id" character varying NULL, "program_subcontrols" character varying NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "user_subcontrols" character varying NULL, PRIMARY KEY ("id"));
-- Create index "subcontrol_control_id_ref_code" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_control_id_ref_code" ON "subcontrols" ("control_id", "ref_code") WHERE (deleted_at IS NULL);
-- Create index "subcontrol_display_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_display_id_owner_id" ON "subcontrols" ("display_id", "owner_id");
-- Create "subcontrol_history" table
CREATE TABLE "subcontrol_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "ref_code" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "subcontrolhistory_history_time" to table: "subcontrol_history"
CREATE INDEX "subcontrolhistory_history_time" ON "subcontrol_history" ("history_time");
-- Create "subscribers" table
CREATE TABLE "subscribers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "phone_number" character varying NULL, "verified_email" boolean NOT NULL DEFAULT false, "verified_phone" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT false, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "secret" bytea NOT NULL, "unsubscribed" boolean NOT NULL DEFAULT false, "send_attempts" bigint NOT NULL DEFAULT 1, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false));
-- Create index "subscribers_token_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_token_key" ON "subscribers" ("token");
-- Create "tfa_settings" table
CREATE TABLE "tfa_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tfa_secret" character varying NULL, "verified" boolean NOT NULL DEFAULT false, "recovery_codes" jsonb NULL, "phone_otp_allowed" boolean NULL DEFAULT false, "email_otp_allowed" boolean NULL DEFAULT false, "totp_allowed" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "tfasetting_owner_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_owner_id" ON "tfa_settings" ("owner_id") WHERE (deleted_at IS NULL);
-- Create "tasks" table
CREATE TABLE "tasks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" text NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "category" character varying NULL, "due" timestamptz NULL, "completed" timestamptz NULL, "owner_id" character varying NULL, "assigner_id" character varying NULL, "assignee_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "task_display_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_display_id_owner_id" ON "tasks" ("display_id", "owner_id");
-- Create "task_history" table
CREATE TABLE "task_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" text NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "category" character varying NULL, "due" timestamptz NULL, "completed" timestamptz NULL, "assignee_id" character varying NULL, "assigner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "taskhistory_history_time" to table: "task_history"
CREATE INDEX "taskhistory_history_time" ON "task_history" ("history_time");
-- Create "templates" table
CREATE TABLE "templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "template_name_owner_id_template_type" to table: "templates"
CREATE UNIQUE INDEX "template_name_owner_id_template_type" ON "templates" ("name", "owner_id", "template_type") WHERE (deleted_at IS NULL);
-- Create "template_history" table
CREATE TABLE "template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "templatehistory_history_time" to table: "template_history"
CREATE INDEX "templatehistory_history_time" ON "template_history" ("history_time");
-- Create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "user_email_auth_provider" to table: "users"
CREATE UNIQUE INDEX "user_email_auth_provider" ON "users" ("email", "auth_provider") WHERE (deleted_at IS NULL);
-- Create index "user_id" to table: "users"
CREATE UNIQUE INDEX "user_id" ON "users" ("id");
-- Create index "users_display_id_key" to table: "users"
CREATE UNIQUE INDEX "users_display_id_key" ON "users" ("display_id");
-- Create index "users_sub_key" to table: "users"
CREATE UNIQUE INDEX "users_sub_key" ON "users" ("sub");
-- Create "user_history" table
CREATE TABLE "user_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_local_file_id" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', PRIMARY KEY ("id"));
-- Create index "userhistory_history_time" to table: "user_history"
CREATE INDEX "userhistory_history_time" ON "user_history" ("history_time");
-- Create "user_settings" table
CREATE TABLE "user_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, "user_id" character varying NULL, "user_setting_default_org" character varying NULL, PRIMARY KEY ("id"));
-- Create index "user_settings_user_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_user_id_key" ON "user_settings" ("user_id");
-- Create "user_setting_history" table
CREATE TABLE "user_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "user_id" character varying NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, PRIMARY KEY ("id"));
-- Create index "usersettinghistory_history_time" to table: "user_setting_history"
CREATE INDEX "usersettinghistory_history_time" ON "user_setting_history" ("history_time");
-- Create "webauthns" table
CREATE TABLE "webauthns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "credential_id" bytea NULL, "public_key" bytea NULL, "attestation_type" character varying NULL, "aaguid" bytea NOT NULL, "sign_count" integer NOT NULL, "transports" jsonb NOT NULL, "backup_eligible" boolean NOT NULL DEFAULT false, "backup_state" boolean NOT NULL DEFAULT false, "user_present" boolean NOT NULL DEFAULT false, "user_verified" boolean NOT NULL DEFAULT false, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "webauthns_aaguid_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_aaguid_key" ON "webauthns" ("aaguid");
-- Create index "webauthns_credential_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_credential_id_key" ON "webauthns" ("credential_id");
-- Create "contact_files" table
CREATE TABLE "contact_files" ("contact_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "file_id"));
-- Create "control_control_objectives" table
CREATE TABLE "control_control_objectives" ("control_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_objective_id"));
-- Create "control_tasks" table
CREATE TABLE "control_tasks" ("control_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_id", "task_id"));
-- Create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- Create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- Create "control_action_plans" table
CREATE TABLE "control_action_plans" ("control_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "action_plan_id"));
-- Create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- Create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- Create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- Create "control_viewers" table
CREATE TABLE "control_viewers" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- Create "control_control_implementations" table
CREATE TABLE "control_control_implementations" ("control_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_implementation_id"));
-- Create "control_objective_blocked_groups" table
CREATE TABLE "control_objective_blocked_groups" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- Create "control_objective_editors" table
CREATE TABLE "control_objective_editors" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- Create "control_objective_viewers" table
CREATE TABLE "control_objective_viewers" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- Create "control_objective_tasks" table
CREATE TABLE "control_objective_tasks" ("control_objective_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "task_id"));
-- Create "document_data_files" table
CREATE TABLE "document_data_files" ("document_data_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("document_data_id", "file_id"));
-- Create "entity_contacts" table
CREATE TABLE "entity_contacts" ("entity_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "contact_id"));
-- Create "entity_documents" table
CREATE TABLE "entity_documents" ("entity_id" character varying NOT NULL, "document_data_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "document_data_id"));
-- Create "entity_files" table
CREATE TABLE "entity_files" ("entity_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "file_id"));
-- Create "evidence_control_objectives" table
CREATE TABLE "evidence_control_objectives" ("evidence_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_objective_id"));
-- Create "evidence_controls" table
CREATE TABLE "evidence_controls" ("evidence_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_id"));
-- Create "evidence_subcontrols" table
CREATE TABLE "evidence_subcontrols" ("evidence_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "subcontrol_id"));
-- Create "evidence_files" table
CREATE TABLE "evidence_files" ("evidence_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "file_id"));
-- Create "file_events" table
CREATE TABLE "file_events" ("file_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("file_id", "event_id"));
-- Create "group_events" table
CREATE TABLE "group_events" ("group_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_id", "event_id"));
-- Create "group_files" table
CREATE TABLE "group_files" ("group_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("group_id", "file_id"));
-- Create "group_tasks" table
CREATE TABLE "group_tasks" ("group_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("group_id", "task_id"));
-- Create "group_membership_events" table
CREATE TABLE "group_membership_events" ("group_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_membership_id", "event_id"));
-- Create "hush_events" table
CREATE TABLE "hush_events" ("hush_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("hush_id", "event_id"));
-- Create "integration_secrets" table
CREATE TABLE "integration_secrets" ("integration_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "hush_id"));
-- Create "integration_events" table
CREATE TABLE "integration_events" ("integration_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "event_id"));
-- Create "internal_policy_blocked_groups" table
CREATE TABLE "internal_policy_blocked_groups" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- Create "internal_policy_editors" table
CREATE TABLE "internal_policy_editors" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- Create "internal_policy_control_objectives" table
CREATE TABLE "internal_policy_control_objectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- Create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- Create "internal_policy_tasks" table
CREATE TABLE "internal_policy_tasks" ("internal_policy_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "task_id"));
-- Create "invite_events" table
CREATE TABLE "invite_events" ("invite_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "event_id"));
-- Create "mapped_control_controls" table
CREATE TABLE "mapped_control_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"));
-- Create "mapped_control_subcontrols" table
CREATE TABLE "mapped_control_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"));
-- Create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- Create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- Create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- Create "org_membership_events" table
CREATE TABLE "org_membership_events" ("org_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_membership_id", "event_id"));
-- Create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"));
-- Create "organization_personal_access_tokens" table
CREATE TABLE "organization_personal_access_tokens" ("organization_id" character varying NOT NULL, "personal_access_token_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "personal_access_token_id"));
-- Create "organization_files" table
CREATE TABLE "organization_files" ("organization_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "file_id"));
-- Create "organization_events" table
CREATE TABLE "organization_events" ("organization_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "event_id"));
-- Create "organization_setting_files" table
CREATE TABLE "organization_setting_files" ("organization_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_setting_id", "file_id"));
-- Create "personal_access_token_events" table
CREATE TABLE "personal_access_token_events" ("personal_access_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("personal_access_token_id", "event_id"));
-- Create "procedure_blocked_groups" table
CREATE TABLE "procedure_blocked_groups" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- Create "procedure_editors" table
CREATE TABLE "procedure_editors" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- Create "procedure_risks" table
CREATE TABLE "procedure_risks" ("procedure_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "risk_id"));
-- Create "procedure_tasks" table
CREATE TABLE "procedure_tasks" ("procedure_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "task_id"));
-- Create "program_blocked_groups" table
CREATE TABLE "program_blocked_groups" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- Create "program_editors" table
CREATE TABLE "program_editors" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- Create "program_viewers" table
CREATE TABLE "program_viewers" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- Create "program_controls" table
CREATE TABLE "program_controls" ("program_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_id"));
-- Create "program_control_objectives" table
CREATE TABLE "program_control_objectives" ("program_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_objective_id"));
-- Create "program_internal_policies" table
CREATE TABLE "program_internal_policies" ("program_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("program_id", "internal_policy_id"));
-- Create "program_procedures" table
CREATE TABLE "program_procedures" ("program_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("program_id", "procedure_id"));
-- Create "program_risks" table
CREATE TABLE "program_risks" ("program_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("program_id", "risk_id"));
-- Create "program_tasks" table
CREATE TABLE "program_tasks" ("program_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("program_id", "task_id"));
-- Create "program_files" table
CREATE TABLE "program_files" ("program_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("program_id", "file_id"));
-- Create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"));
-- Create "program_narratives" table
CREATE TABLE "program_narratives" ("program_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("program_id", "narrative_id"));
-- Create "program_action_plans" table
CREATE TABLE "program_action_plans" ("program_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("program_id", "action_plan_id"));
-- Create "risk_blocked_groups" table
CREATE TABLE "risk_blocked_groups" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- Create "risk_editors" table
CREATE TABLE "risk_editors" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- Create "risk_viewers" table
CREATE TABLE "risk_viewers" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- Create "risk_action_plans" table
CREATE TABLE "risk_action_plans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- Create "subcontrol_control_objectives" table
CREATE TABLE "subcontrol_control_objectives" ("subcontrol_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_objective_id"));
-- Create "subcontrol_tasks" table
CREATE TABLE "subcontrol_tasks" ("subcontrol_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "task_id"));
-- Create "subcontrol_control_implementations" table
CREATE TABLE "subcontrol_control_implementations" ("subcontrol_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_implementation_id"));
-- Create "subscriber_events" table
CREATE TABLE "subscriber_events" ("subscriber_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("subscriber_id", "event_id"));
-- Create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"));
-- Create "template_files" table
CREATE TABLE "template_files" ("template_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("template_id", "file_id"));
-- Create "user_files" table
CREATE TABLE "user_files" ("user_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_id", "file_id"));
-- Create "user_events" table
CREATE TABLE "user_events" ("user_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("user_id", "event_id"));
-- Create "user_action_plans" table
CREATE TABLE "user_action_plans" ("user_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("user_id", "action_plan_id"));
-- Create "user_setting_files" table
CREATE TABLE "user_setting_files" ("user_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_setting_id", "file_id"));
-- Modify "api_tokens" table
ALTER TABLE "api_tokens" ADD CONSTRAINT "api_tokens_organizations_api_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_organizations_action_plans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_subcontrols_action_plans" FOREIGN KEY ("subcontrol_action_plans") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" ADD CONSTRAINT "contacts_organizations_contacts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_internal_policies_controls" FOREIGN KEY ("internal_policy_controls") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_standards_controls" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD CONSTRAINT "control_implementations_organizations_control_implementations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_organizations_control_objectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" ADD CONSTRAINT "document_data_organizations_documents" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD CONSTRAINT "email_verification_tokens_users_email_verification_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_organizations_entities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "entity_types" table
ALTER TABLE "entity_types" ADD CONSTRAINT "entity_types_organizations_entity_types" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "evidences" table
ALTER TABLE "evidences" ADD CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_notes_files" FOREIGN KEY ("note_files") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_organizations_control_creators" FOREIGN KEY ("organization_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_objective_creators" FOREIGN KEY ("organization_control_objective_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_creators" FOREIGN KEY ("organization_group_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_internal_policy_creators" FOREIGN KEY ("organization_internal_policy_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_narrative_creators" FOREIGN KEY ("organization_narrative_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_procedure_creators" FOREIGN KEY ("organization_procedure_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_program_creators" FOREIGN KEY ("organization_program_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_risk_creators" FOREIGN KEY ("organization_risk_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_template_creators" FOREIGN KEY ("organization_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "group_memberships" table
ALTER TABLE "group_memberships" ADD CONSTRAINT "group_memberships_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "group_memberships_org_memberships_orgmembership" FOREIGN KEY ("group_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "group_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "group_settings" table
ALTER TABLE "group_settings" ADD CONSTRAINT "group_settings_groups_setting" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "hushes" table
ALTER TABLE "hushes" ADD CONSTRAINT "hushes_organizations_secrets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_groups_integrations" FOREIGN KEY ("group_integrations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_organizations_integrations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_controls_internal_policies" FOREIGN KEY ("control_internal_policies") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_organizations_internal_policies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_subcontrols_internal_policies" FOREIGN KEY ("subcontrol_internal_policies") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "invites" table
ALTER TABLE "invites" ADD CONSTRAINT "invites_organizations_invites" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "narratives" table
ALTER TABLE "narratives" ADD CONSTRAINT "narratives_control_objectives_narratives" FOREIGN KEY ("control_objective_narratives") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_internal_policies_narratives" FOREIGN KEY ("internal_policy_narratives") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_procedures_narratives" FOREIGN KEY ("procedure_narratives") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_subcontrols_narratives" FOREIGN KEY ("subcontrol_narratives") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "notes" table
ALTER TABLE "notes" ADD CONSTRAINT "notes_entities_notes" FOREIGN KEY ("entity_notes") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_organizations_notes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_programs_notes" FOREIGN KEY ("program_notes") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_tasks_comments" FOREIGN KEY ("task_comments") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "onboardings" table
ALTER TABLE "onboardings" ADD CONSTRAINT "onboardings_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "org_memberships" table
ALTER TABLE "org_memberships" ADD CONSTRAINT "org_memberships_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "org_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
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
-- Modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_organizations_procedures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_subcontrols_procedures" FOREIGN KEY ("subcontrol_procedures") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "programs" table
ALTER TABLE "programs" ADD CONSTRAINT "programs_organizations_programs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" ADD CONSTRAINT "program_memberships_org_memberships_orgmembership" FOREIGN KEY ("program_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "program_memberships_programs_program" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "program_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("stakeholder_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_subcontrols_risks" FOREIGN KEY ("subcontrol_risks") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "standards" table
ALTER TABLE "standards" ADD CONSTRAINT "standards_organizations_standards" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_controls_subcontrols" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_programs_subcontrols" FOREIGN KEY ("program_subcontrols") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_subcontrols" FOREIGN KEY ("user_subcontrols") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subscribers" table
ALTER TABLE "subscribers" ADD CONSTRAINT "subscribers_organizations_subscribers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD CONSTRAINT "tfa_settings_users_tfa_settings" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_organizations_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "user_settings" table
ALTER TABLE "user_settings" ADD CONSTRAINT "user_settings_organizations_default_org" FOREIGN KEY ("user_setting_default_org") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_settings_users_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "webauthns" table
ALTER TABLE "webauthns" ADD CONSTRAINT "webauthns_users_webauthn" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
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
-- Modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" ADD CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_editors" table
ALTER TABLE "control_editors" ADD CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_viewers" table
ALTER TABLE "control_viewers" ADD CONSTRAINT "control_viewers_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "control_control_implementations" table
ALTER TABLE "control_control_implementations" ADD CONSTRAINT "control_control_implementations_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- Modify "entity_contacts" table
ALTER TABLE "entity_contacts" ADD CONSTRAINT "entity_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_contacts_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_documents" table
ALTER TABLE "entity_documents" ADD CONSTRAINT "entity_documents_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_documents_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "entity_files" table
ALTER TABLE "entity_files" ADD CONSTRAINT "entity_files_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_control_objectives" table
ALTER TABLE "evidence_control_objectives" ADD CONSTRAINT "evidence_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_control_objectives_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_controls" table
ALTER TABLE "evidence_controls" ADD CONSTRAINT "evidence_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_controls_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_subcontrols" table
ALTER TABLE "evidence_subcontrols" ADD CONSTRAINT "evidence_subcontrols_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "evidence_files" table
ALTER TABLE "evidence_files" ADD CONSTRAINT "evidence_files_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "file_events" table
ALTER TABLE "file_events" ADD CONSTRAINT "file_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_events_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- Modify "integration_secrets" table
ALTER TABLE "integration_secrets" ADD CONSTRAINT "integration_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_secrets_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "integration_events" table
ALTER TABLE "integration_events" ADD CONSTRAINT "integration_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_events_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_blocked_groups" table
ALTER TABLE "internal_policy_blocked_groups" ADD CONSTRAINT "internal_policy_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_blocked_groups_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_editors" table
ALTER TABLE "internal_policy_editors" ADD CONSTRAINT "internal_policy_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_editors_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_control_objectives" table
ALTER TABLE "internal_policy_control_objectives" ADD CONSTRAINT "internal_policy_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_control_objectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" ADD CONSTRAINT "internal_policy_tasks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "invite_events" table
ALTER TABLE "invite_events" ADD CONSTRAINT "invite_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_events_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_controls" table
ALTER TABLE "mapped_control_controls" ADD CONSTRAINT "mapped_control_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "mapped_control_subcontrols" table
ALTER TABLE "mapped_control_subcontrols" ADD CONSTRAINT "mapped_control_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" ADD CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "narrative_editors" table
ALTER TABLE "narrative_editors" ADD CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" ADD CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "org_membership_events" table
ALTER TABLE "org_membership_events" ADD CONSTRAINT "org_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_membership_events_org_membership_id" FOREIGN KEY ("org_membership_id") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- Modify "procedure_blocked_groups" table
ALTER TABLE "procedure_blocked_groups" ADD CONSTRAINT "procedure_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_blocked_groups_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "procedure_editors" table
ALTER TABLE "procedure_editors" ADD CONSTRAINT "procedure_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_editors_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- Modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" ADD CONSTRAINT "risk_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_blocked_groups_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_editors" table
ALTER TABLE "risk_editors" ADD CONSTRAINT "risk_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_editors_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_viewers" table
ALTER TABLE "risk_viewers" ADD CONSTRAINT "risk_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_viewers_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "risk_action_plans" table
ALTER TABLE "risk_action_plans" ADD CONSTRAINT "risk_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_action_plans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" ADD CONSTRAINT "subcontrol_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_objectives_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" ADD CONSTRAINT "subcontrol_tasks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subcontrol_control_implementations" table
ALTER TABLE "subcontrol_control_implementations" ADD CONSTRAINT "subcontrol_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_implementations_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "subscriber_events" table
ALTER TABLE "subscriber_events" ADD CONSTRAINT "subscriber_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subscriber_events_subscriber_id" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "task_evidence" table
ALTER TABLE "task_evidence" ADD CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "template_files" table
ALTER TABLE "template_files" ADD CONSTRAINT "template_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "template_files_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "user_files" table
ALTER TABLE "user_files" ADD CONSTRAINT "user_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_files_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "user_events" table
ALTER TABLE "user_events" ADD CONSTRAINT "user_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "user_action_plans" table
ALTER TABLE "user_action_plans" ADD CONSTRAINT "user_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_action_plans_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "user_setting_files" table
ALTER TABLE "user_setting_files" ADD CONSTRAINT "user_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_setting_files_user_setting_id" FOREIGN KEY ("user_setting_id") REFERENCES "user_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
