-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "review_frequency" character varying NULL DEFAULT 'YEARLY';
