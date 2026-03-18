-- Modify "directory_account_history" table
ALTER TABLE "directory_account_history" ALTER COLUMN "integration_id" DROP NOT NULL, ALTER COLUMN "directory_sync_run_id" DROP NOT NULL, ADD COLUMN "platform_id" character varying NULL, ADD COLUMN "identity_holder_id" character varying NULL, ADD COLUMN "directory_name" character varying NULL, ADD COLUMN "avatar_remote_url" character varying NULL, ADD COLUMN "avatar_local_file_id" character varying NULL, ADD COLUMN "avatar_updated_at" timestamptz NULL;
-- Modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "platform_id" character varying NULL;
-- Modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "platform_id" character varying NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "platform_id" character varying NULL;
