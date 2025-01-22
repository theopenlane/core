-- +goose Up
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "tags";

-- +goose Down
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD COLUMN "tags" jsonb NULL;
