-- Drop index "control_auditor_reference_id_deleted_at_owner_id" from table: "controls"
DROP INDEX "control_auditor_reference_id_deleted_at_owner_id";
-- Drop index "control_reference_id_deleted_at_owner_id" from table: "controls"
DROP INDEX "control_reference_id_deleted_at_owner_id";
-- Drop index "control_standard_id_deleted_at_owner_id" from table: "controls"
DROP INDEX "control_standard_id_deleted_at_owner_id";
-- Create index "control_auditor_reference_id" to table: "controls"
CREATE INDEX "control_auditor_reference_id" ON "controls" ("auditor_reference_id") WHERE (deleted_at IS NULL);
-- Create index "control_ref_code" to table: "controls"
CREATE INDEX "control_ref_code" ON "controls" ("ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND ((status)::text <> 'ARCHIVED'::text));
-- Create index "control_reference_id" to table: "controls"
CREATE INDEX "control_reference_id" ON "controls" ("reference_id") WHERE (deleted_at IS NULL);
-- Drop index "customtypeenum_object_type" from table: "custom_type_enums"
DROP INDEX "customtypeenum_object_type";
-- Create index "customtypeenum_object_type" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_object_type" ON "custom_type_enums" ("object_type") WHERE (deleted_at IS NULL);
-- Create index "customtypeenum_name_field" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_name_field" ON "custom_type_enums" ("name", "field") WHERE (deleted_at IS NULL);
-- Modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';
