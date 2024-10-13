-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "correlation_id";
-- modify "files" table
ALTER TABLE "files" DROP COLUMN "correlation_id";

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" ADD COLUMN "correlation_id" character varying NULL;
-- reverse: modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "correlation_id" character varying NULL;
