-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ALTER COLUMN "watermarking_enabled" DROP NOT NULL, ALTER COLUMN "watermarking_enabled" DROP DEFAULT;
