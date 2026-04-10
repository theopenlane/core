-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY', ADD COLUMN "next_review_at" timestamptz NULL;
