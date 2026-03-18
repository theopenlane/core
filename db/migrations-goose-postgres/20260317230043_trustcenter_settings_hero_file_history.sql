-- +goose Up
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "hero_image_local_file_id" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "hero_image_local_file_id";
