-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "original_file_id" character varying NULL, ADD COLUMN "watermarking_enabled" boolean NOT NULL DEFAULT false, ADD COLUMN "watermark_status" character varying NULL DEFAULT 'DISABLED';
-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "watermarking_enabled" boolean NOT NULL DEFAULT false, ADD COLUMN "watermark_status" character varying NULL DEFAULT 'DISABLED', ADD COLUMN "original_file_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_files_original_file" FOREIGN KEY ("original_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
