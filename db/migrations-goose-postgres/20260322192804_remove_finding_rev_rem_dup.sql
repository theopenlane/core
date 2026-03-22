-- +goose Up
-- modify "findings" table
ALTER TABLE "findings" DROP COLUMN "remediation_findings", DROP COLUMN "review_findings";
-- modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "finding_remediations";
-- create "remediation_findings" table
CREATE TABLE "remediation_findings" ("remediation_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "finding_id"), CONSTRAINT "remediation_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_findings_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "finding_reviews";
-- create "review_findings" table
CREATE TABLE "review_findings" ("review_id" character varying NOT NULL, "finding_id" character varying NOT NULL, PRIMARY KEY ("review_id", "finding_id"), CONSTRAINT "review_findings_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_findings_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "review_findings" table
DROP TABLE "review_findings";
-- reverse: modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "finding_reviews" character varying NULL;
-- reverse: create "remediation_findings" table
DROP TABLE "remediation_findings";
-- reverse: modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "finding_remediations" character varying NULL;
-- reverse: modify "findings" table
ALTER TABLE "findings" ADD COLUMN "review_findings" character varying NULL, ADD COLUMN "remediation_findings" character varying NULL;
