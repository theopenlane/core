-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "noindex_default_domain" boolean NULL DEFAULT true;
