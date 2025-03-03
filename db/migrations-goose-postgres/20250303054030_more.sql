-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE', ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- modify "control_objectives" table
ALTER TABLE "control_objectives" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- modify "controls" table
ALTER TABLE "controls" ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE', ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED', ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE';
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED', ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE';

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" ALTER COLUMN "control_type" DROP DEFAULT, ALTER COLUMN "source" DROP DEFAULT;
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "control_type" DROP DEFAULT, ALTER COLUMN "source" DROP DEFAULT;
-- reverse: modify "controls" table
ALTER TABLE "controls" ALTER COLUMN "source" DROP DEFAULT, ALTER COLUMN "control_type" DROP DEFAULT;
-- reverse: modify "control_objectives" table
ALTER TABLE "control_objectives" ALTER COLUMN "source" DROP DEFAULT;
-- reverse: modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "source" DROP DEFAULT;
-- reverse: modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "source" DROP DEFAULT, ALTER COLUMN "control_type" DROP DEFAULT;
