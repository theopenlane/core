-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
-- modify "controls" table
ALTER TABLE "controls" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" ALTER COLUMN "status" SET DEFAULT 'NULL';
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "status" SET DEFAULT 'NULL';
-- reverse: modify "controls" table
ALTER TABLE "controls" ALTER COLUMN "status" SET DEFAULT 'NULL';
-- reverse: modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "status" SET DEFAULT 'NULL';
