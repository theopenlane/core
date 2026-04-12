-- Modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "due_date" timestamptz NULL;
