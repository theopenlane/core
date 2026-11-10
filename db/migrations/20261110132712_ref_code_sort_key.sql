-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";
