-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
-- drop index "control_standard_id_ref_code" from table: "controls"
DROP INDEX "control_standard_id_ref_code";
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
-- create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE ((deleted_at IS NULL) AND (owner_id IS NULL));
-- create index "controls_auditor_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_auditor_reference_id_key" ON "controls" ("auditor_reference_id");
-- create index "controls_reference_id_key" to table: "controls"
CREATE UNIQUE INDEX "controls_reference_id_key" ON "controls" ("reference_id");
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "reference_id" character varying NULL, ADD COLUMN "auditor_reference_id" character varying NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "auditor_reference_id", DROP COLUMN "reference_id";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "auditor_reference_id", DROP COLUMN "reference_id";
-- reverse: create index "controls_reference_id_key" to table: "controls"
DROP INDEX "controls_reference_id_key";
-- reverse: create index "controls_auditor_reference_id_key" to table: "controls"
DROP INDEX "controls_auditor_reference_id_key";
-- reverse: create index "control_standard_id_ref_code" to table: "controls"
DROP INDEX "control_standard_id_ref_code";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "auditor_reference_id", DROP COLUMN "reference_id";
-- reverse: drop index "control_standard_id_ref_code" from table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE (deleted_at IS NULL);
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "auditor_reference_id", DROP COLUMN "reference_id";
