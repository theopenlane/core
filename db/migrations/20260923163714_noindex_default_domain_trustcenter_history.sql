-- Modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "noindex_default_domain" boolean NULL DEFAULT true;
