-- +goose Up
-- create index "control_standard_id_ref_code" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code" ON "controls" ("standard_id", "ref_code") WHERE (deleted_at IS NULL);
-- create index "subcontrol_control_id_ref_code" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_control_id_ref_code" ON "subcontrols" ("control_id", "ref_code") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "subcontrol_control_id_ref_code" to table: "subcontrols"
DROP INDEX "subcontrol_control_id_ref_code";
-- reverse: create index "control_standard_id_ref_code" to table: "controls"
DROP INDEX "control_standard_id_ref_code";
