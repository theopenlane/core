-- +goose Up
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "is_suggested" boolean NOT NULL DEFAULT false, ADD COLUMN "priority" bigint NOT NULL DEFAULT 0, ADD COLUMN "available_at" timestamptz NULL, ADD COLUMN "source" character varying NULL, ADD COLUMN "source_key" character varying NULL;

-- +goose Down
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "source_key", DROP COLUMN "source", DROP COLUMN "available_at", DROP COLUMN "priority", DROP COLUMN "is_suggested", DROP COLUMN "metadata";
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "priority";
