-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "example_evidence" text NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "example_evidence" text NULL;
-- Create "evidence_history" table
CREATE TABLE "evidence_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, PRIMARY KEY ("id"));
-- Create index "evidencehistory_history_time" to table: "evidence_history"
CREATE INDEX "evidencehistory_history_time" ON "evidence_history" ("history_time");
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "review_due" timestamptz NULL;
-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "example_evidence" text NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "review_due" timestamptz NULL;
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" ADD COLUMN "example_evidence" text NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "review_due" timestamptz NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "review_due" timestamptz NULL;
-- Create "evidences" table
CREATE TABLE "evidences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "collection_procedure" text NULL, "creation_date" timestamptz NOT NULL, "renewal_date" timestamptz NULL, "source" character varying NULL, "is_automated" boolean NULL DEFAULT false, "url" character varying NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "evidence_display_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_display_id_owner_id" ON "evidences" ("display_id", "owner_id");
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "example_evidence" text NULL, ADD COLUMN "evidence_control_objectives" character varying NULL, ADD CONSTRAINT "control_objectives_evidences_control_objectives" FOREIGN KEY ("evidence_control_objectives") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "example_evidence" text NULL, ADD COLUMN "evidence_controls" character varying NULL, ADD COLUMN "evidence_subcontrols" character varying NULL, ADD CONSTRAINT "controls_evidences_controls" FOREIGN KEY ("evidence_controls") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_evidences_subcontrols" FOREIGN KEY ("evidence_subcontrols") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "evidence_files" character varying NULL, ADD CONSTRAINT "files_evidences_files" FOREIGN KEY ("evidence_files") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "program_evidence" table
CREATE TABLE "program_evidence" ("program_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("program_id", "evidence_id"), CONSTRAINT "program_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_evidence_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "task_evidence" table
CREATE TABLE "task_evidence" ("task_id" character varying NOT NULL, "evidence_id" character varying NOT NULL, PRIMARY KEY ("task_id", "evidence_id"), CONSTRAINT "task_evidence_evidence_id" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "task_evidence_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
