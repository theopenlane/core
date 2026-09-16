-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "actionplan_external_file_id_owner_id" to table: "action_plans"
CREATE INDEX "actionplan_external_file_id_owner_id" ON "action_plans" ("external_file_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "actionplan_integration_run_id" to table: "action_plans"
CREATE INDEX "actionplan_integration_run_id" ON "action_plans" ("integration_run_id");
-- Create index "actionplan_source_instance_id" to table: "action_plans"
CREATE INDEX "actionplan_source_instance_id" ON "action_plans" ("source_instance_id");
-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "organization_directory_sync_run_creators";
-- Modify "integration_runs" table
ALTER TABLE "integration_runs" DROP COLUMN "mapping_version", DROP COLUMN "request_file_id", DROP COLUMN "response_file_id", DROP COLUMN "event_id", DROP COLUMN "assessment_response_id";
-- Create "action_plan_integration_runs" table
CREATE TABLE "action_plan_integration_runs" ("action_plan_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "integration_run_id"), CONSTRAINT "action_plan_integration_runs_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "action_plan_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "action_plan_integration_runs_integration_run_id_idx" to table: "action_plan_integration_runs"
CREATE INDEX "action_plan_integration_runs_integration_run_id_idx" ON "action_plan_integration_runs" ("integration_run_id");
-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "asset_integration_run_id" to table: "assets"
CREATE INDEX "asset_integration_run_id" ON "assets" ("integration_run_id");
-- Create index "asset_source_identifier_owner_id" to table: "assets"
CREATE INDEX "asset_source_identifier_owner_id" ON "assets" ("source_identifier", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "asset_source_instance_id" to table: "assets"
CREATE INDEX "asset_source_instance_id" ON "assets" ("source_instance_id");
-- Create "asset_integration_runs" table
CREATE TABLE "asset_integration_runs" ("asset_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("asset_id", "integration_run_id"), CONSTRAINT "asset_integration_runs_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "asset_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "asset_integration_runs_integration_run_id_idx" to table: "asset_integration_runs"
CREATE INDEX "asset_integration_runs_integration_run_id_idx" ON "asset_integration_runs" ("integration_run_id");
-- Modify "check_results" table
ALTER TABLE "check_results" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "checkresult_integration_run_id" to table: "check_results"
CREATE INDEX "checkresult_integration_run_id" ON "check_results" ("integration_run_id");
-- Create index "checkresult_source_instance_id" to table: "check_results"
CREATE INDEX "checkresult_source_instance_id" ON "check_results" ("source_instance_id");
-- Create "check_result_integration_runs" table
CREATE TABLE "check_result_integration_runs" ("check_result_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("check_result_id", "integration_run_id"), CONSTRAINT "check_result_integration_runs_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "check_result_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "check_result_integration_runs_integration_run_id_idx" to table: "check_result_integration_runs"
CREATE INDEX "check_result_integration_runs_integration_run_id_idx" ON "check_result_integration_runs" ("integration_run_id");
-- Modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "contact_integration_run_id" to table: "contacts"
CREATE INDEX "contact_integration_run_id" ON "contacts" ("integration_run_id");
-- Create index "contact_source_instance_id" to table: "contacts"
CREATE INDEX "contact_source_instance_id" ON "contacts" ("source_instance_id");
-- Create "contact_integration_runs" table
CREATE TABLE "contact_integration_runs" ("contact_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "integration_run_id"), CONSTRAINT "contact_integration_runs_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "contact_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "contact_integration_runs_integration_run_id_idx" to table: "contact_integration_runs"
CREATE INDEX "contact_integration_runs_integration_run_id_idx" ON "contact_integration_runs" ("integration_run_id");
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" DROP COLUMN "directory_instance_id", DROP COLUMN "last_login_at", DROP COLUMN "first_seen_at", DROP COLUMN "last_seen_at", DROP COLUMN "profile_hash", DROP COLUMN "directory_sync_run_id", ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "directoryaccount_integration_run_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_integration_run_id" ON "directory_accounts" ("integration_run_id");
-- Create index "directoryaccount_source_instance_id" to table: "directory_accounts"
CREATE INDEX "directoryaccount_source_instance_id" ON "directory_accounts" ("source_instance_id");
-- Create "directory_account_integration_runs" table
CREATE TABLE "directory_account_integration_runs" ("directory_account_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("directory_account_id", "integration_run_id"), CONSTRAINT "directory_account_integration_runs_directory_account_id" FOREIGN KEY ("directory_account_id") REFERENCES "directory_accounts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "directory_account_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "directory_account_integration_runs_integration_run_id_idx" to table: "directory_account_integration_runs"
CREATE INDEX "directory_account_integration_runs_integration_run_id_idx" ON "directory_account_integration_runs" ("integration_run_id");
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" DROP COLUMN "directory_instance_id", DROP COLUMN "first_seen_at", DROP COLUMN "last_seen_at", DROP COLUMN "profile_hash", DROP COLUMN "directory_sync_run_id", ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "directorygroup_integration_run_id" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_run_id" ON "directory_groups" ("integration_run_id");
-- Create index "directorygroup_source_instance_id" to table: "directory_groups"
CREATE INDEX "directorygroup_source_instance_id" ON "directory_groups" ("source_instance_id");
-- Create "directory_group_integration_runs" table
CREATE TABLE "directory_group_integration_runs" ("directory_group_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("directory_group_id", "integration_run_id"), CONSTRAINT "directory_group_integration_runs_directory_group_id" FOREIGN KEY ("directory_group_id") REFERENCES "directory_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "directory_group_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "directory_group_integration_runs_integration_run_id_idx" to table: "directory_group_integration_runs"
CREATE INDEX "directory_group_integration_runs_integration_run_id_idx" ON "directory_group_integration_runs" ("integration_run_id");
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" DROP COLUMN "directory_instance_id", DROP COLUMN "first_seen_at", DROP COLUMN "last_seen_at", DROP COLUMN "last_confirmed_run_id", DROP COLUMN "directory_sync_run_id", ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "directory_membership_integration_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_integration_id_idx" ON "directory_memberships" ("integration_id");
-- Create index "directory_membership_platform_id_idx" to table: "directory_memberships"
CREATE INDEX "directory_membership_platform_id_idx" ON "directory_memberships" ("platform_id");
-- Create index "directorymembership_integration_run_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_integration_run_id" ON "directory_memberships" ("integration_run_id");
-- Create index "directorymembership_source_instance_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_source_instance_id" ON "directory_memberships" ("source_instance_id");
-- Create "directory_membership_integration_runs" table
CREATE TABLE "directory_membership_integration_runs" ("directory_membership_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("directory_membership_id", "integration_run_id"), CONSTRAINT "directory_membership_integration_runs_directory_membership_id" FOREIGN KEY ("directory_membership_id") REFERENCES "directory_memberships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "directory_membership_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "directory_membership_integration_runs_integration_run_id_idx" to table: "directory_membership_integration_runs"
CREATE INDEX "directory_membership_integration_runs_integration_run_id_idx" ON "directory_membership_integration_runs" ("integration_run_id");
-- Modify "entities" table
ALTER TABLE "entities" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "entity_external_id_owner_id" to table: "entities"
CREATE INDEX "entity_external_id_owner_id" ON "entities" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "entity_integration_run_id" to table: "entities"
CREATE INDEX "entity_integration_run_id" ON "entities" ("integration_run_id");
-- Create index "entity_source_instance_id" to table: "entities"
CREATE INDEX "entity_source_instance_id" ON "entities" ("source_instance_id");
-- Create "entity_integration_runs" table
CREATE TABLE "entity_integration_runs" ("entity_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "integration_run_id"), CONSTRAINT "entity_integration_runs_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "entity_integration_runs_integration_run_id_idx" to table: "entity_integration_runs"
CREATE INDEX "entity_integration_runs_integration_run_id_idx" ON "entity_integration_runs" ("integration_run_id");
-- Modify "findings" table
ALTER TABLE "findings" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "finding_integration_run_id" to table: "findings"
CREATE INDEX "finding_integration_run_id" ON "findings" ("integration_run_id");
-- Create index "finding_source_instance_id" to table: "findings"
CREATE INDEX "finding_source_instance_id" ON "findings" ("source_instance_id");
-- Create "finding_integration_runs" table
CREATE TABLE "finding_integration_runs" ("finding_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "integration_run_id"), CONSTRAINT "finding_integration_runs_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "finding_integration_runs_integration_run_id_idx" to table: "finding_integration_runs"
CREATE INDEX "finding_integration_runs_integration_run_id_idx" ON "finding_integration_runs" ("integration_run_id");
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "internalpolicy_external_file_id_owner_id" to table: "internal_policies"
CREATE INDEX "internalpolicy_external_file_id_owner_id" ON "internal_policies" ("external_file_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "internalpolicy_integration_run_id" to table: "internal_policies"
CREATE INDEX "internalpolicy_integration_run_id" ON "internal_policies" ("integration_run_id");
-- Create index "internalpolicy_source_instance_id" to table: "internal_policies"
CREATE INDEX "internalpolicy_source_instance_id" ON "internal_policies" ("source_instance_id");
-- Create "internal_policy_integration_runs" table
CREATE TABLE "internal_policy_integration_runs" ("internal_policy_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "integration_run_id"), CONSTRAINT "internal_policy_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_integration_runs_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "internal_policy_integration_runs_integration_run_id_idx" to table: "internal_policy_integration_runs"
CREATE INDEX "internal_policy_integration_runs_integration_run_id_idx" ON "internal_policy_integration_runs" ("integration_run_id");
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "procedure_external_file_id_owner_id" to table: "procedures"
CREATE INDEX "procedure_external_file_id_owner_id" ON "procedures" ("external_file_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "procedure_integration_run_id" to table: "procedures"
CREATE INDEX "procedure_integration_run_id" ON "procedures" ("integration_run_id");
-- Create index "procedure_source_instance_id" to table: "procedures"
CREATE INDEX "procedure_source_instance_id" ON "procedures" ("source_instance_id");
-- Create "procedure_integration_runs" table
CREATE TABLE "procedure_integration_runs" ("procedure_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "integration_run_id"), CONSTRAINT "procedure_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "procedure_integration_runs_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "procedure_integration_runs_integration_run_id_idx" to table: "procedure_integration_runs"
CREATE INDEX "procedure_integration_runs_integration_run_id_idx" ON "procedure_integration_runs" ("integration_run_id");
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "risk_external_id_owner_id" to table: "risks"
CREATE INDEX "risk_external_id_owner_id" ON "risks" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "risk_integration_run_id" to table: "risks"
CREATE INDEX "risk_integration_run_id" ON "risks" ("integration_run_id");
-- Create index "risk_source_instance_id" to table: "risks"
CREATE INDEX "risk_source_instance_id" ON "risks" ("source_instance_id");
-- Create "risk_integration_runs" table
CREATE TABLE "risk_integration_runs" ("risk_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "integration_run_id"), CONSTRAINT "risk_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "risk_integration_runs_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "risk_integration_runs_integration_run_id_idx" to table: "risk_integration_runs"
CREATE INDEX "risk_integration_runs_integration_run_id_idx" ON "risk_integration_runs" ("integration_run_id");
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "vulnerability_integration_run_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_integration_run_id" ON "vulnerabilities" ("integration_run_id");
-- Create index "vulnerability_source_instance_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_source_instance_id" ON "vulnerabilities" ("source_instance_id");
-- Create "vulnerability_integration_runs" table
CREATE TABLE "vulnerability_integration_runs" ("vulnerability_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "integration_run_id"), CONSTRAINT "vulnerability_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_integration_runs_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "vulnerability_integration_runs_integration_run_id_idx" to table: "vulnerability_integration_runs"
CREATE INDEX "vulnerability_integration_runs_integration_run_id_idx" ON "vulnerability_integration_runs" ("integration_run_id");
