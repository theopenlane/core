-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" DROP COLUMN "category", ADD COLUMN "trust_center_doc_kind_name" character varying NULL, ADD COLUMN "trust_center_doc_kind_id" character varying NULL;
