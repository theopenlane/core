-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "impact" SET DEFAULT 'MODERATE', ALTER COLUMN "likelihood" SET DEFAULT 'LIKELY';
-- modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "impact" SET DEFAULT 'MODERATE', ALTER COLUMN "likelihood" SET DEFAULT 'LIKELY', ADD COLUMN "program_risks" character varying NOT NULL, ADD CONSTRAINT "risks_programs_risks" FOREIGN KEY ("program_risks") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_programs_risks", DROP COLUMN "program_risks", ALTER COLUMN "likelihood" DROP DEFAULT, ALTER COLUMN "impact" DROP DEFAULT;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "likelihood" DROP DEFAULT, ALTER COLUMN "impact" DROP DEFAULT;
