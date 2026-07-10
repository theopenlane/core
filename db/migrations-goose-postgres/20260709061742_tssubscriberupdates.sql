-- +goose Up
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "allow_subscribers" boolean NULL DEFAULT true;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "allow_subscribers";
