-- +goose Up
-- create index "control_ref_code_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_ref_code_owner_id" ON "controls" ("ref_code", "owner_id") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND (standard_id IS NULL));
-- create index "control_standard_id_ref_code_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_standard_id_ref_code_owner_id" ON "controls" ("standard_id", "ref_code", "owner_id") WHERE ((deleted_at IS NULL) AND (owner_id IS NOT NULL) AND (standard_id IS NOT NULL));

-- +goose Down
-- reverse: create index "control_standard_id_ref_code_owner_id" to table: "controls"
DROP INDEX "control_standard_id_ref_code_owner_id";
-- reverse: create index "control_ref_code_owner_id" to table: "controls"
DROP INDEX "control_ref_code_owner_id";
