-- Modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "sso_authorizations" jsonb NULL;
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "sso_authorizations" jsonb NULL;
