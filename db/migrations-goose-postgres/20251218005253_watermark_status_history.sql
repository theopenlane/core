-- +goose Up
-- modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ALTER COLUMN "watermarking_enabled" DROP NOT NULL, ALTER COLUMN "watermarking_enabled" DROP DEFAULT;

-- +goose Down
-- reverse: modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ALTER COLUMN "watermarking_enabled" SET NOT NULL, ALTER COLUMN "watermarking_enabled" SET DEFAULT false;
