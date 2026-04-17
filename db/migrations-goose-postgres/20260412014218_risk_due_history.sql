-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "due_date" timestamptz NULL;

-- +goose Down
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "due_date", ALTER COLUMN "status" SET DEFAULT 'IDENTIFIED';
