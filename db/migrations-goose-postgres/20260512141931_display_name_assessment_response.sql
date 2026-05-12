-- +goose Up
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "display_name" character varying NULL;

-- +goose Down
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP COLUMN "display_name";
