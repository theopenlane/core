-- +goose Up
-- modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "file_id", ADD COLUMN "avatar_local_file_id" character varying NULL;
-- modify "users" table
ALTER TABLE "users" DROP COLUMN "file_id", ADD COLUMN "avatar_local_file_id" character varying NULL, ADD CONSTRAINT "users_files_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_files_file", DROP COLUMN "avatar_local_file_id", ADD COLUMN "file_id" jsonb NULL;
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "avatar_local_file_id", ADD COLUMN "file_id" jsonb NULL;
