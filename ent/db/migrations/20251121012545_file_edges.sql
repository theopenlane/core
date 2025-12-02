-- Modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" DROP COLUMN "logo_local_file_id", ADD COLUMN "logo_file_id" character varying NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_subprocessor_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_subprocessor_creators" FOREIGN KEY ("organization_subprocessor_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subprocessors" table
ALTER TABLE "subprocessors" DROP CONSTRAINT "subprocessors_files_logo_file", DROP COLUMN "logo_local_file_id", ADD COLUMN "logo_file_id" character varying NULL, ADD CONSTRAINT "subprocessors_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
