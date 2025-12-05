-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "status" SET DEFAULT 'IDENTIFIED';
-- modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "status" SET DEFAULT 'IDENTIFIED';

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "status" SET DEFAULT 'OPEN';
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "status" SET DEFAULT 'OPEN';
