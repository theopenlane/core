-- +goose Up
-- create "api_tokens" table
CREATE TABLE "api_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "api_tokens_token_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_token_key" ON "api_tokens" ("token");
-- create index "apitoken_token" to table: "api_tokens"
CREATE INDEX "apitoken_token" ON "api_tokens" ("token");
-- create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "action_plan_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "due_date" timestamptz NULL, "priority" character varying NULL, "source" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "subcontrol_action_plans" character varying NULL, PRIMARY KEY ("id"));
-- create "action_plan_history" table
CREATE TABLE "action_plan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "action_plan_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "approver_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "due_date" timestamptz NULL, "priority" character varying NULL, "source" character varying NULL, PRIMARY KEY ("id"));
-- create index "actionplanhistory_history_time" to table: "action_plan_history"
CREATE INDEX "actionplanhistory_history_time" ON "action_plan_history" ("history_time");
-- create "contacts" table
CREATE TABLE "contacts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create "contact_history" table
CREATE TABLE "contact_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', PRIMARY KEY ("id"));
-- create index "contacthistory_history_time" to table: "contact_history"
CREATE INDEX "contacthistory_history_time" ON "contact_history" ("history_time");
-- create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "ref_code" character varying NOT NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "internal_policy_controls" character varying NULL, "owner_id" character varying NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "control_display_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_display_id_owner_id" ON "controls" ("display_id", "owner_id");
-- create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NULL));
-- create index "controls_auditor_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_auditor_reference_id_key" ON "controls" ("auditor_reference_id");
-- create index "controls_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_reference_id_key" ON "controls" ("reference_id");
-- create "control_history" table
CREATE TABLE "control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "ref_code" character varying NOT NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "controlhistory_history_time" to table: "control_history"
CREATE INDEX "controlhistory_history_time" ON "control_history" ("history_time");
-- create "control_implementations" table
CREATE TABLE "control_implementations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NULL DEFAULT 'DRAFT', "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create "control_implementation_history" table
CREATE TABLE "control_implementation_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "status" character varying NULL DEFAULT 'DRAFT', "implementation_date" timestamptz NULL, "verified" boolean NULL, "verification_date" timestamptz NULL, "details" text NULL, PRIMARY KEY ("id"));
-- create index "controlimplementationhistory_history_time" to table: "control_implementation_history"
CREATE INDEX "controlimplementationhistory_history_time" ON "control_implementation_history" ("history_time");
-- create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "desired_outcome" text NULL, "status" character varying NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "control_objective_type" character varying NULL, "category" character varying NULL, "subcategory" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "controlobjective_display_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_display_id_owner_id" ON "control_objectives" ("display_id", "owner_id");
-- create "control_objective_history" table
CREATE TABLE "control_objective_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "name" character varying NOT NULL, "desired_outcome" text NULL, "status" character varying NULL, "source" character varying NULL DEFAULT 'USER_DEFINED', "control_objective_type" character varying NULL, "category" character varying NULL, "subcategory" character varying NULL, PRIMARY KEY ("id"));
-- create index "controlobjectivehistory_history_time" to table: "control_objective_history"
CREATE INDEX "controlobjectivehistory_history_time" ON "control_objective_history" ("history_time");
-- create "document_data" table
CREATE TABLE "document_data" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "data" jsonb NOT NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create "document_data_history" table
CREATE TABLE "document_data_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, "data" jsonb NOT NULL, PRIMARY KEY ("id"));
-- create index "documentdatahistory_history_time" to table: "document_data_history"
CREATE INDEX "documentdatahistory_history_time" ON "document_data_history" ("history_time");
-- create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_token_key" ON "email_verification_tokens" ("token");
-- create index "emailverificationtoken_token" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_token" ON "email_verification_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "entities" table
CREATE TABLE "entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "status" character varying NULL DEFAULT 'active', "entity_type_id" character varying NULL, "entity_type_entities" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "entity_name_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "entity_history" table
CREATE TABLE "entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" citext NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "entity_type_id" character varying NULL, "status" character varying NULL DEFAULT 'active', PRIMARY KEY ("id"));
-- create index "entityhistory_history_time" to table: "entity_history"
CREATE INDEX "entityhistory_history_time" ON "entity_history" ("history_time");
-- create "entity_types" table
CREATE TABLE "entity_types" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "entitytype_name_owner_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_name_owner_id" ON "entity_types" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "entity_type_history" table
CREATE TABLE "entity_type_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" citext NOT NULL, PRIMARY KEY ("id"));
-- create index "entitytypehistory_history_time" to table: "entity_type_history"
CREATE INDEX "entitytypehistory_history_time" ON "entity_type_history" ("history_time");
-- create "events" table
CREATE TABLE "events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create "event_history" table
CREATE TABLE "event_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "eventhistory_history_time" to table: "event_history"
CREATE INDEX "eventhistory_history_time" ON "event_history" ("history_time");
-- create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "status" character varying NULL DEFAULT 'READY', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- create "evidence_history" table
CREATE TABLE "evidence_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "status" character varying NULL DEFAULT 'READY', PRIMARY KEY ("id"));
-- create index "evidencehistory_history_time" to table: "evidence_history"
CREATE INDEX "evidencehistory_history_time" ON "evidence_history" ("history_time");
-- create "files" table
CREATE TABLE "files" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, "note_files" character varying NULL, PRIMARY KEY ("id"));
-- create "file_history" table
CREATE TABLE "file_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, PRIMARY KEY ("id"));
-- create index "filehistory_history_time" to table: "file_history"
CREATE INDEX "filehistory_history_time" ON "file_history" ("history_time");
-- create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" citext NOT NULL, "description" character varying NULL, "is_managed" boolean NULL DEFAULT false, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', "organization_control_creators" character varying NULL, "organization_control_objective_creators" character varying NULL, "organization_group_creators" character varying NULL, "organization_internal_policy_creators" character varying NULL, "organization_narrative_creators" character varying NULL, "organization_procedure_creators" character varying NULL, "organization_program_creators" character varying NULL, "organization_risk_creators" character varying NULL, "organization_template_creators" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_display_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_display_id_owner_id" ON "groups" ("display_id", "owner_id");
-- create index "group_name_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_name_owner_id" ON "groups" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "group_history" table
CREATE TABLE "group_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" citext NOT NULL, "description" character varying NULL, "is_managed" boolean NULL DEFAULT false, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', PRIMARY KEY ("id"));
-- create index "grouphistory_history_time" to table: "group_history"
CREATE INDEX "grouphistory_history_time" ON "group_history" ("history_time");
-- create "group_memberships" table
CREATE TABLE "group_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, "group_membership_orgmembership" character varying NULL, PRIMARY KEY ("id"));
-- create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id") WHERE (deleted_at IS NULL);
-- create "group_membership_history" table
CREATE TABLE "group_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "groupmembershiphistory_history_time" to table: "group_membership_history"
CREATE INDEX "groupmembershiphistory_history_time" ON "group_membership_history" ("history_time");
-- create "group_settings" table
CREATE TABLE "group_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_settings_group_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_group_id_key" ON "group_settings" ("group_id");
-- create "group_setting_history" table
CREATE TABLE "group_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "groupsettinghistory_history_time" to table: "group_setting_history"
CREATE INDEX "groupsettinghistory_history_time" ON "group_setting_history" ("history_time");
-- create "hushes" table
CREATE TABLE "hushes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create "hush_history" table
CREATE TABLE "hush_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, PRIMARY KEY ("id"));
-- create index "hushhistory_history_time" to table: "hush_history"
CREATE INDEX "hushhistory_history_time" ON "hush_history" ("history_time");
-- create "integrations" table
CREATE TABLE "integrations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "group_integrations" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create "integration_history" table
CREATE TABLE "integration_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, PRIMARY KEY ("id"));
-- create index "integrationhistory_history_time" to table: "integration_history"
CREATE INDEX "integrationhistory_history_time" ON "integration_history" ("history_time");
-- create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "policy_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "control_internal_policies" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "subcontrol_internal_policies" character varying NULL, PRIMARY KEY ("id"));
-- create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_display_id_owner_id" ON "internal_policies" ("display_id", "owner_id");
-- create "internal_policy_history" table
CREATE TABLE "internal_policy_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "policy_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "approver_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
CREATE INDEX "internalpolicyhistory_history_time" ON "internal_policy_history" ("history_time");
-- create "invites" table
CREATE TABLE "invites" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "expires" timestamptz NULL, "recipient" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'INVITATION_SENT', "role" character varying NOT NULL DEFAULT 'MEMBER', "send_attempts" bigint NOT NULL DEFAULT 1, "requestor_id" character varying NULL, "secret" bytea NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "invite_recipient_owner_id" to table: "invites"
CREATE UNIQUE INDEX "invite_recipient_owner_id" ON "invites" ("recipient", "owner_id") WHERE (deleted_at IS NULL);
-- create index "invites_token_key" to table: "invites"
CREATE UNIQUE INDEX "invites_token_key" ON "invites" ("token");
-- create "mapped_controls" table
CREATE TABLE "mapped_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "mapping_type" character varying NULL, "relation" character varying NULL, PRIMARY KEY ("id"));
-- create "mapped_control_history" table
CREATE TABLE "mapped_control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "mapping_type" character varying NULL, "relation" character varying NULL, PRIMARY KEY ("id"));
-- create index "mappedcontrolhistory_history_time" to table: "mapped_control_history"
CREATE INDEX "mappedcontrolhistory_history_time" ON "mapped_control_history" ("history_time");
-- create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "details" text NULL, "control_objective_narratives" character varying NULL, "internal_policy_narratives" character varying NULL, "owner_id" character varying NULL, "procedure_narratives" character varying NULL, "subcontrol_narratives" character varying NULL, PRIMARY KEY ("id"));
-- create index "narrative_display_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_display_id_owner_id" ON "narratives" ("display_id", "owner_id");
-- create "narrative_history" table
CREATE TABLE "narrative_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "details" text NULL, PRIMARY KEY ("id"));
-- create index "narrativehistory_history_time" to table: "narrative_history"
CREATE INDEX "narrativehistory_history_time" ON "narrative_history" ("history_time");
-- create "notes" table
CREATE TABLE "notes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "text" text NOT NULL, "entity_notes" character varying NULL, "owner_id" character varying NULL, "program_notes" character varying NULL, "task_comments" character varying NULL, PRIMARY KEY ("id"));
-- create index "note_display_id_owner_id" to table: "notes"
CREATE UNIQUE INDEX "note_display_id_owner_id" ON "notes" ("display_id", "owner_id");
-- create "note_history" table
CREATE TABLE "note_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "text" text NOT NULL, PRIMARY KEY ("id"));
-- create index "notehistory_history_time" to table: "note_history"
CREATE INDEX "notehistory_history_time" ON "note_history" ("history_time");
-- create "onboardings" table
CREATE TABLE "onboardings" ("id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "company_name" character varying NOT NULL, "domains" jsonb NULL, "company_details" jsonb NULL, "user_details" jsonb NULL, "compliance" jsonb NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- create "org_memberships" table
CREATE TABLE "org_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id") WHERE (deleted_at IS NULL);
-- create "org_membership_history" table
CREATE TABLE "org_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "orgmembershiphistory_history_time" to table: "org_membership_history"
CREATE INDEX "orgmembershiphistory_history_time" ON "org_membership_history" ("history_time");
-- create "org_subscriptions" table
CREATE TABLE "org_subscriptions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "product_price" jsonb NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "trial_expires_at" timestamptz NULL, "days_until_due" character varying NULL, "payment_method_added" boolean NULL, "features" jsonb NULL, "feature_lookup_keys" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "org_subscriptions_stripe_customer_id_key" to table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_stripe_customer_id_key" ON "org_subscriptions" ("stripe_customer_id");
-- create "org_subscription_history" table
CREATE TABLE "org_subscription_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "product_price" jsonb NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "trial_expires_at" timestamptz NULL, "days_until_due" character varying NULL, "payment_method_added" boolean NULL, "features" jsonb NULL, "feature_lookup_keys" jsonb NULL, PRIMARY KEY ("id"));
-- create index "orgsubscriptionhistory_history_time" to table: "org_subscription_history"
CREATE INDEX "orgsubscriptionhistory_history_time" ON "org_subscription_history" ("history_time");
-- create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "dedicated_db" boolean NOT NULL DEFAULT false, "parent_organization_id" character varying NULL, "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "organization_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_name" ON "organizations" ("name") WHERE (deleted_at IS NULL);
-- create "organization_history" table
CREATE TABLE "organization_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" citext NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "parent_organization_id" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "avatar_local_file_id" character varying NULL, "avatar_updated_at" timestamptz NULL, "dedicated_db" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "organizationhistory_history_time" to table: "organization_history"
CREATE INDEX "organizationhistory_history_time" ON "organization_history" ("history_time");
-- create "organization_settings" table
CREATE TABLE "organization_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" jsonb NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "billing_notifications_enabled" boolean NOT NULL DEFAULT true, "allowed_email_domains" jsonb NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "organization_settings_organization_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_organization_id_key" ON "organization_settings" ("organization_id");
-- create "organization_setting_history" table
CREATE TABLE "organization_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" jsonb NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "organization_id" character varying NULL, "billing_notifications_enabled" boolean NOT NULL DEFAULT true, "allowed_email_domains" jsonb NULL, PRIMARY KEY ("id"));
-- create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
CREATE INDEX "organizationsettinghistory_history_time" ON "organization_setting_history" ("history_time");
-- create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
-- create index "passwordresettoken_token" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_token" ON "password_reset_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "personal_access_tokens" table
CREATE TABLE "personal_access_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_token_key" ON "personal_access_tokens" ("token");
-- create index "personalaccesstoken_token" to table: "personal_access_tokens"
CREATE INDEX "personalaccesstoken_token" ON "personal_access_tokens" ("token");
-- create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "procedure_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "control_objective_procedures" character varying NULL, "owner_id" character varying NULL, "approver_id" character varying NULL, "delegate_id" character varying NULL, "subcontrol_procedures" character varying NULL, PRIMARY KEY ("id"));
-- create index "procedure_display_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_display_id_owner_id" ON "procedures" ("display_id", "owner_id");
-- create "procedure_history" table
CREATE TABLE "procedure_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'DRAFT', "procedure_type" character varying NULL, "details" text NULL, "approval_required" boolean NULL DEFAULT true, "review_due" timestamptz NULL, "review_frequency" character varying NULL DEFAULT 'YEARLY', "approver_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "procedurehistory_history_time" to table: "procedure_history"
CREATE INDEX "procedurehistory_history_time" ON "procedure_history" ("history_time");
-- create "programs" table
CREATE TABLE "programs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "program_display_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_display_id_owner_id" ON "programs" ("display_id", "owner_id");
-- create "program_history" table
CREATE TABLE "program_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "programhistory_history_time" to table: "program_history"
CREATE INDEX "programhistory_history_time" ON "program_history" ("history_time");
-- create "program_memberships" table
CREATE TABLE "program_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, "program_membership_orgmembership" character varying NULL, PRIMARY KEY ("id"));
-- create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id") WHERE (deleted_at IS NULL);
-- create "program_membership_history" table
CREATE TABLE "program_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "programmembershiphistory_history_time" to table: "program_membership_history"
CREATE INDEX "programmembershiphistory_history_time" ON "program_membership_history" ("history_time");
-- create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'OPEN', "risk_type" character varying NULL, "category" character varying NULL, "impact" character varying NULL DEFAULT 'MODERATE', "likelihood" character varying NULL DEFAULT 'LIKELY', "score" bigint NULL, "mitigation" text NULL, "details" text NULL, "business_costs" text NULL, "control_objective_risks" character varying NULL, "owner_id" character varying NULL, "stakeholder_id" character varying NULL, "delegate_id" character varying NULL, "subcontrol_risks" character varying NULL, PRIMARY KEY ("id"));
-- create index "risk_display_id_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_display_id_owner_id" ON "risks" ("display_id", "owner_id");
-- create "risk_history" table
CREATE TABLE "risk_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "status" character varying NULL DEFAULT 'OPEN', "risk_type" character varying NULL, "category" character varying NULL, "impact" character varying NULL DEFAULT 'MODERATE', "likelihood" character varying NULL DEFAULT 'LIKELY', "score" bigint NULL, "mitigation" text NULL, "details" text NULL, "business_costs" text NULL, "stakeholder_id" character varying NULL, "delegate_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "riskhistory_history_time" to table: "risk_history"
CREATE INDEX "riskhistory_history_time" ON "risk_history" ("history_time");
-- create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "system_owned" boolean NULL DEFAULT false, "name" character varying NOT NULL, "short_name" character varying NULL, "framework" text NULL, "description" text NULL, "governing_body_logo_url" character varying NULL, "governing_body" character varying NULL, "domains" jsonb NULL, "link" character varying NULL, "status" character varying NULL DEFAULT 'ACTIVE', "is_public" boolean NULL DEFAULT false, "free_to_use" boolean NULL DEFAULT false, "standard_type" character varying NULL, "version" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create "standard_history" table
CREATE TABLE "standard_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "revision" character varying NULL DEFAULT 'v0.0.1', "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "name" character varying NOT NULL, "short_name" character varying NULL, "framework" text NULL, "description" text NULL, "governing_body_logo_url" character varying NULL, "governing_body" character varying NULL, "domains" jsonb NULL, "link" character varying NULL, "status" character varying NULL DEFAULT 'ACTIVE', "is_public" boolean NULL DEFAULT false, "free_to_use" boolean NULL DEFAULT false, "standard_type" character varying NULL, "version" character varying NULL, PRIMARY KEY ("id"));
-- create index "standardhistory_history_time" to table: "standard_history"
CREATE INDEX "standardhistory_history_time" ON "standard_history" ("history_time");
-- create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "ref_code" character varying NOT NULL, "control_id" character varying NOT NULL, "owner_id" character varying NULL, "program_subcontrols" character varying NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "user_subcontrols" character varying NULL, PRIMARY KEY ("id"));
-- create index "subcontrol_control_id_ref_code" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_control_id_ref_code" ON "subcontrols" ("control_id", "ref_code") WHERE (deleted_at IS NULL);
-- create index "subcontrol_display_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_display_id_owner_id" ON "subcontrols" ("display_id", "owner_id");
-- create "subcontrol_history" table
CREATE TABLE "subcontrol_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "description" text NULL, "reference_id" character varying NULL, "auditor_reference_id" character varying NULL, "status" character varying NULL DEFAULT 'NULL', "source" character varying NULL DEFAULT 'USER_DEFINED', "control_type" character varying NULL DEFAULT 'PREVENTATIVE', "category" character varying NULL, "category_id" character varying NULL, "subcategory" character varying NULL, "mapped_categories" jsonb NULL, "assessment_objectives" jsonb NULL, "assessment_methods" jsonb NULL, "control_questions" jsonb NULL, "implementation_guidance" jsonb NULL, "example_evidence" jsonb NULL, "references" jsonb NULL, "control_owner_id" character varying NULL, "delegate_id" character varying NULL, "owner_id" character varying NULL, "ref_code" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "subcontrolhistory_history_time" to table: "subcontrol_history"
CREATE INDEX "subcontrolhistory_history_time" ON "subcontrol_history" ("history_time");
-- create "subscribers" table
CREATE TABLE "subscribers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "phone_number" character varying NULL, "verified_email" boolean NOT NULL DEFAULT false, "verified_phone" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT false, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "secret" bytea NOT NULL, "unsubscribed" boolean NOT NULL DEFAULT false, "send_attempts" bigint NOT NULL DEFAULT 1, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false));
-- create index "subscribers_token_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_token_key" ON "subscribers" ("token");
-- create "tfa_settings" table
CREATE TABLE "tfa_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tfa_secret" character varying NULL, "verified" boolean NOT NULL DEFAULT false, "recovery_codes" jsonb NULL, "phone_otp_allowed" boolean NULL DEFAULT false, "email_otp_allowed" boolean NULL DEFAULT false, "totp_allowed" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "tfasetting_owner_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_owner_id" ON "tfa_settings" ("owner_id") WHERE (deleted_at IS NULL);
-- create "tasks" table
CREATE TABLE "tasks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" text NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "category" character varying NULL, "due" timestamptz NULL, "completed" timestamptz NULL, "owner_id" character varying NULL, "assigner_id" character varying NULL, "assignee_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "task_display_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_display_id_owner_id" ON "tasks" ("display_id", "owner_id");
-- create "task_history" table
CREATE TABLE "task_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" text NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "category" character varying NULL, "due" timestamptz NULL, "completed" timestamptz NULL, "assignee_id" character varying NULL, "assigner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "taskhistory_history_time" to table: "task_history"
CREATE INDEX "taskhistory_history_time" ON "task_history" ("history_time");
-- create "templates" table
CREATE TABLE "templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "template_name_owner_id_template_type" to table: "templates"
CREATE UNIQUE INDEX "template_name_owner_id_template_type" ON "templates" ("name", "owner_id", "template_type") WHERE (deleted_at IS NULL);
-- create "template_history" table
CREATE TABLE "template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "templatehistory_history_time" to table: "template_history"
CREATE INDEX "templatehistory_history_time" ON "template_history" ("history_time");
-- create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "user_email_auth_provider" to table: "users"
CREATE UNIQUE INDEX "user_email_auth_provider" ON "users" ("email", "auth_provider") WHERE (deleted_at IS NULL);
-- create index "user_id" to table: "users"
CREATE UNIQUE INDEX "user_id" ON "users" ("id");
-- create index "users_display_id_key" to table: "users"
CREATE UNIQUE INDEX "users_display_id_key" ON "users" ("display_id");
-- create index "users_sub_key" to table: "users"
CREATE UNIQUE INDEX "users_sub_key" ON "users" ("sub");
-- create "user_history" table
CREATE TABLE "user_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_local_file_id" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', PRIMARY KEY ("id"));
-- create index "userhistory_history_time" to table: "user_history"
CREATE INDEX "userhistory_history_time" ON "user_history" ("history_time");
-- create "user_settings" table
CREATE TABLE "user_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, "user_id" character varying NULL, "user_setting_default_org" character varying NULL, PRIMARY KEY ("id"));
-- create index "user_settings_user_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_user_id_key" ON "user_settings" ("user_id");
-- create "user_setting_history" table
CREATE TABLE "user_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "user_id" character varying NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, PRIMARY KEY ("id"));
-- create index "usersettinghistory_history_time" to table: "user_setting_history"
CREATE INDEX "usersettinghistory_history_time" ON "user_setting_history" ("history_time");
-- create "webauthns" table
CREATE TABLE "webauthns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "tags" jsonb NULL, "credential_id" bytea NULL, "public_key" bytea NULL, "attestation_type" character varying NULL, "aaguid" bytea NOT NULL, "sign_count" integer NOT NULL, "transports" jsonb NOT NULL, "backup_eligible" boolean NOT NULL DEFAULT false, "backup_state" boolean NOT NULL DEFAULT false, "user_present" boolean NOT NULL DEFAULT false, "user_verified" boolean NOT NULL DEFAULT false, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "webauthns_aaguid_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_aaguid_key" ON "webauthns" ("aaguid");
-- create index "webauthns_credential_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_credential_id_key" ON "webauthns" ("credential_id");
-- create "contact_files" table
CREATE TABLE "contact_files" ("contact_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "file_id"));
-- create "control_control_objectives" table
CREATE TABLE "control_control_objectives" ("control_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_objective_id"));
-- create "control_tasks" table
CREATE TABLE "control_tasks" ("control_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_id", "task_id"));
-- create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- create "control_action_plans" table
CREATE TABLE "control_action_plans" ("control_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "action_plan_id"));
-- create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create "control_viewers" table
CREATE TABLE "control_viewers" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create "control_control_implementations" table
CREATE TABLE "control_control_implementations" ("control_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("control_id", "control_implementation_id"));
-- create "control_objective_blocked_groups" table
CREATE TABLE "control_objective_blocked_groups" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create "control_objective_editors" table
CREATE TABLE "control_objective_editors" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create "control_objective_viewers" table
CREATE TABLE "control_objective_viewers" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create "control_objective_tasks" table
CREATE TABLE "control_objective_tasks" ("control_objective_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "task_id"));
-- create "document_data_files" table
CREATE TABLE "document_data_files" ("document_data_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("document_data_id", "file_id"));
-- create "entity_contacts" table
CREATE TABLE "entity_contacts" ("entity_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "contact_id"));
-- create "entity_documents" table
CREATE TABLE "entity_documents" ("entity_id" character varying NOT NULL, "document_data_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "document_data_id"));
-- create "entity_files" table
CREATE TABLE "entity_files" ("entity_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "file_id"));
-- create "evidence_control_objectives" table
CREATE TABLE "evidence_control_objectives" ("evidence_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_objective_id"));
-- create "evidence_controls" table
CREATE TABLE "evidence_controls" ("evidence_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_id"));
-- create "evidence_subcontrols" table
CREATE TABLE "evidence_subcontrols" ("evidence_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "subcontrol_id"));
-- create "evidence_files" table
CREATE TABLE "evidence_files" ("evidence_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "file_id"));
-- create "file_events" table
CREATE TABLE "file_events" ("file_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("file_id", "event_id"));
-- create "group_events" table
CREATE TABLE "group_events" ("group_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_id", "event_id"));
-- create "group_files" table
CREATE TABLE "group_files" ("group_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("group_id", "file_id"));
-- create "group_tasks" table
CREATE TABLE "group_tasks" ("group_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("group_id", "task_id"));
-- create "group_membership_events" table
CREATE TABLE "group_membership_events" ("group_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_membership_id", "event_id"));
-- create "hush_events" table
CREATE TABLE "hush_events" ("hush_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("hush_id", "event_id"));
-- create "integration_secrets" table
CREATE TABLE "integration_secrets" ("integration_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "hush_id"));
-- create "integration_events" table
CREATE TABLE "integration_events" ("integration_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "event_id"));
-- create "internal_policy_blocked_groups" table
CREATE TABLE "internal_policy_blocked_groups" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- create "internal_policy_editors" table
CREATE TABLE "internal_policy_editors" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"));
-- create "internal_policy_control_objectives" table
CREATE TABLE "internal_policy_control_objectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- create "internal_policy_tasks" table
CREATE TABLE "internal_policy_tasks" ("internal_policy_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "task_id"));
-- create "invite_events" table
CREATE TABLE "invite_events" ("invite_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "event_id"));
-- create "mapped_control_controls" table
CREATE TABLE "mapped_control_controls" ("mapped_control_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "control_id"));
-- create "mapped_control_subcontrols" table
CREATE TABLE "mapped_control_subcontrols" ("mapped_control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("mapped_control_id", "subcontrol_id"));
-- create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create "org_membership_events" table
CREATE TABLE "org_membership_events" ("org_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_membership_id", "event_id"));
-- create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"));
-- create "organization_personal_access_tokens" table
CREATE TABLE "organization_personal_access_tokens" ("organization_id" character varying NOT NULL, "personal_access_token_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "personal_access_token_id"));
-- create "organization_files" table
CREATE TABLE "organization_files" ("organization_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "file_id"));
-- create "organization_events" table
CREATE TABLE "organization_events" ("organization_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "event_id"));
-- create "organization_setting_files" table
CREATE TABLE "organization_setting_files" ("organization_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_setting_id", "file_id"));
-- create "personal_access_token_events" table
CREATE TABLE "personal_access_token_events" ("personal_access_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("personal_access_token_id", "event_id"));
-- create "procedure_blocked_groups" table
CREATE TABLE "procedure_blocked_groups" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- create "procedure_editors" table
CREATE TABLE "procedure_editors" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- create "procedure_risks" table
CREATE TABLE "procedure_risks" ("procedure_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "risk_id"));
-- create "procedure_tasks" table
CREATE TABLE "procedure_tasks" ("procedure_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "task_id"));
-- create "program_blocked_groups" table
CREATE TABLE "program_blocked_groups" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- create "program_editors" table
CREATE TABLE "program_editors" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- create "program_viewers" table
CREATE TABLE "program_viewers" ("program_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("program_id", "group_id"));
-- create "program_controls" table
CREATE TABLE "program_controls" ("program_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_id"));
-- create "program_control_objectives" table
CREATE TABLE "program_control_objectives" ("program_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_objective_id"));
-- create "program_internal_policies" table
CREATE TABLE "program_internal_policies" ("program_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("program_id", "internal_policy_id"));
-- create "program_procedures" table
CREATE TABLE "program_procedures" ("program_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("program_id", "procedure_id"));
-- create "program_risks" table
CREATE TABLE "program_risks" ("program_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("program_id", "risk_id"));
-- create "program_tasks" table
CREATE TABLE "program_tasks" ("program_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("program_id", "task_id"));
-- create "program_files" table
CREATE TABLE "program_files" ("program_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("program_id", "file_id"));
-- create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"));
-- create "program_narratives" table
CREATE TABLE "program_narratives" ("program_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("program_id", "narrative_id"));
-- create "program_action_plans" table
CREATE TABLE "program_action_plans" ("program_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("program_id", "action_plan_id"));
-- create "risk_blocked_groups" table
CREATE TABLE "risk_blocked_groups" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create "risk_editors" table
CREATE TABLE "risk_editors" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create "risk_viewers" table
CREATE TABLE "risk_viewers" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create "risk_action_plans" table
CREATE TABLE "risk_action_plans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- create "subcontrol_control_objectives" table
CREATE TABLE "subcontrol_control_objectives" ("subcontrol_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_objective_id"));
-- create "subcontrol_tasks" table
CREATE TABLE "subcontrol_tasks" ("subcontrol_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "task_id"));
-- create "subcontrol_control_implementations" table
CREATE TABLE "subcontrol_control_implementations" ("subcontrol_id" character varying NOT NULL, "control_implementation_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "control_implementation_id"));
-- create "subscriber_events" table
CREATE TABLE "subscriber_events" ("subscriber_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("subscriber_id", "event_id"));
-- create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"));
-- create "template_files" table
CREATE TABLE "template_files" ("template_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("template_id", "file_id"));
-- create "user_files" table
CREATE TABLE "user_files" ("user_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_id", "file_id"));
-- create "user_events" table
CREATE TABLE "user_events" ("user_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("user_id", "event_id"));
-- create "user_action_plans" table
CREATE TABLE "user_action_plans" ("user_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("user_id", "action_plan_id"));
-- create "user_setting_files" table
CREATE TABLE "user_setting_files" ("user_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_setting_id", "file_id"));
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD CONSTRAINT "api_tokens_organizations_api_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD CONSTRAINT "action_plans_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_organizations_action_plans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_subcontrols_action_plans" FOREIGN KEY ("subcontrol_action_plans") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ADD CONSTRAINT "contacts_organizations_contacts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_internal_policies_controls" FOREIGN KEY ("internal_policy_controls") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_standards_controls" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD CONSTRAINT "control_implementations_organizations_control_implementations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_organizations_control_objectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "document_data" table
ALTER TABLE "document_data" ADD CONSTRAINT "document_data_organizations_documents" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD CONSTRAINT "email_verification_tokens_users_email_verification_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_organizations_entities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" ADD CONSTRAINT "entity_types_organizations_entity_types" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "evidences" table
ALTER TABLE "evidences" ADD CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_notes_files" FOREIGN KEY ("note_files") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_organizations_control_creators" FOREIGN KEY ("organization_control_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_control_objective_creators" FOREIGN KEY ("organization_control_objective_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_group_creators" FOREIGN KEY ("organization_group_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_internal_policy_creators" FOREIGN KEY ("organization_internal_policy_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_narrative_creators" FOREIGN KEY ("organization_narrative_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_procedure_creators" FOREIGN KEY ("organization_procedure_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_program_creators" FOREIGN KEY ("organization_program_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_risk_creators" FOREIGN KEY ("organization_risk_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_template_creators" FOREIGN KEY ("organization_template_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "group_memberships" table
ALTER TABLE "group_memberships" ADD CONSTRAINT "group_memberships_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "group_memberships_org_memberships_orgmembership" FOREIGN KEY ("group_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "group_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "group_settings" table
ALTER TABLE "group_settings" ADD CONSTRAINT "group_settings_groups_setting" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "hushes" table
ALTER TABLE "hushes" ADD CONSTRAINT "hushes_organizations_secrets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_groups_integrations" FOREIGN KEY ("group_integrations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_organizations_integrations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_controls_internal_policies" FOREIGN KEY ("control_internal_policies") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_organizations_internal_policies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_subcontrols_internal_policies" FOREIGN KEY ("subcontrol_internal_policies") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "invites" table
ALTER TABLE "invites" ADD CONSTRAINT "invites_organizations_invites" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD CONSTRAINT "narratives_control_objectives_narratives" FOREIGN KEY ("control_objective_narratives") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_internal_policies_narratives" FOREIGN KEY ("internal_policy_narratives") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_procedures_narratives" FOREIGN KEY ("procedure_narratives") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_subcontrols_narratives" FOREIGN KEY ("subcontrol_narratives") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD CONSTRAINT "notes_entities_notes" FOREIGN KEY ("entity_notes") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_organizations_notes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_programs_notes" FOREIGN KEY ("program_notes") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_tasks_comments" FOREIGN KEY ("task_comments") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "onboardings" table
ALTER TABLE "onboardings" ADD CONSTRAINT "onboardings_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_memberships" table
ALTER TABLE "org_memberships" ADD CONSTRAINT "org_memberships_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "org_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
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
-- modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_approver" FOREIGN KEY ("approver_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_organizations_procedures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_subcontrols_procedures" FOREIGN KEY ("subcontrol_procedures") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD CONSTRAINT "programs_organizations_programs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" ADD CONSTRAINT "program_memberships_org_memberships_orgmembership" FOREIGN KEY ("program_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "program_memberships_programs_program" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "program_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_groups_stakeholder" FOREIGN KEY ("stakeholder_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_subcontrols_risks" FOREIGN KEY ("subcontrol_risks") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD CONSTRAINT "standards_organizations_standards" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_controls_subcontrols" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "subcontrols_groups_control_owner" FOREIGN KEY ("control_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_groups_delegate" FOREIGN KEY ("delegate_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_programs_subcontrols" FOREIGN KEY ("program_subcontrols") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_subcontrols" FOREIGN KEY ("user_subcontrols") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD CONSTRAINT "subscribers_organizations_subscribers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD CONSTRAINT "tfa_settings_users_tfa_settings" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_organizations_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "user_settings" table
ALTER TABLE "user_settings" ADD CONSTRAINT "user_settings_organizations_default_org" FOREIGN KEY ("user_setting_default_org") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_settings_users_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "webauthns" table
ALTER TABLE "webauthns" ADD CONSTRAINT "webauthns_users_webauthn" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
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
-- modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" ADD CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_editors" table
ALTER TABLE "control_editors" ADD CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_viewers" table
ALTER TABLE "control_viewers" ADD CONSTRAINT "control_viewers_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_control_implementations" table
ALTER TABLE "control_control_implementations" ADD CONSTRAINT "control_control_implementations_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "entity_contacts" table
ALTER TABLE "entity_contacts" ADD CONSTRAINT "entity_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_contacts_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_documents" table
ALTER TABLE "entity_documents" ADD CONSTRAINT "entity_documents_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_documents_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_files" table
ALTER TABLE "entity_files" ADD CONSTRAINT "entity_files_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_control_objectives" table
ALTER TABLE "evidence_control_objectives" ADD CONSTRAINT "evidence_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_control_objectives_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_controls" table
ALTER TABLE "evidence_controls" ADD CONSTRAINT "evidence_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_controls_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_subcontrols" table
ALTER TABLE "evidence_subcontrols" ADD CONSTRAINT "evidence_subcontrols_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "evidence_files" table
ALTER TABLE "evidence_files" ADD CONSTRAINT "evidence_files_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "evidence_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "file_events" table
ALTER TABLE "file_events" ADD CONSTRAINT "file_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "file_events_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "integration_secrets" table
ALTER TABLE "integration_secrets" ADD CONSTRAINT "integration_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_secrets_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_events" table
ALTER TABLE "integration_events" ADD CONSTRAINT "integration_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_events_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_blocked_groups" table
ALTER TABLE "internal_policy_blocked_groups" ADD CONSTRAINT "internal_policy_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_blocked_groups_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_editors" table
ALTER TABLE "internal_policy_editors" ADD CONSTRAINT "internal_policy_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_editors_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_control_objectives" table
ALTER TABLE "internal_policy_control_objectives" ADD CONSTRAINT "internal_policy_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_control_objectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" ADD CONSTRAINT "internal_policy_tasks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "invite_events" table
ALTER TABLE "invite_events" ADD CONSTRAINT "invite_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_events_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_controls" table
ALTER TABLE "mapped_control_controls" ADD CONSTRAINT "mapped_control_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_controls_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "mapped_control_subcontrols" table
ALTER TABLE "mapped_control_subcontrols" ADD CONSTRAINT "mapped_control_subcontrols_mapped_control_id" FOREIGN KEY ("mapped_control_id") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "mapped_control_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" ADD CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_editors" table
ALTER TABLE "narrative_editors" ADD CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" ADD CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_membership_events" table
ALTER TABLE "org_membership_events" ADD CONSTRAINT "org_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_membership_events_org_membership_id" FOREIGN KEY ("org_membership_id") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "procedure_blocked_groups" table
ALTER TABLE "procedure_blocked_groups" ADD CONSTRAINT "procedure_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_blocked_groups_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "procedure_editors" table
ALTER TABLE "procedure_editors" ADD CONSTRAINT "procedure_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "procedure_editors_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" ADD CONSTRAINT "risk_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_blocked_groups_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_editors" table
ALTER TABLE "risk_editors" ADD CONSTRAINT "risk_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_editors_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_viewers" table
ALTER TABLE "risk_viewers" ADD CONSTRAINT "risk_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_viewers_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_action_plans" table
ALTER TABLE "risk_action_plans" ADD CONSTRAINT "risk_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_action_plans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" ADD CONSTRAINT "subcontrol_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_objectives_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" ADD CONSTRAINT "subcontrol_tasks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_control_implementations" table
ALTER TABLE "subcontrol_control_implementations" ADD CONSTRAINT "subcontrol_control_implementations_control_implementation_id" FOREIGN KEY ("control_implementation_id") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_control_implementations_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subscriber_events" table
ALTER TABLE "subscriber_events" ADD CONSTRAINT "subscriber_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subscriber_events_subscriber_id" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "task_evidence" table
ALTER TABLE "task_evidence" ADD CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "template_files" table
ALTER TABLE "template_files" ADD CONSTRAINT "template_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "template_files_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_files" table
ALTER TABLE "user_files" ADD CONSTRAINT "user_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_files_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_events" table
ALTER TABLE "user_events" ADD CONSTRAINT "user_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_action_plans" table
ALTER TABLE "user_action_plans" ADD CONSTRAINT "user_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_action_plans_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_setting_files" table
ALTER TABLE "user_setting_files" ADD CONSTRAINT "user_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_setting_files_user_setting_id" FOREIGN KEY ("user_setting_id") REFERENCES "user_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "user_setting_files" table
ALTER TABLE "user_setting_files" DROP CONSTRAINT "user_setting_files_user_setting_id", DROP CONSTRAINT "user_setting_files_file_id";
-- reverse: modify "user_action_plans" table
ALTER TABLE "user_action_plans" DROP CONSTRAINT "user_action_plans_user_id", DROP CONSTRAINT "user_action_plans_action_plan_id";
-- reverse: modify "user_events" table
ALTER TABLE "user_events" DROP CONSTRAINT "user_events_user_id", DROP CONSTRAINT "user_events_event_id";
-- reverse: modify "user_files" table
ALTER TABLE "user_files" DROP CONSTRAINT "user_files_user_id", DROP CONSTRAINT "user_files_file_id";
-- reverse: modify "template_files" table
ALTER TABLE "template_files" DROP CONSTRAINT "template_files_template_id", DROP CONSTRAINT "template_files_file_id";
-- reverse: modify "task_evidence" table
ALTER TABLE "task_evidence" DROP CONSTRAINT "task_evidence_task_id", DROP CONSTRAINT "task_evidence_evidence_id";
-- reverse: modify "subscriber_events" table
ALTER TABLE "subscriber_events" DROP CONSTRAINT "subscriber_events_subscriber_id", DROP CONSTRAINT "subscriber_events_event_id";
-- reverse: modify "subcontrol_control_implementations" table
ALTER TABLE "subcontrol_control_implementations" DROP CONSTRAINT "subcontrol_control_implementations_subcontrol_id", DROP CONSTRAINT "subcontrol_control_implementations_control_implementation_id";
-- reverse: modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" DROP CONSTRAINT "subcontrol_tasks_task_id", DROP CONSTRAINT "subcontrol_tasks_subcontrol_id";
-- reverse: modify "subcontrol_control_objectives" table
ALTER TABLE "subcontrol_control_objectives" DROP CONSTRAINT "subcontrol_control_objectives_subcontrol_id", DROP CONSTRAINT "subcontrol_control_objectives_control_objective_id";
-- reverse: modify "risk_action_plans" table
ALTER TABLE "risk_action_plans" DROP CONSTRAINT "risk_action_plans_risk_id", DROP CONSTRAINT "risk_action_plans_action_plan_id";
-- reverse: modify "risk_viewers" table
ALTER TABLE "risk_viewers" DROP CONSTRAINT "risk_viewers_risk_id", DROP CONSTRAINT "risk_viewers_group_id";
-- reverse: modify "risk_editors" table
ALTER TABLE "risk_editors" DROP CONSTRAINT "risk_editors_risk_id", DROP CONSTRAINT "risk_editors_group_id";
-- reverse: modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" DROP CONSTRAINT "risk_blocked_groups_risk_id", DROP CONSTRAINT "risk_blocked_groups_group_id";
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
-- reverse: modify "procedure_editors" table
ALTER TABLE "procedure_editors" DROP CONSTRAINT "procedure_editors_procedure_id", DROP CONSTRAINT "procedure_editors_group_id";
-- reverse: modify "procedure_blocked_groups" table
ALTER TABLE "procedure_blocked_groups" DROP CONSTRAINT "procedure_blocked_groups_procedure_id", DROP CONSTRAINT "procedure_blocked_groups_group_id";
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
-- reverse: modify "org_membership_events" table
ALTER TABLE "org_membership_events" DROP CONSTRAINT "org_membership_events_org_membership_id", DROP CONSTRAINT "org_membership_events_event_id";
-- reverse: modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" DROP CONSTRAINT "narrative_viewers_narrative_id", DROP CONSTRAINT "narrative_viewers_group_id";
-- reverse: modify "narrative_editors" table
ALTER TABLE "narrative_editors" DROP CONSTRAINT "narrative_editors_narrative_id", DROP CONSTRAINT "narrative_editors_group_id";
-- reverse: modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" DROP CONSTRAINT "narrative_blocked_groups_narrative_id", DROP CONSTRAINT "narrative_blocked_groups_group_id";
-- reverse: modify "mapped_control_subcontrols" table
ALTER TABLE "mapped_control_subcontrols" DROP CONSTRAINT "mapped_control_subcontrols_subcontrol_id", DROP CONSTRAINT "mapped_control_subcontrols_mapped_control_id";
-- reverse: modify "mapped_control_controls" table
ALTER TABLE "mapped_control_controls" DROP CONSTRAINT "mapped_control_controls_mapped_control_id", DROP CONSTRAINT "mapped_control_controls_control_id";
-- reverse: modify "invite_events" table
ALTER TABLE "invite_events" DROP CONSTRAINT "invite_events_invite_id", DROP CONSTRAINT "invite_events_event_id";
-- reverse: modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" DROP CONSTRAINT "internal_policy_tasks_task_id", DROP CONSTRAINT "internal_policy_tasks_internal_policy_id";
-- reverse: modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" DROP CONSTRAINT "internal_policy_procedures_procedure_id", DROP CONSTRAINT "internal_policy_procedures_internal_policy_id";
-- reverse: modify "internal_policy_control_objectives" table
ALTER TABLE "internal_policy_control_objectives" DROP CONSTRAINT "internal_policy_control_objectives_internal_policy_id", DROP CONSTRAINT "internal_policy_control_objectives_control_objective_id";
-- reverse: modify "internal_policy_editors" table
ALTER TABLE "internal_policy_editors" DROP CONSTRAINT "internal_policy_editors_internal_policy_id", DROP CONSTRAINT "internal_policy_editors_group_id";
-- reverse: modify "internal_policy_blocked_groups" table
ALTER TABLE "internal_policy_blocked_groups" DROP CONSTRAINT "internal_policy_blocked_groups_internal_policy_id", DROP CONSTRAINT "internal_policy_blocked_groups_group_id";
-- reverse: modify "integration_events" table
ALTER TABLE "integration_events" DROP CONSTRAINT "integration_events_integration_id", DROP CONSTRAINT "integration_events_event_id";
-- reverse: modify "integration_secrets" table
ALTER TABLE "integration_secrets" DROP CONSTRAINT "integration_secrets_integration_id", DROP CONSTRAINT "integration_secrets_hush_id";
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
-- reverse: modify "file_events" table
ALTER TABLE "file_events" DROP CONSTRAINT "file_events_file_id", DROP CONSTRAINT "file_events_event_id";
-- reverse: modify "evidence_files" table
ALTER TABLE "evidence_files" DROP CONSTRAINT "evidence_files_file_id", DROP CONSTRAINT "evidence_files_evidence_id";
-- reverse: modify "evidence_subcontrols" table
ALTER TABLE "evidence_subcontrols" DROP CONSTRAINT "evidence_subcontrols_subcontrol_id", DROP CONSTRAINT "evidence_subcontrols_evidence_id";
-- reverse: modify "evidence_controls" table
ALTER TABLE "evidence_controls" DROP CONSTRAINT "evidence_controls_evidence_id", DROP CONSTRAINT "evidence_controls_control_id";
-- reverse: modify "evidence_control_objectives" table
ALTER TABLE "evidence_control_objectives" DROP CONSTRAINT "evidence_control_objectives_evidence_id", DROP CONSTRAINT "evidence_control_objectives_control_objective_id";
-- reverse: modify "entity_files" table
ALTER TABLE "entity_files" DROP CONSTRAINT "entity_files_file_id", DROP CONSTRAINT "entity_files_entity_id";
-- reverse: modify "entity_documents" table
ALTER TABLE "entity_documents" DROP CONSTRAINT "entity_documents_entity_id", DROP CONSTRAINT "entity_documents_document_data_id";
-- reverse: modify "entity_contacts" table
ALTER TABLE "entity_contacts" DROP CONSTRAINT "entity_contacts_entity_id", DROP CONSTRAINT "entity_contacts_contact_id";
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
-- reverse: modify "control_control_implementations" table
ALTER TABLE "control_control_implementations" DROP CONSTRAINT "control_control_implementations_control_implementation_id", DROP CONSTRAINT "control_control_implementations_control_id";
-- reverse: modify "control_viewers" table
ALTER TABLE "control_viewers" DROP CONSTRAINT "control_viewers_group_id", DROP CONSTRAINT "control_viewers_control_id";
-- reverse: modify "control_editors" table
ALTER TABLE "control_editors" DROP CONSTRAINT "control_editors_group_id", DROP CONSTRAINT "control_editors_control_id";
-- reverse: modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" DROP CONSTRAINT "control_blocked_groups_group_id", DROP CONSTRAINT "control_blocked_groups_control_id";
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
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_webauthn";
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP CONSTRAINT "user_settings_users_setting", DROP CONSTRAINT "user_settings_organizations_default_org";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_avatar_file";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP CONSTRAINT "templates_organizations_templates";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_assigner_tasks", DROP CONSTRAINT "tasks_users_assignee_tasks", DROP CONSTRAINT "tasks_organizations_tasks";
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP CONSTRAINT "tfa_settings_users_tfa_settings";
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP CONSTRAINT "subscribers_organizations_subscribers";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_users_subcontrols", DROP CONSTRAINT "subcontrols_programs_subcontrols", DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP CONSTRAINT "subcontrols_groups_delegate", DROP CONSTRAINT "subcontrols_groups_control_owner", DROP CONSTRAINT "subcontrols_controls_subcontrols";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP CONSTRAINT "standards_organizations_standards";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_subcontrols_risks", DROP CONSTRAINT "risks_organizations_risks", DROP CONSTRAINT "risks_groups_stakeholder", DROP CONSTRAINT "risks_groups_delegate", DROP CONSTRAINT "risks_control_objectives_risks";
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP CONSTRAINT "program_memberships_users_user", DROP CONSTRAINT "program_memberships_programs_program", DROP CONSTRAINT "program_memberships_org_memberships_orgmembership";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_organizations_programs";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_subcontrols_procedures", DROP CONSTRAINT "procedures_organizations_procedures", DROP CONSTRAINT "procedures_groups_delegate", DROP CONSTRAINT "procedures_groups_approver", DROP CONSTRAINT "procedures_control_objectives_procedures";
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
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP CONSTRAINT "org_memberships_users_user", DROP CONSTRAINT "org_memberships_organizations_organization";
-- reverse: modify "onboardings" table
ALTER TABLE "onboardings" DROP CONSTRAINT "onboardings_organizations_organization";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_tasks_comments", DROP CONSTRAINT "notes_programs_notes", DROP CONSTRAINT "notes_organizations_notes", DROP CONSTRAINT "notes_entities_notes";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_subcontrols_narratives", DROP CONSTRAINT "narratives_procedures_narratives", DROP CONSTRAINT "narratives_organizations_narratives", DROP CONSTRAINT "narratives_internal_policies_narratives", DROP CONSTRAINT "narratives_control_objectives_narratives";
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP CONSTRAINT "invites_organizations_invites";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_subcontrols_internal_policies", DROP CONSTRAINT "internal_policies_organizations_internal_policies", DROP CONSTRAINT "internal_policies_groups_delegate", DROP CONSTRAINT "internal_policies_groups_approver", DROP CONSTRAINT "internal_policies_controls_internal_policies";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_organizations_integrations", DROP CONSTRAINT "integrations_groups_integrations";
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP CONSTRAINT "hushes_organizations_secrets";
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" DROP CONSTRAINT "group_settings_groups_setting";
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP CONSTRAINT "group_memberships_users_user", DROP CONSTRAINT "group_memberships_org_memberships_orgmembership", DROP CONSTRAINT "group_memberships_groups_group";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_organizations_template_creators", DROP CONSTRAINT "groups_organizations_risk_creators", DROP CONSTRAINT "groups_organizations_program_creators", DROP CONSTRAINT "groups_organizations_procedure_creators", DROP CONSTRAINT "groups_organizations_narrative_creators", DROP CONSTRAINT "groups_organizations_internal_policy_creators", DROP CONSTRAINT "groups_organizations_groups", DROP CONSTRAINT "groups_organizations_group_creators", DROP CONSTRAINT "groups_organizations_control_objective_creators", DROP CONSTRAINT "groups_organizations_control_creators";
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_notes_files";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP CONSTRAINT "evidences_organizations_evidence";
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP CONSTRAINT "entity_types_organizations_entity_types";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_organizations_entities", DROP CONSTRAINT "entities_entity_types_entity_type", DROP CONSTRAINT "entities_entity_types_entities";
-- reverse: modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP CONSTRAINT "email_verification_tokens_users_email_verification_tokens";
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_templates_documents", DROP CONSTRAINT "document_data_organizations_documents";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_organizations_control_objectives";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP CONSTRAINT "control_implementations_organizations_control_implementations";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_standards_controls", DROP CONSTRAINT "controls_organizations_controls", DROP CONSTRAINT "controls_internal_policies_controls", DROP CONSTRAINT "controls_groups_delegate", DROP CONSTRAINT "controls_groups_control_owner";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP CONSTRAINT "contacts_organizations_contacts";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_subcontrols_action_plans", DROP CONSTRAINT "action_plans_organizations_action_plans", DROP CONSTRAINT "action_plans_groups_delegate", DROP CONSTRAINT "action_plans_groups_approver";
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP CONSTRAINT "api_tokens_organizations_api_tokens";
-- reverse: create "user_setting_files" table
DROP TABLE "user_setting_files";
-- reverse: create "user_action_plans" table
DROP TABLE "user_action_plans";
-- reverse: create "user_events" table
DROP TABLE "user_events";
-- reverse: create "user_files" table
DROP TABLE "user_files";
-- reverse: create "template_files" table
DROP TABLE "template_files";
-- reverse: create "task_evidence" table
DROP TABLE "task_evidence";
-- reverse: create "subscriber_events" table
DROP TABLE "subscriber_events";
-- reverse: create "subcontrol_control_implementations" table
DROP TABLE "subcontrol_control_implementations";
-- reverse: create "subcontrol_tasks" table
DROP TABLE "subcontrol_tasks";
-- reverse: create "subcontrol_control_objectives" table
DROP TABLE "subcontrol_control_objectives";
-- reverse: create "risk_action_plans" table
DROP TABLE "risk_action_plans";
-- reverse: create "risk_viewers" table
DROP TABLE "risk_viewers";
-- reverse: create "risk_editors" table
DROP TABLE "risk_editors";
-- reverse: create "risk_blocked_groups" table
DROP TABLE "risk_blocked_groups";
-- reverse: create "program_action_plans" table
DROP TABLE "program_action_plans";
-- reverse: create "program_narratives" table
DROP TABLE "program_narratives";
-- reverse: create "program_evidence" table
DROP TABLE "program_evidence";
-- reverse: create "program_files" table
DROP TABLE "program_files";
-- reverse: create "program_tasks" table
DROP TABLE "program_tasks";
-- reverse: create "program_risks" table
DROP TABLE "program_risks";
-- reverse: create "program_procedures" table
DROP TABLE "program_procedures";
-- reverse: create "program_internal_policies" table
DROP TABLE "program_internal_policies";
-- reverse: create "program_control_objectives" table
DROP TABLE "program_control_objectives";
-- reverse: create "program_controls" table
DROP TABLE "program_controls";
-- reverse: create "program_viewers" table
DROP TABLE "program_viewers";
-- reverse: create "program_editors" table
DROP TABLE "program_editors";
-- reverse: create "program_blocked_groups" table
DROP TABLE "program_blocked_groups";
-- reverse: create "procedure_tasks" table
DROP TABLE "procedure_tasks";
-- reverse: create "procedure_risks" table
DROP TABLE "procedure_risks";
-- reverse: create "procedure_editors" table
DROP TABLE "procedure_editors";
-- reverse: create "procedure_blocked_groups" table
DROP TABLE "procedure_blocked_groups";
-- reverse: create "personal_access_token_events" table
DROP TABLE "personal_access_token_events";
-- reverse: create "organization_setting_files" table
DROP TABLE "organization_setting_files";
-- reverse: create "organization_events" table
DROP TABLE "organization_events";
-- reverse: create "organization_files" table
DROP TABLE "organization_files";
-- reverse: create "organization_personal_access_tokens" table
DROP TABLE "organization_personal_access_tokens";
-- reverse: create "org_subscription_events" table
DROP TABLE "org_subscription_events";
-- reverse: create "org_membership_events" table
DROP TABLE "org_membership_events";
-- reverse: create "narrative_viewers" table
DROP TABLE "narrative_viewers";
-- reverse: create "narrative_editors" table
DROP TABLE "narrative_editors";
-- reverse: create "narrative_blocked_groups" table
DROP TABLE "narrative_blocked_groups";
-- reverse: create "mapped_control_subcontrols" table
DROP TABLE "mapped_control_subcontrols";
-- reverse: create "mapped_control_controls" table
DROP TABLE "mapped_control_controls";
-- reverse: create "invite_events" table
DROP TABLE "invite_events";
-- reverse: create "internal_policy_tasks" table
DROP TABLE "internal_policy_tasks";
-- reverse: create "internal_policy_procedures" table
DROP TABLE "internal_policy_procedures";
-- reverse: create "internal_policy_control_objectives" table
DROP TABLE "internal_policy_control_objectives";
-- reverse: create "internal_policy_editors" table
DROP TABLE "internal_policy_editors";
-- reverse: create "internal_policy_blocked_groups" table
DROP TABLE "internal_policy_blocked_groups";
-- reverse: create "integration_events" table
DROP TABLE "integration_events";
-- reverse: create "integration_secrets" table
DROP TABLE "integration_secrets";
-- reverse: create "hush_events" table
DROP TABLE "hush_events";
-- reverse: create "group_membership_events" table
DROP TABLE "group_membership_events";
-- reverse: create "group_tasks" table
DROP TABLE "group_tasks";
-- reverse: create "group_files" table
DROP TABLE "group_files";
-- reverse: create "group_events" table
DROP TABLE "group_events";
-- reverse: create "file_events" table
DROP TABLE "file_events";
-- reverse: create "evidence_files" table
DROP TABLE "evidence_files";
-- reverse: create "evidence_subcontrols" table
DROP TABLE "evidence_subcontrols";
-- reverse: create "evidence_controls" table
DROP TABLE "evidence_controls";
-- reverse: create "evidence_control_objectives" table
DROP TABLE "evidence_control_objectives";
-- reverse: create "entity_files" table
DROP TABLE "entity_files";
-- reverse: create "entity_documents" table
DROP TABLE "entity_documents";
-- reverse: create "entity_contacts" table
DROP TABLE "entity_contacts";
-- reverse: create "document_data_files" table
DROP TABLE "document_data_files";
-- reverse: create "control_objective_tasks" table
DROP TABLE "control_objective_tasks";
-- reverse: create "control_objective_viewers" table
DROP TABLE "control_objective_viewers";
-- reverse: create "control_objective_editors" table
DROP TABLE "control_objective_editors";
-- reverse: create "control_objective_blocked_groups" table
DROP TABLE "control_objective_blocked_groups";
-- reverse: create "control_control_implementations" table
DROP TABLE "control_control_implementations";
-- reverse: create "control_viewers" table
DROP TABLE "control_viewers";
-- reverse: create "control_editors" table
DROP TABLE "control_editors";
-- reverse: create "control_blocked_groups" table
DROP TABLE "control_blocked_groups";
-- reverse: create "control_procedures" table
DROP TABLE "control_procedures";
-- reverse: create "control_action_plans" table
DROP TABLE "control_action_plans";
-- reverse: create "control_risks" table
DROP TABLE "control_risks";
-- reverse: create "control_narratives" table
DROP TABLE "control_narratives";
-- reverse: create "control_tasks" table
DROP TABLE "control_tasks";
-- reverse: create "control_control_objectives" table
DROP TABLE "control_control_objectives";
-- reverse: create "contact_files" table
DROP TABLE "contact_files";
-- reverse: create index "webauthns_credential_id_key" to table: "webauthns"
DROP INDEX "webauthns_credential_id_key";
-- reverse: create index "webauthns_aaguid_key" to table: "webauthns"
DROP INDEX "webauthns_aaguid_key";
-- reverse: create "webauthns" table
DROP TABLE "webauthns";
-- reverse: create index "usersettinghistory_history_time" to table: "user_setting_history"
DROP INDEX "usersettinghistory_history_time";
-- reverse: create "user_setting_history" table
DROP TABLE "user_setting_history";
-- reverse: create index "user_settings_user_id_key" to table: "user_settings"
DROP INDEX "user_settings_user_id_key";
-- reverse: create "user_settings" table
DROP TABLE "user_settings";
-- reverse: create index "userhistory_history_time" to table: "user_history"
DROP INDEX "userhistory_history_time";
-- reverse: create "user_history" table
DROP TABLE "user_history";
-- reverse: create index "users_sub_key" to table: "users"
DROP INDEX "users_sub_key";
-- reverse: create index "users_display_id_key" to table: "users"
DROP INDEX "users_display_id_key";
-- reverse: create index "user_id" to table: "users"
DROP INDEX "user_id";
-- reverse: create index "user_email_auth_provider" to table: "users"
DROP INDEX "user_email_auth_provider";
-- reverse: create "users" table
DROP TABLE "users";
-- reverse: create index "templatehistory_history_time" to table: "template_history"
DROP INDEX "templatehistory_history_time";
-- reverse: create "template_history" table
DROP TABLE "template_history";
-- reverse: create index "template_name_owner_id_template_type" to table: "templates"
DROP INDEX "template_name_owner_id_template_type";
-- reverse: create "templates" table
DROP TABLE "templates";
-- reverse: create index "taskhistory_history_time" to table: "task_history"
DROP INDEX "taskhistory_history_time";
-- reverse: create "task_history" table
DROP TABLE "task_history";
-- reverse: create index "task_display_id_owner_id" to table: "tasks"
DROP INDEX "task_display_id_owner_id";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "tfasetting_owner_id" to table: "tfa_settings"
DROP INDEX "tfasetting_owner_id";
-- reverse: create "tfa_settings" table
DROP TABLE "tfa_settings";
-- reverse: create index "subscribers_token_key" to table: "subscribers"
DROP INDEX "subscribers_token_key";
-- reverse: create index "subscriber_email_owner_id" to table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- reverse: create "subscribers" table
DROP TABLE "subscribers";
-- reverse: create index "subcontrolhistory_history_time" to table: "subcontrol_history"
DROP INDEX "subcontrolhistory_history_time";
-- reverse: create "subcontrol_history" table
DROP TABLE "subcontrol_history";
-- reverse: create index "subcontrol_display_id_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_display_id_owner_id";
-- reverse: create index "subcontrol_control_id_ref_code" to table: "subcontrols"
DROP INDEX "subcontrol_control_id_ref_code";
-- reverse: create "subcontrols" table
DROP TABLE "subcontrols";
-- reverse: create index "standardhistory_history_time" to table: "standard_history"
DROP INDEX "standardhistory_history_time";
-- reverse: create "standard_history" table
DROP TABLE "standard_history";
-- reverse: create "standards" table
DROP TABLE "standards";
-- reverse: create index "riskhistory_history_time" to table: "risk_history"
DROP INDEX "riskhistory_history_time";
-- reverse: create "risk_history" table
DROP TABLE "risk_history";
-- reverse: create index "risk_display_id_owner_id" to table: "risks"
DROP INDEX "risk_display_id_owner_id";
-- reverse: create "risks" table
DROP TABLE "risks";
-- reverse: create index "programmembershiphistory_history_time" to table: "program_membership_history"
DROP INDEX "programmembershiphistory_history_time";
-- reverse: create "program_membership_history" table
DROP TABLE "program_membership_history";
-- reverse: create index "programmembership_user_id_program_id" to table: "program_memberships"
DROP INDEX "programmembership_user_id_program_id";
-- reverse: create "program_memberships" table
DROP TABLE "program_memberships";
-- reverse: create index "programhistory_history_time" to table: "program_history"
DROP INDEX "programhistory_history_time";
-- reverse: create "program_history" table
DROP TABLE "program_history";
-- reverse: create index "program_display_id_owner_id" to table: "programs"
DROP INDEX "program_display_id_owner_id";
-- reverse: create "programs" table
DROP TABLE "programs";
-- reverse: create index "procedurehistory_history_time" to table: "procedure_history"
DROP INDEX "procedurehistory_history_time";
-- reverse: create "procedure_history" table
DROP TABLE "procedure_history";
-- reverse: create index "procedure_display_id_owner_id" to table: "procedures"
DROP INDEX "procedure_display_id_owner_id";
-- reverse: create "procedures" table
DROP TABLE "procedures";
-- reverse: create index "personalaccesstoken_token" to table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_token";
-- reverse: create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_token_key";
-- reverse: create "personal_access_tokens" table
DROP TABLE "personal_access_tokens";
-- reverse: create index "passwordresettoken_token" to table: "password_reset_tokens"
DROP INDEX "passwordresettoken_token";
-- reverse: create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_token_key";
-- reverse: create "password_reset_tokens" table
DROP TABLE "password_reset_tokens";
-- reverse: create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
DROP INDEX "organizationsettinghistory_history_time";
-- reverse: create "organization_setting_history" table
DROP TABLE "organization_setting_history";
-- reverse: create index "organization_settings_organization_id_key" to table: "organization_settings"
DROP INDEX "organization_settings_organization_id_key";
-- reverse: create "organization_settings" table
DROP TABLE "organization_settings";
-- reverse: create index "organizationhistory_history_time" to table: "organization_history"
DROP INDEX "organizationhistory_history_time";
-- reverse: create "organization_history" table
DROP TABLE "organization_history";
-- reverse: create index "organization_name" to table: "organizations"
DROP INDEX "organization_name";
-- reverse: create "organizations" table
DROP TABLE "organizations";
-- reverse: create index "orgsubscriptionhistory_history_time" to table: "org_subscription_history"
DROP INDEX "orgsubscriptionhistory_history_time";
-- reverse: create "org_subscription_history" table
DROP TABLE "org_subscription_history";
-- reverse: create index "org_subscriptions_stripe_customer_id_key" to table: "org_subscriptions"
DROP INDEX "org_subscriptions_stripe_customer_id_key";
-- reverse: create "org_subscriptions" table
DROP TABLE "org_subscriptions";
-- reverse: create index "orgmembershiphistory_history_time" to table: "org_membership_history"
DROP INDEX "orgmembershiphistory_history_time";
-- reverse: create "org_membership_history" table
DROP TABLE "org_membership_history";
-- reverse: create index "orgmembership_user_id_organization_id" to table: "org_memberships"
DROP INDEX "orgmembership_user_id_organization_id";
-- reverse: create "org_memberships" table
DROP TABLE "org_memberships";
-- reverse: create "onboardings" table
DROP TABLE "onboardings";
-- reverse: create index "notehistory_history_time" to table: "note_history"
DROP INDEX "notehistory_history_time";
-- reverse: create "note_history" table
DROP TABLE "note_history";
-- reverse: create index "note_display_id_owner_id" to table: "notes"
DROP INDEX "note_display_id_owner_id";
-- reverse: create "notes" table
DROP TABLE "notes";
-- reverse: create index "narrativehistory_history_time" to table: "narrative_history"
DROP INDEX "narrativehistory_history_time";
-- reverse: create "narrative_history" table
DROP TABLE "narrative_history";
-- reverse: create index "narrative_display_id_owner_id" to table: "narratives"
DROP INDEX "narrative_display_id_owner_id";
-- reverse: create "narratives" table
DROP TABLE "narratives";
-- reverse: create index "mappedcontrolhistory_history_time" to table: "mapped_control_history"
DROP INDEX "mappedcontrolhistory_history_time";
-- reverse: create "mapped_control_history" table
DROP TABLE "mapped_control_history";
-- reverse: create "mapped_controls" table
DROP TABLE "mapped_controls";
-- reverse: create index "invites_token_key" to table: "invites"
DROP INDEX "invites_token_key";
-- reverse: create index "invite_recipient_owner_id" to table: "invites"
DROP INDEX "invite_recipient_owner_id";
-- reverse: create "invites" table
DROP TABLE "invites";
-- reverse: create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
DROP INDEX "internalpolicyhistory_history_time";
-- reverse: create "internal_policy_history" table
DROP TABLE "internal_policy_history";
-- reverse: create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
DROP INDEX "internalpolicy_display_id_owner_id";
-- reverse: create "internal_policies" table
DROP TABLE "internal_policies";
-- reverse: create index "integrationhistory_history_time" to table: "integration_history"
DROP INDEX "integrationhistory_history_time";
-- reverse: create "integration_history" table
DROP TABLE "integration_history";
-- reverse: create "integrations" table
DROP TABLE "integrations";
-- reverse: create index "hushhistory_history_time" to table: "hush_history"
DROP INDEX "hushhistory_history_time";
-- reverse: create "hush_history" table
DROP TABLE "hush_history";
-- reverse: create "hushes" table
DROP TABLE "hushes";
-- reverse: create index "groupsettinghistory_history_time" to table: "group_setting_history"
DROP INDEX "groupsettinghistory_history_time";
-- reverse: create "group_setting_history" table
DROP TABLE "group_setting_history";
-- reverse: create index "group_settings_group_id_key" to table: "group_settings"
DROP INDEX "group_settings_group_id_key";
-- reverse: create "group_settings" table
DROP TABLE "group_settings";
-- reverse: create index "groupmembershiphistory_history_time" to table: "group_membership_history"
DROP INDEX "groupmembershiphistory_history_time";
-- reverse: create "group_membership_history" table
DROP TABLE "group_membership_history";
-- reverse: create index "groupmembership_user_id_group_id" to table: "group_memberships"
DROP INDEX "groupmembership_user_id_group_id";
-- reverse: create "group_memberships" table
DROP TABLE "group_memberships";
-- reverse: create index "grouphistory_history_time" to table: "group_history"
DROP INDEX "grouphistory_history_time";
-- reverse: create "group_history" table
DROP TABLE "group_history";
-- reverse: create index "group_name_owner_id" to table: "groups"
DROP INDEX "group_name_owner_id";
-- reverse: create index "group_display_id_owner_id" to table: "groups"
DROP INDEX "group_display_id_owner_id";
-- reverse: create "groups" table
DROP TABLE "groups";
-- reverse: create index "filehistory_history_time" to table: "file_history"
DROP INDEX "filehistory_history_time";
-- reverse: create "file_history" table
DROP TABLE "file_history";
-- reverse: create "files" table
DROP TABLE "files";
-- reverse: create index "evidencehistory_history_time" to table: "evidence_history"
DROP INDEX "evidencehistory_history_time";
-- reverse: create "evidence_history" table
DROP TABLE "evidence_history";
-- reverse: create index "evidence_display_id_owner_id" to table: "evidences"
DROP INDEX "evidence_display_id_owner_id";
-- reverse: create "evidences" table
DROP TABLE "evidences";
-- reverse: create index "eventhistory_history_time" to table: "event_history"
DROP INDEX "eventhistory_history_time";
-- reverse: create "event_history" table
DROP TABLE "event_history";
-- reverse: create "events" table
DROP TABLE "events";
-- reverse: create index "entitytypehistory_history_time" to table: "entity_type_history"
DROP INDEX "entitytypehistory_history_time";
-- reverse: create "entity_type_history" table
DROP TABLE "entity_type_history";
-- reverse: create index "entitytype_name_owner_id" to table: "entity_types"
DROP INDEX "entitytype_name_owner_id";
-- reverse: create "entity_types" table
DROP TABLE "entity_types";
-- reverse: create index "entityhistory_history_time" to table: "entity_history"
DROP INDEX "entityhistory_history_time";
-- reverse: create "entity_history" table
DROP TABLE "entity_history";
-- reverse: create index "entity_name_owner_id" to table: "entities"
DROP INDEX "entity_name_owner_id";
-- reverse: create "entities" table
DROP TABLE "entities";
-- reverse: create index "emailverificationtoken_token" to table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_token";
-- reverse: create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_token_key";
-- reverse: create "email_verification_tokens" table
DROP TABLE "email_verification_tokens";
-- reverse: create index "documentdatahistory_history_time" to table: "document_data_history"
DROP INDEX "documentdatahistory_history_time";
-- reverse: create "document_data_history" table
DROP TABLE "document_data_history";
-- reverse: create "document_data" table
DROP TABLE "document_data";
-- reverse: create index "controlobjectivehistory_history_time" to table: "control_objective_history"
DROP INDEX "controlobjectivehistory_history_time";
-- reverse: create "control_objective_history" table
DROP TABLE "control_objective_history";
-- reverse: create index "controlobjective_display_id_owner_id" to table: "control_objectives"
DROP INDEX "controlobjective_display_id_owner_id";
-- reverse: create "control_objectives" table
DROP TABLE "control_objectives";
-- reverse: create index "controlimplementationhistory_history_time" to table: "control_implementation_history"
DROP INDEX "controlimplementationhistory_history_time";
-- reverse: create "control_implementation_history" table
DROP TABLE "control_implementation_history";
-- reverse: create "control_implementations" table
DROP TABLE "control_implementations";
-- reverse: create index "controlhistory_history_time" to table: "control_history"
DROP INDEX "controlhistory_history_time";
-- reverse: create "control_history" table
DROP TABLE "control_history";
-- reverse: create index "controls_reference_id_key" to table: "controls"
DROP INDEX "controls_reference_id_key";
-- reverse: create index "controls_auditor_reference_id_key" to table: "controls"
DROP INDEX "controls_auditor_reference_id_key";
-- reverse: create index "control_standard_id_ref_code" to table: "controls"
DROP INDEX "control_standard_id_ref_code";
-- reverse: create index "control_display_id_owner_id" to table: "controls"
DROP INDEX "control_display_id_owner_id";
-- reverse: create "controls" table
DROP TABLE "controls";
-- reverse: create index "contacthistory_history_time" to table: "contact_history"
DROP INDEX "contacthistory_history_time";
-- reverse: create "contact_history" table
DROP TABLE "contact_history";
-- reverse: create "contacts" table
DROP TABLE "contacts";
-- reverse: create index "actionplanhistory_history_time" to table: "action_plan_history"
DROP INDEX "actionplanhistory_history_time";
-- reverse: create "action_plan_history" table
DROP TABLE "action_plan_history";
-- reverse: create "action_plans" table
DROP TABLE "action_plans";
-- reverse: create index "apitoken_token" to table: "api_tokens"
DROP INDEX "apitoken_token";
-- reverse: create index "api_tokens_token_key" to table: "api_tokens"
DROP INDEX "api_tokens_token_key";
-- reverse: create "api_tokens" table
DROP TABLE "api_tokens";
