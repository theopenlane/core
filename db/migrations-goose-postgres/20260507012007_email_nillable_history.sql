-- +goose Up
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "email" DROP NOT NULL;

-- +goose Down
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "email" SET NOT NULL;
