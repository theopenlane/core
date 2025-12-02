-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "theme_mode" character varying NULL DEFAULT 'EASY', ADD COLUMN "font" character varying NULL, ADD COLUMN "foreground_color" character varying NULL, ADD COLUMN "background_color" character varying NULL, ADD COLUMN "accent_color" character varying NULL;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "theme_mode" character varying NULL DEFAULT 'EASY', ADD COLUMN "font" character varying NULL, ADD COLUMN "foreground_color" character varying NULL, ADD COLUMN "background_color" character varying NULL, ADD COLUMN "accent_color" character varying NULL;
