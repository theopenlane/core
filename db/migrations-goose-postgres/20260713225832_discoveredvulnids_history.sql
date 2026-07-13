-- +goose Up
-- modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "scan_history" table
ALTER TABLE "scan_history" RENAME COLUMN "vulnerability_ids" TO "discovered_vulnerability_ids";
-- modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "outcome_metadata" jsonb NULL;
-- modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "finding_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "task_id" character varying NULL, ADD COLUMN "vulnerability_id" character varying NULL;
-- modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "vulnerability_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL;

-- +goose Down
-- reverse: modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" DROP COLUMN "remediation_id", DROP COLUMN "assessment_response_id", DROP COLUMN "assessment_id", DROP COLUMN "risk_id", DROP COLUMN "vulnerability_id";
-- reverse: modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" DROP COLUMN "vulnerability_id", DROP COLUMN "task_id", DROP COLUMN "risk_id", DROP COLUMN "remediation_id", DROP COLUMN "integration_id", DROP COLUMN "finding_id", DROP COLUMN "assessment_response_id", DROP COLUMN "assessment_id";
-- reverse: modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" DROP COLUMN "outcome_metadata";
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "scan_history" table
ALTER TABLE "scan_history" RENAME COLUMN "discovered_vulnerability_ids" TO "vulnerability_ids";
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "remediation_history" table
ALTER TABLE "remediation_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "workflow_eligible_marker";
