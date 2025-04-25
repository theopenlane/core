-- Modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "audit_firm" character varying NULL, ADD COLUMN "auditor" character varying NULL, ADD COLUMN "auditor_email" character varying NULL;
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "audit_firm" character varying NULL, ADD COLUMN "auditor" character varying NULL, ADD COLUMN "auditor_email" character varying NULL;
