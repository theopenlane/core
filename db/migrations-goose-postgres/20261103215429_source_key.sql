-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "source_key" character varying NULL;
-- create index "actionplan_source_key" to table: "action_plans"
CREATE INDEX "actionplan_source_key" ON "action_plans" ("source_key");
-- modify "assessment_policies" table
ALTER TABLE "assessment_policies" ADD COLUMN "source_key" character varying NULL;
-- create index "assessmentpolicy_source_key" to table: "assessment_policies"
CREATE INDEX "assessmentpolicy_source_key" ON "assessment_policies" ("source_key");
-- modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "source_key" character varying NULL;
-- create index "assessment_source_key" to table: "assessments"
CREATE INDEX "assessment_source_key" ON "assessments" ("source_key");
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "source_key" character varying NULL;
-- create index "asset_source_key" to table: "assets"
CREATE INDEX "asset_source_key" ON "assets" ("source_key");
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "source_key" character varying NULL;
-- create index "campaign_source_key" to table: "campaigns"
CREATE INDEX "campaign_source_key" ON "campaigns" ("source_key");
-- modify "check_results" table
ALTER TABLE "check_results" ADD COLUMN "source_key" character varying NULL;
-- create index "checkresult_source_key" to table: "check_results"
CREATE INDEX "checkresult_source_key" ON "check_results" ("source_key");
-- modify "contacts" table
ALTER TABLE "contacts" ADD COLUMN "source_key" character varying NULL;
-- create index "contact_source_key" to table: "contacts"
CREATE INDEX "contact_source_key" ON "contacts" ("source_key");
-- modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "source_key" character varying NULL;
-- create index "controlimplementation_source_key" to table: "control_implementations"
CREATE INDEX "controlimplementation_source_key" ON "control_implementations" ("source_key");
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "source_key" character varying NULL;
-- create index "controlobjective_source_key" to table: "control_objectives"
CREATE INDEX "controlobjective_source_key" ON "control_objectives" ("source_key");
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "source_key" character varying NULL;
-- create index "control_source_key" to table: "controls"
CREATE INDEX "control_source_key" ON "controls" ("source_key");
-- modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "source_key" character varying NULL;
-- create index "customtypeenum_source_key" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_source_key" ON "custom_type_enums" ("source_key");
-- modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "source_key" character varying NULL;
-- create index "directoryaccount_source_key" to table: "directory_accounts"
CREATE INDEX "directoryaccount_source_key" ON "directory_accounts" ("source_key");
-- modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "source_key" character varying NULL;
-- create index "directorygroup_source_key" to table: "directory_groups"
CREATE INDEX "directorygroup_source_key" ON "directory_groups" ("source_key");
-- modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "source_key" character varying NULL;
-- create index "directorymembership_source_key" to table: "directory_memberships"
CREATE INDEX "directorymembership_source_key" ON "directory_memberships" ("source_key");
-- modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "source_key" character varying NULL;
-- create index "emailtemplate_source_key" to table: "email_templates"
CREATE INDEX "emailtemplate_source_key" ON "email_templates" ("source_key");
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "source_key" character varying NULL;
-- create index "entity_source_key" to table: "entities"
CREATE INDEX "entity_source_key" ON "entities" ("source_key");
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "source_key" character varying NULL;
-- create index "evidence_source_key" to table: "evidences"
CREATE INDEX "evidence_source_key" ON "evidences" ("source_key");
-- modify "findings" table
ALTER TABLE "findings" ADD COLUMN "source_key" character varying NULL;
-- create index "finding_source_key" to table: "findings"
CREATE INDEX "finding_source_key" ON "findings" ("source_key");
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "source_key" character varying NULL;
-- create index "group_source_key" to table: "groups"
CREATE INDEX "group_source_key" ON "groups" ("source_key");
-- modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "source_key" character varying NULL;
-- create index "identityholder_source_key" to table: "identity_holders"
CREATE INDEX "identityholder_source_key" ON "identity_holders" ("source_key");
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "source_key" character varying NULL;
-- create index "internalpolicy_source_key" to table: "internal_policies"
CREATE INDEX "internalpolicy_source_key" ON "internal_policies" ("source_key");
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "source_key" character varying NULL;
-- create index "mappedcontrol_source_key" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_source_key" ON "mapped_controls" ("source_key");
-- modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "source_key" character varying NULL;
-- create index "narrative_source_key" to table: "narratives"
CREATE INDEX "narrative_source_key" ON "narratives" ("source_key");
-- modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "source_key" character varying NULL;
-- create index "platform_source_key" to table: "platforms"
CREATE INDEX "platform_source_key" ON "platforms" ("source_key");
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "source_key" character varying NULL;
-- create index "procedure_source_key" to table: "procedures"
CREATE INDEX "procedure_source_key" ON "procedures" ("source_key");
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "source_key" character varying NULL;
-- create index "program_source_key" to table: "programs"
CREATE INDEX "program_source_key" ON "programs" ("source_key");
-- modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "source_key" character varying NULL;
-- create index "remediation_source_key" to table: "remediations"
CREATE INDEX "remediation_source_key" ON "remediations" ("source_key");
-- modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "source_key" character varying NULL;
-- create index "review_source_key" to table: "reviews"
CREATE INDEX "review_source_key" ON "reviews" ("source_key");
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "source_key" character varying NULL;
-- create index "risk_source_key" to table: "risks"
CREATE INDEX "risk_source_key" ON "risks" ("source_key");
-- modify "scans" table
ALTER TABLE "scans" ADD COLUMN "source_key" character varying NULL;
-- create index "scan_source_key" to table: "scans"
CREATE INDEX "scan_source_key" ON "scans" ("source_key");
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "source_key" character varying NULL;
-- create index "subcontrol_source_key" to table: "subcontrols"
CREATE INDEX "subcontrol_source_key" ON "subcontrols" ("source_key");
-- modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "source_key" character varying NULL;
-- create index "systemdetail_source_key" to table: "system_details"
CREATE INDEX "systemdetail_source_key" ON "system_details" ("source_key");
-- modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD COLUMN "source_key" character varying NULL;
-- create index "tagdefinition_source_key" to table: "tag_definitions"
CREATE INDEX "tagdefinition_source_key" ON "tag_definitions" ("source_key");
-- create index "task_source_key" to table: "tasks"
CREATE INDEX "task_source_key" ON "tasks" ("source_key");
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "source_key" character varying NULL;
-- create index "template_source_key" to table: "templates"
CREATE INDEX "template_source_key" ON "templates" ("source_key");
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "source_key" character varying NULL;
-- create index "vulnerability_source_key" to table: "vulnerabilities"
CREATE INDEX "vulnerability_source_key" ON "vulnerabilities" ("source_key");

-- +goose Down
-- reverse: create index "vulnerability_source_key" to table: "vulnerabilities"
DROP INDEX "vulnerability_source_key";
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP COLUMN "source_key";
-- reverse: create index "template_source_key" to table: "templates"
DROP INDEX "template_source_key";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "source_key";
-- reverse: create index "task_source_key" to table: "tasks"
DROP INDEX "task_source_key";
-- reverse: create index "tagdefinition_source_key" to table: "tag_definitions"
DROP INDEX "tagdefinition_source_key";
-- reverse: modify "tag_definitions" table
ALTER TABLE "tag_definitions" DROP COLUMN "source_key";
-- reverse: create index "systemdetail_source_key" to table: "system_details"
DROP INDEX "systemdetail_source_key";
-- reverse: modify "system_details" table
ALTER TABLE "system_details" DROP COLUMN "source_key";
-- reverse: create index "subcontrol_source_key" to table: "subcontrols"
DROP INDEX "subcontrol_source_key";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "source_key";
-- reverse: create index "scan_source_key" to table: "scans"
DROP INDEX "scan_source_key";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP COLUMN "source_key";
-- reverse: create index "risk_source_key" to table: "risks"
DROP INDEX "risk_source_key";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "source_key";
-- reverse: create index "review_source_key" to table: "reviews"
DROP INDEX "review_source_key";
-- reverse: modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "source_key";
-- reverse: create index "remediation_source_key" to table: "remediations"
DROP INDEX "remediation_source_key";
-- reverse: modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "source_key";
-- reverse: create index "program_source_key" to table: "programs"
DROP INDEX "program_source_key";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "source_key";
-- reverse: create index "procedure_source_key" to table: "procedures"
DROP INDEX "procedure_source_key";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "source_key";
-- reverse: create index "platform_source_key" to table: "platforms"
DROP INDEX "platform_source_key";
-- reverse: modify "platforms" table
ALTER TABLE "platforms" DROP COLUMN "source_key";
-- reverse: create index "narrative_source_key" to table: "narratives"
DROP INDEX "narrative_source_key";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP COLUMN "source_key";
-- reverse: create index "mappedcontrol_source_key" to table: "mapped_controls"
DROP INDEX "mappedcontrol_source_key";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "source_key";
-- reverse: create index "internalpolicy_source_key" to table: "internal_policies"
DROP INDEX "internalpolicy_source_key";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "source_key";
-- reverse: create index "identityholder_source_key" to table: "identity_holders"
DROP INDEX "identityholder_source_key";
-- reverse: modify "identity_holders" table
ALTER TABLE "identity_holders" DROP COLUMN "source_key";
-- reverse: create index "group_source_key" to table: "groups"
DROP INDEX "group_source_key";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "source_key";
-- reverse: create index "finding_source_key" to table: "findings"
DROP INDEX "finding_source_key";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP COLUMN "source_key";
-- reverse: create index "evidence_source_key" to table: "evidences"
DROP INDEX "evidence_source_key";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "source_key";
-- reverse: create index "entity_source_key" to table: "entities"
DROP INDEX "entity_source_key";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "source_key";
-- reverse: create index "emailtemplate_source_key" to table: "email_templates"
DROP INDEX "emailtemplate_source_key";
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" DROP COLUMN "source_key";
-- reverse: create index "directorymembership_source_key" to table: "directory_memberships"
DROP INDEX "directorymembership_source_key";
-- reverse: modify "directory_memberships" table
ALTER TABLE "directory_memberships" DROP COLUMN "source_key";
-- reverse: create index "directorygroup_source_key" to table: "directory_groups"
DROP INDEX "directorygroup_source_key";
-- reverse: modify "directory_groups" table
ALTER TABLE "directory_groups" DROP COLUMN "source_key";
-- reverse: create index "directoryaccount_source_key" to table: "directory_accounts"
DROP INDEX "directoryaccount_source_key";
-- reverse: modify "directory_accounts" table
ALTER TABLE "directory_accounts" DROP COLUMN "source_key";
-- reverse: create index "customtypeenum_source_key" to table: "custom_type_enums"
DROP INDEX "customtypeenum_source_key";
-- reverse: modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" DROP COLUMN "source_key";
-- reverse: create index "control_source_key" to table: "controls"
DROP INDEX "control_source_key";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "source_key";
-- reverse: create index "controlobjective_source_key" to table: "control_objectives"
DROP INDEX "controlobjective_source_key";
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" DROP COLUMN "source_key";
-- reverse: create index "controlimplementation_source_key" to table: "control_implementations"
DROP INDEX "controlimplementation_source_key";
-- reverse: modify "control_implementations" table
ALTER TABLE "control_implementations" DROP COLUMN "source_key";
-- reverse: create index "contact_source_key" to table: "contacts"
DROP INDEX "contact_source_key";
-- reverse: modify "contacts" table
ALTER TABLE "contacts" DROP COLUMN "source_key";
-- reverse: create index "checkresult_source_key" to table: "check_results"
DROP INDEX "checkresult_source_key";
-- reverse: modify "check_results" table
ALTER TABLE "check_results" DROP COLUMN "source_key";
-- reverse: create index "campaign_source_key" to table: "campaigns"
DROP INDEX "campaign_source_key";
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP COLUMN "source_key";
-- reverse: create index "asset_source_key" to table: "assets"
DROP INDEX "asset_source_key";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP COLUMN "source_key";
-- reverse: create index "assessment_source_key" to table: "assessments"
DROP INDEX "assessment_source_key";
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "source_key";
-- reverse: create index "assessmentpolicy_source_key" to table: "assessment_policies"
DROP INDEX "assessmentpolicy_source_key";
-- reverse: modify "assessment_policies" table
ALTER TABLE "assessment_policies" DROP COLUMN "source_key";
-- reverse: create index "actionplan_source_key" to table: "action_plans"
DROP INDEX "actionplan_source_key";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "source_key";
