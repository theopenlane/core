-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "sha256_hash" character varying NULL, ADD COLUMN "provenance" jsonb NULL;

-- +goose Down
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "provenance", DROP COLUMN "sha256_hash";
