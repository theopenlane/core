-- +goose Up
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "email" DROP NOT NULL;

-- +goose Down
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "email" SET NOT NULL;
