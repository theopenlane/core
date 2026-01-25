-- +goose Up
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "is_test" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP COLUMN "is_test";
