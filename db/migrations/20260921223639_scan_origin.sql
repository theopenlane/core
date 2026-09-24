-- Modify "scans" table
ALTER TABLE "scans" ADD COLUMN "origin" character varying NOT NULL DEFAULT 'SYSTEM';
