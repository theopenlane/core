-- Modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "pirsch_domain_id" character varying NULL, ADD COLUMN "pirsch_identification_code" character varying NULL;
-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "pirsch_domain_id" character varying NULL, ADD COLUMN "pirsch_identification_code" character varying NULL;
