-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "title" character varying NOT NULL, ADD COLUMN "description" text NULL, ADD COLUMN "completed_at" timestamptz NULL, ADD COLUMN "requires_approval" boolean NOT NULL DEFAULT false, ADD COLUMN "blocked" boolean NOT NULL DEFAULT false, ADD COLUMN "blocker_reason" text NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "raw_payload" jsonb NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "title" character varying NOT NULL, ADD COLUMN "description" text NULL, ADD COLUMN "completed_at" timestamptz NULL, ADD COLUMN "requires_approval" boolean NOT NULL DEFAULT false, ADD COLUMN "blocked" boolean NOT NULL DEFAULT false, ADD COLUMN "blocker_reason" text NULL, ADD COLUMN "metadata" jsonb NULL, ADD COLUMN "raw_payload" jsonb NULL;
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "finding_assets" character varying NULL, ADD COLUMN "remediation_assets" character varying NULL, ADD COLUMN "review_assets" character varying NULL, ADD COLUMN "vulnerability_assets" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "remediation_controls" character varying NULL, ADD COLUMN "review_controls" character varying NULL, ADD COLUMN "vulnerability_controls" character varying NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "finding_entities" character varying NULL, ADD COLUMN "remediation_entities" character varying NULL, ADD COLUMN "review_entities" character varying NULL, ADD COLUMN "vulnerability_entities" character varying NULL;
-- modify "files" table
ALTER TABLE "files" ADD COLUMN "finding_files" character varying NULL, ADD COLUMN "remediation_files" character varying NULL, ADD COLUMN "review_files" character varying NULL, ADD COLUMN "vulnerability_files" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "finding_blocked_groups" character varying NULL, ADD COLUMN "finding_editors" character varying NULL, ADD COLUMN "finding_viewers" character varying NULL, ADD COLUMN "remediation_blocked_groups" character varying NULL, ADD COLUMN "remediation_editors" character varying NULL, ADD COLUMN "remediation_viewers" character varying NULL, ADD COLUMN "review_blocked_groups" character varying NULL, ADD COLUMN "review_editors" character varying NULL, ADD COLUMN "review_viewers" character varying NULL, ADD COLUMN "vulnerability_blocked_groups" character varying NULL, ADD COLUMN "vulnerability_editors" character varying NULL, ADD COLUMN "vulnerability_viewers" character varying NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "finding_comments" character varying NULL, ADD COLUMN "remediation_comments" character varying NULL, ADD COLUMN "review_comments" character varying NULL, ADD COLUMN "vulnerability_comments" character varying NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "finding_programs" character varying NULL, ADD COLUMN "remediation_programs" character varying NULL, ADD COLUMN "review_programs" character varying NULL, ADD COLUMN "vulnerability_programs" character varying NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "finding_risks" character varying NULL, ADD COLUMN "remediation_risks" character varying NULL, ADD COLUMN "review_risks" character varying NULL, ADD COLUMN "vulnerability_risks" character varying NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD COLUMN "finding_scans" character varying NULL, ADD COLUMN "vulnerability_scans" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "finding_subcontrols" character varying NULL, ADD COLUMN "remediation_subcontrols" character varying NULL, ADD COLUMN "review_subcontrols" character varying NULL, ADD COLUMN "vulnerability_subcontrols" character varying NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "finding_tasks" character varying NULL, ADD COLUMN "integration_tasks" character varying NULL, ADD COLUMN "remediation_tasks" character varying NULL, ADD COLUMN "review_tasks" character varying NULL, ADD COLUMN "vulnerability_tasks" character varying NULL;
-- create "findings" table
CREATE TABLE "findings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "source" character varying NULL, "resource_name" character varying NULL, "display_name" character varying NULL, "state" character varying NULL, "category" character varying NULL, "categories" jsonb NULL, "finding_class" character varying NULL, "severity" character varying NULL, "numeric_severity" double precision NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "open" boolean NULL DEFAULT true, "blocks_production" boolean NULL, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "assessment_id" character varying NULL, "description" text NULL, "recommendation" text NULL, "recommended_actions" text NULL, "references" jsonb NULL, "steps_to_reproduce" jsonb NULL, "targets" jsonb NULL, "target_details" jsonb NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "status" character varying NULL, "event_time" timestamptz NULL, "reported_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "owner_id" character varying NULL, "remediation_findings" character varying NULL, "review_findings" character varying NULL, "vulnerability_findings" character varying NULL, PRIMARY KEY ("id"));
-- create index "finding_display_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_display_id_owner_id" ON "findings" ("display_id", "owner_id");
-- create index "finding_external_id_external_owner_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_external_id_external_owner_id_owner_id" ON "findings" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "finding_owner_id" to table: "findings"
CREATE INDEX "finding_owner_id" ON "findings" ("owner_id") WHERE (deleted_at IS NULL);
-- create "finding_controls" table
CREATE TABLE "finding_controls" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "external_standard" character varying NULL, "external_standard_version" character varying NULL, "external_control_id" character varying NULL, "source" character varying NULL, "metadata" jsonb NULL, "discovered_at" timestamptz NULL, "finding_id" character varying NOT NULL, "control_id" character varying NOT NULL, "standard_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "findingcontrol_finding_id_control_id" to table: "finding_controls"
CREATE UNIQUE INDEX "findingcontrol_finding_id_control_id" ON "finding_controls" ("finding_id", "control_id");
-- create "finding_control_history" table
CREATE TABLE "finding_control_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "finding_id" character varying NOT NULL, "control_id" character varying NOT NULL, "standard_id" character varying NULL, "external_standard" character varying NULL, "external_standard_version" character varying NULL, "external_control_id" character varying NULL, "source" character varying NULL, "metadata" jsonb NULL, "discovered_at" timestamptz NULL, PRIMARY KEY ("id"));
-- create index "findingcontrolhistory_history_time" to table: "finding_control_history"
CREATE INDEX "findingcontrolhistory_history_time" ON "finding_control_history" ("history_time");
-- create "finding_history" table
CREATE TABLE "finding_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "source" character varying NULL, "resource_name" character varying NULL, "display_name" character varying NULL, "state" character varying NULL, "category" character varying NULL, "categories" jsonb NULL, "finding_class" character varying NULL, "severity" character varying NULL, "numeric_severity" double precision NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "open" boolean NULL DEFAULT true, "blocks_production" boolean NULL, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "assessment_id" character varying NULL, "description" text NULL, "recommendation" text NULL, "recommended_actions" text NULL, "references" jsonb NULL, "steps_to_reproduce" jsonb NULL, "targets" jsonb NULL, "target_details" jsonb NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "status" character varying NULL, "event_time" timestamptz NULL, "reported_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, PRIMARY KEY ("id"));
-- create index "findinghistory_history_time" to table: "finding_history"
CREATE INDEX "findinghistory_history_time" ON "finding_history" ("history_time");
-- create "remediations" table
CREATE TABLE "remediations" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NULL, "state" character varying NULL, "intent" character varying NULL, "summary" text NULL, "explanation" text NULL, "instructions" text NULL, "owner_reference" character varying NULL, "repository_uri" character varying NULL, "pull_request_uri" character varying NULL, "ticket_reference" character varying NULL, "due_at" timestamptz NULL, "completed_at" timestamptz NULL, "pr_generated_at" timestamptz NULL, "error" text NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "finding_remediations" character varying NULL, "owner_id" character varying NULL, "review_remediations" character varying NULL, "vulnerability_remediations" character varying NULL, PRIMARY KEY ("id"));
-- create index "remediation_display_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_display_id_owner_id" ON "remediations" ("display_id", "owner_id");
-- create index "remediation_external_id_external_owner_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_external_id_external_owner_id_owner_id" ON "remediations" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "remediation_owner_id" to table: "remediations"
CREATE INDEX "remediation_owner_id" ON "remediations" ("owner_id") WHERE (deleted_at IS NULL);
-- create "remediation_history" table
CREATE TABLE "remediation_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NULL, "state" character varying NULL, "intent" character varying NULL, "summary" text NULL, "explanation" text NULL, "instructions" text NULL, "owner_reference" character varying NULL, "repository_uri" character varying NULL, "pull_request_uri" character varying NULL, "ticket_reference" character varying NULL, "due_at" timestamptz NULL, "completed_at" timestamptz NULL, "pr_generated_at" timestamptz NULL, "error" text NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "remediationhistory_history_time" to table: "remediation_history"
CREATE INDEX "remediationhistory_history_time" ON "remediation_history" ("history_time");
-- create "reviews" table
CREATE TABLE "reviews" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NOT NULL, "state" character varying NULL, "category" character varying NULL, "classification" character varying NULL, "summary" text NULL, "details" text NULL, "reporter" character varying NULL, "approved" boolean NULL DEFAULT false, "reviewed_at" timestamptz NULL, "reported_at" timestamptz NULL, "approved_at" timestamptz NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "finding_reviews" character varying NULL, "owner_id" character varying NULL, "remediation_reviews" character varying NULL, "reviewer_id" character varying NULL, "vulnerability_reviews" character varying NULL, PRIMARY KEY ("id"));
-- create index "review_external_id_external_owner_id_owner_id" to table: "reviews"
CREATE UNIQUE INDEX "review_external_id_external_owner_id_owner_id" ON "reviews" ("external_id", "external_owner_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "review_owner_id" to table: "reviews"
CREATE INDEX "review_owner_id" ON "reviews" ("owner_id") WHERE (deleted_at IS NULL);
-- create "review_history" table
CREATE TABLE "review_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_id" character varying NULL, "external_owner_id" character varying NULL, "title" character varying NOT NULL, "state" character varying NULL, "category" character varying NULL, "classification" character varying NULL, "summary" text NULL, "details" text NULL, "reporter" character varying NULL, "approved" boolean NULL DEFAULT false, "reviewed_at" timestamptz NULL, "reported_at" timestamptz NULL, "approved_at" timestamptz NULL, "reviewer_id" character varying NULL, "source" character varying NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, PRIMARY KEY ("id"));
-- create index "reviewhistory_history_time" to table: "review_history"
CREATE INDEX "reviewhistory_history_time" ON "review_history" ("history_time");
-- create "vulnerabilities" table
CREATE TABLE "vulnerabilities" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_owner_id" character varying NULL, "external_id" character varying NOT NULL, "cve_id" character varying NULL, "source" character varying NULL, "display_name" character varying NULL, "category" character varying NULL, "severity" character varying NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "status" character varying NULL, "summary" text NULL, "description" text NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "open" boolean NULL DEFAULT true, "blocking" boolean NULL DEFAULT false, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "references" jsonb NULL, "impacts" jsonb NULL, "published_at" timestamptz NULL, "discovered_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, "finding_vulnerabilities" character varying NULL, "owner_id" character varying NULL, "remediation_vulnerabilities" character varying NULL, "review_vulnerabilities" character varying NULL, PRIMARY KEY ("id"));
-- create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_cve_id_owner_id" ON "vulnerabilities" ("cve_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "vulnerability_display_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_display_id_owner_id" ON "vulnerabilities" ("display_id", "owner_id");
-- create index "vulnerability_external_id_owner_id" to table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_external_id_owner_id" ON "vulnerabilities" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "vulnerability_owner_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_owner_id" ON "vulnerabilities" ("owner_id") WHERE (deleted_at IS NULL);
-- create "vulnerability_history" table
CREATE TABLE "vulnerability_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "internal_notes" character varying NULL, "system_internal_id" character varying NULL, "external_owner_id" character varying NULL, "external_id" character varying NOT NULL, "cve_id" character varying NULL, "source" character varying NULL, "display_name" character varying NULL, "category" character varying NULL, "severity" character varying NULL, "score" double precision NULL, "impact" double precision NULL, "exploitability" double precision NULL, "priority" character varying NULL, "status" character varying NULL, "summary" text NULL, "description" text NULL, "vector" character varying NULL, "remediation_sla" bigint NULL, "open" boolean NULL DEFAULT true, "blocking" boolean NULL DEFAULT false, "production" boolean NULL, "public" boolean NULL, "validated" boolean NULL, "references" jsonb NULL, "impacts" jsonb NULL, "published_at" timestamptz NULL, "discovered_at" timestamptz NULL, "source_updated_at" timestamptz NULL, "external_uri" character varying NULL, "metadata" jsonb NULL, "raw_payload" jsonb NULL, PRIMARY KEY ("id"));
-- create index "vulnerabilityhistory_history_time" to table: "vulnerability_history"
CREATE INDEX "vulnerabilityhistory_history_time" ON "vulnerability_history" ("history_time");
-- create "action_plan_tasks" table
CREATE TABLE "action_plan_tasks" ("action_plan_id" character varying NOT NULL, "task_id" character varying NOT NULL, PRIMARY KEY ("action_plan_id", "task_id"));
-- create "finding_action_plans" table
CREATE TABLE "finding_action_plans" ("finding_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "action_plan_id"));
-- create "integration_findings" table
CREATE TABLE "integration_findings" ("integration_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "finding_id"));
-- create "integration_vulnerabilities" table
CREATE TABLE "integration_vulnerabilities" ("integration_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "vulnerability_id"));
-- create "integration_reviews" table
CREATE TABLE "integration_reviews" ("integration_id" character varying NOT NULL, "review_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "review_id"));
-- create "integration_remediations" table
CREATE TABLE "integration_remediations" ("integration_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "remediation_id"));
-- create "integration_action_plans" table
CREATE TABLE "integration_action_plans" ("integration_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "action_plan_id"));
-- create "remediation_action_plans" table
CREATE TABLE "remediation_action_plans" ("remediation_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "action_plan_id"));
-- create "review_action_plans" table
CREATE TABLE "review_action_plans" ("review_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("review_id", "action_plan_id"));
-- create "vulnerability_action_plans" table
CREATE TABLE "vulnerability_action_plans" ("vulnerability_id" character varying NOT NULL, "action_plan_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "action_plan_id"));
-- modify "assets" table
ALTER TABLE "assets" ADD CONSTRAINT "assets_findings_assets" FOREIGN KEY ("finding_assets") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_remediations_assets" FOREIGN KEY ("remediation_assets") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_reviews_assets" FOREIGN KEY ("review_assets") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_vulnerabilities_assets" FOREIGN KEY ("vulnerability_assets") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD CONSTRAINT "controls_remediations_controls" FOREIGN KEY ("remediation_controls") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_reviews_controls" FOREIGN KEY ("review_controls") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "controls_vulnerabilities_controls" FOREIGN KEY ("vulnerability_controls") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_findings_entities" FOREIGN KEY ("finding_entities") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_remediations_entities" FOREIGN KEY ("remediation_entities") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_reviews_entities" FOREIGN KEY ("review_entities") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_vulnerabilities_entities" FOREIGN KEY ("vulnerability_entities") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "files" table
ALTER TABLE "files" ADD CONSTRAINT "files_findings_files" FOREIGN KEY ("finding_files") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_remediations_files" FOREIGN KEY ("remediation_files") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_reviews_files" FOREIGN KEY ("review_files") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "files_vulnerabilities_files" FOREIGN KEY ("vulnerability_files") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_findings_blocked_groups" FOREIGN KEY ("finding_blocked_groups") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_findings_editors" FOREIGN KEY ("finding_editors") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_findings_viewers" FOREIGN KEY ("finding_viewers") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_remediations_blocked_groups" FOREIGN KEY ("remediation_blocked_groups") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_remediations_editors" FOREIGN KEY ("remediation_editors") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_remediations_viewers" FOREIGN KEY ("remediation_viewers") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_reviews_blocked_groups" FOREIGN KEY ("review_blocked_groups") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_reviews_editors" FOREIGN KEY ("review_editors") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_reviews_viewers" FOREIGN KEY ("review_viewers") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_blocked_groups" FOREIGN KEY ("vulnerability_blocked_groups") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_editors" FOREIGN KEY ("vulnerability_editors") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_vulnerabilities_viewers" FOREIGN KEY ("vulnerability_viewers") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD CONSTRAINT "notes_findings_comments" FOREIGN KEY ("finding_comments") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_remediations_comments" FOREIGN KEY ("remediation_comments") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_reviews_comments" FOREIGN KEY ("review_comments") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_vulnerabilities_comments" FOREIGN KEY ("vulnerability_comments") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD CONSTRAINT "programs_findings_programs" FOREIGN KEY ("finding_programs") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_remediations_programs" FOREIGN KEY ("remediation_programs") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_reviews_programs" FOREIGN KEY ("review_programs") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "programs_vulnerabilities_programs" FOREIGN KEY ("vulnerability_programs") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "risks" table
ALTER TABLE "risks" ADD CONSTRAINT "risks_findings_risks" FOREIGN KEY ("finding_risks") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_remediations_risks" FOREIGN KEY ("remediation_risks") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_reviews_risks" FOREIGN KEY ("review_risks") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "risks_vulnerabilities_risks" FOREIGN KEY ("vulnerability_risks") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD CONSTRAINT "scans_findings_scans" FOREIGN KEY ("finding_scans") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_vulnerabilities_scans" FOREIGN KEY ("vulnerability_scans") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD CONSTRAINT "subcontrols_findings_subcontrols" FOREIGN KEY ("finding_subcontrols") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_remediations_subcontrols" FOREIGN KEY ("remediation_subcontrols") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_reviews_subcontrols" FOREIGN KEY ("review_subcontrols") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subcontrols_vulnerabilities_subcontrols" FOREIGN KEY ("vulnerability_subcontrols") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_findings_tasks" FOREIGN KEY ("finding_tasks") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_integrations_tasks" FOREIGN KEY ("integration_tasks") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_remediations_tasks" FOREIGN KEY ("remediation_tasks") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_reviews_tasks" FOREIGN KEY ("review_tasks") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_vulnerabilities_tasks" FOREIGN KEY ("vulnerability_tasks") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "findings" table
ALTER TABLE "findings" ADD CONSTRAINT "findings_organizations_findings" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_remediations_findings" FOREIGN KEY ("remediation_findings") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_reviews_findings" FOREIGN KEY ("review_findings") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_vulnerabilities_findings" FOREIGN KEY ("vulnerability_findings") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "finding_controls" table
ALTER TABLE "finding_controls" ADD CONSTRAINT "finding_controls_controls_control" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "finding_controls_findings_finding" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "finding_controls_standards_standard" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "remediations" table
ALTER TABLE "remediations" ADD CONSTRAINT "remediations_findings_remediations" FOREIGN KEY ("finding_remediations") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_organizations_remediations" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_reviews_remediations" FOREIGN KEY ("review_remediations") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "remediations_vulnerabilities_remediations" FOREIGN KEY ("vulnerability_remediations") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "reviews" table
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_findings_reviews" FOREIGN KEY ("finding_reviews") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_organizations_reviews" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_remediations_reviews" FOREIGN KEY ("remediation_reviews") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_users_reviewer" FOREIGN KEY ("reviewer_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "reviews_vulnerabilities_reviews" FOREIGN KEY ("vulnerability_reviews") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD CONSTRAINT "vulnerabilities_findings_vulnerabilities" FOREIGN KEY ("finding_vulnerabilities") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_organizations_vulnerabilities" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_remediations_vulnerabilities" FOREIGN KEY ("remediation_vulnerabilities") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_reviews_vulnerabilities" FOREIGN KEY ("review_vulnerabilities") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "action_plan_tasks" table
ALTER TABLE "action_plan_tasks" ADD CONSTRAINT "action_plan_tasks_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "action_plan_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "finding_action_plans" table
ALTER TABLE "finding_action_plans" ADD CONSTRAINT "finding_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "finding_action_plans_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_findings" table
ALTER TABLE "integration_findings" ADD CONSTRAINT "integration_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_findings_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_vulnerabilities" table
ALTER TABLE "integration_vulnerabilities" ADD CONSTRAINT "integration_vulnerabilities_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_reviews" table
ALTER TABLE "integration_reviews" ADD CONSTRAINT "integration_reviews_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_reviews_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_remediations" table
ALTER TABLE "integration_remediations" ADD CONSTRAINT "integration_remediations_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "integration_action_plans" table
ALTER TABLE "integration_action_plans" ADD CONSTRAINT "integration_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "integration_action_plans_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "remediation_action_plans" table
ALTER TABLE "remediation_action_plans" ADD CONSTRAINT "remediation_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "remediation_action_plans_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "review_action_plans" table
ALTER TABLE "review_action_plans" ADD CONSTRAINT "review_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "review_action_plans_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "vulnerability_action_plans" table
ALTER TABLE "vulnerability_action_plans" ADD CONSTRAINT "vulnerability_action_plans_action_plan_id" FOREIGN KEY ("action_plan_id") REFERENCES "action_plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "vulnerability_action_plans_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "vulnerability_action_plans" table
ALTER TABLE "vulnerability_action_plans" DROP CONSTRAINT "vulnerability_action_plans_vulnerability_id", DROP CONSTRAINT "vulnerability_action_plans_action_plan_id";
-- reverse: modify "review_action_plans" table
ALTER TABLE "review_action_plans" DROP CONSTRAINT "review_action_plans_review_id", DROP CONSTRAINT "review_action_plans_action_plan_id";
-- reverse: modify "remediation_action_plans" table
ALTER TABLE "remediation_action_plans" DROP CONSTRAINT "remediation_action_plans_remediation_id", DROP CONSTRAINT "remediation_action_plans_action_plan_id";
-- reverse: modify "integration_action_plans" table
ALTER TABLE "integration_action_plans" DROP CONSTRAINT "integration_action_plans_integration_id", DROP CONSTRAINT "integration_action_plans_action_plan_id";
-- reverse: modify "integration_remediations" table
ALTER TABLE "integration_remediations" DROP CONSTRAINT "integration_remediations_remediation_id", DROP CONSTRAINT "integration_remediations_integration_id";
-- reverse: modify "integration_reviews" table
ALTER TABLE "integration_reviews" DROP CONSTRAINT "integration_reviews_review_id", DROP CONSTRAINT "integration_reviews_integration_id";
-- reverse: modify "integration_vulnerabilities" table
ALTER TABLE "integration_vulnerabilities" DROP CONSTRAINT "integration_vulnerabilities_vulnerability_id", DROP CONSTRAINT "integration_vulnerabilities_integration_id";
-- reverse: modify "integration_findings" table
ALTER TABLE "integration_findings" DROP CONSTRAINT "integration_findings_integration_id", DROP CONSTRAINT "integration_findings_finding_id";
-- reverse: modify "finding_action_plans" table
ALTER TABLE "finding_action_plans" DROP CONSTRAINT "finding_action_plans_finding_id", DROP CONSTRAINT "finding_action_plans_action_plan_id";
-- reverse: modify "action_plan_tasks" table
ALTER TABLE "action_plan_tasks" DROP CONSTRAINT "action_plan_tasks_task_id", DROP CONSTRAINT "action_plan_tasks_action_plan_id";
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP CONSTRAINT "vulnerabilities_reviews_vulnerabilities", DROP CONSTRAINT "vulnerabilities_remediations_vulnerabilities", DROP CONSTRAINT "vulnerabilities_organizations_vulnerabilities", DROP CONSTRAINT "vulnerabilities_findings_vulnerabilities";
-- reverse: modify "reviews" table
ALTER TABLE "reviews" DROP CONSTRAINT "reviews_vulnerabilities_reviews", DROP CONSTRAINT "reviews_users_reviewer", DROP CONSTRAINT "reviews_remediations_reviews", DROP CONSTRAINT "reviews_organizations_reviews", DROP CONSTRAINT "reviews_findings_reviews";
-- reverse: modify "remediations" table
ALTER TABLE "remediations" DROP CONSTRAINT "remediations_vulnerabilities_remediations", DROP CONSTRAINT "remediations_reviews_remediations", DROP CONSTRAINT "remediations_organizations_remediations", DROP CONSTRAINT "remediations_findings_remediations";
-- reverse: modify "finding_controls" table
ALTER TABLE "finding_controls" DROP CONSTRAINT "finding_controls_standards_standard", DROP CONSTRAINT "finding_controls_findings_finding", DROP CONSTRAINT "finding_controls_controls_control";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP CONSTRAINT "findings_vulnerabilities_findings", DROP CONSTRAINT "findings_reviews_findings", DROP CONSTRAINT "findings_remediations_findings", DROP CONSTRAINT "findings_organizations_findings";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_vulnerabilities_tasks", DROP CONSTRAINT "tasks_reviews_tasks", DROP CONSTRAINT "tasks_remediations_tasks", DROP CONSTRAINT "tasks_integrations_tasks", DROP CONSTRAINT "tasks_findings_tasks";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_vulnerabilities_subcontrols", DROP CONSTRAINT "subcontrols_reviews_subcontrols", DROP CONSTRAINT "subcontrols_remediations_subcontrols", DROP CONSTRAINT "subcontrols_findings_subcontrols";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP CONSTRAINT "scans_vulnerabilities_scans", DROP CONSTRAINT "scans_findings_scans";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_vulnerabilities_risks", DROP CONSTRAINT "risks_reviews_risks", DROP CONSTRAINT "risks_remediations_risks", DROP CONSTRAINT "risks_findings_risks";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_vulnerabilities_programs", DROP CONSTRAINT "programs_reviews_programs", DROP CONSTRAINT "programs_remediations_programs", DROP CONSTRAINT "programs_findings_programs";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_vulnerabilities_comments", DROP CONSTRAINT "notes_reviews_comments", DROP CONSTRAINT "notes_remediations_comments", DROP CONSTRAINT "notes_findings_comments";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_vulnerabilities_viewers", DROP CONSTRAINT "groups_vulnerabilities_editors", DROP CONSTRAINT "groups_vulnerabilities_blocked_groups", DROP CONSTRAINT "groups_reviews_viewers", DROP CONSTRAINT "groups_reviews_editors", DROP CONSTRAINT "groups_reviews_blocked_groups", DROP CONSTRAINT "groups_remediations_viewers", DROP CONSTRAINT "groups_remediations_editors", DROP CONSTRAINT "groups_remediations_blocked_groups", DROP CONSTRAINT "groups_findings_viewers", DROP CONSTRAINT "groups_findings_editors", DROP CONSTRAINT "groups_findings_blocked_groups";
-- reverse: modify "files" table
ALTER TABLE "files" DROP CONSTRAINT "files_vulnerabilities_files", DROP CONSTRAINT "files_reviews_files", DROP CONSTRAINT "files_remediations_files", DROP CONSTRAINT "files_findings_files";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_vulnerabilities_entities", DROP CONSTRAINT "entities_reviews_entities", DROP CONSTRAINT "entities_remediations_entities", DROP CONSTRAINT "entities_findings_entities";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_vulnerabilities_controls", DROP CONSTRAINT "controls_reviews_controls", DROP CONSTRAINT "controls_remediations_controls";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP CONSTRAINT "assets_vulnerabilities_assets", DROP CONSTRAINT "assets_reviews_assets", DROP CONSTRAINT "assets_remediations_assets", DROP CONSTRAINT "assets_findings_assets";
-- reverse: create "vulnerability_action_plans" table
DROP TABLE "vulnerability_action_plans";
-- reverse: create "review_action_plans" table
DROP TABLE "review_action_plans";
-- reverse: create "remediation_action_plans" table
DROP TABLE "remediation_action_plans";
-- reverse: create "integration_action_plans" table
DROP TABLE "integration_action_plans";
-- reverse: create "integration_remediations" table
DROP TABLE "integration_remediations";
-- reverse: create "integration_reviews" table
DROP TABLE "integration_reviews";
-- reverse: create "integration_vulnerabilities" table
DROP TABLE "integration_vulnerabilities";
-- reverse: create "integration_findings" table
DROP TABLE "integration_findings";
-- reverse: create "finding_action_plans" table
DROP TABLE "finding_action_plans";
-- reverse: create "action_plan_tasks" table
DROP TABLE "action_plan_tasks";
-- reverse: create index "vulnerabilityhistory_history_time" to table: "vulnerability_history"
DROP INDEX "vulnerabilityhistory_history_time";
-- reverse: create "vulnerability_history" table
DROP TABLE "vulnerability_history";
-- reverse: create index "vulnerability_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_owner_id";
-- reverse: create index "vulnerability_external_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_external_id_owner_id";
-- reverse: create index "vulnerability_display_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_display_id_owner_id";
-- reverse: create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_cve_id_owner_id";
-- reverse: create "vulnerabilities" table
DROP TABLE "vulnerabilities";
-- reverse: create index "reviewhistory_history_time" to table: "review_history"
DROP INDEX "reviewhistory_history_time";
-- reverse: create "review_history" table
DROP TABLE "review_history";
-- reverse: create index "review_owner_id" to table: "reviews"
DROP INDEX "review_owner_id";
-- reverse: create index "review_external_id_external_owner_id_owner_id" to table: "reviews"
DROP INDEX "review_external_id_external_owner_id_owner_id";
-- reverse: create "reviews" table
DROP TABLE "reviews";
-- reverse: create index "remediationhistory_history_time" to table: "remediation_history"
DROP INDEX "remediationhistory_history_time";
-- reverse: create "remediation_history" table
DROP TABLE "remediation_history";
-- reverse: create index "remediation_owner_id" to table: "remediations"
DROP INDEX "remediation_owner_id";
-- reverse: create index "remediation_external_id_external_owner_id_owner_id" to table: "remediations"
DROP INDEX "remediation_external_id_external_owner_id_owner_id";
-- reverse: create index "remediation_display_id_owner_id" to table: "remediations"
DROP INDEX "remediation_display_id_owner_id";
-- reverse: create "remediations" table
DROP TABLE "remediations";
-- reverse: create index "findinghistory_history_time" to table: "finding_history"
DROP INDEX "findinghistory_history_time";
-- reverse: create "finding_history" table
DROP TABLE "finding_history";
-- reverse: create index "findingcontrolhistory_history_time" to table: "finding_control_history"
DROP INDEX "findingcontrolhistory_history_time";
-- reverse: create "finding_control_history" table
DROP TABLE "finding_control_history";
-- reverse: create index "findingcontrol_finding_id_control_id" to table: "finding_controls"
DROP INDEX "findingcontrol_finding_id_control_id";
-- reverse: create "finding_controls" table
DROP TABLE "finding_controls";
-- reverse: create index "finding_owner_id" to table: "findings"
DROP INDEX "finding_owner_id";
-- reverse: create index "finding_external_id_external_owner_id_owner_id" to table: "findings"
DROP INDEX "finding_external_id_external_owner_id_owner_id";
-- reverse: create index "finding_display_id_owner_id" to table: "findings"
DROP INDEX "finding_display_id_owner_id";
-- reverse: create "findings" table
DROP TABLE "findings";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "vulnerability_tasks", DROP COLUMN "review_tasks", DROP COLUMN "remediation_tasks", DROP COLUMN "integration_tasks", DROP COLUMN "finding_tasks";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "vulnerability_subcontrols", DROP COLUMN "review_subcontrols", DROP COLUMN "remediation_subcontrols", DROP COLUMN "finding_subcontrols";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP COLUMN "vulnerability_scans", DROP COLUMN "finding_scans";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "vulnerability_risks", DROP COLUMN "review_risks", DROP COLUMN "remediation_risks", DROP COLUMN "finding_risks";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "vulnerability_programs", DROP COLUMN "review_programs", DROP COLUMN "remediation_programs", DROP COLUMN "finding_programs";
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP COLUMN "vulnerability_comments", DROP COLUMN "review_comments", DROP COLUMN "remediation_comments", DROP COLUMN "finding_comments";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "vulnerability_viewers", DROP COLUMN "vulnerability_editors", DROP COLUMN "vulnerability_blocked_groups", DROP COLUMN "review_viewers", DROP COLUMN "review_editors", DROP COLUMN "review_blocked_groups", DROP COLUMN "remediation_viewers", DROP COLUMN "remediation_editors", DROP COLUMN "remediation_blocked_groups", DROP COLUMN "finding_viewers", DROP COLUMN "finding_editors", DROP COLUMN "finding_blocked_groups";
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "vulnerability_files", DROP COLUMN "review_files", DROP COLUMN "remediation_files", DROP COLUMN "finding_files";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "vulnerability_entities", DROP COLUMN "review_entities", DROP COLUMN "remediation_entities", DROP COLUMN "finding_entities";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "vulnerability_controls", DROP COLUMN "review_controls", DROP COLUMN "remediation_controls";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP COLUMN "vulnerability_assets", DROP COLUMN "review_assets", DROP COLUMN "remediation_assets", DROP COLUMN "finding_assets";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "raw_payload", DROP COLUMN "metadata", DROP COLUMN "blocker_reason", DROP COLUMN "blocked", DROP COLUMN "requires_approval", DROP COLUMN "completed_at", DROP COLUMN "description", DROP COLUMN "title";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "raw_payload", DROP COLUMN "metadata", DROP COLUMN "blocker_reason", DROP COLUMN "blocked", DROP COLUMN "requires_approval", DROP COLUMN "completed_at", DROP COLUMN "description", DROP COLUMN "title";
