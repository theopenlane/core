-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";

-- +goose Down
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "ref_code_sort_key";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "ref_code_sort_key";
