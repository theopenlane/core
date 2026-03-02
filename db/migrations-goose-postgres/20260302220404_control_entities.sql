-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" DROP COLUMN "remediation_controls", DROP COLUMN "review_controls";
-- create "control_campaigns" table
CREATE TABLE "control_campaigns" ("control_id" character varying NOT NULL, "campaign_id" character varying NOT NULL, PRIMARY KEY ("control_id", "campaign_id"), CONSTRAINT "control_campaigns_campaign_id" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_campaigns_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_entities" table
CREATE TABLE "control_entities" ("control_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("control_id", "entity_id"), CONSTRAINT "control_entities_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "control_identity_holders" table
CREATE TABLE "control_identity_holders" ("control_id" character varying NOT NULL, "identity_holder_id" character varying NOT NULL, PRIMARY KEY ("control_id", "identity_holder_id"), CONSTRAINT "control_identity_holders_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "control_identity_holders_identity_holder_id" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "remediation_controls" table
CREATE TABLE "remediation_controls" ("remediation_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "control_id"), CONSTRAINT "remediation_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_controls_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "review_controls" table
CREATE TABLE "review_controls" ("review_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("review_id", "control_id"), CONSTRAINT "review_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "review_controls_review_id" FOREIGN KEY ("review_id") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "review_controls" table
DROP TABLE "review_controls";
-- reverse: create "remediation_controls" table
DROP TABLE "remediation_controls";
-- reverse: create "control_identity_holders" table
DROP TABLE "control_identity_holders";
-- reverse: create "control_entities" table
DROP TABLE "control_entities";
-- reverse: create "control_campaigns" table
DROP TABLE "control_campaigns";
-- reverse: modify "controls" table
ALTER TABLE "controls" ADD COLUMN "review_controls" character varying NULL, ADD COLUMN "remediation_controls" character varying NULL;
