-- +goose Up
-- modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "due_date" timestamptz NULL;

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "due_date", ALTER COLUMN "status" SET DEFAULT 'IDENTIFIED';
