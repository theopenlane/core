-- +goose Up
-- modify "notifications" table
ALTER TABLE "notifications" ADD COLUMN "requestor_id" character varying NULL;

-- +goose Down
-- reverse: modify "notifications" table
ALTER TABLE "notifications" DROP COLUMN "requestor_id";
