-- Drop index "actionplan_id" from table: "action_plans"
DROP INDEX "actionplan_id";
-- Drop index "apitoken_id" from table: "api_tokens"
DROP INDEX "apitoken_id";
-- Drop index "asset_id" from table: "assets"
DROP INDEX "asset_id";
-- Drop index "contact_id" from table: "contacts"
DROP INDEX "contact_id";
-- Drop index "controlimplementation_id" from table: "control_implementations"
DROP INDEX "controlimplementation_id";
-- Drop index "controlobjective_id" from table: "control_objectives"
DROP INDEX "controlobjective_id";
-- Drop index "controlscheduledjob_id" from table: "control_scheduled_jobs"
DROP INDEX "controlscheduledjob_id";
-- Drop index "control_id" from table: "controls"
DROP INDEX "control_id";
-- Create index "control_standard_id_deleted_at_owner_id" to table: "controls"
CREATE INDEX "control_standard_id_deleted_at_owner_id" ON "controls" ("standard_id", "deleted_at", "owner_id");
-- Drop index "customdomain_id" from table: "custom_domains"
DROP INDEX "customdomain_id";
-- Drop index "dnsverification_id" from table: "dns_verifications"
DROP INDEX "dnsverification_id";
-- Drop index "documentdata_id" from table: "document_data"
DROP INDEX "documentdata_id";
-- Drop index "emailverificationtoken_id" from table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_id";
-- Drop index "entity_id" from table: "entities"
DROP INDEX "entity_id";
-- Drop index "entitytype_id" from table: "entity_types"
DROP INDEX "entitytype_id";
-- Drop index "event_id" from table: "events"
DROP INDEX "event_id";
-- Drop index "evidence_id" from table: "evidences"
DROP INDEX "evidence_id";
-- Drop index "file_id" from table: "files"
DROP INDEX "file_id";
-- Drop index "groupmembership_id" from table: "group_memberships"
DROP INDEX "groupmembership_id";
-- Drop index "groupsetting_id" from table: "group_settings"
DROP INDEX "groupsetting_id";
-- Drop index "group_id" from table: "groups"
DROP INDEX "group_id";
-- Drop index "hush_id" from table: "hushes"
DROP INDEX "hush_id";
-- Drop index "integration_id" from table: "integrations"
DROP INDEX "integration_id";
-- Drop index "internalpolicy_id" from table: "internal_policies"
DROP INDEX "internalpolicy_id";
-- Drop index "invite_id" from table: "invites"
DROP INDEX "invite_id";
-- Drop index "jobresult_id" from table: "job_results"
DROP INDEX "jobresult_id";
-- Drop index "jobrunnerregistrationtoken_id" from table: "job_runner_registration_tokens"
DROP INDEX "jobrunnerregistrationtoken_id";
-- Drop index "jobrunnertoken_id" from table: "job_runner_tokens"
DROP INDEX "jobrunnertoken_id";
-- Drop index "jobrunner_id" from table: "job_runners"
DROP INDEX "jobrunner_id";
-- Drop index "mappabledomain_id" from table: "mappable_domains"
DROP INDEX "mappabledomain_id";
-- Drop index "mappedcontrol_id" from table: "mapped_controls"
DROP INDEX "mappedcontrol_id";
-- Drop index "narrative_id" from table: "narratives"
DROP INDEX "narrative_id";
-- Drop index "note_id" from table: "notes"
DROP INDEX "note_id";
-- Drop index "onboarding_id" from table: "onboardings"
DROP INDEX "onboarding_id";
-- Drop index "orgmembership_id" from table: "org_memberships"
DROP INDEX "orgmembership_id";
-- Drop index "orgmodule_id" from table: "org_modules"
DROP INDEX "orgmodule_id";
-- Drop index "orgprice_id" from table: "org_prices"
DROP INDEX "orgprice_id";
-- Drop index "orgproduct_id" from table: "org_products"
DROP INDEX "orgproduct_id";
-- Drop index "orgsubscription_id" from table: "org_subscriptions"
DROP INDEX "orgsubscription_id";
-- Drop index "organizationsetting_id" from table: "organization_settings"
DROP INDEX "organizationsetting_id";
-- Drop index "organization_id" from table: "organizations"
DROP INDEX "organization_id";
-- Drop index "passwordresettoken_id" from table: "password_reset_tokens"
DROP INDEX "passwordresettoken_id";
-- Drop index "personalaccesstoken_id" from table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_id";
-- Drop index "procedure_id" from table: "procedures"
DROP INDEX "procedure_id";
-- Drop index "programmembership_id" from table: "program_memberships"
DROP INDEX "programmembership_id";
-- Drop index "program_id" from table: "programs"
DROP INDEX "program_id";
-- Drop index "risk_id" from table: "risks"
DROP INDEX "risk_id";
-- Drop index "scan_id" from table: "scans"
DROP INDEX "scan_id";
-- Drop index "scheduledjobrun_id" from table: "scheduled_job_runs"
DROP INDEX "scheduledjobrun_id";
-- Drop index "scheduledjob_id" from table: "scheduled_jobs"
DROP INDEX "scheduledjob_id";
-- Drop index "standard_id" from table: "standards"
DROP INDEX "standard_id";
-- Drop index "subcontrol_id" from table: "subcontrols"
DROP INDEX "subcontrol_id";
-- Drop index "subscriber_id" from table: "subscribers"
DROP INDEX "subscriber_id";
-- Drop index "task_id" from table: "tasks"
DROP INDEX "task_id";
-- Drop index "template_id" from table: "templates"
DROP INDEX "template_id";
-- Drop index "tfasetting_id" from table: "tfa_settings"
DROP INDEX "tfasetting_id";
-- Drop index "trustcentersetting_id" from table: "trust_center_settings"
DROP INDEX "trustcentersetting_id";
-- Drop index "trustcenter_id" from table: "trust_centers"
DROP INDEX "trustcenter_id";
-- Drop index "usersetting_id" from table: "user_settings"
DROP INDEX "usersetting_id";
-- Drop index "user_id" from table: "users"
DROP INDEX "user_id";
-- Drop index "webauthn_id" from table: "webauthns"
DROP INDEX "webauthn_id";
