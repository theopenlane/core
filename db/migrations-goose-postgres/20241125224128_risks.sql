-- +goose Up
-- modify "risk_history" table
ALTER TABLE "risk_history" ALTER COLUMN "impact" SET DEFAULT 'MODERATE', ALTER COLUMN "likelihood" SET DEFAULT 'LIKELY', ADD COLUMN "owner_id" character varying NOT NULL;
-- modify "risks" table
ALTER TABLE "risks" ALTER COLUMN "impact" SET DEFAULT 'MODERATE', ALTER COLUMN "likelihood" SET DEFAULT 'LIKELY', ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "risks_organizations_risks" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- create "risk_blocked_groups" table
CREATE TABLE "risk_blocked_groups" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"), CONSTRAINT "risk_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "risk_blocked_groups_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "risk_editors" table
CREATE TABLE "risk_editors" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"), CONSTRAINT "risk_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "risk_editors_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "risk_viewers" table
CREATE TABLE "risk_viewers" ("risk_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "group_id"), CONSTRAINT "risk_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "risk_viewers_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "risk_viewers" table
DROP TABLE "risk_viewers";
-- reverse: create "risk_editors" table
DROP TABLE "risk_editors";
-- reverse: create "risk_blocked_groups" table
DROP TABLE "risk_blocked_groups";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP CONSTRAINT "risks_organizations_risks", DROP COLUMN "owner_id", ALTER COLUMN "likelihood" DROP DEFAULT, ALTER COLUMN "impact" DROP DEFAULT;
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "owner_id", ALTER COLUMN "likelihood" DROP DEFAULT, ALTER COLUMN "impact" DROP DEFAULT;
