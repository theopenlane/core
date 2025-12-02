-- Modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
-- Modify "controls" table
ALTER TABLE "controls" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ALTER COLUMN "status" SET DEFAULT 'NOT_IMPLEMENTED';
