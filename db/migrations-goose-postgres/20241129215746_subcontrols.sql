-- +goose Up
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "owner_id" character varying NOT NULL, ADD COLUMN "user_subcontrols" character varying NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "subcontrols_users_subcontrols" FOREIGN KEY ("user_subcontrols") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "subcontrol_blocked_groups" table
CREATE TABLE "subcontrol_blocked_groups" ("subcontrol_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "group_id"), CONSTRAINT "subcontrol_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_blocked_groups_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "subcontrol_editors" table
CREATE TABLE "subcontrol_editors" ("subcontrol_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "group_id"), CONSTRAINT "subcontrol_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_editors_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "subcontrol_viewers" table
CREATE TABLE "subcontrol_viewers" ("subcontrol_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("subcontrol_id", "group_id"), CONSTRAINT "subcontrol_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "subcontrol_viewers_subcontrol_id" FOREIGN KEY ("subcontrol_id") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "subcontrol_viewers" table
DROP TABLE "subcontrol_viewers";
-- reverse: create "subcontrol_editors" table
DROP TABLE "subcontrol_editors";
-- reverse: create "subcontrol_blocked_groups" table
DROP TABLE "subcontrol_blocked_groups";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_users_subcontrols", DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP COLUMN "user_subcontrols", DROP COLUMN "owner_id";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "owner_id";
