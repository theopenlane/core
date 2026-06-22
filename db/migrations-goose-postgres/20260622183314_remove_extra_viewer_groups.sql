-- +goose Up
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "finding_blocked_groups", DROP COLUMN "finding_editors", DROP COLUMN "finding_viewers", DROP COLUMN "remediation_blocked_groups", DROP COLUMN "remediation_editors", DROP COLUMN "remediation_viewers", DROP COLUMN "review_blocked_groups", DROP COLUMN "review_editors", DROP COLUMN "review_viewers", DROP COLUMN "sla_definition_viewers";
-- create "finding_blocked_groups" table
CREATE TABLE "finding_blocked_groups" ("finding_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "group_id"), CONSTRAINT "finding_blocked_groups_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "finding_editors" table
CREATE TABLE "finding_editors" ("finding_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "group_id"), CONSTRAINT "finding_editors_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "remediation_blocked_groups" table
CREATE TABLE "remediation_blocked_groups" ("remediation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "group_id"), CONSTRAINT "remediation_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_blocked_groups_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "remediation_editors" table
CREATE TABLE "remediation_editors" ("remediation_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "group_id"), CONSTRAINT "remediation_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_editors_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "review_blocked_groups" table
CREATE TABLE "review_blocked_groups" ("review_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("review_id", "group_id"), CONSTRAINT "review_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_blocked_groups_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "review_editors" table
CREATE TABLE "review_editors" ("review_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("review_id", "group_id"), CONSTRAINT "review_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_editors_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "review_editors" table
DROP TABLE "review_editors";
-- reverse: create "review_blocked_groups" table
DROP TABLE "review_blocked_groups";
-- reverse: create "remediation_editors" table
DROP TABLE "remediation_editors";
-- reverse: create "remediation_blocked_groups" table
DROP TABLE "remediation_blocked_groups";
-- reverse: create "finding_editors" table
DROP TABLE "finding_editors";
-- reverse: create "finding_blocked_groups" table
DROP TABLE "finding_blocked_groups";
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "sla_definition_viewers" character varying NULL, ADD COLUMN "review_viewers" character varying NULL, ADD COLUMN "review_editors" character varying NULL, ADD COLUMN "review_blocked_groups" character varying NULL, ADD COLUMN "remediation_viewers" character varying NULL, ADD COLUMN "remediation_editors" character varying NULL, ADD COLUMN "remediation_blocked_groups" character varying NULL, ADD COLUMN "finding_viewers" character varying NULL, ADD COLUMN "finding_editors" character varying NULL, ADD COLUMN "finding_blocked_groups" character varying NULL;
