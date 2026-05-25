-- +goose Up
-- modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "dedicated_db";

-- +goose Down
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "dedicated_db" boolean NOT NULL DEFAULT false;
