-- +goose Up
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "is_template" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "is_template";
