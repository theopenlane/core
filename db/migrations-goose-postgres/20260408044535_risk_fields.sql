-- +goose Up
-- modify "remediations" table
ALTER TABLE "remediations" ADD COLUMN "status" character varying NULL DEFAULT 'IN_PROGRESS';
-- modify "risks" table
ALTER TABLE "risks" DROP COLUMN "remediation_risks", DROP COLUMN "review_risks", ADD COLUMN "mitigated_at" timestamptz NULL, ADD COLUMN "review_required" boolean NULL DEFAULT true, ADD COLUMN "last_reviewed_at" timestamptz NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "next_review_due_at" timestamptz NULL, ADD COLUMN "residual_score" bigint NULL, ADD COLUMN "risk_decision" character varying NULL DEFAULT ' NONE';
-- create "remediation_risks" table
CREATE TABLE "remediation_risks" ("remediation_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "risk_id"), CONSTRAINT "remediation_risks_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "review_risks" table
CREATE TABLE "review_risks" ("review_id" character varying NOT NULL, "risk_id" character varying NOT NULL, PRIMARY KEY ("review_id", "risk_id"), CONSTRAINT "review_risks_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_risks_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "review_risks" table
DROP TABLE "review_risks";
-- reverse: create "remediation_risks" table
DROP TABLE "remediation_risks";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "risk_decision", DROP COLUMN "residual_score", DROP COLUMN "next_review_due_at", DROP COLUMN "review_frequency", DROP COLUMN "last_reviewed_at", DROP COLUMN "review_required", DROP COLUMN "mitigated_at", ADD COLUMN "review_risks" character varying NULL, ADD COLUMN "remediation_risks" character varying NULL;
-- reverse: modify "remediations" table
ALTER TABLE "remediations" DROP COLUMN "status";
