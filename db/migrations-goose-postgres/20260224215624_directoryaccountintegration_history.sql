-- +goose Up
-- modify "directory_account_history" table
ALTER TABLE "directory_account_history" ALTER COLUMN "integration_id" DROP NOT NULL, ALTER COLUMN "directory_sync_run_id" DROP NOT NULL, ADD COLUMN "platform_id" character varying NULL, ADD COLUMN "identity_holder_id" character varying NULL, ADD COLUMN "directory_name" character varying NULL, ADD COLUMN "avatar_remote_url" character varying NULL, ADD COLUMN "avatar_local_file_id" character varying NULL, ADD COLUMN "avatar_updated_at" timestamptz NULL;
-- modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "platform_id" character varying NULL;
-- modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "platform_id" character varying NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "platform_id" character varying NULL;

-- +goose Down
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "platform_id";
-- reverse: modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" DROP COLUMN "platform_id";
-- reverse: modify "directory_group_history" table
ALTER TABLE "directory_group_history" DROP COLUMN "platform_id";
-- reverse: modify "directory_account_history" table
ALTER TABLE "directory_account_history" DROP COLUMN "avatar_updated_at", DROP COLUMN "avatar_local_file_id", DROP COLUMN "avatar_remote_url", DROP COLUMN "directory_name", DROP COLUMN "identity_holder_id", DROP COLUMN "platform_id", ALTER COLUMN "directory_sync_run_id" SET NOT NULL, ALTER COLUMN "integration_id" SET NOT NULL;
