-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "user_action_plans" character varying NULL, ADD CONSTRAINT "action_plans_users_action_plans" FOREIGN KEY ("user_action_plans") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "entity_blocked_groups", DROP COLUMN "entity_editors", DROP COLUMN "entity_viewers";
-- create "entity_blocked_groups" table
CREATE TABLE "entity_blocked_groups" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"), CONSTRAINT "entity_blocked_groups_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "entity_editors" table
CREATE TABLE "entity_editors" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"), CONSTRAINT "entity_editors_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "entity_viewers" table
CREATE TABLE "entity_viewers" ("entity_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "group_id"), CONSTRAINT "entity_viewers_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "group_membership_orgmembership", ADD COLUMN "group_membership_org_membership" character varying NULL, ADD CONSTRAINT "group_memberships_org_memberships_org_membership" FOREIGN KEY ("group_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "program_membership_orgmembership", ADD COLUMN "program_membership_org_membership" character varying NULL, ADD CONSTRAINT "program_memberships_org_memberships_org_membership" FOREIGN KEY ("program_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP CONSTRAINT "program_memberships_org_memberships_org_membership", DROP COLUMN "program_membership_org_membership", ADD COLUMN "program_membership_orgmembership" character varying NULL;
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP CONSTRAINT "group_memberships_org_memberships_org_membership", DROP COLUMN "group_membership_org_membership", ADD COLUMN "group_membership_orgmembership" character varying NULL;
-- reverse: create "entity_viewers" table
DROP TABLE "entity_viewers";
-- reverse: create "entity_editors" table
DROP TABLE "entity_editors";
-- reverse: create "entity_blocked_groups" table
DROP TABLE "entity_blocked_groups";
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "entity_viewers" character varying NULL, ADD COLUMN "entity_editors" character varying NULL, ADD COLUMN "entity_blocked_groups" character varying NULL;
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_users_action_plans", DROP COLUMN "user_action_plans";
