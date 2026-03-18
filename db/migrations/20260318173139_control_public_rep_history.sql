-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;
