-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "trust_center_visibility" character varying NULL DEFAULT 'NOT_VISIBLE', ADD COLUMN "is_trust_center_control" boolean NULL DEFAULT false;
