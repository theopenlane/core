-- +goose Up
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "file_id" character varying NULL, ADD COLUMN "url" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "file_id" character varying NULL, ADD COLUMN "url" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "internal_policies_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "procedures_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_files_file", DROP COLUMN "file_id", DROP COLUMN "url";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_files_file", DROP COLUMN "file_id", DROP COLUMN "url";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "url", DROP COLUMN "file_id";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "url", DROP COLUMN "file_id";
