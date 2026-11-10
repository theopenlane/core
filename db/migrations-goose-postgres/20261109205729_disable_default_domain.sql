-- +goose Up
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "disable_default_domain" boolean NULL DEFAULT true;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "disable_default_domain";
