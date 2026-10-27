-- Create index "finding_external_id_owner_id" to table: "findings"
CREATE UNIQUE INDEX "finding_external_id_owner_id" ON "findings" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_owner_id IS NULL));
-- Modify "assets" table
ALTER TABLE "assets" ADD COLUMN "external_id" character varying NULL;
-- Create index "asset_external_id_owner_id" to table: "assets"
CREATE UNIQUE INDEX "asset_external_id_owner_id" ON "assets" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "contact_external_id_owner_id" to table: "contacts"
CREATE UNIQUE INDEX "contact_external_id_owner_id" ON "contacts" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "external_id" character varying NULL;
-- Create index "actionplan_external_id_owner_id" to table: "action_plans"
CREATE UNIQUE INDEX "actionplan_external_id_owner_id" ON "action_plans" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "external_id" character varying NULL;
-- Create index "procedure_external_id_owner_id" to table: "procedures"
CREATE UNIQUE INDEX "procedure_external_id_owner_id" ON "procedures" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "external_id" character varying NULL;
-- Create index "internalpolicy_external_id_owner_id" to table: "internal_policies"
CREATE UNIQUE INDEX "internalpolicy_external_id_owner_id" ON "internal_policies" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Modify "check_results" table
ALTER TABLE "check_results" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "check_results_organizations_check_results" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "check_result_owner_id_idx" to table: "check_results"
CREATE INDEX "check_result_owner_id_idx" ON "check_results" ("owner_id");
-- Modify "assessment_policies" table
ALTER TABLE "assessment_policies" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "assessmentpolicy_external_id_owner_id" to table: "assessment_policies"
CREATE UNIQUE INDEX "assessmentpolicy_external_id_owner_id" ON "assessment_policies" ("external_id", "owner_id");
-- Create index "assessmentpolicy_integration_run_id" to table: "assessment_policies"
CREATE INDEX "assessmentpolicy_integration_run_id" ON "assessment_policies" ("integration_run_id");
-- Create index "assessmentpolicy_source_instance_id" to table: "assessment_policies"
CREATE INDEX "assessmentpolicy_source_instance_id" ON "assessment_policies" ("source_instance_id");
-- Modify "assessments" table
ALTER TABLE "assessments" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "assessment_external_id_owner_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_external_id_owner_id" ON "assessments" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "assessment_integration_run_id" to table: "assessments"
CREATE INDEX "assessment_integration_run_id" ON "assessments" ("integration_run_id");
-- Create index "assessment_source_instance_id" to table: "assessments"
CREATE INDEX "assessment_source_instance_id" ON "assessments" ("source_instance_id");
-- Modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "campaign_external_id_owner_id" to table: "campaigns"
CREATE UNIQUE INDEX "campaign_external_id_owner_id" ON "campaigns" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "campaign_integration_run_id" to table: "campaigns"
CREATE INDEX "campaign_integration_run_id" ON "campaigns" ("integration_run_id");
-- Create index "campaign_source_instance_id" to table: "campaigns"
CREATE INDEX "campaign_source_instance_id" ON "campaigns" ("source_instance_id");
-- Modify "control_implementations" table
ALTER TABLE "control_implementations" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "controlimplementation_external_id_owner_id" to table: "control_implementations"
CREATE UNIQUE INDEX "controlimplementation_external_id_owner_id" ON "control_implementations" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "controlimplementation_integration_run_id" to table: "control_implementations"
CREATE INDEX "controlimplementation_integration_run_id" ON "control_implementations" ("integration_run_id");
-- Create index "controlimplementation_source_instance_id" to table: "control_implementations"
CREATE INDEX "controlimplementation_source_instance_id" ON "control_implementations" ("source_instance_id");
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "controlobjective_external_id_owner_id" to table: "control_objectives"
CREATE UNIQUE INDEX "controlobjective_external_id_owner_id" ON "control_objectives" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "controlobjective_integration_run_id" to table: "control_objectives"
CREATE INDEX "controlobjective_integration_run_id" ON "control_objectives" ("integration_run_id");
-- Create index "controlobjective_source_instance_id" to table: "control_objectives"
CREATE INDEX "controlobjective_source_instance_id" ON "control_objectives" ("source_instance_id");
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "control_external_id_owner_id" to table: "controls"
CREATE UNIQUE INDEX "control_external_id_owner_id" ON "controls" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "control_integration_run_id" to table: "controls"
CREATE INDEX "control_integration_run_id" ON "controls" ("integration_run_id");
-- Create index "control_source_instance_id" to table: "controls"
CREATE INDEX "control_source_instance_id" ON "controls" ("source_instance_id");
-- Modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "customtypeenum_external_id_owner_id" to table: "custom_type_enums"
CREATE UNIQUE INDEX "customtypeenum_external_id_owner_id" ON "custom_type_enums" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "customtypeenum_integration_run_id" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_integration_run_id" ON "custom_type_enums" ("integration_run_id");
-- Create index "customtypeenum_source_instance_id" to table: "custom_type_enums"
CREATE INDEX "customtypeenum_source_instance_id" ON "custom_type_enums" ("source_instance_id");
-- Modify "email_templates" table
ALTER TABLE "email_templates" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "emailtemplate_external_id_owner_id" to table: "email_templates"
CREATE UNIQUE INDEX "emailtemplate_external_id_owner_id" ON "email_templates" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "emailtemplate_integration_run_id" to table: "email_templates"
CREATE INDEX "emailtemplate_integration_run_id" ON "email_templates" ("integration_run_id");
-- Create index "emailtemplate_source_instance_id" to table: "email_templates"
CREATE INDEX "emailtemplate_source_instance_id" ON "email_templates" ("source_instance_id");
-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "evidence_external_id_owner_id" to table: "evidences"
CREATE UNIQUE INDEX "evidence_external_id_owner_id" ON "evidences" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "evidence_integration_run_id" to table: "evidences"
CREATE INDEX "evidence_integration_run_id" ON "evidences" ("integration_run_id");
-- Create index "evidence_source_instance_id" to table: "evidences"
CREATE INDEX "evidence_source_instance_id" ON "evidences" ("source_instance_id");
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "group_external_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_external_id_owner_id" ON "groups" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "group_integration_run_id" to table: "groups"
CREATE INDEX "group_integration_run_id" ON "groups" ("integration_run_id");
-- Create index "group_source_instance_id" to table: "groups"
CREATE INDEX "group_source_instance_id" ON "groups" ("source_instance_id");
-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "identityholder_external_id_owner_id" to table: "identity_holders"
CREATE UNIQUE INDEX "identityholder_external_id_owner_id" ON "identity_holders" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "identityholder_integration_run_id" to table: "identity_holders"
CREATE INDEX "identityholder_integration_run_id" ON "identity_holders" ("integration_run_id");
-- Create index "identityholder_source_instance_id" to table: "identity_holders"
CREATE INDEX "identityholder_source_instance_id" ON "identity_holders" ("source_instance_id");
-- Modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "mappedcontrol_external_id_owner_id" to table: "mapped_controls"
CREATE UNIQUE INDEX "mappedcontrol_external_id_owner_id" ON "mapped_controls" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "mappedcontrol_integration_run_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_integration_run_id" ON "mapped_controls" ("integration_run_id");
-- Create index "mappedcontrol_source_instance_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_source_instance_id" ON "mapped_controls" ("source_instance_id");
-- Modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "narrative_external_id_owner_id" to table: "narratives"
CREATE UNIQUE INDEX "narrative_external_id_owner_id" ON "narratives" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "narrative_integration_run_id" to table: "narratives"
CREATE INDEX "narrative_integration_run_id" ON "narratives" ("integration_run_id");
-- Create index "narrative_source_instance_id" to table: "narratives"
CREATE INDEX "narrative_source_instance_id" ON "narratives" ("source_instance_id");
-- Modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "platform_external_id_owner_id" to table: "platforms"
CREATE UNIQUE INDEX "platform_external_id_owner_id" ON "platforms" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "platform_integration_run_id" to table: "platforms"
CREATE INDEX "platform_integration_run_id" ON "platforms" ("integration_run_id");
-- Create index "platform_source_instance_id" to table: "platforms"
CREATE INDEX "platform_source_instance_id" ON "platforms" ("source_instance_id");
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "program_external_id_owner_id" to table: "programs"
CREATE UNIQUE INDEX "program_external_id_owner_id" ON "programs" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "program_integration_run_id" to table: "programs"
CREATE INDEX "program_integration_run_id" ON "programs" ("integration_run_id");
-- Create index "program_source_instance_id" to table: "programs"
CREATE INDEX "program_source_instance_id" ON "programs" ("source_instance_id");
-- Modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "remediation_external_id_owner_id" to table: "remediations"
CREATE UNIQUE INDEX "remediation_external_id_owner_id" ON "remediations" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_owner_id IS NULL));
-- Create index "remediation_integration_run_id" to table: "remediations"
CREATE INDEX "remediation_integration_run_id" ON "remediations" ("integration_run_id");
-- Create index "remediation_source_instance_id" to table: "remediations"
CREATE INDEX "remediation_source_instance_id" ON "remediations" ("source_instance_id");
-- Modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "review_external_id_owner_id" to table: "reviews"
CREATE UNIQUE INDEX "review_external_id_owner_id" ON "reviews" ("external_id", "owner_id") WHERE ((deleted_at IS NULL) AND (external_owner_id IS NULL));
-- Create index "review_integration_run_id" to table: "reviews"
CREATE INDEX "review_integration_run_id" ON "reviews" ("integration_run_id");
-- Create index "review_source_instance_id" to table: "reviews"
CREATE INDEX "review_source_instance_id" ON "reviews" ("source_instance_id");
-- Modify "scans" table
ALTER TABLE "scans" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "scan_external_id_owner_id" to table: "scans"
CREATE UNIQUE INDEX "scan_external_id_owner_id" ON "scans" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "scan_integration_run_id" to table: "scans"
CREATE INDEX "scan_integration_run_id" ON "scans" ("integration_run_id");
-- Create index "scan_source_instance_id" to table: "scans"
CREATE INDEX "scan_source_instance_id" ON "scans" ("source_instance_id");
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "subcontrol_external_id_owner_id" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrol_external_id_owner_id" ON "subcontrols" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "subcontrol_integration_run_id" to table: "subcontrols"
CREATE INDEX "subcontrol_integration_run_id" ON "subcontrols" ("integration_run_id");
-- Create index "subcontrol_source_instance_id" to table: "subcontrols"
CREATE INDEX "subcontrol_source_instance_id" ON "subcontrols" ("source_instance_id");
-- Modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "systemdetail_external_id_owner_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_external_id_owner_id" ON "system_details" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "systemdetail_integration_run_id" to table: "system_details"
CREATE INDEX "systemdetail_integration_run_id" ON "system_details" ("integration_run_id");
-- Create index "systemdetail_source_instance_id" to table: "system_details"
CREATE INDEX "systemdetail_source_instance_id" ON "system_details" ("source_instance_id");
-- Modify "tag_definitions" table
ALTER TABLE "tag_definitions" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "tagdefinition_external_id_owner_id" to table: "tag_definitions"
CREATE UNIQUE INDEX "tagdefinition_external_id_owner_id" ON "tag_definitions" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "tagdefinition_integration_run_id" to table: "tag_definitions"
CREATE INDEX "tagdefinition_integration_run_id" ON "tag_definitions" ("integration_run_id");
-- Create index "tagdefinition_source_instance_id" to table: "tag_definitions"
CREATE INDEX "tagdefinition_source_instance_id" ON "tag_definitions" ("source_instance_id");
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "task_external_id_owner_id" to table: "tasks"
CREATE UNIQUE INDEX "task_external_id_owner_id" ON "tasks" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "task_integration_run_id" to table: "tasks"
CREATE INDEX "task_integration_run_id" ON "tasks" ("integration_run_id");
-- Create index "task_source_instance_id" to table: "tasks"
CREATE INDEX "task_source_instance_id" ON "tasks" ("source_instance_id");
-- Modify "templates" table
ALTER TABLE "templates" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "template_external_id_owner_id" to table: "templates"
CREATE UNIQUE INDEX "template_external_id_owner_id" ON "templates" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "template_integration_run_id" to table: "templates"
CREATE INDEX "template_integration_run_id" ON "templates" ("integration_run_id");
-- Create index "template_source_instance_id" to table: "templates"
CREATE INDEX "template_source_instance_id" ON "templates" ("source_instance_id");
-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "assessment_integration_runs" character varying NULL, ADD COLUMN "assessment_policy_integration_runs" character varying NULL, ADD COLUMN "campaign_integration_runs" character varying NULL, ADD COLUMN "control_integration_runs" character varying NULL, ADD COLUMN "control_implementation_integration_runs" character varying NULL, ADD COLUMN "control_objective_integration_runs" character varying NULL, ADD COLUMN "custom_type_enum_integration_runs" character varying NULL, ADD COLUMN "email_template_integration_runs" character varying NULL, ADD COLUMN "evidence_integration_runs" character varying NULL, ADD COLUMN "group_integration_runs" character varying NULL, ADD COLUMN "identity_holder_integration_runs" character varying NULL, ADD COLUMN "mapped_control_integration_runs" character varying NULL, ADD COLUMN "narrative_integration_runs" character varying NULL, ADD COLUMN "platform_integration_runs" character varying NULL, ADD COLUMN "program_integration_runs" character varying NULL, ADD COLUMN "remediation_integration_runs" character varying NULL, ADD COLUMN "review_integration_runs" character varying NULL, ADD COLUMN "scan_integration_runs" character varying NULL, ADD COLUMN "subcontrol_integration_runs" character varying NULL, ADD COLUMN "system_detail_integration_runs" character varying NULL, ADD COLUMN "tag_definition_integration_runs" character varying NULL, ADD COLUMN "task_integration_runs" character varying NULL, ADD COLUMN "template_integration_runs" character varying NULL, ADD CONSTRAINT "integration_runs_assessment_policies_integration_runs" FOREIGN KEY ("assessment_policy_integration_runs") REFERENCES "assessment_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_assessments_integration_runs" FOREIGN KEY ("assessment_integration_runs") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_campaigns_integration_runs" FOREIGN KEY ("campaign_integration_runs") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_control_implementations_integration_runs" FOREIGN KEY ("control_implementation_integration_runs") REFERENCES "control_implementations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_control_objectives_integration_runs" FOREIGN KEY ("control_objective_integration_runs") REFERENCES "control_objectives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_controls_integration_runs" FOREIGN KEY ("control_integration_runs") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_custom_type_enums_integration_runs" FOREIGN KEY ("custom_type_enum_integration_runs") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_email_templates_integration_runs" FOREIGN KEY ("email_template_integration_runs") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_evidences_integration_runs" FOREIGN KEY ("evidence_integration_runs") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_groups_integration_runs" FOREIGN KEY ("group_integration_runs") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_identity_holders_integration_runs" FOREIGN KEY ("identity_holder_integration_runs") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_mapped_controls_integration_runs" FOREIGN KEY ("mapped_control_integration_runs") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_narratives_integration_runs" FOREIGN KEY ("narrative_integration_runs") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_platforms_integration_runs" FOREIGN KEY ("platform_integration_runs") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_programs_integration_runs" FOREIGN KEY ("program_integration_runs") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_remediations_integration_runs" FOREIGN KEY ("remediation_integration_runs") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_reviews_integration_runs" FOREIGN KEY ("review_integration_runs") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_scans_integration_runs" FOREIGN KEY ("scan_integration_runs") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_subcontrols_integration_runs" FOREIGN KEY ("subcontrol_integration_runs") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_system_details_integration_runs" FOREIGN KEY ("system_detail_integration_runs") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_tag_definitions_integration_runs" FOREIGN KEY ("tag_definition_integration_runs") REFERENCES "tag_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_tasks_integration_runs" FOREIGN KEY ("task_integration_runs") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_templates_integration_runs" FOREIGN KEY ("template_integration_runs") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
