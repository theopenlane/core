-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "standard_id" character varying NULL;
-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "standard_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_standards_trust_center_docs" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
