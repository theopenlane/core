-- +goose Up
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "status" SET DEFAULT 'SENT';

-- +goose Down
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "status" SET DEFAULT 'NOT_STARTED';
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "status" SET DEFAULT 'NOT_STARTED';
