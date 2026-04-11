-- Modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "status" character varying NULL DEFAULT 'OPEN';
-- Modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "impact" DROP DEFAULT, ALTER COLUMN "risk_decision" SET DEFAULT 'NONE';
