-- +goose Up
-- modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;

-- +goose Down
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "system_internal_id", DROP COLUMN "internal_notes", DROP COLUMN "system_owned";
