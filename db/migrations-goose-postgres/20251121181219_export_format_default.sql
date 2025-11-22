-- +goose Up
-- modify "exports" table
ALTER TABLE "exports" ALTER COLUMN "format" SET DEFAULT 'CSV';

-- +goose Down
-- reverse: modify "exports" table
ALTER TABLE "exports" ALTER COLUMN "format" DROP DEFAULT;
