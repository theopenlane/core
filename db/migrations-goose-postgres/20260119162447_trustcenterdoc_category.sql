-- +goose Up
-- modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" DROP COLUMN "category", ADD COLUMN "trust_center_doc_kind_name" character varying NULL, ADD COLUMN "trust_center_doc_kind_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_custom_type_enums_trust_center_doc_kind" FOREIGN KEY ("trust_center_doc_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" DROP CONSTRAINT "trust_center_docs_custom_type_enums_trust_center_doc_kind", DROP COLUMN "trust_center_doc_kind_id", DROP COLUMN "trust_center_doc_kind_name", ADD COLUMN "category" character varying NOT NULL;
