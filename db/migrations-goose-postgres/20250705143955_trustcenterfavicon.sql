-- +goose Up
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "favicon_remote_url" character varying NULL, ADD COLUMN "favicon_local_file_id" character varying NULL;
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "favicon_remote_url" character varying NULL, ADD COLUMN "favicon_local_file_id" character varying NULL, ADD CONSTRAINT "trust_center_settings_files_favicon_file" FOREIGN KEY ("favicon_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP CONSTRAINT "trust_center_settings_files_favicon_file", DROP COLUMN "favicon_local_file_id", DROP COLUMN "favicon_remote_url";
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "favicon_local_file_id", DROP COLUMN "favicon_remote_url";
