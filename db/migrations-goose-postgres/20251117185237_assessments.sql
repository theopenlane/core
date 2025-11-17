-- +goose Up
-- modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;

-- +goose Down
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "response_due_duration", ADD COLUMN "assessment_owner_id" character varying NULL;
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "status" SET DEFAULT 'NOT_STARTED';
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "status" SET DEFAULT 'NOT_STARTED';
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "response_due_duration", ADD COLUMN "assessment_owner_id" character varying NULL;
