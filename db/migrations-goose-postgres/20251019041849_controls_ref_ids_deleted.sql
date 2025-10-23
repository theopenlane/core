-- +goose Up
-- drop index "controls_auditor_reference_id_key" from table: "controls"
DROP INDEX "controls_auditor_reference_id_key";
-- drop index "controls_reference_id_key" from table: "controls"
DROP INDEX "controls_reference_id_key";
-- create index "control_auditor_reference_id_deleted_at_owner_id" to table: "controls"
CREATE INDEX "control_auditor_reference_id_deleted_at_owner_id" ON "controls" ("auditor_reference_id", "deleted_at", "owner_id");
-- create index "control_reference_id_deleted_at_owner_id" to table: "controls"
CREATE INDEX "control_reference_id_deleted_at_owner_id" ON "controls" ("reference_id", "deleted_at", "owner_id");
-- create index "subcontrol_auditor_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_auditor_reference_id_deleted_at_owner_id" ON "subcontrols" ("auditor_reference_id", "deleted_at", "owner_id");
-- create index "subcontrol_reference_id_deleted_at_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_reference_id_deleted_at_owner_id" ON "subcontrols" ("reference_id", "deleted_at", "owner_id");

-- +goose Down
-- reverse: create index "subcontrol_reference_id_deleted_at_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_reference_id_deleted_at_owner_id";
-- reverse: create index "subcontrol_auditor_reference_id_deleted_at_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_auditor_reference_id_deleted_at_owner_id";
-- reverse: create index "control_reference_id_deleted_at_owner_id" to table: "controls"
DROP INDEX "control_reference_id_deleted_at_owner_id";
-- reverse: create index "control_auditor_reference_id_deleted_at_owner_id" to table: "controls"
DROP INDEX "control_auditor_reference_id_deleted_at_owner_id";
-- reverse: drop index "controls_reference_id_key" from table: "controls"
CREATE UNIQUE INDEX "controls_reference_id_key" ON "controls" ("reference_id");
-- reverse: drop index "controls_auditor_reference_id_key" from table: "controls"
CREATE UNIQUE INDEX "controls_auditor_reference_id_key" ON "controls" ("auditor_reference_id");
