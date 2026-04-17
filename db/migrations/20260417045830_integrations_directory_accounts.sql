-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "email_aliases" jsonb NULL, ADD COLUMN "phone_number" character varying NULL;
-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "avatar_remote_url" character varying NULL;
