-- Modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "scans" table
ALTER TABLE "scans" DROP COLUMN "vulnerability_ids", ADD COLUMN "discovered_vulnerability_ids" jsonb NULL;
-- Modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "outcome_metadata" jsonb NULL;
-- Modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" ADD COLUMN "proposed_changes" jsonb NULL;
-- Modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "findings" table
ALTER TABLE "findings" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- Modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "finding_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "task_id" character varying NULL, ADD COLUMN "vulnerability_id" character varying NULL, ADD CONSTRAINT "workflow_instances_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "vulnerability_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL, ADD CONSTRAINT "workflow_object_refs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "workflowobjectref_workflow_instance_id_assessment_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_id");
-- Create index "workflowobjectref_workflow_instance_id_assessment_response_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_response_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_response_id");
-- Create index "workflowobjectref_workflow_instance_id_remediation_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_remediation_id" ON "workflow_object_refs" ("workflow_instance_id", "remediation_id");
-- Create index "workflowobjectref_workflow_instance_id_risk_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_risk_id" ON "workflow_object_refs" ("workflow_instance_id", "risk_id");
-- Create index "workflowobjectref_workflow_instance_id_vulnerability_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_vulnerability_id" ON "workflow_object_refs" ("workflow_instance_id", "vulnerability_id");
