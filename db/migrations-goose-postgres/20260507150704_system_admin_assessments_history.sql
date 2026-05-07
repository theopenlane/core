-- +goose Up
-- modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "email" DROP NOT NULL;
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "transform_configuration" jsonb NULL;

-- +goose Down
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "transform_configuration";
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "email" SET NOT NULL;
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
