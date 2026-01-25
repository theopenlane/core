-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "is_test" boolean NOT NULL DEFAULT false;
