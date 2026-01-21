-- +goose Up
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "company_name" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "company_name";
