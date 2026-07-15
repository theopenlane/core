-- +goose Up
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "scans" table
ALTER TABLE "scans" RENAME COLUMN "vulnerability_ids" TO "discovered_vulnerability_ids";
-- modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" ADD COLUMN "outcome_metadata" jsonb NULL;
-- modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" ADD COLUMN "proposed_changes" jsonb NULL;
-- modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "findings" table
ALTER TABLE "findings" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "workflow_eligible_marker" boolean NULL DEFAULT true;
-- modify "workflow_instances" table
ALTER TABLE "workflow_instances" ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "finding_id" character varying NULL, ADD COLUMN "integration_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "task_id" character varying NULL, ADD COLUMN "vulnerability_id" character varying NULL, ADD CONSTRAINT "workflow_instances_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_instances_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" ADD COLUMN "vulnerability_id" character varying NULL, ADD COLUMN "risk_id" character varying NULL, ADD COLUMN "assessment_id" character varying NULL, ADD COLUMN "assessment_response_id" character varying NULL, ADD COLUMN "remediation_id" character varying NULL, ADD CONSTRAINT "workflow_object_refs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_remediations_remediation" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_risks_risk" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "workflow_object_refs_vulnerabilities_vulnerability" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "workflowobjectref_workflow_instance_id_assessment_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_id");
-- create index "workflowobjectref_workflow_instance_id_assessment_response_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_assessment_response_id" ON "workflow_object_refs" ("workflow_instance_id", "assessment_response_id");
-- create index "workflowobjectref_workflow_instance_id_remediation_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_remediation_id" ON "workflow_object_refs" ("workflow_instance_id", "remediation_id");
-- create index "workflowobjectref_workflow_instance_id_risk_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_risk_id" ON "workflow_object_refs" ("workflow_instance_id", "risk_id");
-- create index "workflowobjectref_workflow_instance_id_vulnerability_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_vulnerability_id" ON "workflow_object_refs" ("workflow_instance_id", "vulnerability_id");

-- +goose Down
-- reverse: create index "workflowobjectref_workflow_instance_id_vulnerability_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_vulnerability_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_risk_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_risk_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_remediation_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_remediation_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_assessment_response_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_assessment_response_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_assessment_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_assessment_id";
-- reverse: modify "workflow_object_refs" table
ALTER TABLE "workflow_object_refs" DROP CONSTRAINT "workflow_object_refs_vulnerabilities_vulnerability", DROP CONSTRAINT "workflow_object_refs_risks_risk", DROP CONSTRAINT "workflow_object_refs_remediations_remediation", DROP CONSTRAINT "workflow_object_refs_assessments_assessment", DROP CONSTRAINT "workflow_object_refs_assessment_responses_assessment_response", DROP COLUMN "remediation_id", DROP COLUMN "assessment_response_id", DROP COLUMN "assessment_id", DROP COLUMN "risk_id", DROP COLUMN "vulnerability_id";
-- reverse: modify "workflow_instances" table
ALTER TABLE "workflow_instances" DROP CONSTRAINT "workflow_instances_vulnerabilities_vulnerability", DROP CONSTRAINT "workflow_instances_tasks_task", DROP CONSTRAINT "workflow_instances_risks_risk", DROP CONSTRAINT "workflow_instances_remediations_remediation", DROP CONSTRAINT "workflow_instances_integrations_integration", DROP CONSTRAINT "workflow_instances_findings_finding", DROP CONSTRAINT "workflow_instances_assessments_assessment", DROP CONSTRAINT "workflow_instances_assessment_responses_assessment_response", DROP COLUMN "vulnerability_id", DROP COLUMN "task_id", DROP COLUMN "risk_id", DROP COLUMN "remediation_id", DROP COLUMN "integration_id", DROP COLUMN "finding_id", DROP COLUMN "assessment_response_id", DROP COLUMN "assessment_id";
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "workflow_eligible_marker";
-- reverse: modify "workflow_proposals" table
ALTER TABLE "workflow_proposals" DROP COLUMN "proposed_changes";
-- reverse: modify "workflow_assignments" table
ALTER TABLE "workflow_assignments" DROP COLUMN "outcome_metadata";
-- reverse: modify "scans" table
ALTER TABLE "scans" RENAME COLUMN "discovered_vulnerability_ids" TO "vulnerability_ids";
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP COLUMN "workflow_eligible_marker";
