-- Modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "correlation_id";
-- Modify "files" table
ALTER TABLE "files" DROP COLUMN "correlation_id";
