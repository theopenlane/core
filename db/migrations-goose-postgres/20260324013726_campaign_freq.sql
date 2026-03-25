-- +goose Up
-- modify "campaigns" table
ALTER TABLE "campaigns" ALTER COLUMN "recurrence_frequency" SET DEFAULT 'NONE';

-- +goose Down
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" ALTER COLUMN "recurrence_frequency" DROP DEFAULT;
