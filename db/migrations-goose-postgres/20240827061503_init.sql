-- +goose Up
-- create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "dedicated_db" boolean NOT NULL DEFAULT false, "parent_organization_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organizations_organizations_children" FOREIGN KEY ("parent_organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "organization_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_name" ON "organizations" ("name") WHERE (deleted_at IS NULL);
-- create index "organizations_mapping_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_mapping_id_key" ON "organizations" ("mapping_id");
-- create "file_history" table
CREATE TABLE "file_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "file_name" character varying NOT NULL, "file_extension" character varying NOT NULL, "file_size" bigint NULL, "content_type" character varying NOT NULL, "store_key" character varying NOT NULL, "category" character varying NULL, "annotation" character varying NULL, PRIMARY KEY ("id"));
-- create index "filehistory_history_time" to table: "file_history"
CREATE INDEX "filehistory_history_time" ON "file_history" ("history_time");
-- create "contact_history" table
CREATE TABLE "contact_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', PRIMARY KEY ("id"));
-- create index "contacthistory_history_time" to table: "contact_history"
CREATE INDEX "contacthistory_history_time" ON "contact_history" ("history_time");
-- create "organization_setting_history" table
CREATE TABLE "organization_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" character varying NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "organization_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
CREATE INDEX "organizationsettinghistory_history_time" ON "organization_setting_history" ("history_time");
-- create "document_data_history" table
CREATE TABLE "document_data_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, "data" jsonb NOT NULL, PRIMARY KEY ("id"));
-- create index "documentdatahistory_history_time" to table: "document_data_history"
CREATE INDEX "documentdatahistory_history_time" ON "document_data_history" ("history_time");
-- create "webhook_history" table
CREATE TABLE "webhook_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "destination_url" character varying NOT NULL, "enabled" boolean NOT NULL DEFAULT true, "callback" character varying NULL, "expires_at" timestamptz NULL, "secret" bytea NULL, "failures" bigint NULL DEFAULT 0, "last_error" character varying NULL, "last_response" character varying NULL, PRIMARY KEY ("id"));
-- create index "webhookhistory_history_time" to table: "webhook_history"
CREATE INDEX "webhookhistory_history_time" ON "webhook_history" ("history_time");
-- create "org_membership_history" table
CREATE TABLE "org_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "orgmembershiphistory_history_time" to table: "org_membership_history"
CREATE INDEX "orgmembershiphistory_history_time" ON "org_membership_history" ("history_time");
-- create "entitlement_history" table
CREATE TABLE "entitlement_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "plan_id" character varying NOT NULL, "organization_id" character varying NOT NULL, "external_customer_id" character varying NULL, "external_subscription_id" character varying NULL, "expires" boolean NOT NULL DEFAULT false, "expires_at" timestamptz NULL, "cancelled" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "entitlementhistory_history_time" to table: "entitlement_history"
CREATE INDEX "entitlementhistory_history_time" ON "entitlement_history" ("history_time");
-- create "template_history" table
CREATE TABLE "template_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, PRIMARY KEY ("id"));
-- create index "templatehistory_history_time" to table: "template_history"
CREATE INDEX "templatehistory_history_time" ON "template_history" ("history_time");
-- create "oauth_provider_history" table
CREATE TABLE "oauth_provider_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "client_id" character varying NOT NULL, "client_secret" character varying NOT NULL, "redirect_url" character varying NOT NULL, "scopes" character varying NOT NULL, "auth_url" character varying NOT NULL, "token_url" character varying NOT NULL, "auth_style" smallint NOT NULL, "info_url" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "oauthproviderhistory_history_time" to table: "oauth_provider_history"
CREATE INDEX "oauthproviderhistory_history_time" ON "oauth_provider_history" ("history_time");
-- create "entitlement_plan_feature_history" table
CREATE TABLE "entitlement_plan_feature_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "metadata" jsonb NULL, "plan_id" character varying NOT NULL, "feature_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "entitlementplanfeaturehistory_history_time" to table: "entitlement_plan_feature_history"
CREATE INDEX "entitlementplanfeaturehistory_history_time" ON "entitlement_plan_feature_history" ("history_time");
-- create "entitlement_plan_history" table
CREATE TABLE "entitlement_plan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "display_name" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "version" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "entitlementplanhistory_history_time" to table: "entitlement_plan_history"
CREATE INDEX "entitlementplanhistory_history_time" ON "entitlement_plan_history" ("history_time");
-- create "note_history" table
CREATE TABLE "note_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "text" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "notehistory_history_time" to table: "note_history"
CREATE INDEX "notehistory_history_time" ON "note_history" ("history_time");
-- create "entity_history" table
CREATE TABLE "entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "entity_type_id" character varying NULL, "status" character varying NULL DEFAULT 'active', PRIMARY KEY ("id"));
-- create index "entityhistory_history_time" to table: "entity_history"
CREATE INDEX "entityhistory_history_time" ON "entity_history" ("history_time");
-- create "user_history" table
CREATE TABLE "user_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_local_file" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', PRIMARY KEY ("id"));
-- create index "userhistory_history_time" to table: "user_history"
CREATE INDEX "userhistory_history_time" ON "user_history" ("history_time");
-- create "entity_type_history" table
CREATE TABLE "entity_type_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "entitytypehistory_history_time" to table: "entity_type_history"
CREATE INDEX "entitytypehistory_history_time" ON "entity_type_history" ("history_time");
-- create "user_setting_history" table
CREATE TABLE "user_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "user_id" character varying NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, PRIMARY KEY ("id"));
-- create index "usersettinghistory_history_time" to table: "user_setting_history"
CREATE INDEX "usersettinghistory_history_time" ON "user_setting_history" ("history_time");
-- create "event_history" table
CREATE TABLE "event_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "eventhistory_history_time" to table: "event_history"
CREATE INDEX "eventhistory_history_time" ON "event_history" ("history_time");
-- create "integration_history" table
CREATE TABLE "integration_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, PRIMARY KEY ("id"));
-- create index "integrationhistory_history_time" to table: "integration_history"
CREATE INDEX "integrationhistory_history_time" ON "integration_history" ("history_time");
-- create "feature_history" table
CREATE TABLE "feature_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "display_name" character varying NULL, "enabled" boolean NOT NULL DEFAULT false, "description" character varying NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "featurehistory_history_time" to table: "feature_history"
CREATE INDEX "featurehistory_history_time" ON "feature_history" ("history_time");
-- create "organization_history" table
CREATE TABLE "organization_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "display_name" character varying NOT NULL DEFAULT '', "description" character varying NULL, "parent_organization_id" character varying NULL, "personal_org" boolean NULL DEFAULT false, "avatar_remote_url" character varying NULL, "dedicated_db" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "organizationhistory_history_time" to table: "organization_history"
CREATE INDEX "organizationhistory_history_time" ON "organization_history" ("history_time");
-- create "hush_history" table
CREATE TABLE "hush_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, PRIMARY KEY ("id"));
-- create index "hushhistory_history_time" to table: "hush_history"
CREATE INDEX "hushhistory_history_time" ON "hush_history" ("history_time");
-- create "group_setting_history" table
CREATE TABLE "group_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "groupsettinghistory_history_time" to table: "group_setting_history"
CREATE INDEX "groupsettinghistory_history_time" ON "group_setting_history" ("history_time");
-- create "group_history" table
CREATE TABLE "group_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', PRIMARY KEY ("id"));
-- create index "grouphistory_history_time" to table: "group_history"
CREATE INDEX "grouphistory_history_time" ON "group_history" ("history_time");
-- create "group_membership_history" table
CREATE TABLE "group_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "groupmembershiphistory_history_time" to table: "group_membership_history"
CREATE INDEX "groupmembershiphistory_history_time" ON "group_membership_history" ("history_time");
-- create "api_tokens" table
CREATE TABLE "api_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "api_tokens_organizations_api_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "api_tokens_mapping_id_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_mapping_id_key" ON "api_tokens" ("mapping_id");
-- create index "api_tokens_token_key" to table: "api_tokens"
CREATE UNIQUE INDEX "api_tokens_token_key" ON "api_tokens" ("token");
-- create index "apitoken_token" to table: "api_tokens"
CREATE INDEX "apitoken_token" ON "api_tokens" ("token");
-- create "contacts" table
CREATE TABLE "contacts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "full_name" character varying NOT NULL, "title" character varying NULL, "company" character varying NULL, "email" character varying NULL, "phone_number" character varying NULL, "address" character varying NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "contacts_organizations_contacts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "contacts_mapping_id_key" to table: "contacts"
CREATE UNIQUE INDEX "contacts_mapping_id_key" ON "contacts" ("mapping_id");
-- create "templates" table
CREATE TABLE "templates" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "template_type" character varying NOT NULL DEFAULT 'DOCUMENT', "description" character varying NULL, "jsonconfig" jsonb NOT NULL, "uischema" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "templates_organizations_templates" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "template_name_owner_id_template_type" to table: "templates"
CREATE UNIQUE INDEX "template_name_owner_id_template_type" ON "templates" ("name", "owner_id", "template_type") WHERE (deleted_at IS NULL);
-- create index "templates_mapping_id_key" to table: "templates"
CREATE UNIQUE INDEX "templates_mapping_id_key" ON "templates" ("mapping_id");
-- create "document_data" table
CREATE TABLE "document_data" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "data" jsonb NOT NULL, "owner_id" character varying NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "document_data_organizations_documentdata" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "document_data_mapping_id_key" to table: "document_data"
CREATE UNIQUE INDEX "document_data_mapping_id_key" ON "document_data" ("mapping_id");
-- create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "first_name" character varying NULL, "last_name" character varying NULL, "display_name" character varying NOT NULL, "avatar_remote_url" character varying NULL, "avatar_local_file" character varying NULL, "avatar_updated_at" timestamptz NULL, "last_seen" timestamptz NULL, "password" character varying NULL, "sub" character varying NULL, "auth_provider" character varying NOT NULL DEFAULT 'CREDENTIALS', "role" character varying NULL DEFAULT 'USER', PRIMARY KEY ("id"));
-- create index "user_email_auth_provider" to table: "users"
CREATE UNIQUE INDEX "user_email_auth_provider" ON "users" ("email", "auth_provider") WHERE (deleted_at IS NULL);
-- create index "user_id" to table: "users"
CREATE UNIQUE INDEX "user_id" ON "users" ("id");
-- create index "users_mapping_id_key" to table: "users"
CREATE UNIQUE INDEX "users_mapping_id_key" ON "users" ("mapping_id");
-- create index "users_sub_key" to table: "users"
CREATE UNIQUE INDEX "users_sub_key" ON "users" ("sub");
-- create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "email_verification_tokens_users_email_verification_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "email_verification_tokens_mapping_id_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_mapping_id_key" ON "email_verification_tokens" ("mapping_id");
-- create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "email_verification_tokens_token_key" ON "email_verification_tokens" ("token");
-- create index "emailverificationtoken_token" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_token" ON "email_verification_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "entity_types" table
CREATE TABLE "entity_types" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "entity_types_organizations_entitytypes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "entity_types_mapping_id_key" to table: "entity_types"
CREATE UNIQUE INDEX "entity_types_mapping_id_key" ON "entity_types" ("mapping_id");
-- create index "entitytype_name_owner_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_name_owner_id" ON "entity_types" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "entities" table
CREATE TABLE "entities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "domains" jsonb NULL, "status" character varying NULL DEFAULT 'active', "entity_type_id" character varying NULL, "entity_type_entities" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "entities_entity_types_entity_type" FOREIGN KEY ("entity_type_id") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "entities_organizations_entities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "entities_mapping_id_key" to table: "entities"
CREATE UNIQUE INDEX "entities_mapping_id_key" ON "entities" ("mapping_id");
-- create index "entity_name_owner_id" to table: "entities"
CREATE UNIQUE INDEX "entity_name_owner_id" ON "entities" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create "entitlement_plans" table
CREATE TABLE "entitlement_plans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "display_name" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "version" character varying NOT NULL, "metadata" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "entitlement_plans_organizations_entitlementplans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "entitlement_plans_mapping_id_key" to table: "entitlement_plans"
CREATE UNIQUE INDEX "entitlement_plans_mapping_id_key" ON "entitlement_plans" ("mapping_id");
-- create index "entitlementplan_name_version_owner_id" to table: "entitlement_plans"
CREATE UNIQUE INDEX "entitlementplan_name_version_owner_id" ON "entitlement_plans" ("name", "version", "owner_id") WHERE (deleted_at IS NULL);
-- create "entitlements" table
CREATE TABLE "entitlements" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "external_customer_id" character varying NULL, "external_subscription_id" character varying NULL, "expires" boolean NOT NULL DEFAULT false, "expires_at" timestamptz NULL, "cancelled" boolean NOT NULL DEFAULT false, "plan_id" character varying NOT NULL, "owner_id" character varying NULL, "organization_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "entitlements_entitlement_plans_entitlements" FOREIGN KEY ("plan_id") REFERENCES "entitlement_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "entitlements_organizations_entitlements" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "entitlements_organizations_organization_entitlement" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "entitlement_organization_id_owner_id" to table: "entitlements"
CREATE UNIQUE INDEX "entitlement_organization_id_owner_id" ON "entitlements" ("organization_id", "owner_id") WHERE ((deleted_at IS NULL) AND (cancelled = false));
-- create index "entitlements_mapping_id_key" to table: "entitlements"
CREATE UNIQUE INDEX "entitlements_mapping_id_key" ON "entitlements" ("mapping_id");
-- create "events" table
CREATE TABLE "events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "event_id" character varying NULL, "correlation_id" character varying NULL, "event_type" character varying NOT NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "events_mapping_id_key" to table: "events"
CREATE UNIQUE INDEX "events_mapping_id_key" ON "events" ("mapping_id");
-- create "entitlement_events" table
CREATE TABLE "entitlement_events" ("entitlement_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("entitlement_id", "event_id"), CONSTRAINT "entitlement_events_entitlement_id" FOREIGN KEY ("entitlement_id") REFERENCES "entitlements" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entitlement_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "entitlement_plan_events" table
CREATE TABLE "entitlement_plan_events" ("entitlement_plan_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("entitlement_plan_id", "event_id"), CONSTRAINT "entitlement_plan_events_entitlement_plan_id" FOREIGN KEY ("entitlement_plan_id") REFERENCES "entitlement_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entitlement_plan_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "features" table
CREATE TABLE "features" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "display_name" character varying NULL, "enabled" boolean NOT NULL DEFAULT false, "description" character varying NULL, "metadata" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "features_organizations_features" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "feature_name_owner_id" to table: "features"
CREATE UNIQUE INDEX "feature_name_owner_id" ON "features" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "features_mapping_id_key" to table: "features"
CREATE UNIQUE INDEX "features_mapping_id_key" ON "features" ("mapping_id");
-- create "entitlement_plan_features" table
CREATE TABLE "entitlement_plan_features" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "metadata" jsonb NULL, "plan_id" character varying NOT NULL, "feature_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "entitlement_plan_features_entitlement_plans_plan" FOREIGN KEY ("plan_id") REFERENCES "entitlement_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "entitlement_plan_features_features_feature" FOREIGN KEY ("feature_id") REFERENCES "features" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "entitlement_plan_features_organizations_entitlementplanfeatures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "entitlement_plan_features_mapping_id_key" to table: "entitlement_plan_features"
CREATE UNIQUE INDEX "entitlement_plan_features_mapping_id_key" ON "entitlement_plan_features" ("mapping_id");
-- create index "entitlementplanfeature_feature_id_plan_id" to table: "entitlement_plan_features"
CREATE UNIQUE INDEX "entitlementplanfeature_feature_id_plan_id" ON "entitlement_plan_features" ("feature_id", "plan_id") WHERE (deleted_at IS NULL);
-- create "entitlement_plan_feature_events" table
CREATE TABLE "entitlement_plan_feature_events" ("entitlement_plan_feature_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("entitlement_plan_feature_id", "event_id"), CONSTRAINT "entitlement_plan_feature_events_entitlement_plan_feature_id" FOREIGN KEY ("entitlement_plan_feature_id") REFERENCES "entitlement_plan_features" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entitlement_plan_feature_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "entity_contacts" table
CREATE TABLE "entity_contacts" ("entity_id" character varying NOT NULL, "contact_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "contact_id"), CONSTRAINT "entity_contacts_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_contacts_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "entity_documents" table
CREATE TABLE "entity_documents" ("entity_id" character varying NOT NULL, "document_data_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "document_data_id"), CONSTRAINT "entity_documents_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_documents_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "files" table
CREATE TABLE "files" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "file_name" character varying NOT NULL, "file_extension" character varying NOT NULL, "file_size" bigint NULL, "content_type" character varying NOT NULL, "store_key" character varying NOT NULL, "category" character varying NULL, "annotation" character varying NULL, "user_files" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "files_users_files" FOREIGN KEY ("user_files") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "files_mapping_id_key" to table: "files"
CREATE UNIQUE INDEX "files_mapping_id_key" ON "files" ("mapping_id");
-- create "entity_files" table
CREATE TABLE "entity_files" ("entity_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "file_id"), CONSTRAINT "entity_files_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "feature_events" table
CREATE TABLE "feature_events" ("feature_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("feature_id", "event_id"), CONSTRAINT "feature_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "feature_events_feature_id" FOREIGN KEY ("feature_id") REFERENCES "features" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "gravatar_logo_url" character varying NULL, "logo_url" character varying NULL, "display_name" character varying NOT NULL DEFAULT '', "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "group_name_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_name_owner_id" ON "groups" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "groups_mapping_id_key" to table: "groups"
CREATE UNIQUE INDEX "groups_mapping_id_key" ON "groups" ("mapping_id");
-- create "group_events" table
CREATE TABLE "group_events" ("group_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_id", "event_id"), CONSTRAINT "group_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "group_events_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "group_files" table
CREATE TABLE "group_files" ("group_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("group_id", "file_id"), CONSTRAINT "group_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "group_files_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "group_memberships" table
CREATE TABLE "group_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "group_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "group_memberships_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "group_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "group_memberships_mapping_id_key" to table: "group_memberships"
CREATE UNIQUE INDEX "group_memberships_mapping_id_key" ON "group_memberships" ("mapping_id");
-- create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id") WHERE (deleted_at IS NULL);
-- create "group_membership_events" table
CREATE TABLE "group_membership_events" ("group_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("group_membership_id", "event_id"), CONSTRAINT "group_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "group_membership_events_group_membership_id" FOREIGN KEY ("group_membership_id") REFERENCES "group_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "group_settings" table
CREATE TABLE "group_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "visibility" character varying NOT NULL DEFAULT 'PUBLIC', "join_policy" character varying NOT NULL DEFAULT 'INVITE_OR_APPLICATION', "sync_to_slack" boolean NULL DEFAULT false, "sync_to_github" boolean NULL DEFAULT false, "group_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "group_settings_groups_setting" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "group_settings_group_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_group_id_key" ON "group_settings" ("group_id");
-- create index "group_settings_mapping_id_key" to table: "group_settings"
CREATE UNIQUE INDEX "group_settings_mapping_id_key" ON "group_settings" ("mapping_id");
-- create "hushes" table
CREATE TABLE "hushes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "secret_name" character varying NULL, "secret_value" character varying NULL, PRIMARY KEY ("id"));
-- create index "hushes_mapping_id_key" to table: "hushes"
CREATE UNIQUE INDEX "hushes_mapping_id_key" ON "hushes" ("mapping_id");
-- create "hush_events" table
CREATE TABLE "hush_events" ("hush_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("hush_id", "event_id"), CONSTRAINT "hush_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "hush_events_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "integrations" table
CREATE TABLE "integrations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "kind" character varying NULL, "group_integrations" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "integrations_groups_integrations" FOREIGN KEY ("group_integrations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "integrations_organizations_integrations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "integrations_mapping_id_key" to table: "integrations"
CREATE UNIQUE INDEX "integrations_mapping_id_key" ON "integrations" ("mapping_id");
-- create "integration_events" table
CREATE TABLE "integration_events" ("integration_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "event_id"), CONSTRAINT "integration_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "integration_events_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "oh_auth_too_tokens" table
CREATE TABLE "oh_auth_too_tokens" ("id" character varying NOT NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "client_id" text NOT NULL, "scopes" jsonb NULL, "nonce" text NOT NULL, "claims_user_id" text NOT NULL, "claims_username" text NOT NULL, "claims_email" text NOT NULL, "claims_email_verified" boolean NOT NULL, "claims_groups" jsonb NULL, "claims_preferred_username" text NOT NULL, "connector_id" text NOT NULL, "connector_data" jsonb NULL, "last_used" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- create index "oh_auth_too_tokens_mapping_id_key" to table: "oh_auth_too_tokens"
CREATE UNIQUE INDEX "oh_auth_too_tokens_mapping_id_key" ON "oh_auth_too_tokens" ("mapping_id");
-- create "integration_oauth2tokens" table
CREATE TABLE "integration_oauth2tokens" ("integration_id" character varying NOT NULL, "oh_auth_too_token_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "oh_auth_too_token_id"), CONSTRAINT "integration_oauth2tokens_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "integration_oauth2tokens_oh_auth_too_token_id" FOREIGN KEY ("oh_auth_too_token_id") REFERENCES "oh_auth_too_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "integration_secrets" table
CREATE TABLE "integration_secrets" ("integration_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "hush_id"), CONSTRAINT "integration_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "integration_secrets_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "webhooks" table
CREATE TABLE "webhooks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "destination_url" character varying NOT NULL, "enabled" boolean NOT NULL DEFAULT true, "callback" character varying NULL, "expires_at" timestamptz NULL, "secret" bytea NULL, "failures" bigint NULL DEFAULT 0, "last_error" character varying NULL, "last_response" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "webhooks_organizations_webhooks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "webhook_name_owner_id" to table: "webhooks"
CREATE UNIQUE INDEX "webhook_name_owner_id" ON "webhooks" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "webhooks_callback_key" to table: "webhooks"
CREATE UNIQUE INDEX "webhooks_callback_key" ON "webhooks" ("callback");
-- create index "webhooks_mapping_id_key" to table: "webhooks"
CREATE UNIQUE INDEX "webhooks_mapping_id_key" ON "webhooks" ("mapping_id");
-- create "integration_webhooks" table
CREATE TABLE "integration_webhooks" ("integration_id" character varying NOT NULL, "webhook_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "webhook_id"), CONSTRAINT "integration_webhooks_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "integration_webhooks_webhook_id" FOREIGN KEY ("webhook_id") REFERENCES "webhooks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "invites" table
CREATE TABLE "invites" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "expires" timestamptz NULL, "recipient" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'INVITATION_SENT', "role" character varying NOT NULL DEFAULT 'MEMBER', "send_attempts" bigint NOT NULL DEFAULT 0, "requestor_id" character varying NULL, "secret" bytea NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "invites_organizations_invites" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "invite_recipient_owner_id" to table: "invites"
CREATE UNIQUE INDEX "invite_recipient_owner_id" ON "invites" ("recipient", "owner_id") WHERE (deleted_at IS NULL);
-- create index "invites_mapping_id_key" to table: "invites"
CREATE UNIQUE INDEX "invites_mapping_id_key" ON "invites" ("mapping_id");
-- create index "invites_token_key" to table: "invites"
CREATE UNIQUE INDEX "invites_token_key" ON "invites" ("token");
-- create "invite_events" table
CREATE TABLE "invite_events" ("invite_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("invite_id", "event_id"), CONSTRAINT "invite_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "invite_events_invite_id" FOREIGN KEY ("invite_id") REFERENCES "invites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "notes" table
CREATE TABLE "notes" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "text" character varying NOT NULL, "entity_notes" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "notes_entities_notes" FOREIGN KEY ("entity_notes") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notes_organizations_notes" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "notes_mapping_id_key" to table: "notes"
CREATE UNIQUE INDEX "notes_mapping_id_key" ON "notes" ("mapping_id");
-- create "oauth_providers" table
CREATE TABLE "oauth_providers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NOT NULL, "client_id" character varying NOT NULL, "client_secret" character varying NOT NULL, "redirect_url" character varying NOT NULL, "scopes" character varying NOT NULL, "auth_url" character varying NOT NULL, "token_url" character varying NOT NULL, "auth_style" smallint NOT NULL, "info_url" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "oauth_providers_organizations_oauthprovider" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "oauth_providers_mapping_id_key" to table: "oauth_providers"
CREATE UNIQUE INDEX "oauth_providers_mapping_id_key" ON "oauth_providers" ("mapping_id");
-- create "oh_auth_too_token_events" table
CREATE TABLE "oh_auth_too_token_events" ("oh_auth_too_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("oh_auth_too_token_id", "event_id"), CONSTRAINT "oh_auth_too_token_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "oh_auth_too_token_events_oh_auth_too_token_id" FOREIGN KEY ("oh_auth_too_token_id") REFERENCES "oh_auth_too_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "org_memberships" table
CREATE TABLE "org_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "org_memberships_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "org_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "org_memberships_mapping_id_key" to table: "org_memberships"
CREATE UNIQUE INDEX "org_memberships_mapping_id_key" ON "org_memberships" ("mapping_id");
-- create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id") WHERE (deleted_at IS NULL);
-- create "org_membership_events" table
CREATE TABLE "org_membership_events" ("org_membership_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_membership_id", "event_id"), CONSTRAINT "org_membership_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "org_membership_events_org_membership_id" FOREIGN KEY ("org_membership_id") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "organization_events" table
CREATE TABLE "organization_events" ("organization_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "event_id"), CONSTRAINT "organization_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "organization_events_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "organization_files" table
CREATE TABLE "organization_files" ("organization_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "file_id"), CONSTRAINT "organization_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "organization_files_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "personal_access_tokens" table
CREATE TABLE "personal_access_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "description" character varying NULL, "scopes" jsonb NULL, "last_used_at" timestamptz NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "personal_access_tokens_users_personal_access_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "personal_access_tokens_mapping_id_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_mapping_id_key" ON "personal_access_tokens" ("mapping_id");
-- create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personal_access_tokens_token_key" ON "personal_access_tokens" ("token");
-- create index "personalaccesstoken_token" to table: "personal_access_tokens"
CREATE INDEX "personalaccesstoken_token" ON "personal_access_tokens" ("token");
-- create "organization_personal_access_tokens" table
CREATE TABLE "organization_personal_access_tokens" ("organization_id" character varying NOT NULL, "personal_access_token_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "personal_access_token_id"), CONSTRAINT "organization_personal_access_tokens_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "organization_personal_access_tokens_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "organization_secrets" table
CREATE TABLE "organization_secrets" ("organization_id" character varying NOT NULL, "hush_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "hush_id"), CONSTRAINT "organization_secrets_hush_id" FOREIGN KEY ("hush_id") REFERENCES "hushes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "organization_secrets_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "organization_settings" table
CREATE TABLE "organization_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "domains" jsonb NULL, "billing_contact" character varying NULL, "billing_email" character varying NULL, "billing_phone" character varying NULL, "billing_address" character varying NULL, "tax_identifier" character varying NULL, "geo_location" character varying NULL DEFAULT 'AMER', "organization_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "organization_settings_organizations_setting" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "organization_settings_mapping_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_mapping_id_key" ON "organization_settings" ("mapping_id");
-- create index "organization_settings_organization_id_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_organization_id_key" ON "organization_settings" ("organization_id");
-- create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "password_reset_tokens_users_password_reset_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "password_reset_tokens_mapping_id_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_mapping_id_key" ON "password_reset_tokens" ("mapping_id");
-- create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
-- create index "passwordresettoken_token" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_token" ON "password_reset_tokens" ("token") WHERE (deleted_at IS NULL);
-- create "personal_access_token_events" table
CREATE TABLE "personal_access_token_events" ("personal_access_token_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("personal_access_token_id", "event_id"), CONSTRAINT "personal_access_token_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "personal_access_token_events_personal_access_token_id" FOREIGN KEY ("personal_access_token_id") REFERENCES "personal_access_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "subscribers" table
CREATE TABLE "subscribers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "email" character varying NOT NULL, "phone_number" character varying NULL, "verified_email" boolean NOT NULL DEFAULT false, "verified_phone" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT false, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "secret" bytea NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "subscribers_organizations_subscribers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE (deleted_at IS NULL);
-- create index "subscribers_mapping_id_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_mapping_id_key" ON "subscribers" ("mapping_id");
-- create index "subscribers_token_key" to table: "subscribers"
CREATE UNIQUE INDEX "subscribers_token_key" ON "subscribers" ("token");
-- create "subscriber_events" table
CREATE TABLE "subscriber_events" ("subscriber_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("subscriber_id", "event_id"), CONSTRAINT "subscriber_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subscriber_events_subscriber_id" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "tfa_settings" table
CREATE TABLE "tfa_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "tfa_secret" character varying NULL, "verified" boolean NOT NULL DEFAULT false, "recovery_codes" jsonb NULL, "phone_otp_allowed" boolean NULL DEFAULT false, "email_otp_allowed" boolean NULL DEFAULT false, "totp_allowed" boolean NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "tfa_settings_users_tfa_settings" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "tfa_settings_mapping_id_key" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfa_settings_mapping_id_key" ON "tfa_settings" ("mapping_id");
-- create index "tfasetting_owner_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_owner_id" ON "tfa_settings" ("owner_id") WHERE (deleted_at IS NULL);
-- create "user_events" table
CREATE TABLE "user_events" ("user_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("user_id", "event_id"), CONSTRAINT "user_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "user_settings" table
CREATE TABLE "user_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "locked" boolean NOT NULL DEFAULT false, "silenced_at" timestamptz NULL, "suspended_at" timestamptz NULL, "status" character varying NOT NULL DEFAULT 'ACTIVE', "email_confirmed" boolean NOT NULL DEFAULT false, "is_webauthn_allowed" boolean NULL DEFAULT false, "is_tfa_enabled" boolean NULL DEFAULT false, "phone_number" character varying NULL, "user_id" character varying NULL, "user_setting_default_org" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "user_settings_organizations_default_org" FOREIGN KEY ("user_setting_default_org") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "user_settings_users_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "user_settings_mapping_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_mapping_id_key" ON "user_settings" ("mapping_id");
-- create index "user_settings_user_id_key" to table: "user_settings"
CREATE UNIQUE INDEX "user_settings_user_id_key" ON "user_settings" ("user_id");
-- create "webauthns" table
CREATE TABLE "webauthns" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "credential_id" bytea NULL, "public_key" bytea NULL, "attestation_type" character varying NULL, "aaguid" bytea NOT NULL, "sign_count" integer NOT NULL, "transports" jsonb NOT NULL, "backup_eligible" boolean NOT NULL DEFAULT false, "backup_state" boolean NOT NULL DEFAULT false, "user_present" boolean NOT NULL DEFAULT false, "user_verified" boolean NOT NULL DEFAULT false, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "webauthns_users_webauthn" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "webauthns_aaguid_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_aaguid_key" ON "webauthns" ("aaguid");
-- create index "webauthns_credential_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_credential_id_key" ON "webauthns" ("credential_id");
-- create index "webauthns_mapping_id_key" to table: "webauthns"
CREATE UNIQUE INDEX "webauthns_mapping_id_key" ON "webauthns" ("mapping_id");
-- create "webhook_events" table
CREATE TABLE "webhook_events" ("webhook_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("webhook_id", "event_id"), CONSTRAINT "webhook_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "webhook_events_webhook_id" FOREIGN KEY ("webhook_id") REFERENCES "webhooks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "webhook_events" table
DROP TABLE "webhook_events";
-- reverse: create index "webauthns_mapping_id_key" to table: "webauthns"
DROP INDEX "webauthns_mapping_id_key";
-- reverse: create index "webauthns_credential_id_key" to table: "webauthns"
DROP INDEX "webauthns_credential_id_key";
-- reverse: create index "webauthns_aaguid_key" to table: "webauthns"
DROP INDEX "webauthns_aaguid_key";
-- reverse: create "webauthns" table
DROP TABLE "webauthns";
-- reverse: create index "user_settings_user_id_key" to table: "user_settings"
DROP INDEX "user_settings_user_id_key";
-- reverse: create index "user_settings_mapping_id_key" to table: "user_settings"
DROP INDEX "user_settings_mapping_id_key";
-- reverse: create "user_settings" table
DROP TABLE "user_settings";
-- reverse: create "user_events" table
DROP TABLE "user_events";
-- reverse: create index "tfasetting_owner_id" to table: "tfa_settings"
DROP INDEX "tfasetting_owner_id";
-- reverse: create index "tfa_settings_mapping_id_key" to table: "tfa_settings"
DROP INDEX "tfa_settings_mapping_id_key";
-- reverse: create "tfa_settings" table
DROP TABLE "tfa_settings";
-- reverse: create "subscriber_events" table
DROP TABLE "subscriber_events";
-- reverse: create index "subscribers_token_key" to table: "subscribers"
DROP INDEX "subscribers_token_key";
-- reverse: create index "subscribers_mapping_id_key" to table: "subscribers"
DROP INDEX "subscribers_mapping_id_key";
-- reverse: create index "subscriber_email_owner_id" to table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- reverse: create "subscribers" table
DROP TABLE "subscribers";
-- reverse: create "personal_access_token_events" table
DROP TABLE "personal_access_token_events";
-- reverse: create index "passwordresettoken_token" to table: "password_reset_tokens"
DROP INDEX "passwordresettoken_token";
-- reverse: create index "password_reset_tokens_token_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_token_key";
-- reverse: create index "password_reset_tokens_mapping_id_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_mapping_id_key";
-- reverse: create "password_reset_tokens" table
DROP TABLE "password_reset_tokens";
-- reverse: create index "organization_settings_organization_id_key" to table: "organization_settings"
DROP INDEX "organization_settings_organization_id_key";
-- reverse: create index "organization_settings_mapping_id_key" to table: "organization_settings"
DROP INDEX "organization_settings_mapping_id_key";
-- reverse: create "organization_settings" table
DROP TABLE "organization_settings";
-- reverse: create "organization_secrets" table
DROP TABLE "organization_secrets";
-- reverse: create "organization_personal_access_tokens" table
DROP TABLE "organization_personal_access_tokens";
-- reverse: create index "personalaccesstoken_token" to table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_token";
-- reverse: create index "personal_access_tokens_token_key" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_token_key";
-- reverse: create index "personal_access_tokens_mapping_id_key" to table: "personal_access_tokens"
DROP INDEX "personal_access_tokens_mapping_id_key";
-- reverse: create "personal_access_tokens" table
DROP TABLE "personal_access_tokens";
-- reverse: create "organization_files" table
DROP TABLE "organization_files";
-- reverse: create "organization_events" table
DROP TABLE "organization_events";
-- reverse: create "org_membership_events" table
DROP TABLE "org_membership_events";
-- reverse: create index "orgmembership_user_id_organization_id" to table: "org_memberships"
DROP INDEX "orgmembership_user_id_organization_id";
-- reverse: create index "org_memberships_mapping_id_key" to table: "org_memberships"
DROP INDEX "org_memberships_mapping_id_key";
-- reverse: create "org_memberships" table
DROP TABLE "org_memberships";
-- reverse: create "oh_auth_too_token_events" table
DROP TABLE "oh_auth_too_token_events";
-- reverse: create index "oauth_providers_mapping_id_key" to table: "oauth_providers"
DROP INDEX "oauth_providers_mapping_id_key";
-- reverse: create "oauth_providers" table
DROP TABLE "oauth_providers";
-- reverse: create index "notes_mapping_id_key" to table: "notes"
DROP INDEX "notes_mapping_id_key";
-- reverse: create "notes" table
DROP TABLE "notes";
-- reverse: create "invite_events" table
DROP TABLE "invite_events";
-- reverse: create index "invites_token_key" to table: "invites"
DROP INDEX "invites_token_key";
-- reverse: create index "invites_mapping_id_key" to table: "invites"
DROP INDEX "invites_mapping_id_key";
-- reverse: create index "invite_recipient_owner_id" to table: "invites"
DROP INDEX "invite_recipient_owner_id";
-- reverse: create "invites" table
DROP TABLE "invites";
-- reverse: create "integration_webhooks" table
DROP TABLE "integration_webhooks";
-- reverse: create index "webhooks_mapping_id_key" to table: "webhooks"
DROP INDEX "webhooks_mapping_id_key";
-- reverse: create index "webhooks_callback_key" to table: "webhooks"
DROP INDEX "webhooks_callback_key";
-- reverse: create index "webhook_name_owner_id" to table: "webhooks"
DROP INDEX "webhook_name_owner_id";
-- reverse: create "webhooks" table
DROP TABLE "webhooks";
-- reverse: create "integration_secrets" table
DROP TABLE "integration_secrets";
-- reverse: create "integration_oauth2tokens" table
DROP TABLE "integration_oauth2tokens";
-- reverse: create index "oh_auth_too_tokens_mapping_id_key" to table: "oh_auth_too_tokens"
DROP INDEX "oh_auth_too_tokens_mapping_id_key";
-- reverse: create "oh_auth_too_tokens" table
DROP TABLE "oh_auth_too_tokens";
-- reverse: create "integration_events" table
DROP TABLE "integration_events";
-- reverse: create index "integrations_mapping_id_key" to table: "integrations"
DROP INDEX "integrations_mapping_id_key";
-- reverse: create "integrations" table
DROP TABLE "integrations";
-- reverse: create "hush_events" table
DROP TABLE "hush_events";
-- reverse: create index "hushes_mapping_id_key" to table: "hushes"
DROP INDEX "hushes_mapping_id_key";
-- reverse: create "hushes" table
DROP TABLE "hushes";
-- reverse: create index "group_settings_mapping_id_key" to table: "group_settings"
DROP INDEX "group_settings_mapping_id_key";
-- reverse: create index "group_settings_group_id_key" to table: "group_settings"
DROP INDEX "group_settings_group_id_key";
-- reverse: create "group_settings" table
DROP TABLE "group_settings";
-- reverse: create "group_membership_events" table
DROP TABLE "group_membership_events";
-- reverse: create index "groupmembership_user_id_group_id" to table: "group_memberships"
DROP INDEX "groupmembership_user_id_group_id";
-- reverse: create index "group_memberships_mapping_id_key" to table: "group_memberships"
DROP INDEX "group_memberships_mapping_id_key";
-- reverse: create "group_memberships" table
DROP TABLE "group_memberships";
-- reverse: create "group_files" table
DROP TABLE "group_files";
-- reverse: create "group_events" table
DROP TABLE "group_events";
-- reverse: create index "groups_mapping_id_key" to table: "groups"
DROP INDEX "groups_mapping_id_key";
-- reverse: create index "group_name_owner_id" to table: "groups"
DROP INDEX "group_name_owner_id";
-- reverse: create "groups" table
DROP TABLE "groups";
-- reverse: create "feature_events" table
DROP TABLE "feature_events";
-- reverse: create "entity_files" table
DROP TABLE "entity_files";
-- reverse: create index "files_mapping_id_key" to table: "files"
DROP INDEX "files_mapping_id_key";
-- reverse: create "files" table
DROP TABLE "files";
-- reverse: create "entity_documents" table
DROP TABLE "entity_documents";
-- reverse: create "entity_contacts" table
DROP TABLE "entity_contacts";
-- reverse: create "entitlement_plan_feature_events" table
DROP TABLE "entitlement_plan_feature_events";
-- reverse: create index "entitlementplanfeature_feature_id_plan_id" to table: "entitlement_plan_features"
DROP INDEX "entitlementplanfeature_feature_id_plan_id";
-- reverse: create index "entitlement_plan_features_mapping_id_key" to table: "entitlement_plan_features"
DROP INDEX "entitlement_plan_features_mapping_id_key";
-- reverse: create "entitlement_plan_features" table
DROP TABLE "entitlement_plan_features";
-- reverse: create index "features_mapping_id_key" to table: "features"
DROP INDEX "features_mapping_id_key";
-- reverse: create index "feature_name_owner_id" to table: "features"
DROP INDEX "feature_name_owner_id";
-- reverse: create "features" table
DROP TABLE "features";
-- reverse: create "entitlement_plan_events" table
DROP TABLE "entitlement_plan_events";
-- reverse: create "entitlement_events" table
DROP TABLE "entitlement_events";
-- reverse: create index "events_mapping_id_key" to table: "events"
DROP INDEX "events_mapping_id_key";
-- reverse: create "events" table
DROP TABLE "events";
-- reverse: create index "entitlements_mapping_id_key" to table: "entitlements"
DROP INDEX "entitlements_mapping_id_key";
-- reverse: create index "entitlement_organization_id_owner_id" to table: "entitlements"
DROP INDEX "entitlement_organization_id_owner_id";
-- reverse: create "entitlements" table
DROP TABLE "entitlements";
-- reverse: create index "entitlementplan_name_version_owner_id" to table: "entitlement_plans"
DROP INDEX "entitlementplan_name_version_owner_id";
-- reverse: create index "entitlement_plans_mapping_id_key" to table: "entitlement_plans"
DROP INDEX "entitlement_plans_mapping_id_key";
-- reverse: create "entitlement_plans" table
DROP TABLE "entitlement_plans";
-- reverse: create index "entity_name_owner_id" to table: "entities"
DROP INDEX "entity_name_owner_id";
-- reverse: create index "entities_mapping_id_key" to table: "entities"
DROP INDEX "entities_mapping_id_key";
-- reverse: create "entities" table
DROP TABLE "entities";
-- reverse: create index "entitytype_name_owner_id" to table: "entity_types"
DROP INDEX "entitytype_name_owner_id";
-- reverse: create index "entity_types_mapping_id_key" to table: "entity_types"
DROP INDEX "entity_types_mapping_id_key";
-- reverse: create "entity_types" table
DROP TABLE "entity_types";
-- reverse: create index "emailverificationtoken_token" to table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_token";
-- reverse: create index "email_verification_tokens_token_key" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_token_key";
-- reverse: create index "email_verification_tokens_mapping_id_key" to table: "email_verification_tokens"
DROP INDEX "email_verification_tokens_mapping_id_key";
-- reverse: create "email_verification_tokens" table
DROP TABLE "email_verification_tokens";
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
-- reverse: create index "document_data_mapping_id_key" to table: "document_data"
DROP INDEX "document_data_mapping_id_key";
-- reverse: create "document_data" table
DROP TABLE "document_data";
-- reverse: create index "templates_mapping_id_key" to table: "templates"
DROP INDEX "templates_mapping_id_key";
-- reverse: create index "template_name_owner_id_template_type" to table: "templates"
DROP INDEX "template_name_owner_id_template_type";
-- reverse: create "templates" table
DROP TABLE "templates";
-- reverse: create index "contacts_mapping_id_key" to table: "contacts"
DROP INDEX "contacts_mapping_id_key";
-- reverse: create "contacts" table
DROP TABLE "contacts";
-- reverse: create index "apitoken_token" to table: "api_tokens"
DROP INDEX "apitoken_token";
-- reverse: create index "api_tokens_token_key" to table: "api_tokens"
DROP INDEX "api_tokens_token_key";
-- reverse: create index "api_tokens_mapping_id_key" to table: "api_tokens"
DROP INDEX "api_tokens_mapping_id_key";
-- reverse: create "api_tokens" table
DROP TABLE "api_tokens";
-- reverse: create index "groupmembershiphistory_history_time" to table: "group_membership_history"
DROP INDEX "groupmembershiphistory_history_time";
-- reverse: create "group_membership_history" table
DROP TABLE "group_membership_history";
-- reverse: create index "grouphistory_history_time" to table: "group_history"
DROP INDEX "grouphistory_history_time";
-- reverse: create "group_history" table
DROP TABLE "group_history";
-- reverse: create index "groupsettinghistory_history_time" to table: "group_setting_history"
DROP INDEX "groupsettinghistory_history_time";
-- reverse: create "group_setting_history" table
DROP TABLE "group_setting_history";
-- reverse: create index "hushhistory_history_time" to table: "hush_history"
DROP INDEX "hushhistory_history_time";
-- reverse: create "hush_history" table
DROP TABLE "hush_history";
-- reverse: create index "organizationhistory_history_time" to table: "organization_history"
DROP INDEX "organizationhistory_history_time";
-- reverse: create "organization_history" table
DROP TABLE "organization_history";
-- reverse: create index "featurehistory_history_time" to table: "feature_history"
DROP INDEX "featurehistory_history_time";
-- reverse: create "feature_history" table
DROP TABLE "feature_history";
-- reverse: create index "integrationhistory_history_time" to table: "integration_history"
DROP INDEX "integrationhistory_history_time";
-- reverse: create "integration_history" table
DROP TABLE "integration_history";
-- reverse: create index "eventhistory_history_time" to table: "event_history"
DROP INDEX "eventhistory_history_time";
-- reverse: create "event_history" table
DROP TABLE "event_history";
-- reverse: create index "usersettinghistory_history_time" to table: "user_setting_history"
DROP INDEX "usersettinghistory_history_time";
-- reverse: create "user_setting_history" table
DROP TABLE "user_setting_history";
-- reverse: create index "entitytypehistory_history_time" to table: "entity_type_history"
DROP INDEX "entitytypehistory_history_time";
-- reverse: create "entity_type_history" table
DROP TABLE "entity_type_history";
-- reverse: create index "userhistory_history_time" to table: "user_history"
DROP INDEX "userhistory_history_time";
-- reverse: create "user_history" table
DROP TABLE "user_history";
-- reverse: create index "entityhistory_history_time" to table: "entity_history"
DROP INDEX "entityhistory_history_time";
-- reverse: create "entity_history" table
DROP TABLE "entity_history";
-- reverse: create index "notehistory_history_time" to table: "note_history"
DROP INDEX "notehistory_history_time";
-- reverse: create "note_history" table
DROP TABLE "note_history";
-- reverse: create index "entitlementplanhistory_history_time" to table: "entitlement_plan_history"
DROP INDEX "entitlementplanhistory_history_time";
-- reverse: create "entitlement_plan_history" table
DROP TABLE "entitlement_plan_history";
-- reverse: create index "entitlementplanfeaturehistory_history_time" to table: "entitlement_plan_feature_history"
DROP INDEX "entitlementplanfeaturehistory_history_time";
-- reverse: create "entitlement_plan_feature_history" table
DROP TABLE "entitlement_plan_feature_history";
-- reverse: create index "oauthproviderhistory_history_time" to table: "oauth_provider_history"
DROP INDEX "oauthproviderhistory_history_time";
-- reverse: create "oauth_provider_history" table
DROP TABLE "oauth_provider_history";
-- reverse: create index "templatehistory_history_time" to table: "template_history"
DROP INDEX "templatehistory_history_time";
-- reverse: create "template_history" table
DROP TABLE "template_history";
-- reverse: create index "entitlementhistory_history_time" to table: "entitlement_history"
DROP INDEX "entitlementhistory_history_time";
-- reverse: create "entitlement_history" table
DROP TABLE "entitlement_history";
-- reverse: create index "orgmembershiphistory_history_time" to table: "org_membership_history"
DROP INDEX "orgmembershiphistory_history_time";
-- reverse: create "org_membership_history" table
DROP TABLE "org_membership_history";
-- reverse: create index "webhookhistory_history_time" to table: "webhook_history"
DROP INDEX "webhookhistory_history_time";
-- reverse: create "webhook_history" table
DROP TABLE "webhook_history";
-- reverse: create index "documentdatahistory_history_time" to table: "document_data_history"
DROP INDEX "documentdatahistory_history_time";
-- reverse: create "document_data_history" table
DROP TABLE "document_data_history";
-- reverse: create index "organizationsettinghistory_history_time" to table: "organization_setting_history"
DROP INDEX "organizationsettinghistory_history_time";
-- reverse: create "organization_setting_history" table
DROP TABLE "organization_setting_history";
-- reverse: create index "contacthistory_history_time" to table: "contact_history"
DROP INDEX "contacthistory_history_time";
-- reverse: create "contact_history" table
DROP TABLE "contact_history";
-- reverse: create index "filehistory_history_time" to table: "file_history"
DROP INDEX "filehistory_history_time";
-- reverse: create "file_history" table
DROP TABLE "file_history";
-- reverse: create index "organizations_mapping_id_key" to table: "organizations"
DROP INDEX "organizations_mapping_id_key";
-- reverse: create index "organization_name" to table: "organizations"
DROP INDEX "organization_name";
-- reverse: create "organizations" table
DROP TABLE "organizations";
