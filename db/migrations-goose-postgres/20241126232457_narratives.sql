-- +goose Up
-- modify "narrative_history" table
ALTER TABLE "narrative_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- modify "narratives" table
ALTER TABLE "narratives" ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "narratives_organizations_narratives" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "narrative_blocked_groups" table
CREATE TABLE "narrative_blocked_groups" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"), CONSTRAINT "narrative_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "narrative_blocked_groups_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "narrative_editors" table
CREATE TABLE "narrative_editors" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"), CONSTRAINT "narrative_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "narrative_editors_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "narrative_viewers" table
CREATE TABLE "narrative_viewers" ("narrative_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("narrative_id", "group_id"), CONSTRAINT "narrative_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "narrative_viewers_narrative_id" FOREIGN KEY ("narrative_id") REFERENCES "narratives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "narrative_viewers" table
DROP TABLE "narrative_viewers";
-- reverse: create "narrative_editors" table
DROP TABLE "narrative_editors";
-- reverse: create "narrative_blocked_groups" table
DROP TABLE "narrative_blocked_groups";
-- reverse: modify "narratives" table
ALTER TABLE "narratives" DROP CONSTRAINT "narratives_organizations_narratives", DROP COLUMN "owner_id";
-- reverse: modify "narrative_history" table
ALTER TABLE "narrative_history" DROP COLUMN "owner_id";
