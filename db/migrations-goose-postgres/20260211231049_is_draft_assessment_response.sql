-- +goose Up
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "is_draft" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP COLUMN "is_draft";
