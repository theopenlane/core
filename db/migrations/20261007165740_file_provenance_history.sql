-- Modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "sha256_hash" character varying NULL, ADD COLUMN "provenance" jsonb NULL;
