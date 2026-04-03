-- +goose Up
-- drop index "controls_external_uuid_key" from table: "controls"
DROP INDEX "controls_external_uuid_key";
-- create index "control_external_uuid_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_external_uuid_owner_id" ON "controls" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "evidences_external_uuid_key" from table: "evidences"
DROP INDEX "evidences_external_uuid_key";
-- create index "evidence_external_uuid_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "internal_policies_external_uuid_key" from table: "internal_policies"
DROP INDEX "internal_policies_external_uuid_key";
-- create index "internalpolicy_external_uuid_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_external_uuid_owner_id" ON "internal_policies" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "platforms_external_uuid_key" from table: "platforms"
DROP INDEX "platforms_external_uuid_key";
-- create index "platform_external_uuid_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_external_uuid_owner_id" ON "platforms" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "programs_external_uuid_key" from table: "programs"
DROP INDEX "programs_external_uuid_key";
-- create index "program_external_uuid_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_external_uuid_owner_id" ON "programs" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "risks_external_uuid_key" from table: "risks"
DROP INDEX "risks_external_uuid_key";
-- create index "risk_external_uuid_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_external_uuid_owner_id" ON "risks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "subcontrols_external_uuid_key" from table: "subcontrols"
DROP INDEX "subcontrols_external_uuid_key";
-- create index "subcontrol_external_uuid_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_external_uuid_owner_id" ON "subcontrols" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- drop index "tasks_external_uuid_key" from table: "tasks"
DROP INDEX "tasks_external_uuid_key";
-- create index "task_external_uuid_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_external_uuid_owner_id" ON "tasks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "task_external_uuid_owner_id" to table: "tasks"
DROP INDEX "task_external_uuid_owner_id";
-- reverse: drop index "tasks_external_uuid_key" from table: "tasks"
CREATE UNIQUE INDEX "tasks_external_uuid_key" ON "tasks" ("external_uuid");
-- reverse: create index "subcontrol_external_uuid_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_external_uuid_owner_id";
-- reverse: drop index "subcontrols_external_uuid_key" from table: "subcontrols"
CREATE UNIQUE INDEX "subcontrols_external_uuid_key" ON "subcontrols" ("external_uuid");
-- reverse: create index "risk_external_uuid_owner_id" to table: "risks"
DROP INDEX "risk_external_uuid_owner_id";
-- reverse: drop index "risks_external_uuid_key" from table: "risks"
CREATE UNIQUE INDEX "risks_external_uuid_key" ON "risks" ("external_uuid");
-- reverse: create index "program_external_uuid_owner_id" to table: "programs"
DROP INDEX "program_external_uuid_owner_id";
-- reverse: drop index "programs_external_uuid_key" from table: "programs"
CREATE UNIQUE INDEX "programs_external_uuid_key" ON "programs" ("external_uuid");
-- reverse: create index "platform_external_uuid_owner_id" to table: "platforms"
DROP INDEX "platform_external_uuid_owner_id";
-- reverse: drop index "platforms_external_uuid_key" from table: "platforms"
CREATE UNIQUE INDEX "platforms_external_uuid_key" ON "platforms" ("external_uuid");
-- reverse: create index "internalpolicy_external_uuid_owner_id" to table: "internal_policies"
DROP INDEX "internalpolicy_external_uuid_owner_id";
-- reverse: drop index "internal_policies_external_uuid_key" from table: "internal_policies"
CREATE UNIQUE INDEX "internal_policies_external_uuid_key" ON "internal_policies" ("external_uuid");
-- reverse: create index "evidence_external_uuid_owner_id" to table: "evidences"
DROP INDEX "evidence_external_uuid_owner_id";
-- reverse: drop index "evidences_external_uuid_key" from table: "evidences"
CREATE UNIQUE INDEX "evidences_external_uuid_key" ON "evidences" ("external_uuid");
-- reverse: create index "control_external_uuid_owner_id" to table: "controls"
DROP INDEX "control_external_uuid_owner_id";
-- reverse: drop index "controls_external_uuid_key" from table: "controls"
CREATE UNIQUE INDEX "controls_external_uuid_key" ON "controls" ("external_uuid");
