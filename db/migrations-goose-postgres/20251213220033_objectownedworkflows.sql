-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "testing_procedures" jsonb NULL, ADD COLUMN "evidence_requests" jsonb NULL;
-- modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "control_id" character varying NULL, ADD COLUMN "internal_policy_id" character varying NULL, ADD COLUMN "evidence_id" character varying NULL;
-- modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "evidence_id" character varying NULL;
-- modify "evidences" table
ALTER TABLE "evidences" ALTER COLUMN "status" DROP DEFAULT, ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "proposed_changes" jsonb NULL, ADD COLUMN "proposed_by_user_id" character varying NULL, ADD COLUMN "proposed_at" timestamptz NULL;
-- modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "control_id" character varying NULL, ADD COLUMN "internal_policy_id" character varying NULL, ADD COLUMN "evidence_id" character varying NULL, ADD CONSTRAINT "workflow_instances_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "evidence_id" character varying NULL, ADD CONSTRAINT "workflow_object_refs_evidences_evidence" FOREIGN KEY ("evidence_id") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "workflowobjectref_workflow_instance_id_evidence_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_evidence_id" ON "workflow_object_refs" ("workflow_instance_id", "evidence_id");

-- +goose Down
-- reverse: create index "workflowobjectref_workflow_instance_id_evidence_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_evidence_id";
-- reverse: modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" DROP CONSTRAINT "workflow_object_refs_evidences_evidence", DROP COLUMN "evidence_id";
-- reverse: modify "workflow_instances" table
ALTER TABLE "workflow_instances" DROP CONSTRAINT "workflow_instances_internal_policies_internal_policy", DROP CONSTRAINT "workflow_instances_evidences_evidence", DROP CONSTRAINT "workflow_instances_controls_control", DROP COLUMN "evidence_id", DROP COLUMN "internal_policy_id", DROP COLUMN "control_id";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "proposed_at", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_changes";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "proposed_at", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_changes", ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- reverse: modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" DROP COLUMN "evidence_id";
-- reverse: modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" DROP COLUMN "evidence_id", DROP COLUMN "internal_policy_id", DROP COLUMN "control_id";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "evidence_requests", DROP COLUMN "testing_procedures";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "evidence_requests", DROP COLUMN "testing_procedures";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "proposed_at", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_changes";
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "proposed_at", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_changes", ALTER COLUMN "status" SET DEFAULT 'SUBMITTED';
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "proposed_at", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_changes", DROP COLUMN "evidence_requests", DROP COLUMN "testing_procedures";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "proposed_at", DROP COLUMN "proposed_by_user_id", DROP COLUMN "proposed_changes", DROP COLUMN "evidence_requests", DROP COLUMN "testing_procedures";
