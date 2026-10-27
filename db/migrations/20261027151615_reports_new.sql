-- Modify "reports" table
ALTER TABLE "reports" ALTER COLUMN "visibility" DROP NOT NULL, ALTER COLUMN "visibility" DROP DEFAULT;
