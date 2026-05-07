-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "email" DROP NOT NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "transform_configuration" jsonb NULL;
