-- +goose Up
-- drop "directory_sync_runs" table
DROP TABLE IF EXISTS "directory_sync_runs" CASCADE;

-- +goose Down
-- irreversible: the dropped table cannot be restored from this migration
