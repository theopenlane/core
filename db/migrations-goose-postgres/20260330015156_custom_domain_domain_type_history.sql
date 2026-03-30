-- +goose Up
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "domain_type" character varying NULL;

-- +goose Down
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "domain_type";
