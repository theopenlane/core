-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "is_draft" boolean NOT NULL DEFAULT false;
