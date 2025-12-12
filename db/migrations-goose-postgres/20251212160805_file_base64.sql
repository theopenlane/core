-- +goose Up
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" DROP DEFAULT;
-- modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" DROP DEFAULT;
-- modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "base64_content" character varying NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "base64_content" character varying NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "base64_content";
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "base64_content";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
