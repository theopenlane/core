-- +goose Up
-- modify "discussion_history" table
ALTER TABLE "discussion_history" ALTER COLUMN "external_id" DROP NOT NULL;

-- +goose Down
-- reverse: modify "discussion_history" table
ALTER TABLE "discussion_history" ALTER COLUMN "external_id" SET NOT NULL;
