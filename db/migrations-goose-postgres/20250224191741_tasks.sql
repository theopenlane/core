-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "note_history" table
ALTER TABLE "note_history" ALTER COLUMN "text" TYPE text;
-- modify "task_history" table
ALTER TABLE "task_history" ALTER COLUMN "details" TYPE text, ALTER COLUMN "owner_id" DROP NOT NULL, DROP COLUMN "priority", ALTER COLUMN "assigner_id" DROP NOT NULL, ADD COLUMN "category" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "owner_id" DROP NOT NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_organizations_control_objectives", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "control_objectives_organizations_control_objectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_organizations_controls", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_document_data", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "document_data_organizations_document_data" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "evidences" table
ALTER TABLE "evidences" DROP CONSTRAINT "evidences_organizations_evidence", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_organizations_narratives", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_organizations_tasks", DROP CONSTRAINT "tasks_users_assigner_tasks", ALTER COLUMN "details" TYPE text, DROP COLUMN "priority", ALTER COLUMN "owner_id" DROP NOT NULL, ALTER COLUMN "assigner_id" DROP NOT NULL, ADD COLUMN "category" character varying NULL, ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notes" table
ALTER TABLE "notes" ALTER COLUMN "text" TYPE text, ADD COLUMN "program_notes" character varying NULL, ADD COLUMN "task_comments" character varying NULL, ADD CONSTRAINT "notes_programs_notes" FOREIGN KEY ("program_notes") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_tasks_comments" FOREIGN KEY ("task_comments") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_organizations_risks", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP COLUMN "note_subcontrols", ALTER COLUMN "owner_id" DROP NOT NULL, ADD COLUMN "user_subcontrols" character varying NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_subcontrols" FOREIGN KEY ("user_subcontrols") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_users_subcontrols", DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP COLUMN "user_subcontrols", ALTER COLUMN "owner_id" SET NOT NULL, ADD COLUMN "note_subcontrols" character varying NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_organizations_risks", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_tasks_comments", DROP CONSTRAINT "notes_programs_notes", DROP COLUMN "task_comments", DROP COLUMN "program_notes", ALTER COLUMN "text" TYPE character varying;
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_assigner_tasks", DROP CONSTRAINT "tasks_organizations_tasks", DROP COLUMN "category", ALTER COLUMN "assigner_id" SET NOT NULL, ALTER COLUMN "owner_id" SET NOT NULL, ADD COLUMN "priority" character varying NOT NULL DEFAULT 'MEDIUM', ALTER COLUMN "details" TYPE jsonb, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_organizations_narratives", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP CONSTRAINT "evidences_organizations_evidence", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "evidences_organizations_evidence" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_document_data", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "document_data_organizations_document_data" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_organizations_controls", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_organizations_control_objectives", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "control_objectives_organizations_control_objectives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- reverse: modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "category", ALTER COLUMN "assigner_id" SET NOT NULL, ADD COLUMN "priority" character varying NOT NULL DEFAULT 'MEDIUM', ALTER COLUMN "owner_id" SET NOT NULL, ALTER COLUMN "details" TYPE jsonb;
-- reverse: modify "note_history" table
ALTER TABLE "note_history" ALTER COLUMN "text" TYPE character varying;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "owner_id" SET NOT NULL;
