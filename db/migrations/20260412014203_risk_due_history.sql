-- Modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "due_date" timestamptz NULL;
