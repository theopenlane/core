-- +goose Up
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "secondary_background_color" character varying NULL, ADD COLUMN "secondary_foreground_color" character varying NULL;
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "secondary_background_color" character varying NULL, ADD COLUMN "secondary_foreground_color" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "secondary_foreground_color", DROP COLUMN "secondary_background_color";
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "secondary_foreground_color", DROP COLUMN "secondary_background_color";
