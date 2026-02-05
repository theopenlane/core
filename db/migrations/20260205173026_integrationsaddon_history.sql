-- Modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "config" jsonb NULL, ADD COLUMN "provider_state" jsonb NULL;
