-- +goose Up
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "logo_file_id" character varying NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "logo_file_id" character varying NULL, ADD CONSTRAINT "standards_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP CONSTRAINT "standards_files_logo_file", DROP COLUMN "logo_file_id";
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "logo_file_id";
