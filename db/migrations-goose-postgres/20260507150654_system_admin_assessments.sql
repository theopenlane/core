-- +goose Up
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "email" DROP NOT NULL;
-- modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "transform_configuration" jsonb NULL;

-- +goose Down
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "transform_configuration";
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "email" SET NOT NULL;
