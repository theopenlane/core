-- Modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;
-- Modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- Modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_template", DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800, ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
