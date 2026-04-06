-- Drop index "controls_external_uuid_key" from table: "controls"
DROP INDEX "controls_external_uuid_key";
-- Create index "control_external_uuid_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_external_uuid_owner_id" ON "controls" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "evidences_external_uuid_key" from table: "evidences"
DROP INDEX "evidences_external_uuid_key";
-- Create index "evidence_external_uuid_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "internal_policies_external_uuid_key" from table: "internal_policies"
DROP INDEX "internal_policies_external_uuid_key";
-- Create index "internalpolicy_external_uuid_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_external_uuid_owner_id" ON "internal_policies" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "platforms_external_uuid_key" from table: "platforms"
DROP INDEX "platforms_external_uuid_key";
-- Create index "platform_external_uuid_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_external_uuid_owner_id" ON "platforms" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "programs_external_uuid_key" from table: "programs"
DROP INDEX "programs_external_uuid_key";
-- Create index "program_external_uuid_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_external_uuid_owner_id" ON "programs" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "risks_external_uuid_key" from table: "risks"
DROP INDEX "risks_external_uuid_key";
-- Create index "risk_external_uuid_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_external_uuid_owner_id" ON "risks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "subcontrols_external_uuid_key" from table: "subcontrols"
DROP INDEX "subcontrols_external_uuid_key";
-- Create index "subcontrol_external_uuid_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_external_uuid_owner_id" ON "subcontrols" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
-- Drop index "tasks_external_uuid_key" from table: "tasks"
DROP INDEX "tasks_external_uuid_key";
-- Create index "task_external_uuid_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_external_uuid_owner_id" ON "tasks" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
