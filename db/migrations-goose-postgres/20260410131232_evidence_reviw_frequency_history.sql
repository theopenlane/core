-- +goose Up
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "next_review_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "next_review_at", DROP COLUMN "review_frequency";
