-- +goose Up
-- create "workflow_assignment_history" table
CREATE TABLE "workflow_assignment_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "assignment_key" character varying NOT NULL, "role" character varying NOT NULL DEFAULT 'APPROVER', "label" character varying NULL, "required" boolean NOT NULL DEFAULT true, "status" character varying NOT NULL DEFAULT 'PENDING', "metadata" jsonb NULL, "decided_at" timestamptz NULL, "actor_user_id" character varying NULL, "actor_group_id" character varying NULL, "notes" text NULL, PRIMARY KEY ("id"));
-- create index "workflowassignmenthistory_history_time" to table: "workflow_assignment_history"
CREATE INDEX "workflowassignmenthistory_history_time" ON "workflow_assignment_history" ("history_time");
-- create "workflow_object_ref_history" table
CREATE TABLE "workflow_object_ref_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "control_id" character varying NULL, "task_id" character varying NULL, "internal_policy_id" character varying NULL, "finding_id" character varying NULL, "directory_account_id" character varying NULL, "directory_group_id" character varying NULL, "directory_membership_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflowobjectrefhistory_history_time" to table: "workflow_object_ref_history"
CREATE INDEX "workflowobjectrefhistory_history_time" ON "workflow_object_ref_history" ("history_time");
-- create "workflow_instance_history" table
CREATE TABLE "workflow_instance_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "workflow_definition_id" character varying NOT NULL, "state" character varying NOT NULL DEFAULT 'RUNNING', "context" jsonb NULL, "last_evaluated_at" timestamptz NULL, "definition_snapshot" jsonb NULL, PRIMARY KEY ("id"));
-- create index "workflowinstancehistory_history_time" to table: "workflow_instance_history"
CREATE INDEX "workflowinstancehistory_history_time" ON "workflow_instance_history" ("history_time");
-- create "directory_account_history" table
CREATE TABLE "directory_account_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "integration_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "external_id" character varying NOT NULL, "secondary_key" character varying NULL, "canonical_email" character varying NULL, "display_name" character varying NULL, "given_name" character varying NULL, "family_name" character varying NULL, "job_title" character varying NULL, "department" character varying NULL, "organization_unit" character varying NULL, "account_type" character varying NULL DEFAULT 'USER', "status" character varying NOT NULL DEFAULT 'ACTIVE', "mfa_state" character varying NOT NULL DEFAULT 'UNKNOWN', "last_seen_ip" character varying NULL, "last_login_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, PRIMARY KEY ("id"));
-- create index "directoryaccounthistory_history_time" to table: "directory_account_history"
CREATE INDEX "directoryaccounthistory_history_time" ON "directory_account_history" ("history_time");
-- create "workflow_event_history" table
CREATE TABLE "workflow_event_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "event_type" character varying NOT NULL, "payload" jsonb NULL, PRIMARY KEY ("id"));
-- create index "workfloweventhistory_history_time" to table: "workflow_event_history"
CREATE INDEX "workfloweventhistory_history_time" ON "workflow_event_history" ("history_time");
-- create "directory_group_history" table
CREATE TABLE "directory_group_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "integration_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "external_id" character varying NOT NULL, "email" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "classification" character varying NOT NULL DEFAULT 'TEAM', "status" character varying NOT NULL DEFAULT 'ACTIVE', "external_sharing_allowed" boolean NULL DEFAULT false, "member_count" bigint NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, PRIMARY KEY ("id"));
-- create index "directorygrouphistory_history_time" to table: "directory_group_history"
CREATE INDEX "directorygrouphistory_history_time" ON "directory_group_history" ("history_time");
-- create "workflow_definition_history" table
CREATE TABLE "workflow_definition_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "workflow_kind" character varying NOT NULL, "schema_type" character varying NOT NULL, "revision" bigint NOT NULL DEFAULT 1, "draft" boolean NOT NULL DEFAULT true, "published_at" timestamptz NULL, "cooldown_seconds" bigint NOT NULL DEFAULT 0, "is_default" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT true, "trigger_operations" jsonb NULL, "trigger_fields" jsonb NULL, "definition_json" jsonb NULL, "tracked_fields" jsonb NULL, PRIMARY KEY ("id"));
-- create index "workflowdefinitionhistory_history_time" to table: "workflow_definition_history"
CREATE INDEX "workflowdefinitionhistory_history_time" ON "workflow_definition_history" ("history_time");
-- create "directory_membership_history" table
CREATE TABLE "directory_membership_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "integration_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "directory_account_id" character varying NOT NULL, "directory_group_id" character varying NOT NULL, "role" character varying NULL DEFAULT 'MEMBER', "source" character varying NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "last_confirmed_run_id" character varying NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "directorymembershiphistory_history_time" to table: "directory_membership_history"
CREATE INDEX "directorymembershiphistory_history_time" ON "directory_membership_history" ("history_time");
-- create "workflow_assignment_target_history" table
CREATE TABLE "workflow_assignment_target_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "workflow_assignment_id" character varying NOT NULL, "target_type" character varying NOT NULL, "target_user_id" character varying NULL, "target_group_id" character varying NULL, "resolver_key" character varying NULL, PRIMARY KEY ("id"));
-- create index "workflowassignmenttargethistory_history_time" to table: "workflow_assignment_target_history"
CREATE INDEX "workflowassignmenttargethistory_history_time" ON "workflow_assignment_target_history" ("history_time");
-- create "directory_sync_runs" table
CREATE TABLE "directory_sync_runs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "source_cursor" character varying NULL, "full_count" bigint NOT NULL DEFAULT 0, "delta_count" bigint NOT NULL DEFAULT 0, "error" text NULL, "raw_manifest_file_id" character varying NULL, "stats" jsonb NULL, "integration_id" character varying NOT NULL, "integration_directory_sync_runs" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "directory_sync_runs_integrations_directory_sync_runs" FOREIGN KEY ("integration_directory_sync_runs") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_sync_runs_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_sync_runs_organizations_directory_sync_runs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "directorysyncrun_display_id_owner_id" to table: "directory_sync_runs"
CREATE UNIQUE INDEX "directorysyncrun_display_id_owner_id" ON "directory_sync_runs" ("display_id", "owner_id");
-- create index "directorysyncrun_integration_id_started_at" to table: "directory_sync_runs"
CREATE INDEX "directorysyncrun_integration_id_started_at" ON "directory_sync_runs" ("integration_id", "started_at");
-- create "directory_accounts" table
CREATE TABLE "directory_accounts" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "external_id" character varying NOT NULL, "secondary_key" character varying NULL, "canonical_email" character varying NULL, "display_name" character varying NULL, "given_name" character varying NULL, "family_name" character varying NULL, "job_title" character varying NULL, "department" character varying NULL, "organization_unit" character varying NULL, "account_type" character varying NULL DEFAULT 'USER', "status" character varying NOT NULL DEFAULT 'ACTIVE', "mfa_state" character varying NOT NULL DEFAULT 'UNKNOWN', "last_seen_ip" character varying NULL, "last_login_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, "integration_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "directory_sync_run_directory_accounts" character varying NULL, "integration_directory_accounts" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "directory_accounts_directory_sync_runs_directory_accounts" FOREIGN KEY ("directory_sync_run_directory_accounts") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_accounts_directory_sync_runs_directory_sync_run" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_accounts_integrations_directory_accounts" FOREIGN KEY ("integration_directory_accounts") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_accounts_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_accounts_organizations_directory_accounts" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "directoryaccount_directory_sync_run_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_directory_sync_run_id_canonical_email" ON "directory_accounts" ("directory_sync_run_id", "canonical_email");
-- create index "directoryaccount_display_id_owner_id" to table: "directory_accounts"
CREATE UNIQUE INDEX "directoryaccount_display_id_owner_id" ON "directory_accounts" ("display_id", "owner_id");
-- create index "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" to table: "directory_accounts"
CREATE UNIQUE INDEX "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" ON "directory_accounts" ("integration_id", "external_id", "directory_sync_run_id");
-- create index "directoryaccount_integration_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_integration_id_canonical_email" ON "directory_accounts" ("integration_id", "canonical_email");
-- create index "directoryaccount_owner_id_canonical_email" to table: "directory_accounts"
CREATE INDEX "directoryaccount_owner_id_canonical_email" ON "directory_accounts" ("owner_id", "canonical_email");
-- create "directory_groups" table
CREATE TABLE "directory_groups" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "external_id" character varying NOT NULL, "email" character varying NULL, "display_name" character varying NULL, "description" character varying NULL, "classification" character varying NOT NULL DEFAULT 'TEAM', "status" character varying NOT NULL DEFAULT 'ACTIVE', "external_sharing_allowed" boolean NULL DEFAULT false, "member_count" bigint NULL, "observed_at" timestamptz NOT NULL, "profile_hash" character varying NOT NULL DEFAULT '', "profile" jsonb NULL, "raw_profile_file_id" character varying NULL, "source_version" character varying NULL, "integration_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "directory_sync_run_directory_groups" character varying NULL, "integration_directory_groups" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "directory_groups_directory_sync_runs_directory_groups" FOREIGN KEY ("directory_sync_run_directory_groups") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_groups_directory_sync_runs_directory_sync_run" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("integration_directory_groups") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_groups_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "directorygroup_directory_sync_run_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_directory_sync_run_id_email" ON "directory_groups" ("directory_sync_run_id", "email");
-- create index "directorygroup_display_id_owner_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "owner_id");
-- create index "directorygroup_integration_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("integration_id", "email");
-- create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("integration_id", "external_id", "directory_sync_run_id");
-- create index "directorygroup_owner_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("owner_id", "email");
-- create "directory_memberships" table
CREATE TABLE "directory_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "role" character varying NULL DEFAULT 'MEMBER', "source" character varying NULL, "first_seen_at" timestamptz NULL, "last_seen_at" timestamptz NULL, "observed_at" timestamptz NOT NULL, "last_confirmed_run_id" character varying NULL, "metadata" jsonb NULL, "integration_id" character varying NOT NULL, "directory_sync_run_id" character varying NOT NULL, "directory_account_id" character varying NOT NULL, "directory_group_id" character varying NOT NULL, "directory_sync_run_directory_memberships" character varying NULL, "integration_directory_memberships" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "directory_memberships_directory_accounts_directory_account" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_memberships_directory_groups_directory_group" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_memberships_directory_sync_runs_directory_memberships" FOREIGN KEY ("directory_sync_run_directory_memberships") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_memberships_directory_sync_runs_directory_sync_run" FOREIGN KEY ("directory_sync_run_id") REFERENCES "directory_sync_runs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_memberships_integrations_directory_memberships" FOREIGN KEY ("integration_directory_memberships") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "directory_memberships_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "directory_memberships_organizations_directory_memberships" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" ON "directory_memberships" ("directory_account_id", "directory_group_id", "directory_sync_run_id");
-- create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory_account_id_directory_group_id" ON "directory_memberships" ("directory_account_id", "directory_group_id");
-- create index "directorymembership_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_directory_sync_run_id" ON "directory_memberships" ("directory_sync_run_id");
-- create index "directorymembership_display_id_owner_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_display_id_owner_id" ON "directory_memberships" ("display_id", "owner_id");
-- create index "directorymembership_integration_id_directory_sync_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_integration_id_directory_sync_run_id" ON "directory_memberships" ("integration_id", "directory_sync_run_id");
-- modify "events" table
ALTER TABLE "events" ADD COLUMN "directory_membership_events" character varying NULL, ADD CONSTRAINT "events_directory_memberships_events" FOREIGN KEY ("directory_membership_events") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "workflow_definitions" table
CREATE TABLE "workflow_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "workflow_kind" character varying NOT NULL, "schema_type" character varying NOT NULL, "revision" bigint NOT NULL DEFAULT 1, "draft" boolean NOT NULL DEFAULT true, "published_at" timestamptz NULL, "cooldown_seconds" bigint NOT NULL DEFAULT 0, "is_default" boolean NOT NULL DEFAULT false, "active" boolean NOT NULL DEFAULT true, "trigger_operations" jsonb NULL, "trigger_fields" jsonb NULL, "definition_json" jsonb NULL, "tracked_fields" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_definitions_organizations_workflow_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "workflowdefinition_display_id_owner_id" to table: "workflow_definitions"
CREATE UNIQUE INDEX "workflowdefinition_display_id_owner_id" ON "workflow_definitions" ("display_id", "owner_id");
-- create index "workflowdefinition_owner_id" to table: "workflow_definitions"
CREATE INDEX "workflowdefinition_owner_id" ON "workflow_definitions" ("owner_id") WHERE (deleted_at IS NULL);
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "workflow_definition_groups" character varying NULL, ADD CONSTRAINT "groups_workflow_definitions_groups" FOREIGN KEY ("workflow_definition_groups") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD COLUMN "workflow_definition_tag_definitions" character varying NULL, ADD CONSTRAINT "tag_definitions_workflow_definitions_tag_definitions" FOREIGN KEY ("workflow_definition_tag_definitions") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "workflow_instances" table
CREATE TABLE "workflow_instances" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "state" character varying NOT NULL DEFAULT 'RUNNING', "context" jsonb NULL, "last_evaluated_at" timestamptz NULL, "definition_snapshot" jsonb NULL, "owner_id" character varying NULL, "workflow_definition_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_instances_organizations_workflow_instances" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_instances_workflow_definitions_workflow_definition" FOREIGN KEY ("workflow_definition_id") REFERENCES "workflow_definitions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "workflowinstance_display_id_owner_id" to table: "workflow_instances"
CREATE UNIQUE INDEX "workflowinstance_display_id_owner_id" ON "workflow_instances" ("display_id", "owner_id");
-- create index "workflowinstance_owner_id" to table: "workflow_instances"
CREATE INDEX "workflowinstance_owner_id" ON "workflow_instances" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "workflowinstance_workflow_definition_id" to table: "workflow_instances"
CREATE INDEX "workflowinstance_workflow_definition_id" ON "workflow_instances" ("workflow_definition_id") WHERE (deleted_at IS NULL);
-- create "workflow_assignments" table
CREATE TABLE "workflow_assignments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "assignment_key" character varying NOT NULL, "role" character varying NOT NULL DEFAULT 'APPROVER', "label" character varying NULL, "required" boolean NOT NULL DEFAULT true, "status" character varying NOT NULL DEFAULT 'PENDING', "metadata" jsonb NULL, "decided_at" timestamptz NULL, "notes" text NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "actor_user_id" character varying NULL, "actor_group_id" character varying NULL, "workflow_instance_workflow_assignments" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_assignments_groups_group" FOREIGN KEY ("actor_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignments_organizations_workflow_assignments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignments_users_user" FOREIGN KEY ("actor_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignments_workflow_instances_workflow_assignments" FOREIGN KEY ("workflow_instance_workflow_assignments") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignments_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "workflowassignment_display_id_owner_id" to table: "workflow_assignments"
CREATE UNIQUE INDEX "workflowassignment_display_id_owner_id" ON "workflow_assignments" ("display_id", "owner_id");
-- create index "workflowassignment_owner_id" to table: "workflow_assignments"
CREATE INDEX "workflowassignment_owner_id" ON "workflow_assignments" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "workflowassignment_workflow_instance_id_assignment_key" to table: "workflow_assignments"
CREATE UNIQUE INDEX "workflowassignment_workflow_instance_id_assignment_key" ON "workflow_assignments" ("workflow_instance_id", "assignment_key");
-- create "workflow_assignment_targets" table
CREATE TABLE "workflow_assignment_targets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "target_type" character varying NOT NULL, "resolver_key" character varying NULL, "owner_id" character varying NULL, "workflow_assignment_workflow_assignment_targets" character varying NULL, "workflow_assignment_id" character varying NOT NULL, "target_user_id" character varying NULL, "target_group_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_assignment_targets_groups_group" FOREIGN KEY ("target_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignment_targets_or_8bb74468c70e1b9fcce1d5b038516f9a" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignment_targets_users_user" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignment_targets_wo_35919ebc89c62ef82cb5889ff40ce351" FOREIGN KEY ("workflow_assignment_workflow_assignment_targets") REFERENCES "workflow_assignments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_assignment_targets_wo_6077e6f4bf744947c345bb2733c1c240" FOREIGN KEY ("workflow_assignment_id") REFERENCES "workflow_assignments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "workflowassignmenttarget_display_id_owner_id" to table: "workflow_assignment_targets"
CREATE UNIQUE INDEX "workflowassignmenttarget_display_id_owner_id" ON "workflow_assignment_targets" ("display_id", "owner_id");
-- create index "workflowassignmenttarget_owner_id" to table: "workflow_assignment_targets"
CREATE INDEX "workflowassignmenttarget_owner_id" ON "workflow_assignment_targets" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" to table: "workflow_assignment_targets"
CREATE UNIQUE INDEX "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" ON "workflow_assignment_targets" ("workflow_assignment_id", "target_type", "target_user_id", "target_group_id", "resolver_key") WHERE (deleted_at IS NULL);
-- create index "workflowassignmenttarget_workflow_assignment_id" to table: "workflow_assignment_targets"
CREATE INDEX "workflowassignmenttarget_workflow_assignment_id" ON "workflow_assignment_targets" ("workflow_assignment_id") WHERE (deleted_at IS NULL);
-- create "workflow_events" table
CREATE TABLE "workflow_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "event_type" character varying NOT NULL, "payload" jsonb NULL, "owner_id" character varying NULL, "workflow_instance_id" character varying NOT NULL, "workflow_instance_workflow_events" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_events_organizations_workflow_events" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_events_workflow_instances_workflow_events" FOREIGN KEY ("workflow_instance_workflow_events") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_events_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "workflowevent_display_id_owner_id" to table: "workflow_events"
CREATE UNIQUE INDEX "workflowevent_display_id_owner_id" ON "workflow_events" ("display_id", "owner_id");
-- create index "workflowevent_owner_id" to table: "workflow_events"
CREATE INDEX "workflowevent_owner_id" ON "workflow_events" ("owner_id") WHERE (deleted_at IS NULL);
-- create "workflow_object_refs" table
CREATE TABLE "workflow_object_refs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "display_id" character varying NOT NULL, "owner_id" character varying NULL, "workflow_instance_workflow_object_refs" character varying NULL, "workflow_instance_id" character varying NOT NULL, "control_id" character varying NULL, "task_id" character varying NULL, "internal_policy_id" character varying NULL, "finding_id" character varying NULL, "directory_account_id" character varying NULL, "directory_group_id" character varying NULL, "directory_membership_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "workflow_object_refs_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_directory_accounts_directory_account" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_directory_groups_directory_group" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_directory_memberships_directory_membership" FOREIGN KEY ("directory_membership_id") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_organizations_workflow_object_refs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_tasks_task" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "workflow_object_refs_workflow_instances_workflow_instance" FOREIGN KEY ("workflow_instance_id") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "workflow_object_refs_workflow_instances_workflow_object_refs" FOREIGN KEY ("workflow_instance_workflow_object_refs") REFERENCES "workflow_instances" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "workflowobjectref_display_id_owner_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_display_id_owner_id" ON "workflow_object_refs" ("display_id", "owner_id");
-- create index "workflowobjectref_workflow_instance_id_control_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_control_id" ON "workflow_object_refs" ("workflow_instance_id", "control_id");
-- create index "workflowobjectref_workflow_instance_id_directory_account_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_account_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_account_id");
-- create index "workflowobjectref_workflow_instance_id_directory_group_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_group_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_group_id");
-- create index "workflowobjectref_workflow_instance_id_directory_membership_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_directory_membership_id" ON "workflow_object_refs" ("workflow_instance_id", "directory_membership_id");
-- create index "workflowobjectref_workflow_instance_id_finding_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_finding_id" ON "workflow_object_refs" ("workflow_instance_id", "finding_id");
-- create index "workflowobjectref_workflow_instance_id_internal_policy_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_internal_policy_id" ON "workflow_object_refs" ("workflow_instance_id", "internal_policy_id");
-- create index "workflowobjectref_workflow_instance_id_task_id" to table: "workflow_object_refs"
CREATE UNIQUE INDEX "workflowobjectref_workflow_instance_id_task_id" ON "workflow_object_refs" ("workflow_instance_id", "task_id");

-- +goose Down
-- reverse: create index "workflowobjectref_workflow_instance_id_task_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_task_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_internal_policy_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_internal_policy_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_finding_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_finding_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_directory_membership_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_directory_membership_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_directory_group_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_directory_group_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_directory_account_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_directory_account_id";
-- reverse: create index "workflowobjectref_workflow_instance_id_control_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_workflow_instance_id_control_id";
-- reverse: create index "workflowobjectref_display_id_owner_id" to table: "workflow_object_refs"
DROP INDEX "workflowobjectref_display_id_owner_id";
-- reverse: create "workflow_object_refs" table
DROP TABLE "workflow_object_refs";
-- reverse: create index "workflowevent_owner_id" to table: "workflow_events"
DROP INDEX "workflowevent_owner_id";
-- reverse: create index "workflowevent_display_id_owner_id" to table: "workflow_events"
DROP INDEX "workflowevent_display_id_owner_id";
-- reverse: create "workflow_events" table
DROP TABLE "workflow_events";
-- reverse: create index "workflowassignmenttarget_workflow_assignment_id" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_workflow_assignment_id";
-- reverse: create index "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_workf_699c5ebc2d2baaa6c7e976bd177928fc";
-- reverse: create index "workflowassignmenttarget_owner_id" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_owner_id";
-- reverse: create index "workflowassignmenttarget_display_id_owner_id" to table: "workflow_assignment_targets"
DROP INDEX "workflowassignmenttarget_display_id_owner_id";
-- reverse: create "workflow_assignment_targets" table
DROP TABLE "workflow_assignment_targets";
-- reverse: create index "workflowassignment_workflow_instance_id_assignment_key" to table: "workflow_assignments"
DROP INDEX "workflowassignment_workflow_instance_id_assignment_key";
-- reverse: create index "workflowassignment_owner_id" to table: "workflow_assignments"
DROP INDEX "workflowassignment_owner_id";
-- reverse: create index "workflowassignment_display_id_owner_id" to table: "workflow_assignments"
DROP INDEX "workflowassignment_display_id_owner_id";
-- reverse: create "workflow_assignments" table
DROP TABLE "workflow_assignments";
-- reverse: create index "workflowinstance_workflow_definition_id" to table: "workflow_instances"
DROP INDEX "workflowinstance_workflow_definition_id";
-- reverse: create index "workflowinstance_owner_id" to table: "workflow_instances"
DROP INDEX "workflowinstance_owner_id";
-- reverse: create index "workflowinstance_display_id_owner_id" to table: "workflow_instances"
DROP INDEX "workflowinstance_display_id_owner_id";
-- reverse: create "workflow_instances" table
DROP TABLE "workflow_instances";
-- reverse: modify "tag_definitions" table
ALTER TABLE "tag_definitions" DROP CONSTRAINT "tag_definitions_workflow_definitions_tag_definitions", DROP COLUMN "workflow_definition_tag_definitions";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_workflow_definitions_groups", DROP COLUMN "workflow_definition_groups";
-- reverse: create index "workflowdefinition_owner_id" to table: "workflow_definitions"
DROP INDEX "workflowdefinition_owner_id";
-- reverse: create index "workflowdefinition_display_id_owner_id" to table: "workflow_definitions"
DROP INDEX "workflowdefinition_display_id_owner_id";
-- reverse: create "workflow_definitions" table
DROP TABLE "workflow_definitions";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "events_directory_memberships_events", DROP COLUMN "directory_membership_events";
-- reverse: create index "directorymembership_integration_id_directory_sync_run_id" to table: "directory_memberships"
DROP INDEX "directorymembership_integration_id_directory_sync_run_id";
-- reverse: create index "directorymembership_display_id_owner_id" to table: "directory_memberships"
DROP INDEX "directorymembership_display_id_owner_id";
-- reverse: create index "directorymembership_directory_sync_run_id" to table: "directory_memberships"
DROP INDEX "directorymembership_directory_sync_run_id";
-- reverse: create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
DROP INDEX "directorymembership_directory_account_id_directory_group_id";
-- reverse: create index "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482" to table: "directory_memberships"
DROP INDEX "directorymembership_directory__c4933b3aba6a8094684cc9c233f43482";
-- reverse: create "directory_memberships" table
DROP TABLE "directory_memberships";
-- reverse: create index "directorygroup_owner_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_owner_id_email";
-- reverse: create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
DROP INDEX "directorygroup_integration_id_external_id_directory_sync_run_id";
-- reverse: create index "directorygroup_integration_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_integration_id_email";
-- reverse: create index "directorygroup_display_id_owner_id" to table: "directory_groups"
DROP INDEX "directorygroup_display_id_owner_id";
-- reverse: create index "directorygroup_directory_sync_run_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_directory_sync_run_id_email";
-- reverse: create "directory_groups" table
DROP TABLE "directory_groups";
-- reverse: create index "directoryaccount_owner_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_owner_id_canonical_email";
-- reverse: create index "directoryaccount_integration_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_integration_id_canonical_email";
-- reverse: create index "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d" to table: "directory_accounts"
DROP INDEX "directoryaccount_integration_i_fdd1dd4536589b023ef42f9092fecf7d";
-- reverse: create index "directoryaccount_display_id_owner_id" to table: "directory_accounts"
DROP INDEX "directoryaccount_display_id_owner_id";
-- reverse: create index "directoryaccount_directory_sync_run_id_canonical_email" to table: "directory_accounts"
DROP INDEX "directoryaccount_directory_sync_run_id_canonical_email";
-- reverse: create "directory_accounts" table
DROP TABLE "directory_accounts";
-- reverse: create index "directorysyncrun_integration_id_started_at" to table: "directory_sync_runs"
DROP INDEX "directorysyncrun_integration_id_started_at";
-- reverse: create index "directorysyncrun_display_id_owner_id" to table: "directory_sync_runs"
DROP INDEX "directorysyncrun_display_id_owner_id";
-- reverse: create "directory_sync_runs" table
DROP TABLE "directory_sync_runs";
-- reverse: create index "workflowassignmenttargethistory_history_time" to table: "workflow_assignment_target_history"
DROP INDEX "workflowassignmenttargethistory_history_time";
-- reverse: create "workflow_assignment_target_history" table
DROP TABLE "workflow_assignment_target_history";
-- reverse: create index "directorymembershiphistory_history_time" to table: "directory_membership_history"
DROP INDEX "directorymembershiphistory_history_time";
-- reverse: create "directory_membership_history" table
DROP TABLE "directory_membership_history";
-- reverse: create index "workflowdefinitionhistory_history_time" to table: "workflow_definition_history"
DROP INDEX "workflowdefinitionhistory_history_time";
-- reverse: create "workflow_definition_history" table
DROP TABLE "workflow_definition_history";
-- reverse: create index "directorygrouphistory_history_time" to table: "directory_group_history"
DROP INDEX "directorygrouphistory_history_time";
-- reverse: create "directory_group_history" table
DROP TABLE "directory_group_history";
-- reverse: create index "workfloweventhistory_history_time" to table: "workflow_event_history"
DROP INDEX "workfloweventhistory_history_time";
-- reverse: create "workflow_event_history" table
DROP TABLE "workflow_event_history";
-- reverse: create index "directoryaccounthistory_history_time" to table: "directory_account_history"
DROP INDEX "directoryaccounthistory_history_time";
-- reverse: create "directory_account_history" table
DROP TABLE "directory_account_history";
-- reverse: create index "workflowinstancehistory_history_time" to table: "workflow_instance_history"
DROP INDEX "workflowinstancehistory_history_time";
-- reverse: create "workflow_instance_history" table
DROP TABLE "workflow_instance_history";
-- reverse: create index "workflowobjectrefhistory_history_time" to table: "workflow_object_ref_history"
DROP INDEX "workflowobjectrefhistory_history_time";
-- reverse: create "workflow_object_ref_history" table
DROP TABLE "workflow_object_ref_history";
-- reverse: create index "workflowassignmenthistory_history_time" to table: "workflow_assignment_history"
DROP INDEX "workflowassignmenthistory_history_time";
-- reverse: create "workflow_assignment_history" table
DROP TABLE "workflow_assignment_history";
