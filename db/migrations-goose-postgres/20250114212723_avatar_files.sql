-- +goose Up
-- modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "avatar_local_file_id" character varying NULL, ADD COLUMN "avatar_updated_at" timestamptz NULL;
-- modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "avatar_local_file";
-- modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "avatar_updated_at" timestamptz NULL, ADD COLUMN "avatar_local_file_id" character varying NULL, ADD CONSTRAINT "organizations_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_file", DROP COLUMN "avatar_local_file", ADD CONSTRAINT "users_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_avatar_file", ADD COLUMN "avatar_local_file" character varying NULL, ADD CONSTRAINT "users_files_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP CONSTRAINT "organizations_files_avatar_file", DROP COLUMN "avatar_local_file_id", DROP COLUMN "avatar_updated_at";
-- reverse: modify "user_history" table
ALTER TABLE "user_history" ADD COLUMN "avatar_local_file" character varying NULL;
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "avatar_updated_at", DROP COLUMN "avatar_local_file_id";
