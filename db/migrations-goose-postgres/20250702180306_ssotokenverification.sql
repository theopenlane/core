-- +goose Up
-- modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "sso_authorizations" jsonb NULL;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "sso_authorizations" jsonb NULL;

-- +goose Down
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "sso_authorizations";
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "sso_authorizations";
