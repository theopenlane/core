-- Create "audiences" table
CREATE TABLE "audiences" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "description" character varying NULL, "audience_type" character varying NOT NULL DEFAULT 'MANUAL', "filters" jsonb NULL, "metadata" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "audiences_organizations_audiences" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "audience_display_id_owner_id" to table: "audiences"
CREATE UNIQUE INDEX "audience_display_id_owner_id" ON "audiences" ("display_id", "owner_id");
-- Create index "audience_name_owner_id" to table: "audiences"
CREATE INDEX "audience_name_owner_id" ON "audiences" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "audience_owner_id_idx" to table: "audiences"
CREATE INDEX "audience_owner_id_idx" ON "audiences" ("owner_id");
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_audience_creators" character varying NULL, ADD COLUMN "organization_audience_member_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_audience_creators" FOREIGN KEY ("organization_audience_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_audience_member_creators" FOREIGN KEY ("organization_audience_member_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create "audience_blocked_groups" table
CREATE TABLE "audience_blocked_groups" ("audience_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("audience_id", "group_id"), CONSTRAINT "audience_blocked_groups_audience_id" FOREIGN KEY ("audience_id") REFERENCES "audiences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "audience_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "audience_blocked_groups_group_id_idx" to table: "audience_blocked_groups"
CREATE INDEX "audience_blocked_groups_group_id_idx" ON "audience_blocked_groups" ("group_id");
-- Create "audience_editors" table
CREATE TABLE "audience_editors" ("audience_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("audience_id", "group_id"), CONSTRAINT "audience_editors_audience_id" FOREIGN KEY ("audience_id") REFERENCES "audiences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "audience_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "audience_editors_group_id_idx" to table: "audience_editors"
CREATE INDEX "audience_editors_group_id_idx" ON "audience_editors" ("group_id");
-- Create "audience_members" table
CREATE TABLE "audience_members" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "email" character varying NOT NULL, "full_name" character varying NULL, "metadata" jsonb NULL, "audience_id" character varying NOT NULL, "contact_id" character varying NULL, "group_id" character varying NULL, "identity_holder_id" character varying NULL, "owner_id" character varying NULL, "subscriber_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "audience_members_audiences_audience_members" FOREIGN KEY ("audience_id") REFERENCES "audiences" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "audience_members_contacts_audience_members" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "audience_members_groups_audience_members" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "audience_members_identity_holders_audience_members" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "audience_members_organizations_audience_members" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "audience_members_subscribers_audience_members" FOREIGN KEY ("subscriber_id") REFERENCES "subscribers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "audience_members_users_audience_members" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "audience_member_contact_id_idx" to table: "audience_members"
CREATE INDEX "audience_member_contact_id_idx" ON "audience_members" ("contact_id");
-- Create index "audience_member_group_id_idx" to table: "audience_members"
CREATE INDEX "audience_member_group_id_idx" ON "audience_members" ("group_id");
-- Create index "audience_member_identity_holder_id_idx" to table: "audience_members"
CREATE INDEX "audience_member_identity_holder_id_idx" ON "audience_members" ("identity_holder_id");
-- Create index "audience_member_owner_id_idx" to table: "audience_members"
CREATE INDEX "audience_member_owner_id_idx" ON "audience_members" ("owner_id");
-- Create index "audience_member_subscriber_id_idx" to table: "audience_members"
CREATE INDEX "audience_member_subscriber_id_idx" ON "audience_members" ("subscriber_id");
-- Create index "audience_member_user_id_idx" to table: "audience_members"
CREATE INDEX "audience_member_user_id_idx" ON "audience_members" ("user_id");
-- Create index "audiencemember_audience_id_email" to table: "audience_members"
CREATE UNIQUE INDEX "audiencemember_audience_id_email" ON "audience_members" ("audience_id", "email") WHERE (deleted_at IS NULL);
-- Create index "audiencemember_display_id_owner_id" to table: "audience_members"
CREATE UNIQUE INDEX "audiencemember_display_id_owner_id" ON "audience_members" ("display_id", "owner_id");
-- Create "audience_viewers" table
CREATE TABLE "audience_viewers" ("audience_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("audience_id", "group_id"), CONSTRAINT "audience_viewers_audience_id" FOREIGN KEY ("audience_id") REFERENCES "audiences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "audience_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "audience_viewers_group_id_idx" to table: "audience_viewers"
CREATE INDEX "audience_viewers_group_id_idx" ON "audience_viewers" ("group_id");
-- Create "campaign_audiences" table
CREATE TABLE "campaign_audiences" ("campaign_id" character varying NOT NULL, "audience_id" character varying NOT NULL, PRIMARY KEY ("campaign_id", "audience_id"), CONSTRAINT "campaign_audiences_audience_id" FOREIGN KEY ("audience_id") REFERENCES "audiences" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "campaign_audiences_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "campaign_audiences_audience_id_idx" to table: "campaign_audiences"
CREATE INDEX "campaign_audiences_audience_id_idx" ON "campaign_audiences" ("audience_id");
