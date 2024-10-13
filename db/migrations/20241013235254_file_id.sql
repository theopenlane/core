-- Modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "file_id", ADD COLUMN "avatar_local_file_id" character varying NULL;
-- Modify "users" table
ALTER TABLE "users" DROP COLUMN "file_id", ADD COLUMN "avatar_local_file_id" character varying NULL, ADD CONSTRAINT "users_files_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
