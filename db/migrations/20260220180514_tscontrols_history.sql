-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "trust_center_visibility" character varying NULL DEFAULT 'NOT_VISIBLE', ADD COLUMN "is_trust_center_control" boolean NULL DEFAULT false;
