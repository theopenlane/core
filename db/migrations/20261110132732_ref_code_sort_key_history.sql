-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "ref_code_sort_key" character varying NOT NULL DEFAULT '' COLLATE "C";
