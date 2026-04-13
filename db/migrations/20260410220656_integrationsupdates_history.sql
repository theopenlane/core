-- Modify "identity_holder_history" table
ALTER TABLE "identity_holder_history" ADD COLUMN "email_aliases" jsonb NULL;
