-- Drop index "controls_auditor_reference_id_key" from table: "controls"
DROP INDEX "controls_auditor_reference_id_key";
-- Drop index "controls_reference_id_key" from table: "controls"
DROP INDEX "controls_reference_id_key";
-- Create index "control_auditor_reference_id_deleted_at_owner_id" to table: "controls"
CREATE INDEX "control_auditor_reference_id_deleted_at_owner_id" ON "controls" ("auditor_reference_id", "deleted_at", "owner_id");
-- Create index "control_reference_id_deleted_at_owner_id" to table: "controls"
CREATE INDEX "control_reference_id_deleted_at_owner_id" ON "controls" ("reference_id", "deleted_at", "owner_id");
-- Create index "subcontrol_auditor_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_auditor_reference_id_deleted_at_owner_id" ON "subcontrols" ("auditor_reference_id", "deleted_at", "owner_id");
-- Create index "subcontrol_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_reference_id_deleted_at_owner_id" ON "subcontrols" ("reference_id", "deleted_at", "owner_id");
