-- +goose Up
-- create index "subcontrol_control_id_ref_code_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_control_id_ref_code_owner_id" ON "subcontrols" ("control_id", "ref_code", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "subcontrol_control_id_ref_code_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_control_id_ref_code_owner_id";
