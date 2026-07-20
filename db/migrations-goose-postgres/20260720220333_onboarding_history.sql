-- +goose Up
-- modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "priority" bigint NOT NULL DEFAULT 0;

-- +goose Down
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "priority";
