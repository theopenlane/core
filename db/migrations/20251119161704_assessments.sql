-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ALTER COLUMN "template_id" DROP NOT NULL, ALTER COLUMN "response_due_duration" DROP NOT NULL, ALTER COLUMN "response_due_duration" DROP DEFAULT, ADD COLUMN "jsonconfig" jsonb NULL, ADD COLUMN "uischema" jsonb NULL;
-- Modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "template_id" DROP NOT NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_assessments", ALTER COLUMN "template_id" DROP NOT NULL, ALTER COLUMN "response_due_duration" DROP NOT NULL, ALTER COLUMN "response_due_duration" DROP DEFAULT, ADD COLUMN "jsonconfig" jsonb NULL, ADD COLUMN "uischema" jsonb NULL, ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_templates_documents", ALTER COLUMN "template_id" DROP NOT NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
