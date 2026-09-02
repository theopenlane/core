-- +goose Up
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "backup_state" jsonb NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "backup_state";
