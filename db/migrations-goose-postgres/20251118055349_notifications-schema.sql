-- +goose Up
-- modify "notifications" table
ALTER TABLE "notifications" ADD COLUMN "topic" character varying NULL;

-- +goose Down
-- reverse: modify "notifications" table
ALTER TABLE "notifications" DROP COLUMN "topic";
