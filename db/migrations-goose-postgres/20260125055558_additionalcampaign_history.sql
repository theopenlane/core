-- +goose Up
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "is_test" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" DROP COLUMN "is_test";
