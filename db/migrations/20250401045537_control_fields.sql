-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
-- Drop index "control_standard_id_ref_code" from table: "controls"
DROP INDEX "control_standard_id_ref_code";
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
-- Create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NULL));
-- Create index "controls_auditor_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_auditor_reference_id_key" ON "controls" ("auditor_reference_id");
-- Create index "controls_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_reference_id_key" ON "controls" ("reference_id");
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
