-- +goose Up
-- modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "name" character varying NOT NULL, ADD COLUMN "description" text NULL, ADD COLUMN "logo_remote_url" character varying NULL, ADD COLUMN "logo_local_file_id" character varying NULL;
-- modify "subprocessors" table
ALTER TABLE "subprocessors" ADD COLUMN "system_owned" boolean NULL DEFAULT false, ADD COLUMN "name" character varying NOT NULL, ADD COLUMN "description" text NULL, ADD COLUMN "logo_remote_url" character varying NULL, ADD COLUMN "owner_id" character varying NULL, ADD COLUMN "logo_local_file_id" character varying NULL, ADD CONSTRAINT "subprocessors_files_logo_file" FOREIGN KEY ("logo_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "subprocessors_organizations_subprocessors" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "subprocessor_name_owner_id" to table: "subprocessors"
CREATE UNIQUE INDEX "subprocessor_name_owner_id" ON "subprocessors" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "subprocessor_owner_id" to table: "subprocessors"
CREATE INDEX "subprocessor_owner_id" ON "subprocessors" ("owner_id") WHERE (deleted_at IS NULL);
-- create "subprocessor_files" table
CREATE TABLE "subprocessor_files" ("subprocessor_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("subprocessor_id", "file_id"), CONSTRAINT "subprocessor_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subprocessor_files_subprocessor_id" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "subprocessor_files" table
DROP TABLE "subprocessor_files";
-- reverse: create index "subprocessor_owner_id" to table: "subprocessors"
DROP INDEX "subprocessor_owner_id";
-- reverse: create index "subprocessor_name_owner_id" to table: "subprocessors"
DROP INDEX "subprocessor_name_owner_id";
-- reverse: modify "subprocessors" table
ALTER TABLE "subprocessors" DROP CONSTRAINT "subprocessors_organizations_subprocessors", DROP CONSTRAINT "subprocessors_files_logo_file", DROP COLUMN "logo_local_file_id", DROP COLUMN "owner_id", DROP COLUMN "logo_remote_url", DROP COLUMN "description", DROP COLUMN "name", DROP COLUMN "system_owned";
-- reverse: modify "subprocessor_history" table
ALTER TABLE "subprocessor_history" DROP COLUMN "logo_local_file_id", DROP COLUMN "logo_remote_url", DROP COLUMN "description", DROP COLUMN "name", DROP COLUMN "system_owned", DROP COLUMN "owner_id";
