-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
-- Modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "email" DROP NOT NULL;
-- Modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "transform_configuration" jsonb NULL;
