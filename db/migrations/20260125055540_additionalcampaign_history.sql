-- Modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "is_test" boolean NOT NULL DEFAULT false;
