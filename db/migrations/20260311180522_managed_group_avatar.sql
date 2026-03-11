-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "avatar_local_file_id" character varying NULL, ADD CONSTRAINT "groups_files_avatar_file" FOREIGN KEY ("avatar_local_file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
