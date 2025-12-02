-- +goose Up
-- modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "trust_center_id" character varying NULL, ADD COLUMN "title" character varying NOT NULL, ADD COLUMN "category" character varying NOT NULL, ADD COLUMN "file_id" character varying NULL, ADD COLUMN "visibility" character varying NULL DEFAULT 'NOT_VISIBLE';
-- modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "title" character varying NOT NULL, ADD COLUMN "category" character varying NOT NULL, ADD COLUMN "visibility" character varying NULL DEFAULT 'NOT_VISIBLE', ADD COLUMN "trust_center_id" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_center_docs_trust_centers_trust_center_docs" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" DROP CONSTRAINT "trust_center_docs_trust_centers_trust_center_docs", DROP CONSTRAINT "trust_center_docs_files_file", DROP COLUMN "file_id", DROP COLUMN "trust_center_id", DROP COLUMN "visibility", DROP COLUMN "category", DROP COLUMN "title";
-- reverse: modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" DROP COLUMN "visibility", DROP COLUMN "file_id", DROP COLUMN "category", DROP COLUMN "title", DROP COLUMN "trust_center_id";
