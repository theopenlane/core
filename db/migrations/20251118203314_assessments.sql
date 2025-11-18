-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ALTER COLUMN "template_id" DROP NOT NULL, ADD COLUMN "jsonconfig" jsonb NULL, ADD COLUMN "uischema" jsonb NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_assessments", ALTER COLUMN "template_id" DROP NOT NULL, ADD COLUMN "jsonconfig" jsonb NULL, ADD COLUMN "uischema" jsonb NULL, ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
