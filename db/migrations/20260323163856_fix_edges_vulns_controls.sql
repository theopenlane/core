-- Modify "findings" table
ALTER TABLE "findings" DROP COLUMN "vulnerability_findings";
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP COLUMN "finding_vulnerabilities", DROP COLUMN "remediation_vulnerabilities", DROP COLUMN "review_vulnerabilities";
-- Create "finding_vulnerabilities" table
CREATE TABLE "finding_vulnerabilities" ("finding_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "vulnerability_id"), CONSTRAINT "finding_vulnerabilities_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "review_remediations", DROP COLUMN "vulnerability_remediations";
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "remediation_subcontrols", DROP COLUMN "review_subcontrols";
-- Create "remediation_subcontrols" table
CREATE TABLE "remediation_subcontrols" ("remediation_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "subcontrol_id"), CONSTRAINT "remediation_subcontrols_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "remediation_vulnerabilities" table
CREATE TABLE "remediation_vulnerabilities" ("remediation_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "vulnerability_id"), CONSTRAINT "remediation_vulnerabilities_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "remediation_reviews", DROP COLUMN "vulnerability_reviews";
-- Create "review_remediations" table
CREATE TABLE "review_remediations" ("review_id" character varying NOT NULL, "remediation_id" character varying NOT NULL, PRIMARY KEY ("review_id", "remediation_id"), CONSTRAINT "review_remediations_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_remediations_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "review_subcontrols" table
CREATE TABLE "review_subcontrols" ("review_id" character varying NOT NULL, "subcontrol_id" character varying NOT NULL, PRIMARY KEY ("review_id", "subcontrol_id"), CONSTRAINT "review_subcontrols_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_subcontrols_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "review_vulnerabilities" table
CREATE TABLE "review_vulnerabilities" ("review_id" character varying NOT NULL, "vulnerability_id" character varying NOT NULL, PRIMARY KEY ("review_id", "vulnerability_id"), CONSTRAINT "review_vulnerabilities_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_vulnerabilities_vulnerability_id" FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerabilities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "subcontrol_scans" table
CREATE TABLE "subcontrol_scans" ("subcontrol_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "scan_id"), CONSTRAINT "subcontrol_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_scans_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
