-- Modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "approval_metadata" jsonb NULL, ADD COLUMN "rejection_metadata" jsonb NULL, ADD COLUMN "invalidation_metadata" jsonb NULL;
