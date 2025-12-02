-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "secondary_background_color" character varying NULL, ADD COLUMN "secondary_foreground_color" character varying NULL;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "secondary_background_color" character varying NULL, ADD COLUMN "secondary_foreground_color" character varying NULL;
