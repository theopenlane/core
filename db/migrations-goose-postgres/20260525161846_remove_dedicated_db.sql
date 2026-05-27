-- +goose Up
-- modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "dedicated_db";

-- +goose Down
-- reverse: modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "dedicated_db" boolean NOT NULL DEFAULT false;
