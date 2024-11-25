-- Modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "impact" SET DEFAULT 'MODERATE', ALTER COLUMN "likelihood" SET DEFAULT 'LIKELY';
-- Modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "impact" SET DEFAULT 'MODERATE', ALTER COLUMN "likelihood" SET DEFAULT 'LIKELY', ADD COLUMN "program_risks" character varying NOT NULL, ADD CONSTRAINT "risks_programs_risks" FOREIGN KEY ("program_risks") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
