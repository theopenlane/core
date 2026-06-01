-- +goose Up
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "is_template" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "is_template";
