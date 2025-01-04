-- +goose Up
-- modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "users" table
ALTER TABLE "users" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "created_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;
-- modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "action_plans_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "action_plans_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "api_tokens_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "api_tokens_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "contacts_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "contacts_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "control_objectives_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "control_objectives_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "controls_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "document_data_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "document_data_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "email_verification_tokens_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "email_verification_tokens_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entities" table
ALTER TABLE "entities" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "entities_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "entity_types_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entity_types_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "events" table
ALTER TABLE "events" DROP COLUMN "created_by", DROP COLUMN "updated_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "events_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "events_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "files" table
ALTER TABLE "files" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "files_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "group_memberships_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "group_memberships_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "group_settings_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "group_settings_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "groups_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "hushes_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "hushes_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "integrations_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integrations_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "internal_policies_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "internal_policies_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "invites" table
ALTER TABLE "invites" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "invites_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "invites_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "narratives_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "narratives_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notes" table
ALTER TABLE "notes" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "notes_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "org_memberships_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_memberships_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "org_subscriptions_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "org_subscriptions_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "organization_settings_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "organization_settings_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "organizations_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "organizations_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "password_reset_tokens_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "password_reset_tokens_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "personal_access_tokens_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "personal_access_tokens_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "procedures_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "procedures_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "program_memberships_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "program_memberships_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "programs" table
ALTER TABLE "programs" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "programs_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "risks_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "standards" table
ALTER TABLE "standards" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "standards_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "standards_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "subcontrols_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "subscribers_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subscribers_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "tasks_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "templates" table
ALTER TABLE "templates" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "templates_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "templates_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "tfa_settings_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tfa_settings_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "deleted_by_id" character varying NULL, ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "user_settings_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "user_settings_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "webauthns" table
ALTER TABLE "webauthns" DROP COLUMN "created_by", DROP COLUMN "updated_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD CONSTRAINT "webauthns_users_created_by" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "webauthns_users_updated_by" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" DROP CONSTRAINT "webauthns_users_updated_by", DROP CONSTRAINT "webauthns_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP CONSTRAINT "user_settings_users_updated_by", DROP CONSTRAINT "user_settings_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP CONSTRAINT "tfa_settings_users_updated_by", DROP CONSTRAINT "tfa_settings_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP CONSTRAINT "templates_users_updated_by", DROP CONSTRAINT "templates_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_updated_by", DROP CONSTRAINT "tasks_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP CONSTRAINT "subscribers_users_updated_by", DROP CONSTRAINT "subscribers_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_users_updated_by", DROP CONSTRAINT "subcontrols_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP CONSTRAINT "standards_users_updated_by", DROP CONSTRAINT "standards_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_users_updated_by", DROP CONSTRAINT "risks_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_users_updated_by", DROP CONSTRAINT "programs_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP CONSTRAINT "program_memberships_users_updated_by", DROP CONSTRAINT "program_memberships_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_users_updated_by", DROP CONSTRAINT "procedures_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP CONSTRAINT "personal_access_tokens_users_updated_by", DROP CONSTRAINT "personal_access_tokens_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP CONSTRAINT "password_reset_tokens_users_updated_by", DROP CONSTRAINT "password_reset_tokens_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP CONSTRAINT "organizations_users_updated_by", DROP CONSTRAINT "organizations_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP CONSTRAINT "organization_settings_users_updated_by", DROP CONSTRAINT "organization_settings_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP CONSTRAINT "org_subscriptions_users_updated_by", DROP CONSTRAINT "org_subscriptions_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP CONSTRAINT "org_memberships_users_updated_by", DROP CONSTRAINT "org_memberships_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_users_updated_by", DROP CONSTRAINT "notes_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_users_updated_by", DROP CONSTRAINT "narratives_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP CONSTRAINT "invites_users_updated_by", DROP CONSTRAINT "invites_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_users_updated_by", DROP CONSTRAINT "internal_policies_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_users_updated_by", DROP CONSTRAINT "integrations_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP CONSTRAINT "hushes_users_updated_by", DROP CONSTRAINT "hushes_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_users_updated_by", DROP CONSTRAINT "groups_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" DROP CONSTRAINT "group_settings_users_updated_by", DROP CONSTRAINT "group_settings_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP CONSTRAINT "group_memberships_users_updated_by", DROP CONSTRAINT "group_memberships_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_users_updated_by", DROP CONSTRAINT "files_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "events_users_updated_by", DROP CONSTRAINT "events_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP CONSTRAINT "entity_types_users_updated_by", DROP CONSTRAINT "entity_types_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_users_updated_by", DROP CONSTRAINT "entities_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP CONSTRAINT "email_verification_tokens_users_updated_by", DROP CONSTRAINT "email_verification_tokens_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_users_updated_by", DROP CONSTRAINT "document_data_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_users_updated_by", DROP CONSTRAINT "controls_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP CONSTRAINT "control_objectives_users_updated_by", DROP CONSTRAINT "control_objectives_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP CONSTRAINT "contacts_users_updated_by", DROP CONSTRAINT "contacts_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP CONSTRAINT "api_tokens_users_updated_by", DROP CONSTRAINT "api_tokens_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_users_updated_by", DROP CONSTRAINT "action_plans_users_created_by", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", DROP COLUMN "deleted_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
