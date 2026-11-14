-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "control_origin_name" character varying NULL, ADD COLUMN "control_origin_id" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "control_origin_name" character varying NULL, ADD COLUMN "control_origin_id" character varying NULL;
