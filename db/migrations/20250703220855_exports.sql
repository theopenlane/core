-- Create "export_history" table
CREATE TABLE "export_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "export_type" character varying NOT NULL, "item_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "requestor_id" character varying NULL, "file_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "exporthistory_history_time" to table: "export_history"
CREATE INDEX "exporthistory_history_time" ON "export_history" ("history_time");
-- Create "exports" table
CREATE TABLE "exports" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "export_type" character varying NOT NULL, "item_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "requestor_id" character varying NULL, "file_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "exports_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "exports_organizations_exports" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "export_id" to table: "exports"
CREATE UNIQUE INDEX "export_id" ON "exports" ("id");
-- Create index "export_owner_id" to table: "exports"
CREATE INDEX "export_owner_id" ON "exports" ("owner_id") WHERE (deleted_at IS NULL);
-- Modify "events" table
ALTER TABLE "events" ADD COLUMN "export_events" character varying NULL, ADD CONSTRAINT "events_exports_events" FOREIGN KEY ("export_events") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "export_files" table
CREATE TABLE "export_files" ("export_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("export_id", "file_id"), CONSTRAINT "export_files_export_id" FOREIGN KEY ("export_id") REFERENCES "exports" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "export_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
