-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ALTER COLUMN "description" DROP NOT NULL, ADD COLUMN "owner_id" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "owner_id" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ALTER COLUMN "description" DROP NOT NULL, ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "internal_policies_organizations_internalpolicies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "internal_policy_blocked_groups" table
CREATE TABLE "internal_policy_blocked_groups" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"), CONSTRAINT "internal_policy_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_blocked_groups_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "internal_policy_editors" table
CREATE TABLE "internal_policy_editors" ("internal_policy_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("internal_policy_id", "group_id"), CONSTRAINT "internal_policy_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "internal_policy_editors_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "procedures_organizations_procedures" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "procedure_blocked_groups" table
CREATE TABLE "procedure_blocked_groups" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"), CONSTRAINT "procedure_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "procedure_blocked_groups_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "procedure_editors" table
CREATE TABLE "procedure_editors" ("procedure_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("procedure_id", "group_id"), CONSTRAINT "procedure_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "procedure_editors_procedure_id" FOREIGN KEY ("procedure_id") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
