-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;

-- +goose Down
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "source_name", DROP COLUMN "public_representation";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "source_name", DROP COLUMN "public_representation";
