-- Modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "user_input" jsonb NULL, ADD COLUMN "operation_config" jsonb NULL;
