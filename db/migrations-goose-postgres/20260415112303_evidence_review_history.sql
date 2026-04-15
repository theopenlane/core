-- +goose Up
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY';

-- +goose Down
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "review_frequency";
