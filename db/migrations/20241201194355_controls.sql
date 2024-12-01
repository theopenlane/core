-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "controls_organizations_controls" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "control_blocked_groups" table
CREATE TABLE "control_blocked_groups" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"), CONSTRAINT "control_blocked_groups_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "control_editors" table
CREATE TABLE "control_editors" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"), CONSTRAINT "control_editors_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "control_viewers" table
CREATE TABLE "control_viewers" ("control_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("control_id", "group_id"), CONSTRAINT "control_viewers_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
