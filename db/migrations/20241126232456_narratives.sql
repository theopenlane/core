-- Modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- Modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"), CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"), CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"), CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
