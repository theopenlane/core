-- +goose Up
-- drop index "actionplan_id" from table: "action_plans"
DROP INDEX "actionplan_id";
-- drop index "apitoken_id" from table: "api_tokens"
DROP INDEX "apitoken_id";
-- drop index "asset_id" from table: "assets"
DROP INDEX "asset_id";
-- drop index "contact_id" from table: "contacts"
DROP INDEX "contact_id";
-- drop index "controlimplementation_id" from table: "control_implementations"
DROP INDEX "controlimplementation_id";
-- drop index "controlobjective_id" from table: "control_objectives"
DROP INDEX "controlobjective_id";
-- drop index "controlscheduledjob_id" from table: "control_scheduled_jobs"
DROP INDEX "controlscheduledjob_id";
-- drop index "control_id" from table: "controls"
DROP INDEX "control_id";
-- create index "control_standard_id_deleted_at_owner_id" to table: "controls"
CREATE INDEX "control_standard_id_deleted_at_owner_id" ON "controls" ("standard_id", "deleted_at", "owner_id");
-- drop index "customdomain_id" from table: "custom_domains"
DROP INDEX "customdomain_id";
-- drop index "dnsverification_id" from table: "dns_verifications"
DROP INDEX "dnsverification_id";
-- drop index "documentdata_id" from table: "document_data"
DROP INDEX "documentdata_id";
-- drop index "emailverificationtoken_id" from table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_id";
-- drop index "entity_id" from table: "entities"
DROP INDEX "entity_id";
-- drop index "entitytype_id" from table: "entity_types"
DROP INDEX "entitytype_id";
-- drop index "event_id" from table: "events"
DROP INDEX "event_id";
-- drop index "evidence_id" from table: "evidences"
DROP INDEX "evidence_id";
-- drop index "file_id" from table: "files"
DROP INDEX "file_id";
-- drop index "groupmembership_id" from table: "group_memberships"
DROP INDEX "groupmembership_id";
-- drop index "groupsetting_id" from table: "group_settings"
DROP INDEX "groupsetting_id";
-- drop index "group_id" from table: "groups"
DROP INDEX "group_id";
-- drop index "hush_id" from table: "hushes"
DROP INDEX "hush_id";
-- drop index "integration_id" from table: "integrations"
DROP INDEX "integration_id";
-- drop index "internalpolicy_id" from table: "internal_policies"
DROP INDEX "internalpolicy_id";
-- drop index "invite_id" from table: "invites"
DROP INDEX "invite_id";
-- drop index "jobresult_id" from table: "job_results"
DROP INDEX "jobresult_id";
-- drop index "jobrunnerregistrationtoken_id" from table: "job_runner_registration_tokens"
DROP INDEX "jobrunnerregistrationtoken_id";
-- drop index "jobrunnertoken_id" from table: "job_runner_tokens"
DROP INDEX "jobrunnertoken_id";
-- drop index "jobrunner_id" from table: "job_runners"
DROP INDEX "jobrunner_id";
-- drop index "mappabledomain_id" from table: "mappable_domains"
DROP INDEX "mappabledomain_id";
-- drop index "mappedcontrol_id" from table: "mapped_controls"
DROP INDEX "mappedcontrol_id";
-- drop index "narrative_id" from table: "narratives"
DROP INDEX "narrative_id";
-- drop index "note_id" from table: "notes"
DROP INDEX "note_id";
-- drop index "onboarding_id" from table: "onboardings"
DROP INDEX "onboarding_id";
-- drop index "orgmembership_id" from table: "org_memberships"
DROP INDEX "orgmembership_id";
-- drop index "orgmodule_id" from table: "org_modules"
DROP INDEX "orgmodule_id";
-- drop index "orgprice_id" from table: "org_prices"
DROP INDEX "orgprice_id";
-- drop index "orgproduct_id" from table: "org_products"
DROP INDEX "orgproduct_id";
-- drop index "orgsubscription_id" from table: "org_subscriptions"
DROP INDEX "orgsubscription_id";
-- drop index "organizationsetting_id" from table: "organization_settings"
DROP INDEX "organizationsetting_id";
-- drop index "organization_id" from table: "organizations"
DROP INDEX "organization_id";
-- drop index "passwordresettoken_id" from table: "password_reset_tokens"
DROP INDEX "passwordresettoken_id";
-- drop index "personalaccesstoken_id" from table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_id";
-- drop index "procedure_id" from table: "procedures"
DROP INDEX "procedure_id";
-- drop index "programmembership_id" from table: "program_memberships"
DROP INDEX "programmembership_id";
-- drop index "program_id" from table: "programs"
DROP INDEX "program_id";
-- drop index "risk_id" from table: "risks"
DROP INDEX "risk_id";
-- drop index "scan_id" from table: "scans"
DROP INDEX "scan_id";
-- drop index "scheduledjobrun_id" from table: "scheduled_job_runs"
DROP INDEX "scheduledjobrun_id";
-- drop index "scheduledjob_id" from table: "scheduled_jobs"
DROP INDEX "scheduledjob_id";
-- drop index "standard_id" from table: "standards"
DROP INDEX "standard_id";
-- drop index "subcontrol_id" from table: "subcontrols"
DROP INDEX "subcontrol_id";
-- drop index "subscriber_id" from table: "subscribers"
DROP INDEX "subscriber_id";
-- drop index "task_id" from table: "tasks"
DROP INDEX "task_id";
-- drop index "template_id" from table: "templates"
DROP INDEX "template_id";
-- drop index "tfasetting_id" from table: "tfa_settings"
DROP INDEX "tfasetting_id";
-- drop index "trustcentersetting_id" from table: "trust_center_settings"
DROP INDEX "trustcentersetting_id";
-- drop index "trustcenter_id" from table: "trust_centers"
DROP INDEX "trustcenter_id";
-- drop index "usersetting_id" from table: "user_settings"
DROP INDEX "usersetting_id";
-- drop index "user_id" from table: "users"
DROP INDEX "user_id";
-- drop index "webauthn_id" from table: "webauthns"
DROP INDEX "webauthn_id";

-- +goose Down
-- reverse: drop index "webauthn_id" from table: "webauthns"
CREATE UNIQUE INDEX "webauthn_id" ON "webauthns" ("id");
-- reverse: drop index "user_id" from table: "users"
CREATE UNIQUE INDEX "user_id" ON "users" ("id");
-- reverse: drop index "usersetting_id" from table: "user_settings"
CREATE UNIQUE INDEX "usersetting_id" ON "user_settings" ("id");
-- reverse: drop index "trustcenter_id" from table: "trust_centers"
CREATE UNIQUE INDEX "trustcenter_id" ON "trust_centers" ("id");
-- reverse: drop index "trustcentersetting_id" from table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_id" ON "trust_center_settings" ("id");
-- reverse: drop index "tfasetting_id" from table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_id" ON "tfa_settings" ("id");
-- reverse: drop index "template_id" from table: "templates"
CREATE UNIQUE INDEX "template_id" ON "templates" ("id");
-- reverse: drop index "task_id" from table: "tasks"
CREATE UNIQUE INDEX "task_id" ON "tasks" ("id");
-- reverse: drop index "subscriber_id" from table: "subscribers"
CREATE UNIQUE INDEX "subscriber_id" ON "subscribers" ("id");
-- reverse: drop index "subcontrol_id" from table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_id" ON "subcontrols" ("id");
-- reverse: drop index "standard_id" from table: "standards"
CREATE UNIQUE INDEX "standard_id" ON "standards" ("id");
-- reverse: drop index "scheduledjob_id" from table: "scheduled_jobs"
CREATE UNIQUE INDEX "scheduledjob_id" ON "scheduled_jobs" ("id");
-- reverse: drop index "scheduledjobrun_id" from table: "scheduled_job_runs"
CREATE UNIQUE INDEX "scheduledjobrun_id" ON "scheduled_job_runs" ("id");
-- reverse: drop index "scan_id" from table: "scans"
CREATE UNIQUE INDEX "scan_id" ON "scans" ("id");
-- reverse: drop index "risk_id" from table: "risks"
CREATE UNIQUE INDEX "risk_id" ON "risks" ("id");
-- reverse: drop index "program_id" from table: "programs"
CREATE UNIQUE INDEX "program_id" ON "programs" ("id");
-- reverse: drop index "programmembership_id" from table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_id" ON "program_memberships" ("id");
-- reverse: drop index "procedure_id" from table: "procedures"
CREATE UNIQUE INDEX "procedure_id" ON "procedures" ("id");
-- reverse: drop index "personalaccesstoken_id" from table: "personal_access_tokens"
CREATE UNIQUE INDEX "personalaccesstoken_id" ON "personal_access_tokens" ("id");
-- reverse: drop index "passwordresettoken_id" from table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_id" ON "password_reset_tokens" ("id");
-- reverse: drop index "organization_id" from table: "organizations"
CREATE UNIQUE INDEX "organization_id" ON "organizations" ("id");
-- reverse: drop index "organizationsetting_id" from table: "organization_settings"
CREATE UNIQUE INDEX "organizationsetting_id" ON "organization_settings" ("id");
-- reverse: drop index "orgsubscription_id" from table: "org_subscriptions"
CREATE UNIQUE INDEX "orgsubscription_id" ON "org_subscriptions" ("id");
-- reverse: drop index "orgproduct_id" from table: "org_products"
CREATE UNIQUE INDEX "orgproduct_id" ON "org_products" ("id");
-- reverse: drop index "orgprice_id" from table: "org_prices"
CREATE UNIQUE INDEX "orgprice_id" ON "org_prices" ("id");
-- reverse: drop index "orgmodule_id" from table: "org_modules"
CREATE UNIQUE INDEX "orgmodule_id" ON "org_modules" ("id");
-- reverse: drop index "orgmembership_id" from table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_id" ON "org_memberships" ("id");
-- reverse: drop index "onboarding_id" from table: "onboardings"
CREATE UNIQUE INDEX "onboarding_id" ON "onboardings" ("id");
-- reverse: drop index "note_id" from table: "notes"
CREATE UNIQUE INDEX "note_id" ON "notes" ("id");
-- reverse: drop index "narrative_id" from table: "narratives"
CREATE UNIQUE INDEX "narrative_id" ON "narratives" ("id");
-- reverse: drop index "mappedcontrol_id" from table: "mapped_controls"
CREATE UNIQUE INDEX "mappedcontrol_id" ON "mapped_controls" ("id");
-- reverse: drop index "mappabledomain_id" from table: "mappable_domains"
CREATE UNIQUE INDEX "mappabledomain_id" ON "mappable_domains" ("id");
-- reverse: drop index "jobrunner_id" from table: "job_runners"
CREATE UNIQUE INDEX "jobrunner_id" ON "job_runners" ("id");
-- reverse: drop index "jobrunnertoken_id" from table: "job_runner_tokens"
CREATE UNIQUE INDEX "jobrunnertoken_id" ON "job_runner_tokens" ("id");
-- reverse: drop index "jobrunnerregistrationtoken_id" from table: "job_runner_registration_tokens"
CREATE UNIQUE INDEX "jobrunnerregistrationtoken_id" ON "job_runner_registration_tokens" ("id");
-- reverse: drop index "jobresult_id" from table: "job_results"
CREATE UNIQUE INDEX "jobresult_id" ON "job_results" ("id");
-- reverse: drop index "invite_id" from table: "invites"
CREATE UNIQUE INDEX "invite_id" ON "invites" ("id");
-- reverse: drop index "internalpolicy_id" from table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_id" ON "internal_policies" ("id");
-- reverse: drop index "integration_id" from table: "integrations"
CREATE UNIQUE INDEX "integration_id" ON "integrations" ("id");
-- reverse: drop index "hush_id" from table: "hushes"
CREATE UNIQUE INDEX "hush_id" ON "hushes" ("id");
-- reverse: drop index "group_id" from table: "groups"
CREATE UNIQUE INDEX "group_id" ON "groups" ("id");
-- reverse: drop index "groupsetting_id" from table: "group_settings"
CREATE UNIQUE INDEX "groupsetting_id" ON "group_settings" ("id");
-- reverse: drop index "groupmembership_id" from table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_id" ON "group_memberships" ("id");
-- reverse: drop index "file_id" from table: "files"
CREATE UNIQUE INDEX "file_id" ON "files" ("id");
-- reverse: drop index "evidence_id" from table: "evidences"
CREATE UNIQUE INDEX "evidence_id" ON "evidences" ("id");
-- reverse: drop index "event_id" from table: "events"
CREATE UNIQUE INDEX "event_id" ON "events" ("id");
-- reverse: drop index "entitytype_id" from table: "entity_types"
CREATE UNIQUE INDEX "entitytype_id" ON "entity_types" ("id");
-- reverse: drop index "entity_id" from table: "entities"
CREATE UNIQUE INDEX "entity_id" ON "entities" ("id");
-- reverse: drop index "emailverificationtoken_id" from table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_id" ON "email_verification_tokens" ("id");
-- reverse: drop index "documentdata_id" from table: "document_data"
CREATE UNIQUE INDEX "documentdata_id" ON "document_data" ("id");
-- reverse: drop index "dnsverification_id" from table: "dns_verifications"
CREATE UNIQUE INDEX "dnsverification_id" ON "dns_verifications" ("id");
-- reverse: drop index "customdomain_id" from table: "custom_domains"
CREATE UNIQUE INDEX "customdomain_id" ON "custom_domains" ("id");
-- reverse: create index "control_standard_id_deleted_at_owner_id" to table: "controls"
DROP INDEX "control_standard_id_deleted_at_owner_id";
-- reverse: drop index "control_id" from table: "controls"
CREATE UNIQUE INDEX "control_id" ON "controls" ("id");
-- reverse: drop index "controlscheduledjob_id" from table: "control_scheduled_jobs"
CREATE UNIQUE INDEX "controlscheduledjob_id" ON "control_scheduled_jobs" ("id");
-- reverse: drop index "controlobjective_id" from table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_id" ON "control_objectives" ("id");
-- reverse: drop index "controlimplementation_id" from table: "control_implementations"
CREATE UNIQUE INDEX "controlimplementation_id" ON "control_implementations" ("id");
-- reverse: drop index "contact_id" from table: "contacts"
CREATE UNIQUE INDEX "contact_id" ON "contacts" ("id");
-- reverse: drop index "asset_id" from table: "assets"
CREATE UNIQUE INDEX "asset_id" ON "assets" ("id");
-- reverse: drop index "apitoken_id" from table: "api_tokens"
CREATE UNIQUE INDEX "apitoken_id" ON "api_tokens" ("id");
-- reverse: drop index "actionplan_id" from table: "action_plans"
CREATE UNIQUE INDEX "actionplan_id" ON "action_plans" ("id");
