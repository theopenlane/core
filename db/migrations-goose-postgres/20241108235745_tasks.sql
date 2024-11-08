-- +goose Up
-- create "task_history" table
CREATE TABLE "task_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" jsonb NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "due" timestamptz NULL, "completed" timestamptz NULL, "assignee" character varying NULL, "assigner" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "taskhistory_history_time" to table: "task_history"
CREATE INDEX "taskhistory_history_time" ON "task_history" ("history_time");
-- create "tasks" table
CREATE TABLE "tasks" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "description" character varying NULL, "details" jsonb NULL, "status" character varying NOT NULL DEFAULT 'OPEN', "due" timestamptz NULL, "completed" timestamptz NULL, "assignee" character varying NULL, "assigner" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "tasks_users_tasks" FOREIGN KEY ("assigner") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "tasks_mapping_id_key" to table: "tasks"
CREATE UNIQUE INDEX "tasks_mapping_id_key" ON "tasks" ("mapping_id");
-- create "control_objective_tasks" table
CREATE TABLE "control_objective_tasks" ("control_objective_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_objective_id", "task_id"), CONSTRAINT "control_objective_tasks_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_objective_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_tasks" table
CREATE TABLE "control_tasks" ("control_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("control_id", "task_id"), CONSTRAINT "control_tasks_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "group_tasks" table
CREATE TABLE "group_tasks" ("group_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("group_id", "task_id"), CONSTRAINT "group_tasks_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "group_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "internal_policy_tasks" table
CREATE TABLE "internal_policy_tasks" ("internal_policy_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "task_id"), CONSTRAINT "internal_policy_tasks_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "organization_tasks" table
CREATE TABLE "organization_tasks" ("organization_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("organization_id", "task_id"), CONSTRAINT "organization_tasks_organization_id" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "organization_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "procedure_tasks" table
CREATE TABLE "procedure_tasks" ("procedure_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "task_id"), CONSTRAINT "procedure_tasks_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "procedure_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "subcontrol_tasks" table
CREATE TABLE "subcontrol_tasks" ("subcontrol_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "task_id"), CONSTRAINT "subcontrol_tasks_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "subcontrol_tasks" table
DROP TABLE "subcontrol_tasks";
-- reverse: create "procedure_tasks" table
DROP TABLE "procedure_tasks";
-- reverse: create "organization_tasks" table
DROP TABLE "organization_tasks";
-- reverse: create "internal_policy_tasks" table
DROP TABLE "internal_policy_tasks";
-- reverse: create "group_tasks" table
DROP TABLE "group_tasks";
-- reverse: create "control_tasks" table
DROP TABLE "control_tasks";
-- reverse: create "control_objective_tasks" table
DROP TABLE "control_objective_tasks";
-- reverse: create index "tasks_mapping_id_key" to table: "tasks"
DROP INDEX "tasks_mapping_id_key";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "taskhistory_history_time" to table: "task_history"
DROP INDEX "taskhistory_history_time";
-- reverse: create "task_history" table
DROP TABLE "task_history";
