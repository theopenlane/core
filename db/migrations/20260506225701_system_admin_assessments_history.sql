-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "system_internal_id" character varying NULL;
