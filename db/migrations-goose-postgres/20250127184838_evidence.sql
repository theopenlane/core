-- +goose Up
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "example_evidence" text NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "example_evidence" text NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "example_evidence" text NULL;
-- create "evidence_history" table
CREATE TABLE "evidence_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, PRIMARY KEY ("id"));
-- create index "evidencehistory_history_time" to table: "evidence_history"
CREATE INDEX "evidencehistory_history_time" ON "evidence_history" ("history_time");
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "review_due" timestamptz NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "review_due" timestamptz NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "review_due" timestamptz NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "review_due" timestamptz NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "example_evidence" text NULL;
-- create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- create "evidence_control_objectives" table
CREATE TABLE "evidence_control_objectives" ("evidence_id" character varying NOT NULL, "control_objective_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_objective_id"), CONSTRAINT "evidence_control_objectives_control_objective_id" FOREIGN KEY ("control_objective_id") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "evidence_control_objectives_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "example_evidence" text NULL;
-- create "evidence_controls" table
CREATE TABLE "evidence_controls" ("evidence_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "control_id"), CONSTRAINT "evidence_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "evidence_controls_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "evidence_files" table
CREATE TABLE "evidence_files" ("evidence_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "file_id"), CONSTRAINT "evidence_files_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "evidence_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "example_evidence" text NULL;
-- create "evidence_subcontrols" table
CREATE TABLE "evidence_subcontrols" ("evidence_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("evidence_id", "subcontrol_id"), CONSTRAINT "evidence_subcontrols_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "evidence_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"), CONSTRAINT "program_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_evidence_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"), CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "task_evidence" table
DROP TABLE "task_evidence";
-- reverse: create "program_evidence" table
DROP TABLE "program_evidence";
-- reverse: create "evidence_subcontrols" table
DROP TABLE "evidence_subcontrols";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "example_evidence";
-- reverse: create "evidence_files" table
DROP TABLE "evidence_files";
-- reverse: create "evidence_controls" table
DROP TABLE "evidence_controls";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "example_evidence";
-- reverse: create "evidence_control_objectives" table
DROP TABLE "evidence_control_objectives";
-- reverse: create index "evidence_display_id_owner_id" to table: "evidences"
DROP INDEX "evidence_display_id_owner_id";
-- reverse: create "evidences" table
DROP TABLE "evidences";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "example_evidence";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "review_due";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "review_due";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "review_due";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "review_due";
-- reverse: create index "evidencehistory_history_time" to table: "evidence_history"
DROP INDEX "evidencehistory_history_time";
-- reverse: create "evidence_history" table
DROP TABLE "evidence_history";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "example_evidence";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "example_evidence";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "example_evidence";
