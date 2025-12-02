-- Modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "kind" character varying NULL DEFAULT 'QUESTIONNAIRE';
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "kind" character varying NULL DEFAULT 'QUESTIONNAIRE';
