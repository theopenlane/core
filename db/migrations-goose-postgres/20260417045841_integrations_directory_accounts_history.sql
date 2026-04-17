-- +goose Up
-- modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "email_aliases" jsonb NULL, ADD COLUMN "phone_number" character varying NULL;
-- modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "avatar_remote_url" character varying NULL;

-- +goose Down
-- reverse: modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" DROP COLUMN "avatar_remote_url";
-- reverse: modify "directory_account_history" table
ALTER TABLE "directory_account_history" DROP COLUMN "phone_number", DROP COLUMN "email_aliases";
