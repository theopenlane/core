-- +goose Up
-- modify "events" table
ALTER TABLE "events" DROP COLUMN "org_subscription_events";
-- create "org_subscription_events" table
CREATE TABLE "org_subscription_events" ("org_subscription_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("org_subscription_id", "event_id"), CONSTRAINT "org_subscription_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "org_subscription_events_org_subscription_id" FOREIGN KEY ("org_subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "org_subscription_events" table
DROP TABLE "org_subscription_events";
-- reverse: modify "events" table
ALTER TABLE "events" ADD COLUMN "org_subscription_events" character varying NULL;
