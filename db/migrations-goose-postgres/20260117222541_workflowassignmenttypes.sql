-- +goose Up
-- modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "approval_metadata" jsonb NULL, ADD COLUMN "rejection_metadata" jsonb NULL, ADD COLUMN "invalidation_metadata" jsonb NULL;

-- +goose Down
-- reverse: modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" DROP COLUMN "invalidation_metadata", DROP COLUMN "rejection_metadata", DROP COLUMN "approval_metadata";
