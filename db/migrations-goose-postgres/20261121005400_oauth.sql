-- +goose Up
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_oauth_grant_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_oauth_grant_creators" FOREIGN KEY ("organization_oauth_grant_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "oauth_grants" table
CREATE TABLE "oauth_grants" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "client_id" character varying NOT NULL, "client_name" character varying NOT NULL, "client_uri" character varying NULL, "logo_uri" character varying NULL, "kind" character varying NOT NULL DEFAULT 'USER', "access_level" character varying NULL DEFAULT 'READ', "scopes" jsonb NULL, "installed_by_user_id" character varying NOT NULL, "last_used_at" timestamptz NULL, "is_active" boolean NOT NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "user_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "oauth_grants_organizations_oauth_grants" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "oauth_grants_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "oauth_grant_owner_id_idx" to table: "oauth_grants"
CREATE INDEX "oauth_grant_owner_id_idx" ON "oauth_grants" ("owner_id");
-- create index "oauth_grant_user_id_idx" to table: "oauth_grants"
CREATE INDEX "oauth_grant_user_id_idx" ON "oauth_grants" ("user_id");
-- create index "oauthgrant_client_id_owner_id_kind_user_id" to table: "oauth_grants"
CREATE UNIQUE INDEX "oauthgrant_client_id_owner_id_kind_user_id" ON "oauth_grants" ("client_id", "owner_id", "kind", "user_id") WHERE (deleted_at IS NULL);
-- create "oauth_tokens" table
CREATE TABLE "oauth_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "kind" character varying NOT NULL DEFAULT 'ACCESS', "token_hash" character varying NOT NULL, "key_version" character varying NOT NULL, "expires_at" timestamptz NOT NULL, "consumed_at" timestamptz NULL, "client_id" character varying NOT NULL, "redirect_uri" character varying NULL, "code_challenge" character varying NULL, "state" character varying NULL, "resource" character varying NULL, "scopes" jsonb NULL, "user_id" character varying NULL, "oauth_grant_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "oauth_tokens_oauth_grants_oauth_tokens" FOREIGN KEY ("oauth_grant_id") REFERENCES "oauth_grants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "oauth_tokens_token_hash_key" to table: "oauth_tokens"
CREATE UNIQUE INDEX "oauth_tokens_token_hash_key" ON "oauth_tokens" ("token_hash");
-- create index "oauthtoken_expires_at" to table: "oauth_tokens"
CREATE INDEX "oauthtoken_expires_at" ON "oauth_tokens" ("expires_at");
-- create index "oauthtoken_oauth_grant_id_kind" to table: "oauth_tokens"
CREATE INDEX "oauthtoken_oauth_grant_id_kind" ON "oauth_tokens" ("oauth_grant_id", "kind");

-- +goose Down
-- reverse: create index "oauthtoken_oauth_grant_id_kind" to table: "oauth_tokens"
DROP INDEX "oauthtoken_oauth_grant_id_kind";
-- reverse: create index "oauthtoken_expires_at" to table: "oauth_tokens"
DROP INDEX "oauthtoken_expires_at";
-- reverse: create index "oauth_tokens_token_hash_key" to table: "oauth_tokens"
DROP INDEX "oauth_tokens_token_hash_key";
-- reverse: create "oauth_tokens" table
DROP TABLE "oauth_tokens";
-- reverse: create index "oauthgrant_client_id_owner_id_kind_user_id" to table: "oauth_grants"
DROP INDEX "oauthgrant_client_id_owner_id_kind_user_id";
-- reverse: create index "oauth_grant_user_id_idx" to table: "oauth_grants"
DROP INDEX "oauth_grant_user_id_idx";
-- reverse: create index "oauth_grant_owner_id_idx" to table: "oauth_grants"
DROP INDEX "oauth_grant_owner_id_idx";
-- reverse: create "oauth_grants" table
DROP TABLE "oauth_grants";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_organizations_oauth_grant_creators", DROP COLUMN "organization_oauth_grant_creators";
