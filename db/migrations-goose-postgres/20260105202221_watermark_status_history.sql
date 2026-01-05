-- +goose Up
-- modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ALTER COLUMN "watermark_status" SET DEFAULT 'PENDING';

-- +goose Down
-- reverse: modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ALTER COLUMN "watermark_status" SET DEFAULT 'DISABLED';
