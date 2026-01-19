-- +goose Up
-- modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "approval_metadata" jsonb NULL, ADD COLUMN "rejection_metadata" jsonb NULL, ADD COLUMN "invalidation_metadata" jsonb NULL;

-- +goose Down
-- reverse: modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" DROP COLUMN "invalidation_metadata", DROP COLUMN "rejection_metadata", DROP COLUMN "approval_metadata";
