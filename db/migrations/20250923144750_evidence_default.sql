-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- Modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
