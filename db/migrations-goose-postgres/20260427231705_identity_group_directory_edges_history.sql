-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" SET DEFAULT 'LOW';

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';
