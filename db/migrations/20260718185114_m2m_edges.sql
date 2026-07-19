-- Modify "assets" table
ALTER TABLE "assets" DROP COLUMN "finding_assets", DROP COLUMN "remediation_assets", DROP COLUMN "review_assets", DROP COLUMN "vulnerability_assets";
-- Create "finding_assets" table
CREATE TABLE "finding_assets" ("finding_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "asset_id"), CONSTRAINT "finding_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_assets_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "entities" table
ALTER TABLE "entities" DROP COLUMN "finding_entities", DROP COLUMN "remediation_entities", DROP COLUMN "review_entities", DROP COLUMN "vulnerability_entities";
-- Create "finding_entities" table
CREATE TABLE "finding_entities" ("finding_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "entity_id"), CONSTRAINT "finding_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_entities_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "programs" table
ALTER TABLE "programs" DROP COLUMN "finding_programs", DROP COLUMN "remediation_programs", DROP COLUMN "review_programs", DROP COLUMN "vulnerability_programs";
-- Create "finding_programs" table
CREATE TABLE "finding_programs" ("finding_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "program_id"), CONSTRAINT "finding_programs_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "risks" table
ALTER TABLE "risks" DROP COLUMN "finding_risks", DROP COLUMN "vulnerability_risks";
-- Create "finding_risks" table
CREATE TABLE "finding_risks" ("finding_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "risk_id"), CONSTRAINT "finding_risks_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "finding_subcontrols", DROP COLUMN "vulnerability_subcontrols";
-- Create "finding_subcontrols" table
CREATE TABLE "finding_subcontrols" ("finding_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "subcontrol_id"), CONSTRAINT "finding_subcontrols_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "remediation_assets" table
CREATE TABLE "remediation_assets" ("remediation_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "asset_id"), CONSTRAINT "remediation_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_assets_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "remediation_entities" table
CREATE TABLE "remediation_entities" ("remediation_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "entity_id"), CONSTRAINT "remediation_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_entities_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "remediation_programs" table
CREATE TABLE "remediation_programs" ("remediation_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "program_id"), CONSTRAINT "remediation_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_programs_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "review_assets" table
CREATE TABLE "review_assets" ("review_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("review_id", "asset_id"), CONSTRAINT "review_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_assets_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "review_entities" table
CREATE TABLE "review_entities" ("review_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("review_id", "entity_id"), CONSTRAINT "review_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_entities_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "review_programs" table
CREATE TABLE "review_programs" ("review_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("review_id", "program_id"), CONSTRAINT "review_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_programs_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "vulnerability_assets" table
CREATE TABLE "vulnerability_assets" ("vulnerability_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "asset_id"), CONSTRAINT "vulnerability_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_assets_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "controls" table
ALTER TABLE "controls" DROP COLUMN "vulnerability_controls";
-- Create "vulnerability_controls" table
CREATE TABLE "vulnerability_controls" ("vulnerability_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "control_id"), CONSTRAINT "vulnerability_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_controls_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "vulnerability_entities" table
CREATE TABLE "vulnerability_entities" ("vulnerability_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "entity_id"), CONSTRAINT "vulnerability_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_entities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "vulnerability_programs" table
CREATE TABLE "vulnerability_programs" ("vulnerability_id" character varying NOT NULL, "program_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "program_id"), CONSTRAINT "vulnerability_programs_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_programs_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "vulnerability_risks" table
CREATE TABLE "vulnerability_risks" ("vulnerability_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "risk_id"), CONSTRAINT "vulnerability_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_risks_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "vulnerability_subcontrols" table
CREATE TABLE "vulnerability_subcontrols" ("vulnerability_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("vulnerability_id", "subcontrol_id"), CONSTRAINT "vulnerability_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "vulnerability_subcontrols_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
