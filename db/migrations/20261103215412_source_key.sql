-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "source_key" character varying NULL;
-- Create index "actionplan_source_key" to table: "action_plans"
CREATE INDEX "actionplan_source_key" ON "action_plans" ("source_key");
-- Modify "assessment_policies" table
ALTER TABLE "assessment_policies" ADD COLUMN "source_key" character varying NULL;
-- Create index "assessmentpolicy_source_key" to table: "assessment_policies"
CREATE INDEX "assessmentpolicy_source_key" ON "assessment_policies" ("source_key");
-- Modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "source_key" character varying NULL;
-- Create index "assessment_source_key" to table: "assessments"
CREATE INDEX "assessment_source_key" ON "assessments" ("source_key");
-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "source_key" character varying NULL;
-- Create index "asset_source_key" to table: "assets"
CREATE INDEX "asset_source_key" ON "assets" ("source_key");
-- Modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "source_key" character varying NULL;
-- Create index "campaign_source_key" to table: "campaigns"
CREATE INDEX "campaign_source_key" ON "campaigns" ("source_key");
-- Modify "check_results" table
ALTER TABLE "check_results" ADD COLUMN "source_key" character varying NULL;
-- Create index "checkresult_source_key" to table: "check_results"
CREATE INDEX "checkresult_source_key" ON "check_results" ("source_key");
-- Modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "source_key" character varying NULL;
-- Create index "contact_source_key" to table: "contacts"
CREATE INDEX "contact_source_key" ON "contacts" ("source_key");
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "source_key" character varying NULL;
-- Create index "controlimplementation_source_key" to table: "control_implementations"
CREATE INDEX "controlimplementation_source_key" ON "control_implementations" ("source_key");
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "source_key" character varying NULL;
-- Create index "controlobjective_source_key" to table: "control_objectives"
CREATE INDEX "controlobjective_source_key" ON "control_objectives" ("source_key");
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "source_key" character varying NULL;
-- Create index "control_source_key" to table: "controls"
CREATE INDEX "control_source_key" ON "controls" ("source_key");
-- Modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "source_key" character varying NULL;
-- Create index "customtypeenum_source_key" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_source_key" ON "custom_type_enums" ("source_key");
-- Modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "source_key" character varying NULL;
-- Create index "directoryaccount_source_key" to table: "directory_accounts"
CREATE INDEX "directoryaccount_source_key" ON "directory_accounts" ("source_key");
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "source_key" character varying NULL;
-- Create index "directorygroup_source_key" to table: "directory_groups"
CREATE INDEX "directorygroup_source_key" ON "directory_groups" ("source_key");
-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "source_key" character varying NULL;
-- Create index "directorymembership_source_key" to table: "directory_memberships"
CREATE INDEX "directorymembership_source_key" ON "directory_memberships" ("source_key");
-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "source_key" character varying NULL;
-- Create index "emailtemplate_source_key" to table: "email_templates"
CREATE INDEX "emailtemplate_source_key" ON "email_templates" ("source_key");
-- Modify "entities" table
ALTER TABLE "entities" ADD COLUMN "source_key" character varying NULL;
-- Create index "entity_source_key" to table: "entities"
CREATE INDEX "entity_source_key" ON "entities" ("source_key");
-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "source_key" character varying NULL;
-- Create index "evidence_source_key" to table: "evidences"
CREATE INDEX "evidence_source_key" ON "evidences" ("source_key");
-- Modify "findings" table
ALTER TABLE "findings" ADD COLUMN "source_key" character varying NULL;
-- Create index "finding_source_key" to table: "findings"
CREATE INDEX "finding_source_key" ON "findings" ("source_key");
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "source_key" character varying NULL;
-- Create index "group_source_key" to table: "groups"
CREATE INDEX "group_source_key" ON "groups" ("source_key");
-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "source_key" character varying NULL;
-- Create index "identityholder_source_key" to table: "identity_holders"
CREATE INDEX "identityholder_source_key" ON "identity_holders" ("source_key");
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "source_key" character varying NULL;
-- Create index "internalpolicy_source_key" to table: "internal_policies"
CREATE INDEX "internalpolicy_source_key" ON "internal_policies" ("source_key");
-- Modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "source_key" character varying NULL;
-- Create index "mappedcontrol_source_key" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_source_key" ON "mapped_controls" ("source_key");
-- Modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "source_key" character varying NULL;
-- Create index "narrative_source_key" to table: "narratives"
CREATE INDEX "narrative_source_key" ON "narratives" ("source_key");
-- Modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "source_key" character varying NULL;
-- Create index "platform_source_key" to table: "platforms"
CREATE INDEX "platform_source_key" ON "platforms" ("source_key");
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "source_key" character varying NULL;
-- Create index "procedure_source_key" to table: "procedures"
CREATE INDEX "procedure_source_key" ON "procedures" ("source_key");
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "source_key" character varying NULL;
-- Create index "program_source_key" to table: "programs"
CREATE INDEX "program_source_key" ON "programs" ("source_key");
-- Modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "source_key" character varying NULL;
-- Create index "remediation_source_key" to table: "remediations"
CREATE INDEX "remediation_source_key" ON "remediations" ("source_key");
-- Modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "source_key" character varying NULL;
-- Create index "review_source_key" to table: "reviews"
CREATE INDEX "review_source_key" ON "reviews" ("source_key");
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "source_key" character varying NULL;
-- Create index "risk_source_key" to table: "risks"
CREATE INDEX "risk_source_key" ON "risks" ("source_key");
-- Modify "scans" table
ALTER TABLE "scans" ADD COLUMN "source_key" character varying NULL;
-- Create index "scan_source_key" to table: "scans"
CREATE INDEX "scan_source_key" ON "scans" ("source_key");
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "source_key" character varying NULL;
-- Create index "subcontrol_source_key" to table: "subcontrols"
CREATE INDEX "subcontrol_source_key" ON "subcontrols" ("source_key");
-- Modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "source_key" character varying NULL;
-- Create index "systemdetail_source_key" to table: "system_details"
CREATE INDEX "systemdetail_source_key" ON "system_details" ("source_key");
-- Modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD COLUMN "source_key" character varying NULL;
-- Create index "tagdefinition_source_key" to table: "tag_definitions"
CREATE INDEX "tagdefinition_source_key" ON "tag_definitions" ("source_key");
-- Create index "task_source_key" to table: "tasks"
CREATE INDEX "task_source_key" ON "tasks" ("source_key");
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "source_key" character varying NULL;
-- Create index "template_source_key" to table: "templates"
CREATE INDEX "template_source_key" ON "templates" ("source_key");
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "source_key" character varying NULL;
-- Create index "vulnerability_source_key" to table: "vulnerabilities"
CREATE INDEX "vulnerability_source_key" ON "vulnerabilities" ("source_key");
