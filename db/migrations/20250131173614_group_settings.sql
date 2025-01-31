-- Modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "display_id" character varying NOT NULL;
-- Modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "tags";
-- Modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "tags";
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "display_id" character varying NOT NULL;
-- Create index "group_display_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_display_id_owner_id" ON "groups" ("display_id", "owner_id");
-- Modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "tags";
-- Modify "notes" table
ALTER TABLE "notes" DROP COLUMN "tags";
