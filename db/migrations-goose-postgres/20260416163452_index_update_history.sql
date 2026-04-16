-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" DROP DEFAULT;
