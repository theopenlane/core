-- +goose Up
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "display_name" character varying NULL;

-- +goose Down
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" DROP COLUMN "display_name";
