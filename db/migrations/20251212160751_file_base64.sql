-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" DROP DEFAULT;
-- Modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" DROP DEFAULT;
-- Modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "base64_content" character varying NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "base64_content" character varying NULL;
