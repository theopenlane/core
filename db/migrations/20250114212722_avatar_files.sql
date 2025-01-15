-- Modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "avatar_local_file_id" character varying NULL, ADD COLUMN "avatar_updated_at" timestamptz NULL;
-- Modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "avatar_local_file";
-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "avatar_updated_at" timestamptz NULL, ADD COLUMN "avatar_local_file_id" character varying NULL, ADD CONSTRAINT "organizations_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_file", DROP COLUMN "avatar_local_file", ADD CONSTRAINT "users_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
