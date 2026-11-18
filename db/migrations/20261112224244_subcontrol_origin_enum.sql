-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "control_origin_name" character varying NULL, ADD COLUMN "control_origin_id" character varying NULL, ADD CONSTRAINT "controls_custom_type_enums_control_origin" FOREIGN KEY ("control_origin_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "control_origin_name" character varying NULL, ADD COLUMN "control_origin_id" character varying NULL, ADD CONSTRAINT "subcontrols_custom_type_enums_control_origin" FOREIGN KEY ("control_origin_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
