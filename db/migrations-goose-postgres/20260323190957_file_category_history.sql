-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "file_category_name" character varying NULL, ADD COLUMN "file_category_id" character varying NULL;

-- +goose Down
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "file_category_id", DROP COLUMN "file_category_name";
