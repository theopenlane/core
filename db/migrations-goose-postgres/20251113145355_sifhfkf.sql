-- +goose Up
-- modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;
-- modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "assessment_blocked_groups", DROP COLUMN "assessment_editors", DROP COLUMN "assessment_viewers";

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "assessment_viewers" character varying NULL, ADD COLUMN "assessment_editors" character varying NULL, ADD COLUMN "assessment_blocked_groups" character varying NULL;
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "response_due_duration", ADD COLUMN "assessment_owner_id" character varying NULL;
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "response_due_duration", ADD COLUMN "assessment_owner_id" character varying NULL;
