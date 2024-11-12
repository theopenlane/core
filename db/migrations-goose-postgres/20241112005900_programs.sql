-- +goose Up
-- create "program_history" table
CREATE TABLE "program_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "programhistory_history_time" to table: "program_history"
CREATE INDEX "programhistory_history_time" ON "program_history" ("history_time");
-- create "program_membership_history" table
CREATE TABLE "program_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "programmembershiphistory_history_time" to table: "program_membership_history"
CREATE INDEX "programmembershiphistory_history_time" ON "program_membership_history" ("history_time");
-- create "programs" table
CREATE TABLE "programs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "start_date" timestamptz NULL, "end_date" timestamptz NULL, "auditor_ready" boolean NOT NULL DEFAULT false, "auditor_write_comments" boolean NOT NULL DEFAULT false, "auditor_read_comments" boolean NOT NULL DEFAULT false, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "programs_organizations_programs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "programs_mapping_id_key" to table: "programs"
CREATE UNIQUE INDEX "programs_mapping_id_key" ON "programs" ("mapping_id");
-- create "program_actionplans" table
CREATE TABLE "program_actionplans" ("program_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("program_id", "action_plan_id"), CONSTRAINT "program_actionplans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_actionplans_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_controlobjectives" table
CREATE TABLE "program_controlobjectives" ("program_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_objective_id"), CONSTRAINT "program_controlobjectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_controlobjectives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_controls" table
CREATE TABLE "program_controls" ("program_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("program_id", "control_id"), CONSTRAINT "program_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_controls_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_files" table
CREATE TABLE "program_files" ("program_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("program_id", "file_id"), CONSTRAINT "program_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_files_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_memberships" table
CREATE TABLE "program_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "program_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "program_memberships_programs_program" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "program_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "program_memberships_mapping_id_key" to table: "program_memberships"
CREATE UNIQUE INDEX "program_memberships_mapping_id_key" ON "program_memberships" ("mapping_id");
-- create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id") WHERE (deleted_at IS NULL);
-- create "program_narratives" table
CREATE TABLE "program_narratives" ("program_id" character varying NOT NULL, "narrative_id" character varying NOT NULL, PRIMARY KEY ("program_id", "narrative_id"), CONSTRAINT "program_narratives_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_narratives_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_notes" table
CREATE TABLE "program_notes" ("program_id" character varying NOT NULL, "note_id" character varying NOT NULL, PRIMARY KEY ("program_id", "note_id"), CONSTRAINT "program_notes_note_id" FOREIGN KEY ("note_id") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_notes_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_policies" table
CREATE TABLE "program_policies" ("program_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("program_id", "internal_policy_id"), CONSTRAINT "program_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_policies_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_procedures" table
CREATE TABLE "program_procedures" ("program_id" character varying NOT NULL, "procedure_id" character varying NOT NULL, PRIMARY KEY ("program_id", "procedure_id"), CONSTRAINT "program_procedures_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_procedures_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_risks" table
CREATE TABLE "program_risks" ("program_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("program_id", "risk_id"), CONSTRAINT "program_risks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_subcontrols" table
CREATE TABLE "program_subcontrols" ("program_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("program_id", "subcontrol_id"), CONSTRAINT "program_subcontrols_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_tasks" table
CREATE TABLE "program_tasks" ("program_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("program_id", "task_id"), CONSTRAINT "program_tasks_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "standard_programs" table
CREATE TABLE "standard_programs" ("standard_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("standard_id", "program_id"), CONSTRAINT "standard_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "standard_programs_standard_id" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "standard_programs" table
DROP TABLE "standard_programs";
-- reverse: create "program_tasks" table
DROP TABLE "program_tasks";
-- reverse: create "program_subcontrols" table
DROP TABLE "program_subcontrols";
-- reverse: create "program_risks" table
DROP TABLE "program_risks";
-- reverse: create "program_procedures" table
DROP TABLE "program_procedures";
-- reverse: create "program_policies" table
DROP TABLE "program_policies";
-- reverse: create "program_notes" table
DROP TABLE "program_notes";
-- reverse: create "program_narratives" table
DROP TABLE "program_narratives";
-- reverse: create index "programmembership_user_id_program_id" to table: "program_memberships"
DROP INDEX "programmembership_user_id_program_id";
-- reverse: create index "program_memberships_mapping_id_key" to table: "program_memberships"
DROP INDEX "program_memberships_mapping_id_key";
-- reverse: create "program_memberships" table
DROP TABLE "program_memberships";
-- reverse: create "program_files" table
DROP TABLE "program_files";
-- reverse: create "program_controls" table
DROP TABLE "program_controls";
-- reverse: create "program_controlobjectives" table
DROP TABLE "program_controlobjectives";
-- reverse: create "program_actionplans" table
DROP TABLE "program_actionplans";
-- reverse: create index "programs_mapping_id_key" to table: "programs"
DROP INDEX "programs_mapping_id_key";
-- reverse: create "programs" table
DROP TABLE "programs";
-- reverse: create index "programmembershiphistory_history_time" to table: "program_membership_history"
DROP INDEX "programmembershiphistory_history_time";
-- reverse: create "program_membership_history" table
DROP TABLE "program_membership_history";
-- reverse: create index "programhistory_history_time" to table: "program_history"
DROP INDEX "programhistory_history_time";
-- reverse: create "program_history" table
DROP TABLE "program_history";
