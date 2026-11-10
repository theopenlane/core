-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "disable_default_domain" boolean NULL DEFAULT true;
