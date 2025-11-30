-- Modify "notes" table
ALTER TABLE "notes" ADD COLUMN "evidence_comments" character varying NULL, ADD CONSTRAINT "notes_evidences_comments" FOREIGN KEY ("evidence_comments") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
