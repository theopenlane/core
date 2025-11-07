-- +goose Up
-- drop index "programs_program_owner_id_key" from table: "programs"
DROP INDEX "programs_program_owner_id_key";
-- modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_users_program_owner", ADD CONSTRAINT "programs_users_programs_owned" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_users_programs_owned", ADD CONSTRAINT "programs_users_program_owner" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: drop index "programs_program_owner_id_key" from table: "programs"
CREATE UNIQUE INDEX "programs_program_owner_id_key" ON "programs" ("program_owner_id");
