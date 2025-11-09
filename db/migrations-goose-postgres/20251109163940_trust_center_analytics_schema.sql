-- +goose Up
-- modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "pirsch_domain_id" character varying NULL, ADD COLUMN "pirsch_identification_code" character varying NULL;
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "pirsch_domain_id" character varying NULL, ADD COLUMN "pirsch_identification_code" character varying NULL;

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP COLUMN "pirsch_identification_code", DROP COLUMN "pirsch_domain_id";
-- reverse: modify "trust_center_history" table
ALTER TABLE "trust_center_history" DROP COLUMN "pirsch_identification_code", DROP COLUMN "pirsch_domain_id";
