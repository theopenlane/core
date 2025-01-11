-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "controls" table
ALTER TABLE "controls" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "entities" table
ALTER TABLE "entities" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "created_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;
-- Modify "events" table
ALTER TABLE "events" DROP COLUMN "created_by", DROP COLUMN "updated_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;
-- Modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "files" table
ALTER TABLE "files" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "invites" table
ALTER TABLE "invites" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "notes" table
ALTER TABLE "notes" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "programs" table
ALTER TABLE "programs" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "risks" table
ALTER TABLE "risks" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "standards" table
ALTER TABLE "standards" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "templates" table
ALTER TABLE "templates" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "users" table
ALTER TABLE "users" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- Modify "webauthns" table
ALTER TABLE "webauthns" DROP COLUMN "created_by", DROP COLUMN "updated_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;
