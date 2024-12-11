-- +goose Up
-- create "api_tokens" table
CREATE TABLE "api_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "api_tokens_mapping_id_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_mapping_id_key" ON "api_tokens" ("mapping_id");
-- create index "api_tokens_token_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_token_key" ON "api_tokens" ("token");
-- create index "apitoken_token" to table: "api_tokens"
CREATE INDEX "apitoken_token" ON "api_tokens" ("token");
-- create "action_plans" table
CREATE TABLE "action_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "due_date" timestamptz NULL, "priority" character varying NULL, "source" character varying NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "action_plans_mapping_id_key" to table: "action_plans"
CREATE UNIQUE INDEX "action_plans_mapping_id_key" ON "action_plans" ("mapping_id");
-- create "action_plan_history" table
CREATE TABLE "action_plan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "due_date" timestamptz NULL, "priority" character varying NULL, "source" character varying NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "actionplanhistory_history_time" to table: "action_plan_history"
CREATE INDEX "actionplanhistory_history_time" ON "action_plan_history" ("history_time");
-- create "contacts" table
CREATE TABLE "contacts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "contacts_mapping_id_key" to table: "contacts"
CREATE UNIQUE INDEX "contacts_mapping_id_key" ON "contacts" ("mapping_id");
-- create "contact_history" table
CREATE TABLE "contact_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', PRIMARY KEY ("id"));
-- create index "contacthistory_history_time" to table: "contact_history"
CREATE INDEX "contacthistory_history_time" ON "contact_history" ("history_time");
-- create "controls" table
CREATE TABLE "controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_type" character varying NULL, "version" character varying NULL, "control_number" character varying NULL, "family" text NULL, "class" character varying NULL, "source" character varying NULL, "satisfies" text NULL, "mapped_frameworks" text NULL, "details" jsonb NULL, "control_objective_controls" character varying NULL, "internal_policy_controls" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "controls_mapping_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_mapping_id_key" ON "controls" ("mapping_id");
-- create "control_history" table
CREATE TABLE "control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_type" character varying NULL, "version" character varying NULL, "control_number" character varying NULL, "family" text NULL, "class" character varying NULL, "source" character varying NULL, "satisfies" text NULL, "mapped_frameworks" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "controlhistory_history_time" to table: "control_history"
CREATE INDEX "controlhistory_history_time" ON "control_history" ("history_time");
-- create "control_objectives" table
CREATE TABLE "control_objectives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_objective_type" character varying NULL, "version" character varying NULL, "control_number" character varying NULL, "family" text NULL, "class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "details" jsonb NULL, "control_controlobjectives" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "control_objectives_mapping_id_key" to table: "control_objectives"
CREATE UNIQUE INDEX "control_objectives_mapping_id_key" ON "control_objectives" ("mapping_id");
-- create "control_objective_history" table
CREATE TABLE "control_objective_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "control_objective_type" character varying NULL, "version" character varying NULL, "control_number" character varying NULL, "family" text NULL, "class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "controlobjectivehistory_history_time" to table: "control_objective_history"
CREATE INDEX "controlobjectivehistory_history_time" ON "control_objective_history" ("history_time");
-- create "document_data" table
CREATE TABLE "document_data" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "data" jsonb NOT NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "document_data_mapping_id_key" to table: "document_data"
CREATE UNIQUE INDEX "document_data_mapping_id_key" ON "document_data" ("mapping_id");
-- create "document_data_history" table
CREATE TABLE "document_data_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, "data" jsonb NOT NULL, PRIMARY KEY ("id"));
-- create index "documentdatahistory_history_time" to table: "document_data_history"
CREATE INDEX "documentdatahistory_history_time" ON "document_data_history" ("history_time");
-- create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "email_verification_tokens_mapping_id_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_mapping_id_key" ON "email_verification_tokens" ("mapping_id");
-- create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_token_key" ON "email_verification_tokens" ("token");
-- create index "emailverificationtoken_token" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_token" ON "email_verification_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "entities" table
CREATE TABLE "entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "status" character varying NULL DEFAULT 'active', "entity_type_id" character varying NULL, "entity_type_entities" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "entities_mapping_id_key" to table: "entities"
CREATE UNIQUE INDEX "entities_mapping_id_key" ON "entities" ("mapping_id");
-- create index "entity_name_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "entity_history" table
CREATE TABLE "entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "entity_type_id" character varying NULL, "status" character varying NULL DEFAULT 'active', PRIMARY KEY ("id"));
-- create index "entityhistory_history_time" to table: "entity_history"
CREATE INDEX "entityhistory_history_time" ON "entity_history" ("history_time");
-- create "entity_types" table
CREATE TABLE "entity_types" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "entity_types_mapping_id_key" to table: "entity_types"
CREATE UNIQUE INDEX "entity_types_mapping_id_key" ON "entity_types" ("mapping_id");
-- create index "entitytype_name_owner_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_name_owner_id" ON "entity_types" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "entity_type_history" table
CREATE TABLE "entity_type_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "entitytypehistory_history_time" to table: "entity_type_history"
CREATE INDEX "entitytypehistory_history_time" ON "entity_type_history" ("history_time");
-- create "events" table
CREATE TABLE "events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "events_mapping_id_key" to table: "events"
CREATE UNIQUE INDEX "events_mapping_id_key" ON "events" ("mapping_id");
-- create "event_history" table
CREATE TABLE "event_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "eventhistory_history_time" to table: "event_history"
CREATE INDEX "eventhistory_history_time" ON "event_history" ("history_time");
-- create "files" table
CREATE TABLE "files" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, PRIMARY KEY ("id"));
-- create index "files_mapping_id_key" to table: "files"
CREATE UNIQUE INDEX "files_mapping_id_key" ON "files" ("mapping_id");
-- create "file_history" table
CREATE TABLE "file_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "provided_file_name" character varying NOT NULL, "provided_file_extension" character varying NOT NULL, "provided_file_size" bigint NULL, "persisted_file_size" bigint NULL, "detected_mime_type" character varying NULL, "md5_hash" character varying NULL, "detected_content_type" character varying NOT NULL, "store_key" character varying NULL, "category_type" character varying NULL, "uri" character varying NULL, "storage_scheme" character varying NULL, "storage_volume" character varying NULL, "storage_path" character varying NULL, "file_contents" bytea NULL, PRIMARY KEY ("id"));
-- create index "filehistory_history_time" to table: "file_history"
CREATE INDEX "filehistory_history_time" ON "file_history" ("history_time");
-- create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_name_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_name_owner_id" ON "groups" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "groups_mapping_id_key" to table: "groups"
CREATE UNIQUE INDEX "groups_mapping_id_key" ON "groups" ("mapping_id");
-- create "group_history" table
CREATE TABLE "group_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', PRIMARY KEY ("id"));
-- create index "grouphistory_history_time" to table: "group_history"
CREATE INDEX "grouphistory_history_time" ON "group_history" ("history_time");
-- create "group_memberships" table
CREATE TABLE "group_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "group_memberships_mapping_id_key" to table: "group_memberships"
CREATE UNIQUE INDEX "group_memberships_mapping_id_key" ON "group_memberships" ("mapping_id");
-- create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id") WHERE (deleted_at IS NULL);
-- create "group_membership_history" table
CREATE TABLE "group_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "groupmembershiphistory_history_time" to table: "group_membership_history"
CREATE INDEX "groupmembershiphistory_history_time" ON "group_membership_history" ("history_time");
-- create "group_settings" table
CREATE TABLE "group_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "group_settings_group_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_group_id_key" ON "group_settings" ("group_id");
-- create index "group_settings_mapping_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_mapping_id_key" ON "group_settings" ("mapping_id");
-- create "group_setting_history" table
CREATE TABLE "group_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "groupsettinghistory_history_time" to table: "group_setting_history"
CREATE INDEX "groupsettinghistory_history_time" ON "group_setting_history" ("history_time");
-- create "hushes" table
CREATE TABLE "hushes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, PRIMARY KEY ("id"));
-- create index "hushes_mapping_id_key" to table: "hushes"
CREATE UNIQUE INDEX "hushes_mapping_id_key" ON "hushes" ("mapping_id");
-- create "hush_history" table
CREATE TABLE "hush_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, PRIMARY KEY ("id"));
-- create index "hushhistory_history_time" to table: "hush_history"
CREATE INDEX "hushhistory_history_time" ON "hush_history" ("history_time");
-- create "integrations" table
CREATE TABLE "integrations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "group_integrations" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "integrations_mapping_id_key" to table: "integrations"
CREATE UNIQUE INDEX "integrations_mapping_id_key" ON "integrations" ("mapping_id");
-- create "integration_history" table
CREATE TABLE "integration_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, PRIMARY KEY ("id"));
-- create index "integrationhistory_history_time" to table: "integration_history"
CREATE INDEX "integrationhistory_history_time" ON "integration_history" ("history_time");
-- create "internal_policies" table
CREATE TABLE "internal_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "policy_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "details" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "internal_policies_mapping_id_key" to table: "internal_policies"
CREATE UNIQUE INDEX "internal_policies_mapping_id_key" ON "internal_policies" ("mapping_id");
-- create "internal_policy_history" table
CREATE TABLE "internal_policy_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "policy_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
CREATE INDEX "internalpolicyhistory_history_time" ON "internal_policy_history" ("history_time");
-- create "invites" table
CREATE TABLE "invites" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "expires" timestamptz NULL, "recipient" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'INVITATION_SENT', "role" character varying NOT NULL DEFAULT 'MEMBER', "send_attempts" bigint NOT NULL DEFAULT 0, "requestor_id" character varying NULL, "secret" bytea NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "invite_recipient_owner_id" to table: "invites"
CREATE UNIQUE INDEX "invite_recipient_owner_id" ON "invites" ("recipient", "owner_id") WHERE (deleted_at IS NULL);
-- create index "invites_mapping_id_key" to table: "invites"
CREATE UNIQUE INDEX "invites_mapping_id_key" ON "invites" ("mapping_id");
-- create index "invites_token_key" to table: "invites"
CREATE UNIQUE INDEX "invites_token_key" ON "invites" ("token");
-- create "narratives" table
CREATE TABLE "narratives" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "satisfies" text NULL, "details" jsonb NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "narratives_mapping_id_key" to table: "narratives"
CREATE UNIQUE INDEX "narratives_mapping_id_key" ON "narratives" ("mapping_id");
-- create "narrative_history" table
CREATE TABLE "narrative_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" text NULL, "satisfies" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "narrativehistory_history_time" to table: "narrative_history"
CREATE INDEX "narrativehistory_history_time" ON "narrative_history" ("history_time");
-- create "notes" table
CREATE TABLE "notes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "text" character varying NOT NULL, "entity_notes" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "notes_mapping_id_key" to table: "notes"
CREATE UNIQUE INDEX "notes_mapping_id_key" ON "notes" ("mapping_id");
-- create "note_history" table
CREATE TABLE "note_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "text" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "notehistory_history_time" to table: "note_history"
CREATE INDEX "notehistory_history_time" ON "note_history" ("history_time");
-- create "org_memberships" table
CREATE TABLE "org_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "org_memberships_mapping_id_key" to table: "org_memberships"
CREATE UNIQUE INDEX "org_memberships_mapping_id_key" ON "org_memberships" ("mapping_id");
-- create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id") WHERE (deleted_at IS NULL);
-- create "org_membership_history" table
CREATE TABLE "org_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "orgmembershiphistory_history_time" to table: "org_membership_history"
CREATE INDEX "orgmembershiphistory_history_time" ON "org_membership_history" ("history_time");
-- create "org_subscriptions" table
CREATE TABLE "org_subscriptions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "features" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "org_subscriptions_mapping_id_key" to table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_mapping_id_key" ON "org_subscriptions" ("mapping_id");
-- create index "org_subscriptions_stripe_customer_id_key" to table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_stripe_customer_id_key" ON "org_subscriptions" ("stripe_customer_id");
-- create "org_subscription_history" table
CREATE TABLE "org_subscription_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "features" jsonb NULL, PRIMARY KEY ("id"));
-- create index "orgsubscriptionhistory_history_time" to table: "org_subscription_history"
CREATE INDEX "orgsubscriptionhistory_history_time" ON "org_subscription_history" ("history_time");
-- create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "dedicated_db" boolean NOT NULL DEFAULT false, "parent_organization_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "organization_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_name" ON "organizations" ("name") WHERE (deleted_at IS NULL);
-- create index "organizations_mapping_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_mapping_id_key" ON "organizations" ("mapping_id");
-- create "organization_history" table
CREATE TABLE "organization_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "parent_organization_id" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "dedicated_db" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "organizationhistory_history_time" to table: "organization_history"
CREATE INDEX "organizationhistory_history_time" ON "organization_history" ("history_time");
-- create "organization_settings" table
CREATE TABLE "organization_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" character varying NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "stripe_id" character varying NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "organization_settings_mapping_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_mapping_id_key" ON "organization_settings" ("mapping_id");
-- create index "organization_settings_organization_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_organization_id_key" ON "organization_settings" ("organization_id");
-- create "organization_setting_history" table
CREATE TABLE "organization_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" character varying NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "organization_id" character varying NULL, "stripe_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
CREATE INDEX "organizationsettinghistory_history_time" ON "organization_setting_history" ("history_time");
-- create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "password_reset_tokens_mapping_id_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_mapping_id_key" ON "password_reset_tokens" ("mapping_id");
-- create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
-- create index "passwordresettoken_token" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_token" ON "password_reset_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "personal_access_tokens" table
CREATE TABLE "personal_access_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "personal_access_tokens_mapping_id_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_mapping_id_key" ON "personal_access_tokens" ("mapping_id");
-- create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_token_key" ON "personal_access_tokens" ("token");
-- create index "personalaccesstoken_token" to table: "personal_access_tokens"
CREATE INDEX "personalaccesstoken_token" ON "personal_access_tokens" ("token");
-- create "procedures" table
CREATE TABLE "procedures" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "procedure_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "details" jsonb NULL, "control_objective_procedures" character varying NULL, "owner_id" character varying NULL, "standard_procedures" character varying NULL, PRIMARY KEY ("id"));
-- create index "procedures_mapping_id_key" to table: "procedures"
CREATE UNIQUE INDEX "procedures_mapping_id_key" ON "procedures" ("mapping_id");
-- create "procedure_history" table
CREATE TABLE "procedure_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "procedure_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "procedurehistory_history_time" to table: "procedure_history"
CREATE INDEX "procedurehistory_history_time" ON "procedure_history" ("history_time");
-- create "programs" table
CREATE TABLE "programs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "programs_mapping_id_key" to table: "programs"
CREATE UNIQUE INDEX "programs_mapping_id_key" ON "programs" ("mapping_id");
-- create "program_history" table
CREATE TABLE "program_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "programhistory_history_time" to table: "program_history"
CREATE INDEX "programhistory_history_time" ON "program_history" ("history_time");
-- create "program_memberships" table
CREATE TABLE "program_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "program_memberships_mapping_id_key" to table: "program_memberships"
CREATE UNIQUE INDEX "program_memberships_mapping_id_key" ON "program_memberships" ("mapping_id");
-- create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id") WHERE (deleted_at IS NULL);
-- create "program_membership_history" table
CREATE TABLE "program_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "programmembershiphistory_history_time" to table: "program_membership_history"
CREATE INDEX "programmembershiphistory_history_time" ON "program_membership_history" ("history_time");
-- create "risks" table
CREATE TABLE "risks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "risk_type" character varying NULL, "business_costs" text NULL, "impact" character varying NULL DEFAULT 'MODERATE', "likelihood" character varying NULL DEFAULT 'LIKELY', "mitigation" text NULL, "satisfies" text NULL, "details" jsonb NULL, "control_objective_risks" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "risks_mapping_id_key" to table: "risks"
CREATE UNIQUE INDEX "risks_mapping_id_key" ON "risks" ("mapping_id");
-- create "risk_history" table
CREATE TABLE "risk_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "risk_type" character varying NULL, "business_costs" text NULL, "impact" character varying NULL DEFAULT 'MODERATE', "likelihood" character varying NULL DEFAULT 'LIKELY', "mitigation" text NULL, "satisfies" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "riskhistory_history_time" to table: "risk_history"
CREATE INDEX "riskhistory_history_time" ON "risk_history" ("history_time");
-- create "standards" table
CREATE TABLE "standards" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "family" character varying NULL, "status" character varying NULL, "standard_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "standards_mapping_id_key" to table: "standards"
CREATE UNIQUE INDEX "standards_mapping_id_key" ON "standards" ("mapping_id");
-- create "standard_history" table
CREATE TABLE "standard_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "family" character varying NULL, "status" character varying NULL, "standard_type" character varying NULL, "version" character varying NULL, "purpose_and_scope" text NULL, "background" text NULL, "satisfies" text NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "standardhistory_history_time" to table: "standard_history"
CREATE INDEX "standardhistory_history_time" ON "standard_history" ("history_time");
-- create "subcontrols" table
CREATE TABLE "subcontrols" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "subcontrol_type" character varying NULL, "version" character varying NULL, "subcontrol_number" character varying NULL, "family" text NULL, "class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "implementation_evidence" character varying NULL, "implementation_status" character varying NULL, "implementation_date" timestamptz NULL, "implementation_verification" character varying NULL, "implementation_verification_date" timestamptz NULL, "details" jsonb NULL, "control_objective_subcontrols" character varying NULL, "note_subcontrols" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "subcontrols_mapping_id_key" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrols_mapping_id_key" ON "subcontrols" ("mapping_id");
-- create "subcontrol_history" table
CREATE TABLE "subcontrol_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" text NULL, "status" character varying NULL, "subcontrol_type" character varying NULL, "version" character varying NULL, "subcontrol_number" character varying NULL, "family" text NULL, "class" character varying NULL, "source" character varying NULL, "mapped_frameworks" text NULL, "implementation_evidence" character varying NULL, "implementation_status" character varying NULL, "implementation_date" timestamptz NULL, "implementation_verification" character varying NULL, "implementation_verification_date" timestamptz NULL, "details" jsonb NULL, PRIMARY KEY ("id"));
-- create index "subcontrolhistory_history_time" to table: "subcontrol_history"
CREATE INDEX "subcontrolhistory_history_time" ON "subcontrol_history" ("history_time");
-- create "subscribers" table
CREATE TABLE "subscribers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "email" character varying NOT NULL, "phone_number" character varying NULL, "verified_email" boolean NOT NULL DEFAULT false, "verified_phone" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT false, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE (deleted_at IS NULL);
-- create index "subscribers_mapping_id_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_mapping_id_key" ON "subscribers" ("mapping_id");
-- create index "subscribers_token_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_token_key" ON "subscribers" ("token");
-- create "tfa_settings" table
CREATE TABLE "tfa_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "tfa_secret" character varying NULL, "verified" boolean NOT NULL DEFAULT false, "recovery_codes" jsonb NULL, "phone_otp_allowed" boolean NULL DEFAULT false, "email_otp_allowed" boolean NULL DEFAULT false, "totp_allowed" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "tfa_settings_mapping_id_key" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfa_settings_mapping_id_key" ON "tfa_settings" ("mapping_id");
-- create index "tfasetting_owner_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_owner_id" ON "tfa_settings" ("owner_id") WHERE (deleted_at IS NULL);
-- create "tasks" table
CREATE TABLE "tasks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" jsonb NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "due" timestamptz NULL, "completed" timestamptz NULL, "user_assigner_tasks" character varying NOT NULL, "user_assignee_tasks" character varying NULL, PRIMARY KEY ("id"));
-- create index "tasks_mapping_id_key" to table: "tasks"
CREATE UNIQUE INDEX "tasks_mapping_id_key" ON "tasks" ("mapping_id");
-- create "task_history" table
CREATE TABLE "task_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" jsonb NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "due" timestamptz NULL, "completed" timestamptz NULL, PRIMARY KEY ("id"));
-- create index "taskhistory_history_time" to table: "task_history"
CREATE INDEX "taskhistory_history_time" ON "task_history" ("history_time");
-- create "templates" table
CREATE TABLE "templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "template_name_owner_id_template_type" to table: "templates"
CREATE UNIQUE INDEX "template_name_owner_id_template_type" ON "templates" ("name", "owner_id", "template_type") WHERE (deleted_at IS NULL);
-- create index "templates_mapping_id_key" to table: "templates"
CREATE UNIQUE INDEX "templates_mapping_id_key" ON "templates" ("mapping_id");
-- create "template_history" table
CREATE TABLE "template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "templatehistory_history_time" to table: "template_history"
CREATE INDEX "templatehistory_history_time" ON "template_history" ("history_time");
-- create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_local_file" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', "avatar_local_file_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "user_email_auth_provider" to table: "users"
CREATE UNIQUE INDEX "user_email_auth_provider" ON "users" ("email", "auth_provider") WHERE (deleted_at IS NULL);
-- create index "user_id" to table: "users"
CREATE UNIQUE INDEX "user_id" ON "users" ("id");
-- create index "users_mapping_id_key" to table: "users"
CREATE UNIQUE INDEX "users_mapping_id_key" ON "users" ("mapping_id");
-- create index "users_sub_key" to table: "users"
CREATE UNIQUE INDEX "users_sub_key" ON "users" ("sub");
-- create "user_history" table
CREATE TABLE "user_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_local_file" character varying NULL, "avatar_local_file_id" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', PRIMARY KEY ("id"));
-- create index "userhistory_history_time" to table: "user_history"
CREATE INDEX "userhistory_history_time" ON "user_history" ("history_time");
-- create "user_settings" table
CREATE TABLE "user_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, "user_id" character varying NULL, "user_setting_default_org" character varying NULL, PRIMARY KEY ("id"));
-- create index "user_settings_mapping_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_mapping_id_key" ON "user_settings" ("mapping_id");
-- create index "user_settings_user_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_user_id_key" ON "user_settings" ("user_id");
-- create "user_setting_history" table
CREATE TABLE "user_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "user_id" character varying NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, PRIMARY KEY ("id"));
-- create index "usersettinghistory_history_time" to table: "user_setting_history"
CREATE INDEX "usersettinghistory_history_time" ON "user_setting_history" ("history_time");
-- create "webauthns" table
CREATE TABLE "webauthns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "credential_id" bytea NULL, "public_key" bytea NULL, "attestation_type" character varying NULL, "aaguid" bytea NOT NULL, "sign_count" integer NOT NULL, "transports" jsonb NOT NULL, "backup_eligible" boolean NOT NULL DEFAULT false, "backup_state" boolean NOT NULL DEFAULT false, "user_present" boolean NOT NULL DEFAULT false, "user_verified" boolean NOT NULL DEFAULT false, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "webauthns_aaguid_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_aaguid_key" ON "webauthns" ("aaguid");
-- create index "webauthns_credential_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_credential_id_key" ON "webauthns" ("credential_id");
-- create index "webauthns_mapping_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_mapping_id_key" ON "webauthns" ("mapping_id");
-- create "contact_files" table
CREATE TABLE "contact_files" ("contact_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "file_id"));
-- create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create "control_viewers" table
CREATE TABLE "control_viewers" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"));
-- create "control_procedures" table
CREATE TABLE "control_procedures" ("control_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("control_id", "procedure_id"));
-- create "control_subcontrols" table
CREATE TABLE "control_subcontrols" ("control_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("control_id", "subcontrol_id"));
-- create "control_narratives" table
CREATE TABLE "control_narratives" ("control_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_id", "narrative_id"));
-- create "control_risks" table
CREATE TABLE "control_risks" ("control_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("control_id", "risk_id"));
-- create "control_actionplans" table
CREATE TABLE "control_actionplans" ("control_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("control_id", "action_plan_id"));
-- create "control_tasks" table
CREATE TABLE "control_tasks" ("control_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_id", "task_id"));
-- create "control_objective_blocked_groups" table
CREATE TABLE "control_objective_blocked_groups" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create "control_objective_editors" table
CREATE TABLE "control_objective_editors" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create "control_objective_viewers" table
CREATE TABLE "control_objective_viewers" ("control_objective_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "group_id"));
-- create "control_objective_narratives" table
CREATE TABLE "control_objective_narratives" ("control_objective_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "narrative_id"));
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
-- create "internal_policy_controlobjectives" table
CREATE TABLE "internal_policy_controlobjectives" ("internal_policy_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "control_objective_id"));
-- create "internal_policy_procedures" table
CREATE TABLE "internal_policy_procedures" ("internal_policy_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "procedure_id"));
-- create "internal_policy_narratives" table
CREATE TABLE "internal_policy_narratives" ("internal_policy_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "narrative_id"));
-- create "internal_policy_tasks" table
CREATE TABLE "internal_policy_tasks" ("internal_policy_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "task_id"));
-- create "invite_events" table
CREATE TABLE "invite_events" ("invite_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "event_id"));
-- create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"));
-- create "org_membership_events" table
CREATE TABLE "org_membership_events" ("org_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_membership_id", "event_id"));
-- create "organization_control_creators" table
CREATE TABLE "organization_control_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_control_objective_creators" table
CREATE TABLE "organization_control_objective_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_group_creators" table
CREATE TABLE "organization_group_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_internal_policy_creators" table
CREATE TABLE "organization_internal_policy_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_narrative_creators" table
CREATE TABLE "organization_narrative_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_procedure_creators" table
CREATE TABLE "organization_procedure_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_program_creators" table
CREATE TABLE "organization_program_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_risk_creators" table
CREATE TABLE "organization_risk_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_template_creators" table
CREATE TABLE "organization_template_creators" ("organization_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "group_id"));
-- create "organization_personal_access_tokens" table
CREATE TABLE "organization_personal_access_tokens" ("organization_id" character varying NOT NULL, "personal_access_token_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "personal_access_token_id"));
-- create "organization_events" table
CREATE TABLE "organization_events" ("organization_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "event_id"));
-- create "organization_secrets" table
CREATE TABLE "organization_secrets" ("organization_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "hush_id"));
-- create "organization_files" table
CREATE TABLE "organization_files" ("organization_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "file_id"));
-- create "organization_tasks" table
CREATE TABLE "organization_tasks" ("organization_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "task_id"));
-- create "organization_setting_files" table
CREATE TABLE "organization_setting_files" ("organization_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_setting_id", "file_id"));
-- create "personal_access_token_events" table
CREATE TABLE "personal_access_token_events" ("personal_access_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("personal_access_token_id", "event_id"));
-- create "procedure_blocked_groups" table
CREATE TABLE "procedure_blocked_groups" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- create "procedure_editors" table
CREATE TABLE "procedure_editors" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"));
-- create "procedure_narratives" table
CREATE TABLE "procedure_narratives" ("procedure_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "narrative_id"));
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
-- create "program_subcontrols" table
CREATE TABLE "program_subcontrols" ("program_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("program_id", "subcontrol_id"));
-- create "program_controlobjectives" table
CREATE TABLE "program_controlobjectives" ("program_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_objective_id"));
-- create "program_policies" table
CREATE TABLE "program_policies" ("program_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("program_id", "internal_policy_id"));
-- create "program_procedures" table
CREATE TABLE "program_procedures" ("program_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("program_id", "procedure_id"));
-- create "program_risks" table
CREATE TABLE "program_risks" ("program_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("program_id", "risk_id"));
-- create "program_tasks" table
CREATE TABLE "program_tasks" ("program_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("program_id", "task_id"));
-- create "program_notes" table
CREATE TABLE "program_notes" ("program_id" character varying NOT NULL, "note_id" character varying NOT NULL, PRIMARY KEY ("program_id", "note_id"));
-- create "program_files" table
CREATE TABLE "program_files" ("program_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("program_id", "file_id"));
-- create "program_narratives" table
CREATE TABLE "program_narratives" ("program_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("program_id", "narrative_id"));
-- create "program_actionplans" table
CREATE TABLE "program_actionplans" ("program_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("program_id", "action_plan_id"));
-- create "risk_blocked_groups" table
CREATE TABLE "risk_blocked_groups" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create "risk_editors" table
CREATE TABLE "risk_editors" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create "risk_viewers" table
CREATE TABLE "risk_viewers" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"));
-- create "risk_actionplans" table
CREATE TABLE "risk_actionplans" ("risk_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "action_plan_id"));
-- create "standard_controlobjectives" table
CREATE TABLE "standard_controlobjectives" ("standard_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "control_objective_id"));
-- create "standard_controls" table
CREATE TABLE "standard_controls" ("standard_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "control_id"));
-- create "standard_actionplans" table
CREATE TABLE "standard_actionplans" ("standard_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "action_plan_id"));
-- create "standard_programs" table
CREATE TABLE "standard_programs" ("standard_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "program_id"));
-- create "subcontrol_tasks" table
CREATE TABLE "subcontrol_tasks" ("subcontrol_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "task_id"));
-- create "subscriber_events" table
CREATE TABLE "subscriber_events" ("subscriber_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("subscriber_id", "event_id"));
-- create "template_files" table
CREATE TABLE "template_files" ("template_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("template_id", "file_id"));
-- create "user_files" table
CREATE TABLE "user_files" ("user_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_id", "file_id"));
-- create "user_events" table
CREATE TABLE "user_events" ("user_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("user_id", "event_id"));
-- create "user_actionplans" table
CREATE TABLE "user_actionplans" ("user_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("user_id", "action_plan_id"));
-- create "user_subcontrols" table
CREATE TABLE "user_subcontrols" ("user_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("user_id", "subcontrol_id"));
-- create "user_setting_files" table
CREATE TABLE "user_setting_files" ("user_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_setting_id", "file_id"));
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD CONSTRAINT "api_tokens_organizations_api_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ADD CONSTRAINT "contacts_organizations_contacts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_control_objectives_controls" FOREIGN KEY ("control_objective_controls") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_internal_policies_controls" FOREIGN KEY ("internal_policy_controls") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD CONSTRAINT "control_objectives_controls_controlobjectives" FOREIGN KEY ("control_controlobjectives") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_objectives_organizations_controlobjectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "document_data" table
ALTER TABLE "document_data" ADD CONSTRAINT "document_data_organizations_documentdata" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD CONSTRAINT "email_verification_tokens_users_email_verification_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_organizations_entities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" ADD CONSTRAINT "entity_types_organizations_entitytypes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "group_memberships" table
ALTER TABLE "group_memberships" ADD CONSTRAINT "group_memberships_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "group_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "group_settings" table
ALTER TABLE "group_settings" ADD CONSTRAINT "group_settings_groups_setting" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_groups_integrations" FOREIGN KEY ("group_integrations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_organizations_integrations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD CONSTRAINT "internal_policies_organizations_internalpolicies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "invites" table
ALTER TABLE "invites" ADD CONSTRAINT "invites_organizations_invites" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "notes" table
ALTER TABLE "notes" ADD CONSTRAINT "notes_entities_notes" FOREIGN KEY ("entity_notes") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_organizations_notes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_memberships" table
ALTER TABLE "org_memberships" ADD CONSTRAINT "org_memberships_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "org_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD CONSTRAINT "org_subscriptions_organizations_orgsubscriptions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD CONSTRAINT "organization_settings_organizations_setting" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" ADD CONSTRAINT "password_reset_tokens_users_password_reset_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD CONSTRAINT "personal_access_tokens_users_personal_access_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "procedures" table
ALTER TABLE "procedures" ADD CONSTRAINT "procedures_control_objectives_procedures" FOREIGN KEY ("control_objective_procedures") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_organizations_procedures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_standards_procedures" FOREIGN KEY ("standard_procedures") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD CONSTRAINT "programs_organizations_programs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" ADD CONSTRAINT "program_memberships_programs_program" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "program_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_control_objectives_risks" FOREIGN KEY ("control_objective_risks") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_control_objectives_subcontrols" FOREIGN KEY ("control_objective_subcontrols") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_notes_subcontrols" FOREIGN KEY ("note_subcontrols") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD CONSTRAINT "subscribers_organizations_subscribers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD CONSTRAINT "tfa_settings_users_tfa_settings" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("user_assignee_tasks") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("user_assigner_tasks") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_organizations_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_files_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "user_settings" table
ALTER TABLE "user_settings" ADD CONSTRAINT "user_settings_organizations_default_org" FOREIGN KEY ("user_setting_default_org") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_settings_users_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "webauthns" table
ALTER TABLE "webauthns" ADD CONSTRAINT "webauthns_users_webauthn" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "contact_files" table
ALTER TABLE "contact_files" ADD CONSTRAINT "contact_files_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "contact_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" ADD CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_editors" table
ALTER TABLE "control_editors" ADD CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_viewers" table
ALTER TABLE "control_viewers" ADD CONSTRAINT "control_viewers_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_procedures" table
ALTER TABLE "control_procedures" ADD CONSTRAINT "control_procedures_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_subcontrols" table
ALTER TABLE "control_subcontrols" ADD CONSTRAINT "control_subcontrols_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_narratives" table
ALTER TABLE "control_narratives" ADD CONSTRAINT "control_narratives_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_risks" table
ALTER TABLE "control_risks" ADD CONSTRAINT "control_risks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_actionplans" table
ALTER TABLE "control_actionplans" ADD CONSTRAINT "control_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_actionplans_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_tasks" table
ALTER TABLE "control_tasks" ADD CONSTRAINT "control_tasks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_blocked_groups" table
ALTER TABLE "control_objective_blocked_groups" ADD CONSTRAINT "control_objective_blocked_groups_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_editors" table
ALTER TABLE "control_objective_editors" ADD CONSTRAINT "control_objective_editors_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_viewers" table
ALTER TABLE "control_objective_viewers" ADD CONSTRAINT "control_objective_viewers_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "control_objective_narratives" table
ALTER TABLE "control_objective_narratives" ADD CONSTRAINT "control_objective_narratives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_objective_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "internal_policy_controlobjectives" table
ALTER TABLE "internal_policy_controlobjectives" ADD CONSTRAINT "internal_policy_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_controlobjectives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" ADD CONSTRAINT "internal_policy_procedures_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" ADD CONSTRAINT "internal_policy_narratives_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" ADD CONSTRAINT "internal_policy_tasks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "internal_policy_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "invite_events" table
ALTER TABLE "invite_events" ADD CONSTRAINT "invite_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "invite_events_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" ADD CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_editors" table
ALTER TABLE "narrative_editors" ADD CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" ADD CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_membership_events" table
ALTER TABLE "org_membership_events" ADD CONSTRAINT "org_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_membership_events_org_membership_id" FOREIGN KEY ("org_membership_id") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_control_creators" table
ALTER TABLE "organization_control_creators" ADD CONSTRAINT "organization_control_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_control_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_control_objective_creators" table
ALTER TABLE "organization_control_objective_creators" ADD CONSTRAINT "organization_control_objective_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_control_objective_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_group_creators" table
ALTER TABLE "organization_group_creators" ADD CONSTRAINT "organization_group_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_group_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_internal_policy_creators" table
ALTER TABLE "organization_internal_policy_creators" ADD CONSTRAINT "organization_internal_policy_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_internal_policy_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_narrative_creators" table
ALTER TABLE "organization_narrative_creators" ADD CONSTRAINT "organization_narrative_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_narrative_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_procedure_creators" table
ALTER TABLE "organization_procedure_creators" ADD CONSTRAINT "organization_procedure_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_procedure_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_program_creators" table
ALTER TABLE "organization_program_creators" ADD CONSTRAINT "organization_program_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_program_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_risk_creators" table
ALTER TABLE "organization_risk_creators" ADD CONSTRAINT "organization_risk_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_risk_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_template_creators" table
ALTER TABLE "organization_template_creators" ADD CONSTRAINT "organization_template_creators_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_template_creators_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_personal_access_tokens" table
ALTER TABLE "organization_personal_access_tokens" ADD CONSTRAINT "organization_personal_access_tokens_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_personal_access_tokens_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_events" table
ALTER TABLE "organization_events" ADD CONSTRAINT "organization_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_events_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_secrets" table
ALTER TABLE "organization_secrets" ADD CONSTRAINT "organization_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_secrets_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_files" table
ALTER TABLE "organization_files" ADD CONSTRAINT "organization_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_files_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_tasks" table
ALTER TABLE "organization_tasks" ADD CONSTRAINT "organization_tasks_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "organization_setting_files" table
ALTER TABLE "organization_setting_files" ADD CONSTRAINT "organization_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "organization_setting_files_organization_setting_id" FOREIGN KEY ("organization_setting_id") REFERENCES "organization_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "personal_access_token_events" table
ALTER TABLE "personal_access_token_events" ADD CONSTRAINT "personal_access_token_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "personal_access_token_events_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "program_subcontrols" table
ALTER TABLE "program_subcontrols" ADD CONSTRAINT "program_subcontrols_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_controlobjectives" table
ALTER TABLE "program_controlobjectives" ADD CONSTRAINT "program_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_controlobjectives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_policies" table
ALTER TABLE "program_policies" ADD CONSTRAINT "program_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_policies_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_procedures" table
ALTER TABLE "program_procedures" ADD CONSTRAINT "program_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_procedures_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_risks" table
ALTER TABLE "program_risks" ADD CONSTRAINT "program_risks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_tasks" table
ALTER TABLE "program_tasks" ADD CONSTRAINT "program_tasks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_notes" table
ALTER TABLE "program_notes" ADD CONSTRAINT "program_notes_note_id" FOREIGN KEY ("note_id") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_notes_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_files" table
ALTER TABLE "program_files" ADD CONSTRAINT "program_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_files_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_narratives" table
ALTER TABLE "program_narratives" ADD CONSTRAINT "program_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_narratives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "program_actionplans" table
ALTER TABLE "program_actionplans" ADD CONSTRAINT "program_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "program_actionplans_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" ADD CONSTRAINT "risk_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_blocked_groups_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_editors" table
ALTER TABLE "risk_editors" ADD CONSTRAINT "risk_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_editors_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_viewers" table
ALTER TABLE "risk_viewers" ADD CONSTRAINT "risk_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_viewers_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "risk_actionplans" table
ALTER TABLE "risk_actionplans" ADD CONSTRAINT "risk_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "risk_actionplans_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_controlobjectives" table
ALTER TABLE "standard_controlobjectives" ADD CONSTRAINT "standard_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_controlobjectives_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_controls" table
ALTER TABLE "standard_controls" ADD CONSTRAINT "standard_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_controls_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_actionplans" table
ALTER TABLE "standard_actionplans" ADD CONSTRAINT "standard_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_actionplans_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "standard_programs" table
ALTER TABLE "standard_programs" ADD CONSTRAINT "standard_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "standard_programs_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" ADD CONSTRAINT "subcontrol_tasks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subcontrol_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "subscriber_events" table
ALTER TABLE "subscriber_events" ADD CONSTRAINT "subscriber_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "subscriber_events_subscriber_id" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "template_files" table
ALTER TABLE "template_files" ADD CONSTRAINT "template_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "template_files_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_files" table
ALTER TABLE "user_files" ADD CONSTRAINT "user_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_files_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_events" table
ALTER TABLE "user_events" ADD CONSTRAINT "user_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_actionplans" table
ALTER TABLE "user_actionplans" ADD CONSTRAINT "user_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_actionplans_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_subcontrols" table
ALTER TABLE "user_subcontrols" ADD CONSTRAINT "user_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_subcontrols_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_setting_files" table
ALTER TABLE "user_setting_files" ADD CONSTRAINT "user_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_setting_files_user_setting_id" FOREIGN KEY ("user_setting_id") REFERENCES "user_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "user_setting_files" table
ALTER TABLE "user_setting_files" DROP CONSTRAINT "user_setting_files_user_setting_id", DROP CONSTRAINT "user_setting_files_file_id";
-- reverse: modify "user_subcontrols" table
ALTER TABLE "user_subcontrols" DROP CONSTRAINT "user_subcontrols_user_id", DROP CONSTRAINT "user_subcontrols_subcontrol_id";
-- reverse: modify "user_actionplans" table
ALTER TABLE "user_actionplans" DROP CONSTRAINT "user_actionplans_user_id", DROP CONSTRAINT "user_actionplans_action_plan_id";
-- reverse: modify "user_events" table
ALTER TABLE "user_events" DROP CONSTRAINT "user_events_user_id", DROP CONSTRAINT "user_events_event_id";
-- reverse: modify "user_files" table
ALTER TABLE "user_files" DROP CONSTRAINT "user_files_user_id", DROP CONSTRAINT "user_files_file_id";
-- reverse: modify "template_files" table
ALTER TABLE "template_files" DROP CONSTRAINT "template_files_template_id", DROP CONSTRAINT "template_files_file_id";
-- reverse: modify "subscriber_events" table
ALTER TABLE "subscriber_events" DROP CONSTRAINT "subscriber_events_subscriber_id", DROP CONSTRAINT "subscriber_events_event_id";
-- reverse: modify "subcontrol_tasks" table
ALTER TABLE "subcontrol_tasks" DROP CONSTRAINT "subcontrol_tasks_task_id", DROP CONSTRAINT "subcontrol_tasks_subcontrol_id";
-- reverse: modify "standard_programs" table
ALTER TABLE "standard_programs" DROP CONSTRAINT "standard_programs_standard_id", DROP CONSTRAINT "standard_programs_program_id";
-- reverse: modify "standard_actionplans" table
ALTER TABLE "standard_actionplans" DROP CONSTRAINT "standard_actionplans_standard_id", DROP CONSTRAINT "standard_actionplans_action_plan_id";
-- reverse: modify "standard_controls" table
ALTER TABLE "standard_controls" DROP CONSTRAINT "standard_controls_standard_id", DROP CONSTRAINT "standard_controls_control_id";
-- reverse: modify "standard_controlobjectives" table
ALTER TABLE "standard_controlobjectives" DROP CONSTRAINT "standard_controlobjectives_standard_id", DROP CONSTRAINT "standard_controlobjectives_control_objective_id";
-- reverse: modify "risk_actionplans" table
ALTER TABLE "risk_actionplans" DROP CONSTRAINT "risk_actionplans_risk_id", DROP CONSTRAINT "risk_actionplans_action_plan_id";
-- reverse: modify "risk_viewers" table
ALTER TABLE "risk_viewers" DROP CONSTRAINT "risk_viewers_risk_id", DROP CONSTRAINT "risk_viewers_group_id";
-- reverse: modify "risk_editors" table
ALTER TABLE "risk_editors" DROP CONSTRAINT "risk_editors_risk_id", DROP CONSTRAINT "risk_editors_group_id";
-- reverse: modify "risk_blocked_groups" table
ALTER TABLE "risk_blocked_groups" DROP CONSTRAINT "risk_blocked_groups_risk_id", DROP CONSTRAINT "risk_blocked_groups_group_id";
-- reverse: modify "program_actionplans" table
ALTER TABLE "program_actionplans" DROP CONSTRAINT "program_actionplans_program_id", DROP CONSTRAINT "program_actionplans_action_plan_id";
-- reverse: modify "program_narratives" table
ALTER TABLE "program_narratives" DROP CONSTRAINT "program_narratives_program_id", DROP CONSTRAINT "program_narratives_narrative_id";
-- reverse: modify "program_files" table
ALTER TABLE "program_files" DROP CONSTRAINT "program_files_program_id", DROP CONSTRAINT "program_files_file_id";
-- reverse: modify "program_notes" table
ALTER TABLE "program_notes" DROP CONSTRAINT "program_notes_program_id", DROP CONSTRAINT "program_notes_note_id";
-- reverse: modify "program_tasks" table
ALTER TABLE "program_tasks" DROP CONSTRAINT "program_tasks_task_id", DROP CONSTRAINT "program_tasks_program_id";
-- reverse: modify "program_risks" table
ALTER TABLE "program_risks" DROP CONSTRAINT "program_risks_risk_id", DROP CONSTRAINT "program_risks_program_id";
-- reverse: modify "program_procedures" table
ALTER TABLE "program_procedures" DROP CONSTRAINT "program_procedures_program_id", DROP CONSTRAINT "program_procedures_procedure_id";
-- reverse: modify "program_policies" table
ALTER TABLE "program_policies" DROP CONSTRAINT "program_policies_program_id", DROP CONSTRAINT "program_policies_internal_policy_id";
-- reverse: modify "program_controlobjectives" table
ALTER TABLE "program_controlobjectives" DROP CONSTRAINT "program_controlobjectives_program_id", DROP CONSTRAINT "program_controlobjectives_control_objective_id";
-- reverse: modify "program_subcontrols" table
ALTER TABLE "program_subcontrols" DROP CONSTRAINT "program_subcontrols_subcontrol_id", DROP CONSTRAINT "program_subcontrols_program_id";
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
-- reverse: modify "personal_access_token_events" table
ALTER TABLE "personal_access_token_events" DROP CONSTRAINT "personal_access_token_events_personal_access_token_id", DROP CONSTRAINT "personal_access_token_events_event_id";
-- reverse: modify "organization_setting_files" table
ALTER TABLE "organization_setting_files" DROP CONSTRAINT "organization_setting_files_organization_setting_id", DROP CONSTRAINT "organization_setting_files_file_id";
-- reverse: modify "organization_tasks" table
ALTER TABLE "organization_tasks" DROP CONSTRAINT "organization_tasks_task_id", DROP CONSTRAINT "organization_tasks_organization_id";
-- reverse: modify "organization_files" table
ALTER TABLE "organization_files" DROP CONSTRAINT "organization_files_organization_id", DROP CONSTRAINT "organization_files_file_id";
-- reverse: modify "organization_secrets" table
ALTER TABLE "organization_secrets" DROP CONSTRAINT "organization_secrets_organization_id", DROP CONSTRAINT "organization_secrets_hush_id";
-- reverse: modify "organization_events" table
ALTER TABLE "organization_events" DROP CONSTRAINT "organization_events_organization_id", DROP CONSTRAINT "organization_events_event_id";
-- reverse: modify "organization_personal_access_tokens" table
ALTER TABLE "organization_personal_access_tokens" DROP CONSTRAINT "organization_personal_access_tokens_personal_access_token_id", DROP CONSTRAINT "organization_personal_access_tokens_organization_id";
-- reverse: modify "organization_template_creators" table
ALTER TABLE "organization_template_creators" DROP CONSTRAINT "organization_template_creators_organization_id", DROP CONSTRAINT "organization_template_creators_group_id";
-- reverse: modify "organization_risk_creators" table
ALTER TABLE "organization_risk_creators" DROP CONSTRAINT "organization_risk_creators_organization_id", DROP CONSTRAINT "organization_risk_creators_group_id";
-- reverse: modify "organization_program_creators" table
ALTER TABLE "organization_program_creators" DROP CONSTRAINT "organization_program_creators_organization_id", DROP CONSTRAINT "organization_program_creators_group_id";
-- reverse: modify "organization_procedure_creators" table
ALTER TABLE "organization_procedure_creators" DROP CONSTRAINT "organization_procedure_creators_organization_id", DROP CONSTRAINT "organization_procedure_creators_group_id";
-- reverse: modify "organization_narrative_creators" table
ALTER TABLE "organization_narrative_creators" DROP CONSTRAINT "organization_narrative_creators_organization_id", DROP CONSTRAINT "organization_narrative_creators_group_id";
-- reverse: modify "organization_internal_policy_creators" table
ALTER TABLE "organization_internal_policy_creators" DROP CONSTRAINT "organization_internal_policy_creators_organization_id", DROP CONSTRAINT "organization_internal_policy_creators_group_id";
-- reverse: modify "organization_group_creators" table
ALTER TABLE "organization_group_creators" DROP CONSTRAINT "organization_group_creators_organization_id", DROP CONSTRAINT "organization_group_creators_group_id";
-- reverse: modify "organization_control_objective_creators" table
ALTER TABLE "organization_control_objective_creators" DROP CONSTRAINT "organization_control_objective_creators_organization_id", DROP CONSTRAINT "organization_control_objective_creators_group_id";
-- reverse: modify "organization_control_creators" table
ALTER TABLE "organization_control_creators" DROP CONSTRAINT "organization_control_creators_organization_id", DROP CONSTRAINT "organization_control_creators_group_id";
-- reverse: modify "org_membership_events" table
ALTER TABLE "org_membership_events" DROP CONSTRAINT "org_membership_events_org_membership_id", DROP CONSTRAINT "org_membership_events_event_id";
-- reverse: modify "narrative_viewers" table
ALTER TABLE "narrative_viewers" DROP CONSTRAINT "narrative_viewers_narrative_id", DROP CONSTRAINT "narrative_viewers_group_id";
-- reverse: modify "narrative_editors" table
ALTER TABLE "narrative_editors" DROP CONSTRAINT "narrative_editors_narrative_id", DROP CONSTRAINT "narrative_editors_group_id";
-- reverse: modify "narrative_blocked_groups" table
ALTER TABLE "narrative_blocked_groups" DROP CONSTRAINT "narrative_blocked_groups_narrative_id", DROP CONSTRAINT "narrative_blocked_groups_group_id";
-- reverse: modify "invite_events" table
ALTER TABLE "invite_events" DROP CONSTRAINT "invite_events_invite_id", DROP CONSTRAINT "invite_events_event_id";
-- reverse: modify "internal_policy_tasks" table
ALTER TABLE "internal_policy_tasks" DROP CONSTRAINT "internal_policy_tasks_task_id", DROP CONSTRAINT "internal_policy_tasks_internal_policy_id";
-- reverse: modify "internal_policy_narratives" table
ALTER TABLE "internal_policy_narratives" DROP CONSTRAINT "internal_policy_narratives_narrative_id", DROP CONSTRAINT "internal_policy_narratives_internal_policy_id";
-- reverse: modify "internal_policy_procedures" table
ALTER TABLE "internal_policy_procedures" DROP CONSTRAINT "internal_policy_procedures_procedure_id", DROP CONSTRAINT "internal_policy_procedures_internal_policy_id";
-- reverse: modify "internal_policy_controlobjectives" table
ALTER TABLE "internal_policy_controlobjectives" DROP CONSTRAINT "internal_policy_controlobjectives_internal_policy_id", DROP CONSTRAINT "internal_policy_controlobjectives_control_objective_id";
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
-- reverse: modify "control_objective_narratives" table
ALTER TABLE "control_objective_narratives" DROP CONSTRAINT "control_objective_narratives_narrative_id", DROP CONSTRAINT "control_objective_narratives_control_objective_id";
-- reverse: modify "control_objective_viewers" table
ALTER TABLE "control_objective_viewers" DROP CONSTRAINT "control_objective_viewers_group_id", DROP CONSTRAINT "control_objective_viewers_control_objective_id";
-- reverse: modify "control_objective_editors" table
ALTER TABLE "control_objective_editors" DROP CONSTRAINT "control_objective_editors_group_id", DROP CONSTRAINT "control_objective_editors_control_objective_id";
-- reverse: modify "control_objective_blocked_groups" table
ALTER TABLE "control_objective_blocked_groups" DROP CONSTRAINT "control_objective_blocked_groups_group_id", DROP CONSTRAINT "control_objective_blocked_groups_control_objective_id";
-- reverse: modify "control_tasks" table
ALTER TABLE "control_tasks" DROP CONSTRAINT "control_tasks_task_id", DROP CONSTRAINT "control_tasks_control_id";
-- reverse: modify "control_actionplans" table
ALTER TABLE "control_actionplans" DROP CONSTRAINT "control_actionplans_control_id", DROP CONSTRAINT "control_actionplans_action_plan_id";
-- reverse: modify "control_risks" table
ALTER TABLE "control_risks" DROP CONSTRAINT "control_risks_risk_id", DROP CONSTRAINT "control_risks_control_id";
-- reverse: modify "control_narratives" table
ALTER TABLE "control_narratives" DROP CONSTRAINT "control_narratives_narrative_id", DROP CONSTRAINT "control_narratives_control_id";
-- reverse: modify "control_subcontrols" table
ALTER TABLE "control_subcontrols" DROP CONSTRAINT "control_subcontrols_subcontrol_id", DROP CONSTRAINT "control_subcontrols_control_id";
-- reverse: modify "control_procedures" table
ALTER TABLE "control_procedures" DROP CONSTRAINT "control_procedures_procedure_id", DROP CONSTRAINT "control_procedures_control_id";
-- reverse: modify "control_viewers" table
ALTER TABLE "control_viewers" DROP CONSTRAINT "control_viewers_group_id", DROP CONSTRAINT "control_viewers_control_id";
-- reverse: modify "control_editors" table
ALTER TABLE "control_editors" DROP CONSTRAINT "control_editors_group_id", DROP CONSTRAINT "control_editors_control_id";
-- reverse: modify "control_blocked_groups" table
ALTER TABLE "control_blocked_groups" DROP CONSTRAINT "control_blocked_groups_group_id", DROP CONSTRAINT "control_blocked_groups_control_id";
-- reverse: modify "contact_files" table
ALTER TABLE "contact_files" DROP CONSTRAINT "contact_files_file_id", DROP CONSTRAINT "contact_files_contact_id";
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_webauthn";
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP CONSTRAINT "user_settings_users_setting", DROP CONSTRAINT "user_settings_organizations_default_org";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_file";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP CONSTRAINT "templates_organizations_templates";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_assigner_tasks", DROP CONSTRAINT "tasks_users_assignee_tasks";
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP CONSTRAINT "tfa_settings_users_tfa_settings";
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP CONSTRAINT "subscribers_organizations_subscribers";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP CONSTRAINT "subcontrols_notes_subcontrols", DROP CONSTRAINT "subcontrols_control_objectives_subcontrols";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_organizations_risks", DROP CONSTRAINT "risks_control_objectives_risks";
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP CONSTRAINT "program_memberships_users_user", DROP CONSTRAINT "program_memberships_programs_program";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_organizations_programs";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_standards_procedures", DROP CONSTRAINT "procedures_organizations_procedures", DROP CONSTRAINT "procedures_control_objectives_procedures";
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP CONSTRAINT "personal_access_tokens_users_personal_access_tokens";
-- reverse: modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP CONSTRAINT "password_reset_tokens_users_password_reset_tokens";
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP CONSTRAINT "organization_settings_organizations_setting";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP CONSTRAINT "org_subscriptions_organizations_orgsubscriptions";
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP CONSTRAINT "org_memberships_users_user", DROP CONSTRAINT "org_memberships_organizations_organization";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_organizations_notes", DROP CONSTRAINT "notes_entities_notes";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_organizations_narratives";
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP CONSTRAINT "invites_organizations_invites";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_organizations_internalpolicies";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_organizations_integrations", DROP CONSTRAINT "integrations_groups_integrations";
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" DROP CONSTRAINT "group_settings_groups_setting";
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP CONSTRAINT "group_memberships_users_user", DROP CONSTRAINT "group_memberships_groups_group";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_organizations_groups";
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP CONSTRAINT "entity_types_organizations_entitytypes";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_organizations_entities", DROP CONSTRAINT "entities_entity_types_entity_type", DROP CONSTRAINT "entities_entity_types_entities";
-- reverse: modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP CONSTRAINT "email_verification_tokens_users_email_verification_tokens";
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_templates_documents", DROP CONSTRAINT "document_data_organizations_documentdata";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_organizations_controlobjectives", DROP CONSTRAINT "control_objectives_controls_controlobjectives";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_organizations_controls", DROP CONSTRAINT "controls_internal_policies_controls", DROP CONSTRAINT "controls_control_objectives_controls";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP CONSTRAINT "contacts_organizations_contacts";
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP CONSTRAINT "api_tokens_organizations_api_tokens";
-- reverse: create "user_setting_files" table
DROP TABLE "user_setting_files";
-- reverse: create "user_subcontrols" table
DROP TABLE "user_subcontrols";
-- reverse: create "user_actionplans" table
DROP TABLE "user_actionplans";
-- reverse: create "user_events" table
DROP TABLE "user_events";
-- reverse: create "user_files" table
DROP TABLE "user_files";
-- reverse: create "template_files" table
DROP TABLE "template_files";
-- reverse: create "subscriber_events" table
DROP TABLE "subscriber_events";
-- reverse: create "subcontrol_tasks" table
DROP TABLE "subcontrol_tasks";
-- reverse: create "standard_programs" table
DROP TABLE "standard_programs";
-- reverse: create "standard_actionplans" table
DROP TABLE "standard_actionplans";
-- reverse: create "standard_controls" table
DROP TABLE "standard_controls";
-- reverse: create "standard_controlobjectives" table
DROP TABLE "standard_controlobjectives";
-- reverse: create "risk_actionplans" table
DROP TABLE "risk_actionplans";
-- reverse: create "risk_viewers" table
DROP TABLE "risk_viewers";
-- reverse: create "risk_editors" table
DROP TABLE "risk_editors";
-- reverse: create "risk_blocked_groups" table
DROP TABLE "risk_blocked_groups";
-- reverse: create "program_actionplans" table
DROP TABLE "program_actionplans";
-- reverse: create "program_narratives" table
DROP TABLE "program_narratives";
-- reverse: create "program_files" table
DROP TABLE "program_files";
-- reverse: create "program_notes" table
DROP TABLE "program_notes";
-- reverse: create "program_tasks" table
DROP TABLE "program_tasks";
-- reverse: create "program_risks" table
DROP TABLE "program_risks";
-- reverse: create "program_procedures" table
DROP TABLE "program_procedures";
-- reverse: create "program_policies" table
DROP TABLE "program_policies";
-- reverse: create "program_controlobjectives" table
DROP TABLE "program_controlobjectives";
-- reverse: create "program_subcontrols" table
DROP TABLE "program_subcontrols";
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
-- reverse: create "procedure_narratives" table
DROP TABLE "procedure_narratives";
-- reverse: create "procedure_editors" table
DROP TABLE "procedure_editors";
-- reverse: create "procedure_blocked_groups" table
DROP TABLE "procedure_blocked_groups";
-- reverse: create "personal_access_token_events" table
DROP TABLE "personal_access_token_events";
-- reverse: create "organization_setting_files" table
DROP TABLE "organization_setting_files";
-- reverse: create "organization_tasks" table
DROP TABLE "organization_tasks";
-- reverse: create "organization_files" table
DROP TABLE "organization_files";
-- reverse: create "organization_secrets" table
DROP TABLE "organization_secrets";
-- reverse: create "organization_events" table
DROP TABLE "organization_events";
-- reverse: create "organization_personal_access_tokens" table
DROP TABLE "organization_personal_access_tokens";
-- reverse: create "organization_template_creators" table
DROP TABLE "organization_template_creators";
-- reverse: create "organization_risk_creators" table
DROP TABLE "organization_risk_creators";
-- reverse: create "organization_program_creators" table
DROP TABLE "organization_program_creators";
-- reverse: create "organization_procedure_creators" table
DROP TABLE "organization_procedure_creators";
-- reverse: create "organization_narrative_creators" table
DROP TABLE "organization_narrative_creators";
-- reverse: create "organization_internal_policy_creators" table
DROP TABLE "organization_internal_policy_creators";
-- reverse: create "organization_group_creators" table
DROP TABLE "organization_group_creators";
-- reverse: create "organization_control_objective_creators" table
DROP TABLE "organization_control_objective_creators";
-- reverse: create "organization_control_creators" table
DROP TABLE "organization_control_creators";
-- reverse: create "org_membership_events" table
DROP TABLE "org_membership_events";
-- reverse: create "narrative_viewers" table
DROP TABLE "narrative_viewers";
-- reverse: create "narrative_editors" table
DROP TABLE "narrative_editors";
-- reverse: create "narrative_blocked_groups" table
DROP TABLE "narrative_blocked_groups";
-- reverse: create "invite_events" table
DROP TABLE "invite_events";
-- reverse: create "internal_policy_tasks" table
DROP TABLE "internal_policy_tasks";
-- reverse: create "internal_policy_narratives" table
DROP TABLE "internal_policy_narratives";
-- reverse: create "internal_policy_procedures" table
DROP TABLE "internal_policy_procedures";
-- reverse: create "internal_policy_controlobjectives" table
DROP TABLE "internal_policy_controlobjectives";
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
-- reverse: create "control_objective_narratives" table
DROP TABLE "control_objective_narratives";
-- reverse: create "control_objective_viewers" table
DROP TABLE "control_objective_viewers";
-- reverse: create "control_objective_editors" table
DROP TABLE "control_objective_editors";
-- reverse: create "control_objective_blocked_groups" table
DROP TABLE "control_objective_blocked_groups";
-- reverse: create "control_tasks" table
DROP TABLE "control_tasks";
-- reverse: create "control_actionplans" table
DROP TABLE "control_actionplans";
-- reverse: create "control_risks" table
DROP TABLE "control_risks";
-- reverse: create "control_narratives" table
DROP TABLE "control_narratives";
-- reverse: create "control_subcontrols" table
DROP TABLE "control_subcontrols";
-- reverse: create "control_procedures" table
DROP TABLE "control_procedures";
-- reverse: create "control_viewers" table
DROP TABLE "control_viewers";
-- reverse: create "control_editors" table
DROP TABLE "control_editors";
-- reverse: create "control_blocked_groups" table
DROP TABLE "control_blocked_groups";
-- reverse: create "contact_files" table
DROP TABLE "contact_files";
-- reverse: create index "webauthns_mapping_id_key" to table: "webauthns"
DROP INDEX "webauthns_mapping_id_key";
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
-- reverse: create index "user_settings_mapping_id_key" to table: "user_settings"
DROP INDEX "user_settings_mapping_id_key";
-- reverse: create "user_settings" table
DROP TABLE "user_settings";
-- reverse: create index "userhistory_history_time" to table: "user_history"
DROP INDEX "userhistory_history_time";
-- reverse: create "user_history" table
DROP TABLE "user_history";
-- reverse: create index "users_sub_key" to table: "users"
DROP INDEX "users_sub_key";
-- reverse: create index "users_mapping_id_key" to table: "users"
DROP INDEX "users_mapping_id_key";
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
-- reverse: create index "templates_mapping_id_key" to table: "templates"
DROP INDEX "templates_mapping_id_key";
-- reverse: create index "template_name_owner_id_template_type" to table: "templates"
DROP INDEX "template_name_owner_id_template_type";
-- reverse: create "templates" table
DROP TABLE "templates";
-- reverse: create index "taskhistory_history_time" to table: "task_history"
DROP INDEX "taskhistory_history_time";
-- reverse: create "task_history" table
DROP TABLE "task_history";
-- reverse: create index "tasks_mapping_id_key" to table: "tasks"
DROP INDEX "tasks_mapping_id_key";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "tfasetting_owner_id" to table: "tfa_settings"
DROP INDEX "tfasetting_owner_id";
-- reverse: create index "tfa_settings_mapping_id_key" to table: "tfa_settings"
DROP INDEX "tfa_settings_mapping_id_key";
-- reverse: create "tfa_settings" table
DROP TABLE "tfa_settings";
-- reverse: create index "subscribers_token_key" to table: "subscribers"
DROP INDEX "subscribers_token_key";
-- reverse: create index "subscribers_mapping_id_key" to table: "subscribers"
DROP INDEX "subscribers_mapping_id_key";
-- reverse: create index "subscriber_email_owner_id" to table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- reverse: create "subscribers" table
DROP TABLE "subscribers";
-- reverse: create index "subcontrolhistory_history_time" to table: "subcontrol_history"
DROP INDEX "subcontrolhistory_history_time";
-- reverse: create "subcontrol_history" table
DROP TABLE "subcontrol_history";
-- reverse: create index "subcontrols_mapping_id_key" to table: "subcontrols"
DROP INDEX "subcontrols_mapping_id_key";
-- reverse: create "subcontrols" table
DROP TABLE "subcontrols";
-- reverse: create index "standardhistory_history_time" to table: "standard_history"
DROP INDEX "standardhistory_history_time";
-- reverse: create "standard_history" table
DROP TABLE "standard_history";
-- reverse: create index "standards_mapping_id_key" to table: "standards"
DROP INDEX "standards_mapping_id_key";
-- reverse: create "standards" table
DROP TABLE "standards";
-- reverse: create index "riskhistory_history_time" to table: "risk_history"
DROP INDEX "riskhistory_history_time";
-- reverse: create "risk_history" table
DROP TABLE "risk_history";
-- reverse: create index "risks_mapping_id_key" to table: "risks"
DROP INDEX "risks_mapping_id_key";
-- reverse: create "risks" table
DROP TABLE "risks";
-- reverse: create index "programmembershiphistory_history_time" to table: "program_membership_history"
DROP INDEX "programmembershiphistory_history_time";
-- reverse: create "program_membership_history" table
DROP TABLE "program_membership_history";
-- reverse: create index "programmembership_user_id_program_id" to table: "program_memberships"
DROP INDEX "programmembership_user_id_program_id";
-- reverse: create index "program_memberships_mapping_id_key" to table: "program_memberships"
DROP INDEX "program_memberships_mapping_id_key";
-- reverse: create "program_memberships" table
DROP TABLE "program_memberships";
-- reverse: create index "programhistory_history_time" to table: "program_history"
DROP INDEX "programhistory_history_time";
-- reverse: create "program_history" table
DROP TABLE "program_history";
-- reverse: create index "programs_mapping_id_key" to table: "programs"
DROP INDEX "programs_mapping_id_key";
-- reverse: create "programs" table
DROP TABLE "programs";
-- reverse: create index "procedurehistory_history_time" to table: "procedure_history"
DROP INDEX "procedurehistory_history_time";
-- reverse: create "procedure_history" table
DROP TABLE "procedure_history";
-- reverse: create index "procedures_mapping_id_key" to table: "procedures"
DROP INDEX "procedures_mapping_id_key";
-- reverse: create "procedures" table
DROP TABLE "procedures";
-- reverse: create index "personalaccesstoken_token" to table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_token";
-- reverse: create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_token_key";
-- reverse: create index "personal_access_tokens_mapping_id_key" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_mapping_id_key";
-- reverse: create "personal_access_tokens" table
DROP TABLE "personal_access_tokens";
-- reverse: create index "passwordresettoken_token" to table: "password_reset_tokens"
DROP INDEX "passwordresettoken_token";
-- reverse: create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_token_key";
-- reverse: create index "password_reset_tokens_mapping_id_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_mapping_id_key";
-- reverse: create "password_reset_tokens" table
DROP TABLE "password_reset_tokens";
-- reverse: create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
DROP INDEX "organizationsettinghistory_history_time";
-- reverse: create "organization_setting_history" table
DROP TABLE "organization_setting_history";
-- reverse: create index "organization_settings_organization_id_key" to table: "organization_settings"
DROP INDEX "organization_settings_organization_id_key";
-- reverse: create index "organization_settings_mapping_id_key" to table: "organization_settings"
DROP INDEX "organization_settings_mapping_id_key";
-- reverse: create "organization_settings" table
DROP TABLE "organization_settings";
-- reverse: create index "organizationhistory_history_time" to table: "organization_history"
DROP INDEX "organizationhistory_history_time";
-- reverse: create "organization_history" table
DROP TABLE "organization_history";
-- reverse: create index "organizations_mapping_id_key" to table: "organizations"
DROP INDEX "organizations_mapping_id_key";
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
-- reverse: create index "org_subscriptions_mapping_id_key" to table: "org_subscriptions"
DROP INDEX "org_subscriptions_mapping_id_key";
-- reverse: create "org_subscriptions" table
DROP TABLE "org_subscriptions";
-- reverse: create index "orgmembershiphistory_history_time" to table: "org_membership_history"
DROP INDEX "orgmembershiphistory_history_time";
-- reverse: create "org_membership_history" table
DROP TABLE "org_membership_history";
-- reverse: create index "orgmembership_user_id_organization_id" to table: "org_memberships"
DROP INDEX "orgmembership_user_id_organization_id";
-- reverse: create index "org_memberships_mapping_id_key" to table: "org_memberships"
DROP INDEX "org_memberships_mapping_id_key";
-- reverse: create "org_memberships" table
DROP TABLE "org_memberships";
-- reverse: create index "notehistory_history_time" to table: "note_history"
DROP INDEX "notehistory_history_time";
-- reverse: create "note_history" table
DROP TABLE "note_history";
-- reverse: create index "notes_mapping_id_key" to table: "notes"
DROP INDEX "notes_mapping_id_key";
-- reverse: create "notes" table
DROP TABLE "notes";
-- reverse: create index "narrativehistory_history_time" to table: "narrative_history"
DROP INDEX "narrativehistory_history_time";
-- reverse: create "narrative_history" table
DROP TABLE "narrative_history";
-- reverse: create index "narratives_mapping_id_key" to table: "narratives"
DROP INDEX "narratives_mapping_id_key";
-- reverse: create "narratives" table
DROP TABLE "narratives";
-- reverse: create index "invites_token_key" to table: "invites"
DROP INDEX "invites_token_key";
-- reverse: create index "invites_mapping_id_key" to table: "invites"
DROP INDEX "invites_mapping_id_key";
-- reverse: create index "invite_recipient_owner_id" to table: "invites"
DROP INDEX "invite_recipient_owner_id";
-- reverse: create "invites" table
DROP TABLE "invites";
-- reverse: create index "internalpolicyhistory_history_time" to table: "internal_policy_history"
DROP INDEX "internalpolicyhistory_history_time";
-- reverse: create "internal_policy_history" table
DROP TABLE "internal_policy_history";
-- reverse: create index "internal_policies_mapping_id_key" to table: "internal_policies"
DROP INDEX "internal_policies_mapping_id_key";
-- reverse: create "internal_policies" table
DROP TABLE "internal_policies";
-- reverse: create index "integrationhistory_history_time" to table: "integration_history"
DROP INDEX "integrationhistory_history_time";
-- reverse: create "integration_history" table
DROP TABLE "integration_history";
-- reverse: create index "integrations_mapping_id_key" to table: "integrations"
DROP INDEX "integrations_mapping_id_key";
-- reverse: create "integrations" table
DROP TABLE "integrations";
-- reverse: create index "hushhistory_history_time" to table: "hush_history"
DROP INDEX "hushhistory_history_time";
-- reverse: create "hush_history" table
DROP TABLE "hush_history";
-- reverse: create index "hushes_mapping_id_key" to table: "hushes"
DROP INDEX "hushes_mapping_id_key";
-- reverse: create "hushes" table
DROP TABLE "hushes";
-- reverse: create index "groupsettinghistory_history_time" to table: "group_setting_history"
DROP INDEX "groupsettinghistory_history_time";
-- reverse: create "group_setting_history" table
DROP TABLE "group_setting_history";
-- reverse: create index "group_settings_mapping_id_key" to table: "group_settings"
DROP INDEX "group_settings_mapping_id_key";
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
-- reverse: create index "group_memberships_mapping_id_key" to table: "group_memberships"
DROP INDEX "group_memberships_mapping_id_key";
-- reverse: create "group_memberships" table
DROP TABLE "group_memberships";
-- reverse: create index "grouphistory_history_time" to table: "group_history"
DROP INDEX "grouphistory_history_time";
-- reverse: create "group_history" table
DROP TABLE "group_history";
-- reverse: create index "groups_mapping_id_key" to table: "groups"
DROP INDEX "groups_mapping_id_key";
-- reverse: create index "group_name_owner_id" to table: "groups"
DROP INDEX "group_name_owner_id";
-- reverse: create "groups" table
DROP TABLE "groups";
-- reverse: create index "filehistory_history_time" to table: "file_history"
DROP INDEX "filehistory_history_time";
-- reverse: create "file_history" table
DROP TABLE "file_history";
-- reverse: create index "files_mapping_id_key" to table: "files"
DROP INDEX "files_mapping_id_key";
-- reverse: create "files" table
DROP TABLE "files";
-- reverse: create index "eventhistory_history_time" to table: "event_history"
DROP INDEX "eventhistory_history_time";
-- reverse: create "event_history" table
DROP TABLE "event_history";
-- reverse: create index "events_mapping_id_key" to table: "events"
DROP INDEX "events_mapping_id_key";
-- reverse: create "events" table
DROP TABLE "events";
-- reverse: create index "entitytypehistory_history_time" to table: "entity_type_history"
DROP INDEX "entitytypehistory_history_time";
-- reverse: create "entity_type_history" table
DROP TABLE "entity_type_history";
-- reverse: create index "entitytype_name_owner_id" to table: "entity_types"
DROP INDEX "entitytype_name_owner_id";
-- reverse: create index "entity_types_mapping_id_key" to table: "entity_types"
DROP INDEX "entity_types_mapping_id_key";
-- reverse: create "entity_types" table
DROP TABLE "entity_types";
-- reverse: create index "entityhistory_history_time" to table: "entity_history"
DROP INDEX "entityhistory_history_time";
-- reverse: create "entity_history" table
DROP TABLE "entity_history";
-- reverse: create index "entity_name_owner_id" to table: "entities"
DROP INDEX "entity_name_owner_id";
-- reverse: create index "entities_mapping_id_key" to table: "entities"
DROP INDEX "entities_mapping_id_key";
-- reverse: create "entities" table
DROP TABLE "entities";
-- reverse: create index "emailverificationtoken_token" to table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_token";
-- reverse: create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_token_key";
-- reverse: create index "email_verification_tokens_mapping_id_key" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_mapping_id_key";
-- reverse: create "email_verification_tokens" table
DROP TABLE "email_verification_tokens";
-- reverse: create index "documentdatahistory_history_time" to table: "document_data_history"
DROP INDEX "documentdatahistory_history_time";
-- reverse: create "document_data_history" table
DROP TABLE "document_data_history";
-- reverse: create index "document_data_mapping_id_key" to table: "document_data"
DROP INDEX "document_data_mapping_id_key";
-- reverse: create "document_data" table
DROP TABLE "document_data";
-- reverse: create index "controlobjectivehistory_history_time" to table: "control_objective_history"
DROP INDEX "controlobjectivehistory_history_time";
-- reverse: create "control_objective_history" table
DROP TABLE "control_objective_history";
-- reverse: create index "control_objectives_mapping_id_key" to table: "control_objectives"
DROP INDEX "control_objectives_mapping_id_key";
-- reverse: create "control_objectives" table
DROP TABLE "control_objectives";
-- reverse: create index "controlhistory_history_time" to table: "control_history"
DROP INDEX "controlhistory_history_time";
-- reverse: create "control_history" table
DROP TABLE "control_history";
-- reverse: create index "controls_mapping_id_key" to table: "controls"
DROP INDEX "controls_mapping_id_key";
-- reverse: create "controls" table
DROP TABLE "controls";
-- reverse: create index "contacthistory_history_time" to table: "contact_history"
DROP INDEX "contacthistory_history_time";
-- reverse: create "contact_history" table
DROP TABLE "contact_history";
-- reverse: create index "contacts_mapping_id_key" to table: "contacts"
DROP INDEX "contacts_mapping_id_key";
-- reverse: create "contacts" table
DROP TABLE "contacts";
-- reverse: create index "actionplanhistory_history_time" to table: "action_plan_history"
DROP INDEX "actionplanhistory_history_time";
-- reverse: create "action_plan_history" table
DROP TABLE "action_plan_history";
-- reverse: create index "action_plans_mapping_id_key" to table: "action_plans"
DROP INDEX "action_plans_mapping_id_key";
-- reverse: create "action_plans" table
DROP TABLE "action_plans";
-- reverse: create index "apitoken_token" to table: "api_tokens"
DROP INDEX "apitoken_token";
-- reverse: create index "api_tokens_token_key" to table: "api_tokens"
DROP INDEX "api_tokens_token_key";
-- reverse: create index "api_tokens_mapping_id_key" to table: "api_tokens"
DROP INDEX "api_tokens_mapping_id_key";
-- reverse: create "api_tokens" table
DROP TABLE "api_tokens";
