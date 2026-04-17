-- +goose Up
-- modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "email_aliases" jsonb NULL, ADD COLUMN "phone_number" character varying NULL;
-- modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "avatar_remote_url" character varying NULL;

-- +goose Down
-- reverse: modify "identity_holders" table
ALTER TABLE "identity_holders" DROP COLUMN "avatar_remote_url";
-- reverse: modify "directory_accounts" table
ALTER TABLE "directory_accounts" DROP COLUMN "phone_number", DROP COLUMN "email_aliases";
