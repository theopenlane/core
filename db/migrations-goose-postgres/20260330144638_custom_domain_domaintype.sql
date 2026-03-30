-- +goose Up
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "domain_type" character varying NOT NULL DEFAULT 'UNKNOWN';

-- +goose Down
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "domain_type";
