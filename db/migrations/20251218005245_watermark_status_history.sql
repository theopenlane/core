-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ALTER COLUMN "watermarking_enabled" DROP NOT NULL, ALTER COLUMN "watermarking_enabled" DROP DEFAULT;
