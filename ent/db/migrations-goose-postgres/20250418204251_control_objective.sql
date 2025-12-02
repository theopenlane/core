-- +goose Up
-- create index "actionplan_id" to table: "action_plans"
CREATE UNIQUE INDEX "actionplan_id" ON "action_plans" ("id");
-- create index "apitoken_id" to table: "api_tokens"
CREATE UNIQUE INDEX "apitoken_id" ON "api_tokens" ("id");
-- create index "contact_id" to table: "contacts"
CREATE UNIQUE INDEX "contact_id" ON "contacts" ("id");
-- create index "controlimplementation_id" to table: "control_implementations"
CREATE UNIQUE INDEX "controlimplementation_id" ON "control_implementations" ("id");
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "status" SET DEFAULT 'DRAFT';
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ALTER COLUMN "status" SET DEFAULT 'DRAFT';
-- create index "controlobjective_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_id" ON "control_objectives" ("id");
-- create index "control_id" to table: "controls"
CREATE UNIQUE INDEX "control_id" ON "controls" ("id");
-- create index "documentdata_id" to table: "document_data"
CREATE UNIQUE INDEX "documentdata_id" ON "document_data" ("id");
-- create index "emailverificationtoken_id" to table: "email_verification_tokens"
CREATE UNIQUE INDEX "emailverificationtoken_id" ON "email_verification_tokens" ("id");
-- create index "entity_id" to table: "entities"
CREATE UNIQUE INDEX "entity_id" ON "entities" ("id");
-- create index "entitytype_id" to table: "entity_types"
CREATE UNIQUE INDEX "entitytype_id" ON "entity_types" ("id");
-- create index "event_id" to table: "events"
CREATE UNIQUE INDEX "event_id" ON "events" ("id");
-- create index "evidence_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_id" ON "evidences" ("id");
-- create index "file_id" to table: "files"
CREATE UNIQUE INDEX "file_id" ON "files" ("id");
-- create index "groupmembership_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_id" ON "group_memberships" ("id");
-- create index "groupsetting_id" to table: "group_settings"
CREATE UNIQUE INDEX "groupsetting_id" ON "group_settings" ("id");
-- create index "group_id" to table: "groups"
CREATE UNIQUE INDEX "group_id" ON "groups" ("id");
-- create index "hush_id" to table: "hushes"
CREATE UNIQUE INDEX "hush_id" ON "hushes" ("id");
-- create index "integration_id" to table: "integrations"
CREATE UNIQUE INDEX "integration_id" ON "integrations" ("id");
-- create index "internalpolicy_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_id" ON "internal_policies" ("id");
-- create index "invite_id" to table: "invites"
CREATE UNIQUE INDEX "invite_id" ON "invites" ("id");
-- create index "mappedcontrol_id" to table: "mapped_controls"
CREATE UNIQUE INDEX "mappedcontrol_id" ON "mapped_controls" ("id");
-- create index "narrative_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_id" ON "narratives" ("id");
-- create index "note_id" to table: "notes"
CREATE UNIQUE INDEX "note_id" ON "notes" ("id");
-- create index "onboarding_id" to table: "onboardings"
CREATE UNIQUE INDEX "onboarding_id" ON "onboardings" ("id");
-- create index "orgmembership_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_id" ON "org_memberships" ("id");
-- create index "orgsubscription_id" to table: "org_subscriptions"
CREATE UNIQUE INDEX "orgsubscription_id" ON "org_subscriptions" ("id");
-- create index "organizationsetting_id" to table: "organization_settings"
CREATE UNIQUE INDEX "organizationsetting_id" ON "organization_settings" ("id");
-- create index "organization_id" to table: "organizations"
CREATE UNIQUE INDEX "organization_id" ON "organizations" ("id");
-- create index "passwordresettoken_id" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "passwordresettoken_id" ON "password_reset_tokens" ("id");
-- create index "personalaccesstoken_id" to table: "personal_access_tokens"
CREATE UNIQUE INDEX "personalaccesstoken_id" ON "personal_access_tokens" ("id");
-- create index "procedure_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_id" ON "procedures" ("id");
-- create index "programmembership_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_id" ON "program_memberships" ("id");
-- create index "program_id" to table: "programs"
CREATE UNIQUE INDEX "program_id" ON "programs" ("id");
-- create index "risk_id" to table: "risks"
CREATE UNIQUE INDEX "risk_id" ON "risks" ("id");
-- create index "standard_id" to table: "standards"
CREATE UNIQUE INDEX "standard_id" ON "standards" ("id");
-- create index "subcontrol_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_id" ON "subcontrols" ("id");
-- create index "subscriber_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_id" ON "subscribers" ("id");
-- create index "task_id" to table: "tasks"
CREATE UNIQUE INDEX "task_id" ON "tasks" ("id");
-- create index "template_id" to table: "templates"
CREATE UNIQUE INDEX "template_id" ON "templates" ("id");
-- create index "tfasetting_id" to table: "tfa_settings"
CREATE UNIQUE INDEX "tfasetting_id" ON "tfa_settings" ("id");
-- create index "usersetting_id" to table: "user_settings"
CREATE UNIQUE INDEX "usersetting_id" ON "user_settings" ("id");
-- drop index "user_email_auth_provider" from table: "users"
DROP INDEX "user_email_auth_provider";
-- create index "user_email" to table: "users"
CREATE UNIQUE INDEX "user_email" ON "users" ("email") WHERE (deleted_at IS NULL);
-- create index "webauthn_id" to table: "webauthns"
CREATE UNIQUE INDEX "webauthn_id" ON "webauthns" ("id");

-- +goose Down
-- reverse: create index "webauthn_id" to table: "webauthns"
DROP INDEX "webauthn_id";
-- reverse: create index "user_email" to table: "users"
DROP INDEX "user_email";
-- reverse: drop index "user_email_auth_provider" from table: "users"
CREATE UNIQUE INDEX "user_email_auth_provider" ON "users" ("email", "auth_provider") WHERE (deleted_at IS NULL);
-- reverse: create index "usersetting_id" to table: "user_settings"
DROP INDEX "usersetting_id";
-- reverse: create index "tfasetting_id" to table: "tfa_settings"
DROP INDEX "tfasetting_id";
-- reverse: create index "template_id" to table: "templates"
DROP INDEX "template_id";
-- reverse: create index "task_id" to table: "tasks"
DROP INDEX "task_id";
-- reverse: create index "subscriber_id" to table: "subscribers"
DROP INDEX "subscriber_id";
-- reverse: create index "subcontrol_id" to table: "subcontrols"
DROP INDEX "subcontrol_id";
-- reverse: create index "standard_id" to table: "standards"
DROP INDEX "standard_id";
-- reverse: create index "risk_id" to table: "risks"
DROP INDEX "risk_id";
-- reverse: create index "program_id" to table: "programs"
DROP INDEX "program_id";
-- reverse: create index "programmembership_id" to table: "program_memberships"
DROP INDEX "programmembership_id";
-- reverse: create index "procedure_id" to table: "procedures"
DROP INDEX "procedure_id";
-- reverse: create index "personalaccesstoken_id" to table: "personal_access_tokens"
DROP INDEX "personalaccesstoken_id";
-- reverse: create index "passwordresettoken_id" to table: "password_reset_tokens"
DROP INDEX "passwordresettoken_id";
-- reverse: create index "organization_id" to table: "organizations"
DROP INDEX "organization_id";
-- reverse: create index "organizationsetting_id" to table: "organization_settings"
DROP INDEX "organizationsetting_id";
-- reverse: create index "orgsubscription_id" to table: "org_subscriptions"
DROP INDEX "orgsubscription_id";
-- reverse: create index "orgmembership_id" to table: "org_memberships"
DROP INDEX "orgmembership_id";
-- reverse: create index "onboarding_id" to table: "onboardings"
DROP INDEX "onboarding_id";
-- reverse: create index "note_id" to table: "notes"
DROP INDEX "note_id";
-- reverse: create index "narrative_id" to table: "narratives"
DROP INDEX "narrative_id";
-- reverse: create index "mappedcontrol_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_id";
-- reverse: create index "invite_id" to table: "invites"
DROP INDEX "invite_id";
-- reverse: create index "internalpolicy_id" to table: "internal_policies"
DROP INDEX "internalpolicy_id";
-- reverse: create index "integration_id" to table: "integrations"
DROP INDEX "integration_id";
-- reverse: create index "hush_id" to table: "hushes"
DROP INDEX "hush_id";
-- reverse: create index "group_id" to table: "groups"
DROP INDEX "group_id";
-- reverse: create index "groupsetting_id" to table: "group_settings"
DROP INDEX "groupsetting_id";
-- reverse: create index "groupmembership_id" to table: "group_memberships"
DROP INDEX "groupmembership_id";
-- reverse: create index "file_id" to table: "files"
DROP INDEX "file_id";
-- reverse: create index "evidence_id" to table: "evidences"
DROP INDEX "evidence_id";
-- reverse: create index "event_id" to table: "events"
DROP INDEX "event_id";
-- reverse: create index "entitytype_id" to table: "entity_types"
DROP INDEX "entitytype_id";
-- reverse: create index "entity_id" to table: "entities"
DROP INDEX "entity_id";
-- reverse: create index "emailverificationtoken_id" to table: "email_verification_tokens"
DROP INDEX "emailverificationtoken_id";
-- reverse: create index "documentdata_id" to table: "document_data"
DROP INDEX "documentdata_id";
-- reverse: create index "control_id" to table: "controls"
DROP INDEX "control_id";
-- reverse: create index "controlobjective_id" to table: "control_objectives"
DROP INDEX "controlobjective_id";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" ALTER COLUMN "status" DROP DEFAULT;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "status" DROP DEFAULT;
-- reverse: create index "controlimplementation_id" to table: "control_implementations"
DROP INDEX "controlimplementation_id";
-- reverse: create index "contact_id" to table: "contacts"
DROP INDEX "contact_id";
-- reverse: create index "apitoken_id" to table: "api_tokens"
DROP INDEX "apitoken_id";
-- reverse: create index "actionplan_id" to table: "action_plans"
DROP INDEX "actionplan_id";
