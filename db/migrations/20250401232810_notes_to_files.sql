-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "note_files" character varying NULL, ADD CONSTRAINT "files_notes_files" FOREIGN KEY ("note_files") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
