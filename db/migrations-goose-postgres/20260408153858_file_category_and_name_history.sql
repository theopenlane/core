-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "category_name" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD COLUMN "name" character varying NULL;

-- +goose Down
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "name", DROP COLUMN "category_id", DROP COLUMN "category_name";
