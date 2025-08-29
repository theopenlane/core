-- +goose Up
-- create "internal_policy_files" table
CREATE TABLE "internal_policy_files" ("internal_policy_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "file_id"), CONSTRAINT "internal_policy_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_files_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "procedure_files" table
CREATE TABLE "procedure_files" ("procedure_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "file_id"), CONSTRAINT "procedure_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "procedure_files_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "procedure_files" table
DROP TABLE "procedure_files";
-- reverse: create "internal_policy_files" table
DROP TABLE "internal_policy_files";
