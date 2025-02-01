-- +goose Up
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "display_id" character varying NOT NULL;
-- modify "group_setting_history" table
ALTER TABLE "group_setting_history" DROP COLUMN "tags";
-- modify "group_settings" table
ALTER TABLE "group_settings" DROP COLUMN "tags";
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "display_id" character varying NOT NULL;
-- create index "group_display_id_owner_id" to table: "groups"
CREATE UNIQUE INDEX "group_display_id_owner_id" ON "groups" ("display_id", "owner_id");
-- modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "tags";
-- modify "notes" table
ALTER TABLE "notes" DROP COLUMN "tags";

-- +goose Down
-- reverse: modify "notes" table
ALTER TABLE "notes" ADD COLUMN "tags" jsonb NULL;
-- reverse: modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "tags" jsonb NULL;
-- reverse: create index "group_display_id_owner_id" to table: "groups"
DROP INDEX "group_display_id_owner_id";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "display_id";
-- reverse: modify "group_settings" table
ALTER TABLE "group_settings" ADD COLUMN "tags" jsonb NULL;
-- reverse: modify "group_setting_history" table
ALTER TABLE "group_setting_history" ADD COLUMN "tags" jsonb NULL;
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "display_id";
