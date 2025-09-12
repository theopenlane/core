-- +goose Up
-- modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "internal_id" character varying NULL;
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "internal_notes" character varying NULL, ADD COLUMN "internal_id" character varying NULL;

-- +goose Down
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "internal_id", DROP COLUMN "internal_notes";
-- reverse: modify "mapped_control_history" table
ALTER TABLE "mapped_control_history" DROP COLUMN "internal_id", DROP COLUMN "internal_notes";
