-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "file_id" character varying NULL, ADD COLUMN "url" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "file_id" character varying NULL, ADD COLUMN "url" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "internal_policies_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "procedures_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
