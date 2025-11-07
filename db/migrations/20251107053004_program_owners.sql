-- Drop index "programs_program_owner_id_key" from table: "programs"
DROP INDEX "programs_program_owner_id_key";
-- Modify "programs" table
ALTER TABLE "programs" DROP CONSTRAINT "programs_users_program_owner", ADD CONSTRAINT "programs_users_programs_owned" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
