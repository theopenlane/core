-- +goose Up
-- modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "program_owner_id" character varying NULL;
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "program_owner_id" character varying NULL, ADD CONSTRAINT "programs_users_program_owner" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "programs_program_owner_id_key" to table: "programs"
CREATE UNIQUE INDEX "programs_program_owner_id_key" ON "programs" ("program_owner_id");

-- +goose Down
-- reverse: create index "programs_program_owner_id_key" to table: "programs"
DROP INDEX "programs_program_owner_id_key";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_users_program_owner", DROP COLUMN "program_owner_id";
-- reverse: modify "program_history" table
ALTER TABLE "program_history" DROP COLUMN "program_owner_id";
