-- +goose Up
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "company_name" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "company_name";
