-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ALTER COLUMN "template_id" DROP NOT NULL, DROP COLUMN "assessment_owner_id", ADD COLUMN "jsonconfig" jsonb NULL, ADD COLUMN "uischema" jsonb NULL, ADD COLUMN "response_due_duration" bigint NULL;
-- Modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ALTER COLUMN "status" SET DEFAULT 'SENT';
-- Modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "template_id" DROP NOT NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_template", DROP COLUMN "assessment_owner_id", ALTER COLUMN "template_id" DROP NOT NULL, ADD COLUMN "jsonconfig" jsonb NULL, ADD COLUMN "uischema" jsonb NULL, ADD COLUMN "response_due_duration" bigint NULL, ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_templates_documents", ALTER COLUMN "template_id" DROP NOT NULL, ADD CONSTRAINT "document_data_templates_documents" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
