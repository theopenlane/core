-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;
