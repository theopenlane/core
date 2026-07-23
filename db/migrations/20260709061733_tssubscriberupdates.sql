-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "allow_subscribers" boolean NULL DEFAULT true;
