-- +goose Up
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "next_review_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "next_review_at", DROP COLUMN "review_frequency";
