-- +goose Up
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "audit_firm" character varying NULL, ADD COLUMN "auditor" character varying NULL, ADD COLUMN "auditor_email" character varying NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "audit_firm" character varying NULL, ADD COLUMN "auditor" character varying NULL, ADD COLUMN "auditor_email" character varying NULL;

-- +goose Down
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "auditor_email", DROP COLUMN "auditor", DROP COLUMN "audit_firm";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "auditor_email", DROP COLUMN "auditor", DROP COLUMN "audit_firm";
