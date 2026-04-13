-- +goose Up
-- modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "email_aliases" jsonb NULL;

-- +goose Down
-- reverse: modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" DROP COLUMN "email_aliases";
