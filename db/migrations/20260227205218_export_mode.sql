-- Modify "exports" table
ALTER TABLE "exports" ADD COLUMN "mode" character varying NOT NULL DEFAULT 'FLAT', ADD COLUMN "export_metadata" jsonb NULL;
