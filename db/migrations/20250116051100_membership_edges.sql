-- Modify "group_memberships" table
ALTER TABLE "group_memberships" ADD COLUMN "group_membership_orgmembership" character varying NULL, ADD CONSTRAINT "group_memberships_org_memberships_orgmembership" FOREIGN KEY ("group_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" ADD COLUMN "program_membership_orgmembership" character varying NULL, ADD CONSTRAINT "program_memberships_org_memberships_orgmembership" FOREIGN KEY ("program_membership_orgmembership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
