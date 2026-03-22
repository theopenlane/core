-- Modify "findings" table
ALTER TABLE "findings" DROP COLUMN "remediation_findings", DROP COLUMN "review_findings";
-- Modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "finding_remediations";
-- Create "remediation_findings" table
CREATE TABLE "remediation_findings" ("remediation_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "finding_id"), CONSTRAINT "remediation_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_findings_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "finding_reviews";
-- Create "review_findings" table
CREATE TABLE "review_findings" ("review_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("review_id", "finding_id"), CONSTRAINT "review_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_findings_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
