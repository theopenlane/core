-- +goose Up
-- modify "files" table
ALTER TABLE "files" DROP COLUMN "category_type";

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" ADD COLUMN "category_type" character varying NULL;
