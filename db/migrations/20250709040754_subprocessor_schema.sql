-- Modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "name" character varying NOT NULL, ADD COLUMN "description" text NULL, ADD COLUMN "logo_remote_url" character varying NULL, ADD COLUMN "logo_local_file_id" character varying NULL;
-- Modify "subprocessors" table
ALTER TABLE "subprocessors" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "name" character varying NOT NULL, ADD COLUMN "description" text NULL, ADD COLUMN "logo_remote_url" character varying NULL, ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "logo_local_file_id" character varying NULL, ADD CONSTRAINT "subprocessors_files_logo_file" FOREIGN KEY ("logo_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subprocessors_organizations_subprocessors" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "subprocessor_name_owner_id" to table: "subprocessors"
CREATE UNIQUE INDEX "subprocessor_name_owner_id" ON "subprocessors" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "subprocessor_owner_id" to table: "subprocessors"
CREATE INDEX "subprocessor_owner_id" ON "subprocessors" ("owner_id") WHERE (deleted_at IS NULL);
-- Create "subprocessor_files" table
CREATE TABLE "subprocessor_files" ("subprocessor_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("subprocessor_id", "file_id"), CONSTRAINT "subprocessor_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subprocessor_files_subprocessor_id" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
