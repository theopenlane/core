-- Modify "report_history" table
ALTER TABLE "report_history" ALTER COLUMN "visibility" DROP NOT NULL, ALTER COLUMN "visibility" DROP DEFAULT;
