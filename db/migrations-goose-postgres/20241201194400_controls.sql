-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"), CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"), CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_viewers" table
CREATE TABLE "control_viewers" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"), CONSTRAINT "control_viewers_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "control_viewers" table
DROP TABLE "control_viewers";
-- reverse: create "control_editors" table
DROP TABLE "control_editors";
-- reverse: create "control_blocked_groups" table
DROP TABLE "control_blocked_groups";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_organizations_controls", DROP COLUMN "owner_id";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "owner_id";
