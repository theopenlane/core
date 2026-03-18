-- +goose Up
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "status_page_url" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "status_page_url";
