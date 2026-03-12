-- +goose Up
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "avatar_local_file_id" character varying NULL, ADD CONSTRAINT "groups_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_files_avatar_file", DROP COLUMN "avatar_local_file_id";
