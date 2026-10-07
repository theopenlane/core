-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "sha256_hash" character varying NULL, ADD COLUMN "provenance" jsonb NULL;
