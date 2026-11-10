-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "ref_code_sort_key";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "ref_code_sort_key";
