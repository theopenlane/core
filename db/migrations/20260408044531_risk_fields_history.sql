-- Modify "remediation_history" table
ALTER TABLE "remediation_history" ADD COLUMN "status" character varying NULL DEFAULT 'IN_PROGRESS';
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "mitigated_at" timestamptz NULL, ADD COLUMN "review_required" boolean NULL DEFAULT true, ADD COLUMN "last_reviewed_at" timestamptz NULL, ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "next_review_due_at" timestamptz NULL, ADD COLUMN "residual_score" bigint NULL, ADD COLUMN "risk_decision" character varying NULL DEFAULT ' NONE';
