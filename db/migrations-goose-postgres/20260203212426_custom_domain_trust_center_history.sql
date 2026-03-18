-- +goose Up
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "trust_center_id" character varying NULL;

-- +goose Down
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "trust_center_id";
