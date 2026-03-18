-- +goose Up
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "avatar_local_file_id" character varying NULL;

-- +goose Down
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "avatar_local_file_id";
