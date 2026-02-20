-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "trust_center_visibility" character varying NULL DEFAULT 'NOT_VISIBLE', ADD COLUMN "is_trust_center_control" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "is_trust_center_control", DROP COLUMN "trust_center_visibility";
