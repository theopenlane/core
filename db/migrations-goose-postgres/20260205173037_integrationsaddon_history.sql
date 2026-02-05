-- +goose Up
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "config" jsonb NULL, ADD COLUMN "provider_state" jsonb NULL;

-- +goose Down
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "provider_state", DROP COLUMN "config";
