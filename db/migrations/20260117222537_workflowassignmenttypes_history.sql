-- Modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "approval_metadata" jsonb NULL, ADD COLUMN "rejection_metadata" jsonb NULL, ADD COLUMN "invalidation_metadata" jsonb NULL;
