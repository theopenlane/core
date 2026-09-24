-- +goose Up
-- modify "scans" table
ALTER TABLE "scans" ADD COLUMN "origin" character varying NOT NULL DEFAULT 'SYSTEM';

-- +goose Down
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP COLUMN "origin";
