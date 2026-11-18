-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "control_origin_name" character varying NULL, ADD COLUMN "control_origin_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "control_origin_name" character varying NULL, ADD COLUMN "control_origin_id" character varying NULL;

-- +goose Down
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "control_origin_id", DROP COLUMN "control_origin_name";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "control_origin_id", DROP COLUMN "control_origin_name";
