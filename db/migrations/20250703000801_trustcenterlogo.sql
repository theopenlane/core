-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "logo_remote_url" character varying NULL, ADD COLUMN "logo_local_file_id" character varying NULL;
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "logo_remote_url" character varying NULL, ADD COLUMN "logo_local_file_id" character varying NULL, ADD CONSTRAINT "trust_center_settings_files_logo_file" FOREIGN KEY ("logo_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "trust_center_setting_files" table
CREATE TABLE "trust_center_setting_files" ("trust_center_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("trust_center_setting_id", "file_id"), CONSTRAINT "trust_center_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "trust_center_setting_files_trust_center_setting_id" FOREIGN KEY ("trust_center_setting_id") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
