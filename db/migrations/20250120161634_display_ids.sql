-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "mapping_id";
-- Modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "mapping_id";
-- Modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "mapping_id";
-- Modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "mapping_id";
-- Modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "mapping_id";
-- Modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "controlobjective_display_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_display_id_owner_id" ON "control_objectives" ("display_id", "owner_id");
-- Modify "controls" table
ALTER TABLE "controls" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "control_display_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_display_id_owner_id" ON "controls" ("display_id", "owner_id");
-- Modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "mapping_id";
-- Modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "mapping_id";
-- Modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP COLUMN "mapping_id";
-- Modify "entities" table
ALTER TABLE "entities" DROP COLUMN "mapping_id";
-- Modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "mapping_id";
-- Modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "mapping_id";
-- Modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "mapping_id";
-- Modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "mapping_id";
-- Modify "events" table
ALTER TABLE "events" DROP COLUMN "mapping_id";
-- Modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "mapping_id";
-- Modify "files" table
ALTER TABLE "files" DROP COLUMN "mapping_id";
-- Modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "mapping_id";
-- Modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "mapping_id";
-- Modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "mapping_id";
-- Modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "mapping_id";
-- Modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "mapping_id";
-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "mapping_id";
-- Modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "mapping_id";
-- Modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "mapping_id";
-- Modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "mapping_id";
-- Modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "mapping_id";
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "internalpolicy_display_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_display_id_owner_id" ON "internal_policies" ("display_id", "owner_id");
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "invites" table
ALTER TABLE "invites" DROP COLUMN "mapping_id";
-- Modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "narrative_display_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_display_id_owner_id" ON "narratives" ("display_id", "owner_id");
-- Modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "notes" table
ALTER TABLE "notes" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "note_display_id_owner_id" to table: "notes"
CREATE UNIQUE INDEX "note_display_id_owner_id" ON "notes" ("display_id", "owner_id");
-- Modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "mapping_id";
-- Modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "mapping_id";
-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "mapping_id";
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "mapping_id";
-- Modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "mapping_id";
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "mapping_id";
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "mapping_id";
-- Modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP COLUMN "mapping_id";
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "mapping_id";
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "procedure_display_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_display_id_owner_id" ON "procedures" ("display_id", "owner_id");
-- Modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "mapping_id";
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "mapping_id";
-- Modify "programs" table
ALTER TABLE "programs" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "program_display_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_display_id_owner_id" ON "programs" ("display_id", "owner_id");
-- Modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "risks" table
ALTER TABLE "risks" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "risk_display_id_owner_id" to table: "risks"
CREATE UNIQUE INDEX "risk_display_id_owner_id" ON "risks" ("display_id", "owner_id");
-- Modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "mapping_id";
-- Modify "standards" table
ALTER TABLE "standards" DROP COLUMN "mapping_id";
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "subcontrol_display_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_display_id_owner_id" ON "subcontrols" ("display_id", "owner_id");
-- Modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "mapping_id";
-- Modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL, ADD COLUMN "owner_id" character varying NOT NULL, ADD COLUMN "priority" character varying NOT NULL DEFAULT 'MEDIUM', ADD COLUMN "assignee_id" character varying NULL, ADD COLUMN "assigner_id" character varying NOT NULL;
-- Modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "mapping_id";
-- Modify "templates" table
ALTER TABLE "templates" DROP COLUMN "mapping_id";
-- Modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "mapping_id";
-- Modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "mapping_id";
-- Modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "mapping_id";
-- Modify "users" table
ALTER TABLE "users" DROP COLUMN "mapping_id", ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "users_display_id_key" to table: "users"
CREATE UNIQUE INDEX "users_display_id_key" ON "users" ("display_id");
-- Modify "webauthns" table
ALTER TABLE "webauthns" DROP COLUMN "mapping_id";
-- Modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "mapping_id";
-- Modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_assignee_tasks", DROP CONSTRAINT "tasks_users_assigner_tasks", DROP COLUMN "mapping_id", DROP COLUMN "user_assigner_tasks", DROP COLUMN "user_assignee_tasks", ADD COLUMN "display_id" character varying NOT NULL, ADD COLUMN "priority" character varying NOT NULL DEFAULT 'MEDIUM', ADD COLUMN "owner_id" character varying NOT NULL, ADD COLUMN "assigner_id" character varying NOT NULL, ADD COLUMN "assignee_id" character varying NULL, ADD CONSTRAINT "tasks_users_assignee_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_assigner_tasks" FOREIGN KEY ("assigner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "tasks_organizations_tasks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "task_display_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_display_id_owner_id" ON "tasks" ("display_id", "owner_id");
