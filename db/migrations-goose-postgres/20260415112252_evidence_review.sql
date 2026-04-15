-- +goose Up
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY';

-- +goose Down
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "review_frequency";
