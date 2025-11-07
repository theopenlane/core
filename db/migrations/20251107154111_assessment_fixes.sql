-- Modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_template", ADD CONSTRAINT "assessments_templates_assessments" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
