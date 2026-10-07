-- +goose Up
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "sha256_hash" character varying NULL, ADD COLUMN "provenance" jsonb NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "provenance", DROP COLUMN "sha256_hash";
