-- +goose Up
-- modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ADD COLUMN "metadata" jsonb NULL;
-- modify "job_results" table
ALTER TABLE "job_results" ALTER COLUMN "exit_code" DROP NOT NULL, ADD COLUMN "compliance_job_id" character varying NULL, ADD COLUMN "metadata" jsonb NULL;
-- modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" ADD COLUMN "metadata" jsonb NULL;
-- modify "job_template_history" table
ALTER TABLE "job_template_history" ALTER COLUMN "download_url" DROP NOT NULL, ADD COLUMN "runtime_platform" character varying NULL, ADD COLUMN "script_path" character varying NULL, ADD COLUMN "metadata" jsonb NULL;
-- modify "job_runners" table
ALTER TABLE "job_runners" ADD COLUMN "metadata" jsonb NULL;
-- modify "job_templates" table
ALTER TABLE "job_templates" ALTER COLUMN "download_url" DROP NOT NULL, ADD COLUMN "runtime_platform" character varying NULL, ADD COLUMN "script_path" character varying NULL, ADD COLUMN "metadata" jsonb NULL;
-- modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" ADD COLUMN "metadata" jsonb NULL;
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "job_result_assets" character varying NULL, ADD COLUMN "job_template_assets" character varying NULL, ADD COLUMN "scheduled_job_assets" character varying NULL, ADD CONSTRAINT "assets_job_results_assets" FOREIGN KEY ("job_result_assets") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_job_templates_assets" FOREIGN KEY ("job_template_assets") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_scheduled_jobs_assets" FOREIGN KEY ("scheduled_job_assets") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "job_result_contacts" character varying NULL, ADD COLUMN "job_template_contacts" character varying NULL, ADD COLUMN "scheduled_job_contacts" character varying NULL, ADD CONSTRAINT "contacts_job_results_contacts" FOREIGN KEY ("job_result_contacts") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "contacts_job_templates_contacts" FOREIGN KEY ("job_template_contacts") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "contacts_scheduled_jobs_contacts" FOREIGN KEY ("scheduled_job_contacts") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "job_result_controls" character varying NULL, ADD COLUMN "job_template_controls" character varying NULL, ADD CONSTRAINT "controls_job_results_controls" FOREIGN KEY ("job_result_controls") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_job_templates_controls" FOREIGN KEY ("job_template_controls") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "job_result_entities" character varying NULL, ADD COLUMN "job_template_entities" character varying NULL, ADD COLUMN "scheduled_job_entities" character varying NULL, ADD CONSTRAINT "entities_job_results_entities" FOREIGN KEY ("job_result_entities") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_job_templates_entities" FOREIGN KEY ("job_template_entities") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_scheduled_jobs_entities" FOREIGN KEY ("scheduled_job_entities") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "job_result_evidence" character varying NULL, ADD COLUMN "job_template_evidence" character varying NULL, ADD COLUMN "scheduled_job_evidence" character varying NULL, ADD CONSTRAINT "evidences_job_results_evidence" FOREIGN KEY ("job_result_evidence") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_job_templates_evidence" FOREIGN KEY ("job_template_evidence") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "evidences_scheduled_jobs_evidence" FOREIGN KEY ("scheduled_job_evidence") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "findings" table
ALTER TABLE "findings" ADD COLUMN "job_result_findings" character varying NULL, ADD COLUMN "job_template_findings" character varying NULL, ADD COLUMN "scheduled_job_findings" character varying NULL, ADD CONSTRAINT "findings_job_results_findings" FOREIGN KEY ("job_result_findings") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_job_templates_findings" FOREIGN KEY ("job_template_findings") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_scheduled_jobs_findings" FOREIGN KEY ("scheduled_job_findings") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "job_result_risks" character varying NULL, ADD COLUMN "job_template_risks" character varying NULL, ADD COLUMN "scheduled_job_risks" character varying NULL, ADD CONSTRAINT "risks_job_results_risks" FOREIGN KEY ("job_result_risks") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_job_templates_risks" FOREIGN KEY ("job_template_risks") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_scheduled_jobs_risks" FOREIGN KEY ("scheduled_job_risks") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "standards" table
ALTER TABLE "standards" ADD COLUMN "job_result_standards" character varying NULL, ADD COLUMN "job_template_standards" character varying NULL, ADD COLUMN "scheduled_job_standards" character varying NULL, ADD CONSTRAINT "standards_job_results_standards" FOREIGN KEY ("job_result_standards") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "standards_job_templates_standards" FOREIGN KEY ("job_template_standards") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "standards_scheduled_jobs_standards" FOREIGN KEY ("scheduled_job_standards") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "job_result_tasks" character varying NULL, ADD COLUMN "job_template_tasks" character varying NULL, ADD COLUMN "scheduled_job_tasks" character varying NULL, ADD CONSTRAINT "tasks_job_results_tasks" FOREIGN KEY ("job_result_tasks") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_job_templates_tasks" FOREIGN KEY ("job_template_tasks") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_scheduled_jobs_tasks" FOREIGN KEY ("scheduled_job_tasks") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "job_result_vulnerabilities" character varying NULL, ADD COLUMN "job_template_vulnerabilities" character varying NULL, ADD COLUMN "scheduled_job_vulnerabilities" character varying NULL, ADD CONSTRAINT "vulnerabilities_job_results_vulnerabilities" FOREIGN KEY ("job_result_vulnerabilities") REFERENCES "job_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_job_templates_vulnerabilities" FOREIGN KEY ("job_template_vulnerabilities") REFERENCES "job_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_scheduled_jobs_vulnerabilities" FOREIGN KEY ("scheduled_job_vulnerabilities") REFERENCES "scheduled_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP CONSTRAINT "vulnerabilities_scheduled_jobs_vulnerabilities", DROP CONSTRAINT "vulnerabilities_job_templates_vulnerabilities", DROP CONSTRAINT "vulnerabilities_job_results_vulnerabilities", DROP COLUMN "scheduled_job_vulnerabilities", DROP COLUMN "job_template_vulnerabilities", DROP COLUMN "job_result_vulnerabilities";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_scheduled_jobs_tasks", DROP CONSTRAINT "tasks_job_templates_tasks", DROP CONSTRAINT "tasks_job_results_tasks", DROP COLUMN "scheduled_job_tasks", DROP COLUMN "job_template_tasks", DROP COLUMN "job_result_tasks";
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP CONSTRAINT "standards_scheduled_jobs_standards", DROP CONSTRAINT "standards_job_templates_standards", DROP CONSTRAINT "standards_job_results_standards", DROP COLUMN "scheduled_job_standards", DROP COLUMN "job_template_standards", DROP COLUMN "job_result_standards";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_scheduled_jobs_risks", DROP CONSTRAINT "risks_job_templates_risks", DROP CONSTRAINT "risks_job_results_risks", DROP COLUMN "scheduled_job_risks", DROP COLUMN "job_template_risks", DROP COLUMN "job_result_risks";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP CONSTRAINT "findings_scheduled_jobs_findings", DROP CONSTRAINT "findings_job_templates_findings", DROP CONSTRAINT "findings_job_results_findings", DROP COLUMN "scheduled_job_findings", DROP COLUMN "job_template_findings", DROP COLUMN "job_result_findings";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP CONSTRAINT "evidences_scheduled_jobs_evidence", DROP CONSTRAINT "evidences_job_templates_evidence", DROP CONSTRAINT "evidences_job_results_evidence", DROP COLUMN "scheduled_job_evidence", DROP COLUMN "job_template_evidence", DROP COLUMN "job_result_evidence";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_scheduled_jobs_entities", DROP CONSTRAINT "entities_job_templates_entities", DROP CONSTRAINT "entities_job_results_entities", DROP COLUMN "scheduled_job_entities", DROP COLUMN "job_template_entities", DROP COLUMN "job_result_entities";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_job_templates_controls", DROP CONSTRAINT "controls_job_results_controls", DROP COLUMN "job_template_controls", DROP COLUMN "job_result_controls";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP CONSTRAINT "contacts_scheduled_jobs_contacts", DROP CONSTRAINT "contacts_job_templates_contacts", DROP CONSTRAINT "contacts_job_results_contacts", DROP COLUMN "scheduled_job_contacts", DROP COLUMN "job_template_contacts", DROP COLUMN "job_result_contacts";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP CONSTRAINT "assets_scheduled_jobs_assets", DROP CONSTRAINT "assets_job_templates_assets", DROP CONSTRAINT "assets_job_results_assets", DROP COLUMN "scheduled_job_assets", DROP COLUMN "job_template_assets", DROP COLUMN "job_result_assets";
-- reverse: modify "scheduled_jobs" table
ALTER TABLE "scheduled_jobs" DROP COLUMN "metadata";
-- reverse: modify "job_templates" table
ALTER TABLE "job_templates" DROP COLUMN "metadata", DROP COLUMN "script_path", DROP COLUMN "runtime_platform", ALTER COLUMN "download_url" SET NOT NULL;
-- reverse: modify "job_runners" table
ALTER TABLE "job_runners" DROP COLUMN "metadata";
-- reverse: modify "job_template_history" table
ALTER TABLE "job_template_history" DROP COLUMN "metadata", DROP COLUMN "script_path", DROP COLUMN "runtime_platform", ALTER COLUMN "download_url" SET NOT NULL;
-- reverse: modify "scheduled_job_history" table
ALTER TABLE "scheduled_job_history" DROP COLUMN "metadata";
-- reverse: modify "job_results" table
ALTER TABLE "job_results" DROP COLUMN "metadata", DROP COLUMN "compliance_job_id", ALTER COLUMN "exit_code" SET NOT NULL;
-- reverse: modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" DROP COLUMN "metadata";
