-- +goose Up
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "pirsch_access_link" character varying NULL;

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP COLUMN "pirsch_access_link";
