-- +goose Up
-- modify "review_history" table
ALTER TABLE "review_history" ADD COLUMN "status" character varying NULL DEFAULT 'OPEN';
-- modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "impact" DROP DEFAULT, ALTER COLUMN "risk_decision" SET DEFAULT 'NONE';

-- +goose Down
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "risk_decision" SET DEFAULT ' NONE', ALTER COLUMN "impact" SET DEFAULT 'MODERATE';
-- reverse: modify "review_history" table
ALTER TABLE "review_history" DROP COLUMN "status";
