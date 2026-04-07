-- +goose Up
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "file_category_name" character varying NULL, ADD COLUMN "file_category_id" character varying NULL, ADD CONSTRAINT "files_custom_type_enums_file_category" FOREIGN KEY ("file_category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_custom_type_enums_file_category", DROP COLUMN "file_category_id", DROP COLUMN "file_category_name";
