-- Modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "status" SET DEFAULT 'IDENTIFIED';
-- Modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "status" SET DEFAULT 'IDENTIFIED';
