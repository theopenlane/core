-- Modify "review_history" table
ALTER TABLE "review_history" ADD COLUMN "status" character varying NULL DEFAULT 'OPEN';
-- Modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "impact" DROP DEFAULT, ALTER COLUMN "risk_decision" SET DEFAULT 'NONE';
