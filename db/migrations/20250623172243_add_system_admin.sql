-- Modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false;
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "system_owned" boolean NULL DEFAULT false;
