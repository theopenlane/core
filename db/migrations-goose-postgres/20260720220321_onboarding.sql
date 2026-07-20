-- +goose Up
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;

-- +goose Down
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "priority";
