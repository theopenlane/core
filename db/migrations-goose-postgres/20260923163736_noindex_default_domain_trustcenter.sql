-- +goose Up
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "noindex_default_domain" boolean NULL DEFAULT true;

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP COLUMN "noindex_default_domain";
