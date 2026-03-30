-- +goose Up
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "domain_type" character varying NULL;

-- +goose Down
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "domain_type";
