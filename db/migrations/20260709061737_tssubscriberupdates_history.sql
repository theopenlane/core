-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "allow_subscribers" boolean NULL DEFAULT true;
