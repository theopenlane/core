-- Modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "email_aliases" jsonb NULL, ADD COLUMN "phone_number" character varying NULL;
-- Modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "avatar_remote_url" character varying NULL;
