-- +goose Up
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "hero_image_local_file_id" character varying NULL, ADD CONSTRAINT "trust_center_settings_files_hero_image_file" FOREIGN KEY ("hero_image_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP CONSTRAINT "trust_center_settings_files_hero_image_file", DROP COLUMN "hero_image_local_file_id";
