-- +goose Up
-- modify "group_memberships" table
ALTER TABLE "group_memberships" ADD COLUMN "group_membership_orgmembership" character varying NULL, ADD CONSTRAINT "group_memberships_org_memberships_orgmembership" FOREIGN KEY ("group_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "program_memberships" table
ALTER TABLE "program_memberships" ADD COLUMN "program_membership_orgmembership" character varying NULL, ADD CONSTRAINT "program_memberships_org_memberships_orgmembership" FOREIGN KEY ("program_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "program_memberships" table
ALTER TABLE "program_memberships" DROP CONSTRAINT "program_memberships_org_memberships_orgmembership", DROP COLUMN "program_membership_orgmembership";
-- reverse: modify "group_memberships" table
ALTER TABLE "group_memberships" DROP CONSTRAINT "group_memberships_org_memberships_orgmembership", DROP COLUMN "group_membership_orgmembership";
