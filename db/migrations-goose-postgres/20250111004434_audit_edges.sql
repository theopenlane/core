-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "entities" table
ALTER TABLE "entities" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "created_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;
-- modify "events" table
ALTER TABLE "events" DROP COLUMN "created_by", DROP COLUMN "updated_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;
-- modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "files" table
ALTER TABLE "files" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "invites" table
ALTER TABLE "invites" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "notes" table
ALTER TABLE "notes" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "programs" table
ALTER TABLE "programs" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "standards" table
ALTER TABLE "standards" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "templates" table
ALTER TABLE "templates" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "created_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "users" table
ALTER TABLE "users" DROP COLUMN "created_by", DROP COLUMN "updated_by", DROP COLUMN "deleted_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL, ADD COLUMN "deleted_by_id" character varying NULL;
-- modify "webauthns" table
ALTER TABLE "webauthns" DROP COLUMN "created_by", DROP COLUMN "updated_by", ADD COLUMN "created_by_id" character varying NULL, ADD COLUMN "updated_by_id" character varying NULL;

-- +goose Down
-- reverse: modify "webauthns" table
ALTER TABLE "webauthns" DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "user_settings" table
ALTER TABLE "user_settings" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "user_setting_history" table
ALTER TABLE "user_setting_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "tfa_settings" table
ALTER TABLE "tfa_settings" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "task_history" table
ALTER TABLE "task_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "standards" table
ALTER TABLE "standards" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "personal_access_tokens" table
ALTER TABLE "personal_access_tokens" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "events" table
ALTER TABLE "events" DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "event_history" table
ALTER TABLE "event_history" DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entity_type_history" table
ALTER TABLE "entity_type_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "document_data_history" table
ALTER TABLE "document_data_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "api_tokens" table
ALTER TABLE "api_tokens" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "deleted_by_id", DROP COLUMN "updated_by_id", DROP COLUMN "created_by_id", ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "created_by" character varying NULL;
