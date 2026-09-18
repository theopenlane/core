-- +goose Up
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "expires_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "expires_at";
