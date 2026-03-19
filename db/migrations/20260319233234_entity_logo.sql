-- Modify "entities" table
ALTER TABLE "entities" ADD COLUMN "logo_file_id" character varying NULL, ADD CONSTRAINT "entities_files_logo_file" FOREIGN KEY ("logo_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
