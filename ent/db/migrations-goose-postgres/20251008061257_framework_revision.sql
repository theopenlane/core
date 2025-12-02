-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "reference_framework_revision" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "reference_framework_revision" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "reference_framework_revision" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "reference_framework_revision" character varying NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "reference_framework_revision";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "reference_framework_revision";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "reference_framework_revision";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "reference_framework_revision";
