-- Modify "assessment_history" table
ALTER TABLE "assessment_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "assessment_response_history" table
ALTER TABLE "assessment_response_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "scan_history" table
ALTER TABLE "scan_history" DROP COLUMN "vulnerability_ids", ADD COLUMN "discovered_vulnerability_ids" jsonb NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "outcome_metadata" jsonb NULL;
-- Modify "workflow_instance_history" table
ALTER TABLE "workflow_instance_history" ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "finding_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "task_id" character varying NULL, ADD COLUMN "vulnerability_id" character varying NULL;
-- Modify "workflow_object_ref_history" table
ALTER TABLE "workflow_object_ref_history" ADD COLUMN "vulnerability_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL;
