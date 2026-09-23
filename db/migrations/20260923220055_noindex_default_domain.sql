-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "noindex_default_domain" boolean NULL DEFAULT true;
