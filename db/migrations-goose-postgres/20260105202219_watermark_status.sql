-- +goose Up
-- modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ALTER COLUMN "watermark_status" SET DEFAULT 'PENDING';

-- +goose Down
-- reverse: modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ALTER COLUMN "watermark_status" SET DEFAULT 'DISABLED';
