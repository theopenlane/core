-- Create "exports" table
CREATE TABLE "exports" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "export_type" character varying NOT NULL, "format" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "requestor_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "exports_organizations_exports" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "export_owner_id" to table: "exports"
CREATE INDEX "export_owner_id" ON "exports" ("owner_id") WHERE (deleted_at IS NULL);
-- Modify "events" table
ALTER TABLE "events" ADD COLUMN "export_events" character varying NULL, ADD CONSTRAINT "events_exports_events" FOREIGN KEY ("export_events") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "export_files" character varying NULL, ADD CONSTRAINT "files_exports_files" FOREIGN KEY ("export_files") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
