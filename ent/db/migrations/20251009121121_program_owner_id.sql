-- Modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "program_owner_id" character varying NULL;
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "program_owner_id" character varying NULL, ADD CONSTRAINT "programs_users_program_owner" FOREIGN KEY ("program_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "programs_program_owner_id_key" to table: "programs"
CREATE UNIQUE INDEX "programs_program_owner_id_key" ON "programs" ("program_owner_id");
