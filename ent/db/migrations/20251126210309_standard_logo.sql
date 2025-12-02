-- Modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "logo_file_id" character varying NULL;
-- Modify "standards" table
ALTER TABLE "standards" ADD COLUMN "logo_file_id" character varying NULL, ADD CONSTRAINT "standards_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
