-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "reference_framework_revision" character varying NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "reference_framework_revision" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "reference_framework_revision" character varying NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "reference_framework_revision" character varying NULL;
