-- +goose Up
-- modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "preview_status" character varying NULL DEFAULT 'NONE';
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "preview_status" character varying NULL DEFAULT 'NONE';

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP COLUMN "preview_status";
-- reverse: modify "trust_center_history" table
ALTER TABLE "trust_center_history" DROP COLUMN "preview_status";
