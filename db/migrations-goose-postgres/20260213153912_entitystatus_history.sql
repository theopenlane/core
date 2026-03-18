-- +goose Up
-- modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "status" SET DEFAULT 'ACTIVE';

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "status" SET DEFAULT 'active';
