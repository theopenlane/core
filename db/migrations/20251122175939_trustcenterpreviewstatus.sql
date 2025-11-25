-- Modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "preview_status" character varying NULL DEFAULT 'NONE';
-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "preview_status" character varying NULL DEFAULT 'NONE';
