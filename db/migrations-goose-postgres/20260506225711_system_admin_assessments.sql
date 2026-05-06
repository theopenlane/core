-- +goose Up
-- modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;

-- +goose Down
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
