-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "expires_at" timestamptz NULL;
