-- Modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'LOW';
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "identity_holder_id" character varying NULL, ADD CONSTRAINT "directory_groups_identity_holders_directory_groups" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "directorygroup_identity_holder_id" to table: "directory_groups"
CREATE INDEX "directorygroup_identity_holder_id" ON "directory_groups" ("identity_holder_id");
