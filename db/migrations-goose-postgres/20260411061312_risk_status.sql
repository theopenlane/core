-- +goose Up
-- modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "status" character varying NULL DEFAULT 'OPEN';
-- modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "impact" DROP DEFAULT, ALTER COLUMN "risk_decision" SET DEFAULT 'NONE';

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "risk_decision" SET DEFAULT ' NONE', ALTER COLUMN "impact" SET DEFAULT 'MODERATE';
-- reverse: modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "status";
