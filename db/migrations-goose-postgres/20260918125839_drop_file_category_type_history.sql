-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "category_type";

-- +goose Down
-- reverse: modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "category_type" character varying NULL;
