-- +goose Up
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "note_files" character varying NULL, ADD CONSTRAINT "files_notes_files" FOREIGN KEY ("note_files") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_notes_files", DROP COLUMN "note_files";
