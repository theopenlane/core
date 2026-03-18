-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "status" SET DEFAULT 'ACTIVE';

-- +goose Down
-- reverse: modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "status" SET DEFAULT 'active';
