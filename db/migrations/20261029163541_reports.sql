-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_report_creators" character varying NULL, ADD COLUMN "organization_report_membership_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_report_creators" FOREIGN KEY ("organization_report_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_report_membership_creators" FOREIGN KEY ("organization_report_membership_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "reports" table
CREATE TABLE "reports" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "requestor_id" character varying NULL, "report_kind_name" character varying NULL, "name" character varying NOT NULL, "query" text NOT NULL, "variables" jsonb NULL, "status" character varying NOT NULL DEFAULT 'DRAFT', "visibility" character varying NULL, "owner_id" character varying NULL, "report_kind_id" character varying NULL, "report_owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "reports_custom_type_enums_report_kind" FOREIGN KEY ("report_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "reports_organizations_reports" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "reports_users_reports_owned" FOREIGN KEY ("report_owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "report_display_id_owner_id" to table: "reports"
CREATE UNIQUE INDEX "report_display_id_owner_id" ON "reports" ("display_id", "owner_id");
-- Create index "report_owner_id_idx" to table: "reports"
CREATE INDEX "report_owner_id_idx" ON "reports" ("owner_id");
-- Create index "report_report_owner_id_idx" to table: "reports"
CREATE INDEX "report_report_owner_id_idx" ON "reports" ("report_owner_id");
-- Create "report_blocked_groups" table
CREATE TABLE "report_blocked_groups" ("report_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("report_id", "group_id"), CONSTRAINT "report_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "report_blocked_groups_report_id" FOREIGN KEY ("report_id") REFERENCES "reports" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "report_blocked_groups_group_id_idx" to table: "report_blocked_groups"
CREATE INDEX "report_blocked_groups_group_id_idx" ON "report_blocked_groups" ("group_id");
-- Create "report_editors" table
CREATE TABLE "report_editors" ("report_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("report_id", "group_id"), CONSTRAINT "report_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "report_editors_report_id" FOREIGN KEY ("report_id") REFERENCES "reports" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "report_editors_group_id_idx" to table: "report_editors"
CREATE INDEX "report_editors_group_id_idx" ON "report_editors" ("group_id");
-- Create "report_memberships" table
CREATE TABLE "report_memberships" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "role" character varying NOT NULL DEFAULT 'MEMBER', "report_id" character varying NOT NULL, "user_id" character varying NOT NULL, "report_membership_org_membership" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "report_memberships_org_memberships_org_membership" FOREIGN KEY ("report_membership_org_membership") REFERENCES "org_memberships" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "report_memberships_reports_report" FOREIGN KEY ("report_id") REFERENCES "reports" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "report_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "report_membership_report_id_idx" to table: "report_memberships"
CREATE INDEX "report_membership_report_id_idx" ON "report_memberships" ("report_id");
-- Create index "reportmembership_user_id_report_id" to table: "report_memberships"
CREATE UNIQUE INDEX "reportmembership_user_id_report_id" ON "report_memberships" ("user_id", "report_id");
-- Create "report_viewers" table
CREATE TABLE "report_viewers" ("report_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("report_id", "group_id"), CONSTRAINT "report_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "report_viewers_report_id" FOREIGN KEY ("report_id") REFERENCES "reports" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "report_viewers_group_id_idx" to table: "report_viewers"
CREATE INDEX "report_viewers_group_id_idx" ON "report_viewers" ("group_id");
