-- +goose Up
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "trust_center_id" character varying NULL;

-- +goose Down
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "trust_center_id";
