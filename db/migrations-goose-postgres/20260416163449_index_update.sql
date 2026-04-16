-- +goose Up
-- drop index "control_auditor_reference_id_deleted_at_owner_id" from table: "controls"
DROP INDEX "control_auditor_reference_id_deleted_at_owner_id";
-- drop index "control_reference_id_deleted_at_owner_id" from table: "controls"
DROP INDEX "control_reference_id_deleted_at_owner_id";
-- drop index "control_standard_id_deleted_at_owner_id" from table: "controls"
DROP INDEX "control_standard_id_deleted_at_owner_id";
-- create index "control_auditor_reference_id" to table: "controls"
CREATE INDEX "control_auditor_reference_id" ON "controls" ("auditor_reference_id") WHERE (deleted_at IS NULL);
-- create index "control_ref_code" to table: "controls"
CREATE INDEX "control_ref_code" ON "controls" ("ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND ((status)::text <> 'ARCHIVED'::text));
-- create index "control_reference_id" to table: "controls"
CREATE INDEX "control_reference_id" ON "controls" ("reference_id") WHERE (deleted_at IS NULL);
-- drop index "customtypeenum_object_type" from table: "custom_type_enums"
DROP INDEX "customtypeenum_object_type";
-- create index "customtypeenum_object_type" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_object_type" ON "custom_type_enums" ("object_type") WHERE (deleted_at IS NULL);
-- create index "customtypeenum_name_field" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_name_field" ON "custom_type_enums" ("name", "field") WHERE (deleted_at IS NULL);
-- modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';

-- +goose Down
-- reverse: modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" DROP DEFAULT;
-- reverse: create index "customtypeenum_name_field" to table: "custom_type_enums"
DROP INDEX "customtypeenum_name_field";
-- reverse: create index "customtypeenum_object_type" to table: "custom_type_enums"
DROP INDEX "customtypeenum_object_type";
-- reverse: drop index "customtypeenum_object_type" from table: "custom_type_enums"
CREATE INDEX "customtypeenum_object_type" ON "custom_type_enums" ("object_type");
-- reverse: create index "control_reference_id" to table: "controls"
DROP INDEX "control_reference_id";
-- reverse: create index "control_ref_code" to table: "controls"
DROP INDEX "control_ref_code";
-- reverse: create index "control_auditor_reference_id" to table: "controls"
DROP INDEX "control_auditor_reference_id";
-- reverse: drop index "control_standard_id_deleted_at_owner_id" from table: "controls"
CREATE INDEX "control_standard_id_deleted_at_owner_id" ON "controls" ("standard_id", "deleted_at", "owner_id");
-- reverse: drop index "control_reference_id_deleted_at_owner_id" from table: "controls"
CREATE INDEX "control_reference_id_deleted_at_owner_id" ON "controls" ("reference_id", "deleted_at", "owner_id");
-- reverse: drop index "control_auditor_reference_id_deleted_at_owner_id" from table: "controls"
CREATE INDEX "control_auditor_reference_id_deleted_at_owner_id" ON "controls" ("auditor_reference_id", "deleted_at", "owner_id");
