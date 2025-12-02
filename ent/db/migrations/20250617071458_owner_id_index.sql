-- Create index "actionplan_owner_id" to table: "action_plans"
CREATE INDEX "actionplan_owner_id" ON "action_plans" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "apitoken_owner_id" to table: "api_tokens"
CREATE INDEX "apitoken_owner_id" ON "api_tokens" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "contact_owner_id" to table: "contacts"
CREATE INDEX "contact_owner_id" ON "contacts" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "controlimplementation_owner_id" to table: "control_implementations"
CREATE INDEX "controlimplementation_owner_id" ON "control_implementations" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "controlobjective_owner_id" to table: "control_objectives"
CREATE INDEX "controlobjective_owner_id" ON "control_objectives" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "controlscheduledjob_owner_id" to table: "control_scheduled_jobs"
CREATE INDEX "controlscheduledjob_owner_id" ON "control_scheduled_jobs" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "control_owner_id" to table: "controls"
CREATE INDEX "control_owner_id" ON "controls" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "dnsverification_owner_id" to table: "dns_verifications"
CREATE INDEX "dnsverification_owner_id" ON "dns_verifications" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "documentdata_owner_id" to table: "document_data"
CREATE INDEX "documentdata_owner_id" ON "document_data" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "entity_owner_id" to table: "entities"
CREATE INDEX "entity_owner_id" ON "entities" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "entitytype_owner_id" to table: "entity_types"
CREATE INDEX "entitytype_owner_id" ON "entity_types" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "evidence_owner_id" to table: "evidences"
CREATE INDEX "evidence_owner_id" ON "evidences" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "group_owner_id" to table: "groups"
CREATE INDEX "group_owner_id" ON "groups" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "hush_owner_id" to table: "hushes"
CREATE INDEX "hush_owner_id" ON "hushes" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "integration_owner_id" to table: "integrations"
CREATE INDEX "integration_owner_id" ON "integrations" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "internalpolicy_owner_id" to table: "internal_policies"
CREATE INDEX "internalpolicy_owner_id" ON "internal_policies" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "invite_owner_id" to table: "invites"
CREATE INDEX "invite_owner_id" ON "invites" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "jobresult_owner_id" to table: "job_results"
CREATE INDEX "jobresult_owner_id" ON "job_results" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "jobrunnerregistrationtoken_owner_id" to table: "job_runner_registration_tokens"
CREATE INDEX "jobrunnerregistrationtoken_owner_id" ON "job_runner_registration_tokens" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "jobrunnertoken_owner_id" to table: "job_runner_tokens"
CREATE INDEX "jobrunnertoken_owner_id" ON "job_runner_tokens" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "jobrunner_owner_id" to table: "job_runners"
CREATE INDEX "jobrunner_owner_id" ON "job_runners" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "mappedcontrol_owner_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_owner_id" ON "mapped_controls" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "narrative_owner_id" to table: "narratives"
CREATE INDEX "narrative_owner_id" ON "narratives" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "note_owner_id" to table: "notes"
CREATE INDEX "note_owner_id" ON "notes" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "orgsubscription_owner_id" to table: "org_subscriptions"
CREATE INDEX "orgsubscription_owner_id" ON "org_subscriptions" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "procedure_owner_id" to table: "procedures"
CREATE INDEX "procedure_owner_id" ON "procedures" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "program_owner_id" to table: "programs"
CREATE INDEX "program_owner_id" ON "programs" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "risk_owner_id" to table: "risks"
CREATE INDEX "risk_owner_id" ON "risks" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "scheduledjobrun_owner_id" to table: "scheduled_job_runs"
CREATE INDEX "scheduledjobrun_owner_id" ON "scheduled_job_runs" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "scheduledjob_owner_id" to table: "scheduled_jobs"
CREATE INDEX "scheduledjob_owner_id" ON "scheduled_jobs" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "standard_owner_id" to table: "standards"
CREATE INDEX "standard_owner_id" ON "standards" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "subcontrol_owner_id" to table: "subcontrols"
CREATE INDEX "subcontrol_owner_id" ON "subcontrols" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "subscriber_owner_id" to table: "subscribers"
CREATE INDEX "subscriber_owner_id" ON "subscribers" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "task_owner_id" to table: "tasks"
CREATE INDEX "task_owner_id" ON "tasks" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "template_owner_id" to table: "templates"
CREATE INDEX "template_owner_id" ON "templates" ("owner_id") WHERE (deleted_at IS NULL);
