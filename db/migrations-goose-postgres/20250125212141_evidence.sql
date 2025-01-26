-- +goose Up
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "example_evidence" text NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "example_evidence" text NULL;
-- create "evidence_history" table
CREATE TABLE "evidence_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, PRIMARY KEY ("id"));
-- create index "evidencehistory_history_time" to table: "evidence_history"
CREATE INDEX "evidencehistory_history_time" ON "evidence_history" ("history_time");
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "review_due" timestamptz NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "example_evidence" text NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "review_due" timestamptz NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "example_evidence" text NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "review_due" timestamptz NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "review_due" timestamptz NULL;
-- create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "example_evidence" text NULL, ADD COLUMN "evidence_control_objectives" character varying NULL, ADD CONSTRAINT "control_objectives_evidences_control_objectives" FOREIGN KEY ("evidence_control_objectives") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "example_evidence" text NULL, ADD COLUMN "evidence_controls" character varying NULL, ADD COLUMN "evidence_subcontrols" character varying NULL, ADD CONSTRAINT "controls_evidences_controls" FOREIGN KEY ("evidence_controls") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_evidences_subcontrols" FOREIGN KEY ("evidence_subcontrols") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "evidence_files" character varying NULL, ADD CONSTRAINT "files_evidences_files" FOREIGN KEY ("evidence_files") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"), CONSTRAINT "program_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_evidence_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"), CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "task_evidence" table
DROP TABLE "task_evidence";
-- reverse: create "program_evidence" table
DROP TABLE "program_evidence";
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_evidences_files", DROP COLUMN "evidence_files";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_evidences_subcontrols", DROP CONSTRAINT "controls_evidences_controls", DROP COLUMN "evidence_subcontrols", DROP COLUMN "evidence_controls", DROP COLUMN "example_evidence";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_evidences_control_objectives", DROP COLUMN "evidence_control_objectives", DROP COLUMN "example_evidence";
-- reverse: create index "evidence_display_id_owner_id" to table: "evidences"
DROP INDEX "evidence_display_id_owner_id";
-- reverse: create "evidences" table
DROP TABLE "evidences";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "review_due";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "review_due";
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "example_evidence";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "review_due";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "example_evidence";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "review_due";
-- reverse: create index "evidencehistory_history_time" to table: "evidence_history"
DROP INDEX "evidencehistory_history_time";
-- reverse: create "evidence_history" table
DROP TABLE "evidence_history";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "example_evidence";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "example_evidence";
