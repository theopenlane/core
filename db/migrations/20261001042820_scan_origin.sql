-- Modify "scans" table
ALTER TABLE "scans" ADD COLUMN "document_kind_name" character varying NULL, ADD COLUMN "origin" character varying NOT NULL DEFAULT 'SYSTEM', ADD COLUMN "document_kind_id" character varying NULL, ADD CONSTRAINT "scans_custom_type_enums_document_kind" FOREIGN KEY ("document_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
