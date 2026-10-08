-- +goose Up
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "user_input" jsonb NULL, ADD COLUMN "operation_config" jsonb NULL;

-- +goose Down
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "operation_config", DROP COLUMN "user_input";
