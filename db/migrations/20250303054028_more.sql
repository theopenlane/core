-- Modify "control_history" table
ALTER TABLE "control_history" ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE', ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- Modify "control_objective_history" table
ALTER TABLE "control_objective_history" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- Modify "control_objectives" table
ALTER TABLE "control_objectives" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- Modify "controls" table
ALTER TABLE "controls" ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE', ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED';
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED', ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE';
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ALTER COLUMN "source" SET DEFAULT 'USER_DEFINED', ALTER COLUMN "control_type" SET DEFAULT 'PREVENTATIVE';
