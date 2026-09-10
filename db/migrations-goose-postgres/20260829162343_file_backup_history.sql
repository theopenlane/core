-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "backup_state" jsonb NULL;

-- +goose Down
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "backup_state";
