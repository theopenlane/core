-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "mapping_id";
-- modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "mapping_id";
-- modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "mapping_id";
-- modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "mapping_id";
-- modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "mapping_id";
-- modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "controlobjective_display_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_display_id_owner_id" ON "control_objectives" ("display_id", "owner_id");
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "control_display_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_display_id_owner_id" ON "controls" ("display_id", "owner_id");
-- modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "mapping_id";
-- modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "mapping_id";
-- modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP COLUMN "mapping_id";
-- modify "entities" table
ALTER TABLE "entities" DROP COLUMN "mapping_id";
-- modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "mapping_id";
-- modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "mapping_id";
-- modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "mapping_id";
-- modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "mapping_id";
-- modify "events" table
ALTER TABLE "events" DROP COLUMN "mapping_id";
-- modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "mapping_id";
-- modify "files" table
ALTER TABLE "files" DROP COLUMN "mapping_id";
-- modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "mapping_id";
-- modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "mapping_id";
-- modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "mapping_id";
-- modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "mapping_id";
-- modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "mapping_id";
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "mapping_id";
-- modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "mapping_id";
-- modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "mapping_id";
-- modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "mapping_id";
-- modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "mapping_id";
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_display_id_owner_id" ON "internal_policies" ("display_id", "owner_id");
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "invites" table
ALTER TABLE "invites" DROP COLUMN "mapping_id";
-- modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "narrative_display_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_display_id_owner_id" ON "narratives" ("display_id", "owner_id");
-- modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "notes" table
ALTER TABLE "notes" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "note_display_id_owner_id" to table: "notes"
CREATE UNIQUE INDEX "note_display_id_owner_id" ON "notes" ("display_id", "owner_id");
-- modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "mapping_id";
-- modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "mapping_id";
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "mapping_id";
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "mapping_id";
-- modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "mapping_id";
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "mapping_id";
-- modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "mapping_id";
-- modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP COLUMN "mapping_id";
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "mapping_id";
-- modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "procedure_display_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_display_id_owner_id" ON "procedures" ("display_id", "owner_id");
-- modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "mapping_id";
-- modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "mapping_id";
-- modify "programs" table
ALTER TABLE "programs" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "program_display_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_display_id_owner_id" ON "programs" ("display_id", "owner_id");
-- modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "risk_display_id_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_display_id_owner_id" ON "risks" ("display_id", "owner_id");
-- modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "mapping_id";
-- modify "standards" table
ALTER TABLE "standards" DROP COLUMN "mapping_id";
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "subcontrol_display_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_display_id_owner_id" ON "subcontrols" ("display_id", "owner_id");
-- modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "mapping_id";
-- modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL, ADD COLUMN "owner_id" character varying NOT NULL, ADD COLUMN "priority" character varying NOT NULL DEFAULT 'MEDIUM', ADD COLUMN "assignee_id" character varying NULL, ADD COLUMN "assigner_id" character varying NOT NULL;
-- modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "mapping_id";
-- modify "templates" table
ALTER TABLE "templates" DROP COLUMN "mapping_id";
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "mapping_id";
-- modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "mapping_id";
-- modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "mapping_id";
-- modify "users" table
ALTER TABLE "users" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- create index "users_display_id_key" to table: "users"
CREATE UNIQUE INDEX "users_display_id_key" ON "users" ("display_id");
-- modify "webauthns" table
ALTER TABLE "webauthns" DROP COLUMN "mapping_id";
-- modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "mapping_id";
-- modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_assignee_tasks", DROP CONSTRAINT "tasks_users_assigner_tasks", DROP COLUMN "mapping_id", DROP COLUMN "user_assigner_tasks", DROP COLUMN "user_assignee_tasks", ADD COLUMN "display_id" character varying NOT NULL, ADD COLUMN "priority" character varying NOT NULL DEFAULT 'MEDIUM', ADD COLUMN "owner_id" character varying NOT NULL, ADD COLUMN "assigner_id" character varying NOT NULL, ADD COLUMN "assignee_id" character varying NULL, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create index "task_display_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_display_id_owner_id" ON "tasks" ("display_id", "owner_id");

-- +goose Down
-- reverse: create index "task_display_id_owner_id" to table: "tasks"
DROP INDEX "task_display_id_owner_id";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_organizations_tasks", DROP CONSTRAINT "tasks_users_assigner_tasks", DROP CONSTRAINT "tasks_users_assignee_tasks", DROP COLUMN "assignee_id", DROP COLUMN "assigner_id", DROP COLUMN "owner_id", DROP COLUMN "priority", DROP COLUMN "display_id", ADD COLUMN "user_assignee_tasks" character varying NULL, ADD COLUMN "user_assigner_tasks" character varying NOT NULL, ADD COLUMN "mapping_id" character varying NOT NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("user_assigner_tasks") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("user_assignee_tasks") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "users_display_id_key" to table: "users"
DROP INDEX "users_display_id_key";
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "user_setting_history" table
ALTER TABLE "user_setting_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "templates" table
ALTER TABLE "templates" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "assigner_id", DROP COLUMN "assignee_id", DROP COLUMN "priority", DROP COLUMN "owner_id", DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "subcontrol_display_id_owner_id" to table: "subcontrols"
DROP INDEX "subcontrol_display_id_owner_id";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "standards" table
ALTER TABLE "standards" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "risk_display_id_owner_id" to table: "risks"
DROP INDEX "risk_display_id_owner_id";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "program_display_id_owner_id" to table: "programs"
DROP INDEX "program_display_id_owner_id";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "program_membership_history" table
ALTER TABLE "program_membership_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "procedure_display_id_owner_id" to table: "procedures"
DROP INDEX "procedure_display_id_owner_id";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "note_display_id_owner_id" to table: "notes"
DROP INDEX "note_display_id_owner_id";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "narrative_display_id_owner_id" to table: "narratives"
DROP INDEX "narrative_display_id_owner_id";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "invites" table
ALTER TABLE "invites" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
DROP INDEX "internalpolicy_display_id_owner_id";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "group_setting_history" table
ALTER TABLE "group_setting_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "group_membership_history" table
ALTER TABLE "group_membership_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "files" table
ALTER TABLE "files" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "file_history" table
ALTER TABLE "file_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "events" table
ALTER TABLE "events" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "event_history" table
ALTER TABLE "event_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "entity_type_history" table
ALTER TABLE "entity_type_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "entities" table
ALTER TABLE "entities" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "document_data_history" table
ALTER TABLE "document_data_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "document_data" table
ALTER TABLE "document_data" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "control_display_id_owner_id" to table: "controls"
DROP INDEX "control_display_id_owner_id";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: create index "controlobjective_display_id_owner_id" to table: "control_objectives"
DROP INDEX "controlobjective_display_id_owner_id";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "display_id", ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "mapping_id" character varying NOT NULL;
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "mapping_id" character varying NOT NULL;
