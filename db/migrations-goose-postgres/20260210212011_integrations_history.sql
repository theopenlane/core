-- +goose Up
-- modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" ALTER COLUMN "approval_submission_mode" SET DEFAULT 'AUTO_SUBMIT';

-- +goose Down
-- reverse: modify "workflow_definition_history" table
ALTER TABLE "workflow_definition_history" ALTER COLUMN "approval_submission_mode" SET DEFAULT 'MANUAL_SUBMIT';
